package v1

import (
	"testing"
	"time"

	"github.com/Podre-Henrique/my-ratelimit/middleware/mylimiter"
	"github.com/Podre-Henrique/my-ratelimit/middleware/mylimiter/config"
	"github.com/gofiber/fiber/v2"
)

func Test_myLimiter(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(mylimiter.New(config.Config{LimitRequest: 2, DurationBan: 1 * time.Minute, DurationInMemory: 30 * time.Second}))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("Bom dia/tarde/noite")
	})
}
