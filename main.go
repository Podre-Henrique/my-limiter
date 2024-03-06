package main

import (
	"log"

	"github.com/Podre-Henrique/my-ratelimit/middleware/mylimiter"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	api := fiber.New()
	api.Use(logger.New(logger.Config{}))

	// api.Use(mylimiter.New(config.Config{LimitRequest: 2, DurationBan: 1 * time.Minute, DurationInMemory: 1 * time.Minute}))
	api.Use(mylimiter.New())
	api.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("Bom dia/tarde/noite")
	})
	log.Fatal(api.Listen("localhost:888"))
}
