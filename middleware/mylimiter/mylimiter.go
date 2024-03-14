package mylimiter

import (
	"time"

	"github.com/Podre-Henrique/my-ratelimit/middleware/mylimiter/config"
	v1 "github.com/Podre-Henrique/my-ratelimit/middleware/mylimiter/v1"
	"github.com/gofiber/fiber/v2"
)

var configDefault = config.Config{
	LimitRequest:     50,
	RenewIn:          1 * time.Minute,
	DurationInMemory: 5 * time.Minute,
	DurationBan:      5 * time.Minute,
}

func configure(config ...config.Config) config.Config {
	if len(config) < 1 {
		return configDefault
	}

	cfg := config[0]
	if cfg.LimitRequest <= 0 {
		cfg.LimitRequest = configDefault.LimitRequest
	}
	if cfg.DurationBan <= 0 {
		cfg.DurationBan = configDefault.DurationBan
	}
	if cfg.DurationInMemory <= 0 {
		cfg.DurationInMemory = configDefault.DurationInMemory
	}
	if cfg.RenewIn <= 0 {
		cfg.RenewIn = configDefault.RenewIn
	}
	return cfg
}

func New(config ...config.Config) fiber.Handler {
	cfg := configure(config...)
	// return v1.New(cfg)
	return v1.New(cfg)
}
