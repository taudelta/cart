package main

import (
	"context"
	"flag"
	"net/http"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/taudelta/cart/params"
	"github.com/taudelta/cart/service"
	"github.com/taudelta/cart/service/dto"
)

func main() {
	var serviceAddr string
	var redisURL string
	flag.Parse()
	flag.StringVar(&serviceAddr, "addr", ":8000", "service listen address")
	flag.StringVar(&redisURL, "redis_url", "redis://localhost:6379/1", "redis database listen address")

	app := fiber.New()

	redisOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(redisOptions)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	cartService := &service.Cart{
		Storage: redisClient,
	}

	app.Get("/", func(c *fiber.Ctx) error {
		userID := c.Query("user_id")
		cart, err := cartService.GetCart(userID)
		if err != nil {
			return err
		}
		return c.Status(http.StatusOK).JSON(&cart)
	})

	app.Post("/add-item", func(c *fiber.Ctx) error {
		var param params.CartAdd

		err := c.BodyParser(&param)
		if err != nil {
			return err
		}

		items := make([]dto.CartItem, 0, len(param.Items))
		for _, i := range param.Items {
			items = append(items, dto.CartItem{
				Sku:      i.Sku,
				Quantity: int(i.Quantity),
			})
		}

		return cartService.AddItem(param.UserID, items)
	})

	app.Post("/remove-item", func(c *fiber.Ctx) error {
		var param params.CartRemove

		err := c.BodyParser(&param)
		if err != nil {
			return err
		}

		removeItems := make([]dto.CartItem, 0, len(param.Items))
		for _, i := range param.Items {
			removeItems = append(removeItems, dto.CartItem{
				Sku:      i.Sku,
				Quantity: int(i.Quantity),
			})
		}

		return cartService.RemoveItem(param.UserID, removeItems, param.RemoveAll)
	})

	app.Post("/delete", func(c *fiber.Ctx) error {
		var param params.Delete
		if err := c.BodyParser(&param); err != nil {
			return err
		}

		return cartService.DeleteCart(param.UserID)
	})

	app.Listen(serviceAddr)
}
