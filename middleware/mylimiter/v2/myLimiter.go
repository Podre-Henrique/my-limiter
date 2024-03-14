package v2

import (
	"sync"
	"time"

	"github.com/Podre-Henrique/my-ratelimit/middleware/mylimiter/config"
	"github.com/gofiber/fiber/v2"
)

/*
Esta versão consome mais memoria pois para cada usuario é criada uma goroutine para gerenciar seu tempo em memoria e a
quantidade de requisições feita no intervalo de tempo
*/

var (
	// limite de requisições
	limitRequest uint16
	// renova a cada minuto
	renewIn time.Duration
	// tempo em memoria
	durationInMemory time.Duration
	// duração do timeout no usuario
	durationBan time.Duration
)

type user struct {
	ip    string
	block bool
	// quantidade de requisições
	counts  uint16
	blockIn time.Time
	sync.Mutex
	// usado para renovar as requisiçoes do usuario
	ticker *time.Ticker
	// usado para gerenciar o tempo do usuario em memoria
	duration *time.Timer
}

type users struct {
	usersMap map[string]*user
	sync.Mutex
}

var limiterUsers = &users{
	usersMap: make(map[string]*user),
}

func (i *user) timeout() {
	for {
		select {
		case <-i.duration.C:
			// deleta o usuario caso atinja o tempo limite em memoria
			delete(limiterUsers.usersMap, i.ip)
			return
		case <-i.ticker.C:
			// caso o tempo de block tenha excedido(ou nao esteja bloqueado), desbloqueia o usuario e reseta suas requisições
			if !i.block || time.Since(i.blockIn) > durationBan {
				i.counts = 0
				i.block = false
			}
		}
	}
}
func myLimiter(c *fiber.Ctx) error {
	ip := c.IP()

	// ao fazer testes com racecondition percebi a necessidade de utilizar mutex
	limiterUsers.Lock()
	req, ok := limiterUsers.usersMap[ip]
	if !ok {
		duration := time.NewTimer(durationInMemory)
		ticker := time.NewTicker(renewIn)
		// insere um novo usuario no map
		limiterUsers.usersMap[ip] = &user{ip: ip, counts: 1, ticker: ticker, duration: duration}
		// cria uma goroutine para este ussuario
		go limiterUsers.usersMap[ip].timeout()
		limiterUsers.Unlock()
		// retorna o proximo middleware/handler
		return c.Next()
	}
	limiterUsers.Unlock()

	req.Lock()
	defer req.Unlock()
	// reseta o tempo em que o usuario ficara em memoria
	req.duration.Reset(durationInMemory)

	// caso o usuario esteja bloqueado
	if req.block {
		// verifica se ainda passou o tempo de "bloqueamento"
		if time.Since(req.blockIn) < durationBan {
			return fiber.ErrTooManyRequests // retorna 429
		} else {
			req.block = false
			req.counts = 0
		}
	}
	// Adiciona uma requisição para o usuario
	req.counts++
	if req.counts >= limitRequest {
		req.block = true
		req.blockIn = time.Now()
	}
	return c.Next()
}

func New(config config.Config) fiber.Handler {
	limitRequest = config.LimitRequest
	renewIn = config.RenewIn
	durationInMemory = config.DurationInMemory
	durationBan = config.DurationBan
	return myLimiter
}
