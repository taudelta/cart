package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/taudelta/cart/model"
	"github.com/taudelta/cart/params"
	"github.com/taudelta/cart/response"
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

	app.Get("/", func(c *fiber.Ctx) error {
		userID := c.Query("user_id")
		cartData := redisClient.HGetAll(context.Background(), userID)
		if err := cartData.Err(); err != nil {
			log.Println("get cart error", err)
			return err
		}

		cartValue := cartData.Val()
		if cartValue == nil {
			return errors.New("no cart is found")
		}

		items := make([]model.CartItem, 0)
		if cartValue["items"] != "" {
			if err := json.Unmarshal([]byte(cartValue["items"]), &items); err != nil {
				return err
			}
		}

		var cart model.Cart
		cart.Items = items
		return c.Status(http.StatusOK).JSON(&cart)
	})

	app.Post("/add-item", func(c *fiber.Ctx) error {
		var param params.CartAdd

		err := c.BodyParser(&param)
		if err != nil {
			return err
		}

		ctx := context.Background()

		cartData := redisClient.HGetAll(ctx, param.UserID)
		if err := cartData.Err(); err != nil {
			return err
		}

		cartValue := cartData.Val()

		log.Println("cart", param.UserID, cartValue)

		var items []*model.CartItem

		if cartValue["items"] != "" {
			if err := json.Unmarshal([]byte(cartValue["items"]), &items); err != nil {
				return err
			}
		}

		itemsMap := make(map[string]*model.CartItem)
		for _, i := range items {
			log.Printf("item: %+v\n", i)
			itemsMap[i.Sku] = i
		}

		for _, addItem := range param.Items {
			if _, ok := itemsMap[addItem.Sku]; ok {
				itemsMap[addItem.Sku].Quantity += int(addItem.Quantity)
			} else {
				items = append(items, &model.CartItem{
					Sku:      addItem.Sku,
					Quantity: int(addItem.Quantity),
				})
			}
		}

		log.Println("add to cart", param.UserID)

		itemBody, err := json.Marshal(&items)
		if err != nil {
			return err
		}

		pipeline := redisClient.Pipeline()
		pipeline.HSet(ctx, param.UserID, "items", string(itemBody))

		if _, err := pipeline.Exec(ctx); err != nil {
			log.Println("failed to add items in cart", err)
			return err
		}

		var resp response.CartAddOk
		return c.Status(http.StatusOK).JSON(&resp)
	})

	app.Post("/remove-item", func(c *fiber.Ctx) error {
		return nil
	})

	app.Post("/delete", func(c *fiber.Ctx) error {
		return nil
	})

	app.Post("/set-attribute", func(c *fiber.Ctx) error {
		return nil
	})

	app.Listen(serviceAddr)
}
