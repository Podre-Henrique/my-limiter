package v1

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

/*
Esta versão consome mais memoria pois para cada usuario é criada uma goroutine para gerenciar seu tempo em memoria e a
quantidade de requisições feita no intervalo de tempo
*/

const (
	// limite de requisições
	limitRequest uint8 = 50
	// renova a cada minuto
	renewIn = time.Minute * 1
	// tempo em memoria
	durationInMemory = time.Minute * 5
	// duração do timeout no usuario
	durationTimeout = time.Minute * 5
)

type user struct {
	ip    string
	block bool
	// quantidade de requisições
	counts  uint8
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

var limiter = &users{
	usersMap: make(map[string]*user),
}

func (i *user) timeout() {
	for {
		select {
		case <-i.duration.C:
			// deleta o usuario caso atinja o tempo limite em memoria
			delete(limiter.usersMap, i.ip)
			return
		case <-i.ticker.C:
			// caso o tempo de block tenha excedido(ou nao esteja bloqueado), desbloqueia o usuario e reseta suas requisições
			if !i.block || time.Since(i.blockIn) > durationTimeout {
				i.counts = 0
				i.block = false
			}
		}
	}
}
func MyLimiter(c *fiber.Ctx) error {
	ip := c.IP()

	// ao fazer testes com racecondition percebi a necessidade de utilizar mutex
	limiter.Lock()
	req, ok := limiter.usersMap[ip]
	if !ok {
		duration := time.NewTimer(durationInMemory)
		ticker := time.NewTicker(renewIn)
		// insere um novo usuario no map
		limiter.usersMap[ip] = &user{ip: ip, counts: 1, ticker: ticker, duration: duration}
		// cria uma goroutine para este ussuario
		go limiter.usersMap[ip].timeout()
		limiter.Unlock()
		// retorna o proximo middleware/handler
		return c.Next()
	}
	limiter.Unlock()

	req.Lock()
	defer req.Unlock()
	// reseta o tempo em que o usuario ficara em memoria
	req.duration.Reset(durationInMemory)

	// caso o usuario esteja bloqueado
	if req.block {
		// verifica se ainda passou o tempo de "bloqueamento"
		if time.Since(req.blockIn) < durationTimeout {
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
