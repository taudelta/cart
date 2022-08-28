package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/taudelta/cart/params"
	"github.com/taudelta/cart/service"
	"github.com/taudelta/cart/service/dto"
)

var (
	requestsCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "requests_count",
	}, []string{"method"})

	requestsErrorCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "requests_error_count",
	}, []string{"method"})
)

func main() {
	var serviceAddr string
	var redisURL string
	flag.Parse()
	flag.StringVar(&serviceAddr, "addr", ":8000", "service listen address")
	flag.StringVar(&redisURL, "redis_url", "redis://localhost:6379/1", "redis database listen address")

	registry := prometheus.NewRegistry()

	registry.MustRegister(
		requestsCount,
		requestsErrorCount,
	)

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
		requestsCount.With(prometheus.Labels{
			"method": "get",
		}).Add(1)
		userID := c.Query("user_id")
		cart, err := cartService.GetCart(userID)
		if err != nil {
			requestsErrorCount.With(prometheus.Labels{
				"method": "get",
			}).Add(1)
			return err
		}
		return c.Status(http.StatusOK).JSON(&cart)
	})

	app.Get("/metrics", func(c *fiber.Ctx) error {
		metrics, err := registry.Gather()
		if err != nil {
			return err
		}
		return c.Status(http.StatusOK).JSON(metrics)
	})

	app.Post("/add-item", func(c *fiber.Ctx) error {
		requestsCount.With(prometheus.Labels{
			"method": "add-item",
		}).Add(1)

		var param params.CartAdd
		var err error
		defer func() {
			if err != nil {
				requestsErrorCount.With(prometheus.Labels{
					"method": "add-item",
				}).Add(1)
			}
		}()

		err = c.BodyParser(&param)
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

		err = cartService.AddItem(param.UserID, items)
		return err
	})

	app.Post("/remove-item", func(c *fiber.Ctx) error {
		requestsCount.With(prometheus.Labels{
			"method": "remove-item",
		}).Add(1)

		var err error
		defer func() {
			if err != nil {
				requestsErrorCount.With(prometheus.Labels{
					"method": "remove-item",
				}).Add(1)
			}
		}()

		var param params.CartRemove

		err = c.BodyParser(&param)
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

		err = cartService.RemoveItem(param.UserID, removeItems, param.RemoveAll)
		return err
	})

	app.Post("/delete", func(c *fiber.Ctx) error {
		requestsCount.With(prometheus.Labels{
			"method": "delete",
		}).Add(1)

		var err error
		defer func() {
			if err != nil {
				requestsErrorCount.With(prometheus.Labels{
					"method": "delete",
				}).Add(1)
			}
		}()

		var param params.Delete
		err = c.BodyParser(&param)
		if err != nil {
			return err
		}

		err = cartService.DeleteCart(param.UserID)
		return err
	})

	if err := app.Listen(serviceAddr); err != nil {
		log.Panic(err)
	}
}
