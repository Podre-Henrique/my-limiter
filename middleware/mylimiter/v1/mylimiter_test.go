package v1

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Podre-Henrique/my-ratelimit/middleware/mylimiter"
	"github.com/Podre-Henrique/my-ratelimit/middleware/mylimiter/config"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func Test_myLimiter(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(mylimiter.New(config.Config{LimitRequest: 2, DurationBan: 1 * time.Minute, DurationInMemory: 10 * time.Second}))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("Bom dia/tarde/noite")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "Bom dia/tarde/noite", string(body))
	// Adicionar uma forma de inserir varios usuarios e fazer uma requisição enquanto estiver deletando os inativos
}
