package main

import (
	"log"

	limiter "github.com/Podre-Henrique/my-ratelimit/middleware/limiter/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	api := fiber.New()
	api.Use(logger.New())
	// a.Use(v1.MyLimiter)
	api.Use(limiter.MyLimiter)
	api.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("Bom dia/tarde/noite")
	})
	log.Fatal(api.Listen("localhost:888"))
}
