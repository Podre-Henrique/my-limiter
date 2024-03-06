package config

import "time"

type Config struct {
	// limite de requisições. Default: 50 req
	LimitRequest uint16
	// renova a cada minuto. Default: 1 minuto
	RenewIn time.Duration
	// tempo em memoria. Default: 5 minutos
	DurationInMemory time.Duration
	// duração da penalidade no usuario. Default: 5 minutos
	DurationBan time.Duration
}
