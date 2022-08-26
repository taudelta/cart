package main

import (
	"bytes"
	"encoding/json"
	"flag"

	"github.com/gofiber/fiber/v2"
	"github.com/taudelta/cart/params"
	"github.com/taudelta/cart/response"
)

func main() {
	var serviceAddr string
	var redisURL string
	flag.Parse()
	flag.StringVar(&serviceAddr, "addr", ":8000", "service listen address")
	flag.StringVar(&redisURL, "redis_url", "redis://localhost:6789/1", "redis database listen address")

	app := fiber.New()

	// redisClient := redis.ParseURL(redisURL, "redis_url", "redis://localhost:6789/1")
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	app.Post("/add-item", func(c *fiber.Ctx) error {
		var param params.CartAdd

		err := json.NewDecoder(bytes.NewBuffer(c.Body())).Decode(&param)
		if err != nil {
			return err
		}

		var resp response.CartAddOk

		buf := bytes.NewBuffer([]byte{})
		err = json.NewEncoder(buf).Encode(&resp)
		if err != nil {
			return err
		}
		return c.Send(buf.Bytes())
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
