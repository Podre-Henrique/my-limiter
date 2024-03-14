package v1

import (
	"fmt"
	"sync"
	"time"

	"github.com/Podre-Henrique/my-ratelimit/middleware/mylimiter/config"
	"github.com/gofiber/fiber/v2"
)

/*
Esta versão consome menos memoria pois utiliza apenas uma goroutine para deletar ou para resetar as requisições do usuario
Todavia, supondo um limite de 20 requisições por minuto e o usuario faça 19 requisições em apenas segundo e o contador resete
o usuario podera fazer mais 20 requisições novamente, o que somaria 39 requisições em menos de um minuto no pior dos casos
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

var (
	// usado para renovar as requisiçoes do usuario
	renew *time.Ticker
	// usado para gerenciar o tempo do usuario em memoria
	duration *time.Ticker
)

type user struct {
	ip    string
	block bool
	// quantidade de requisições
	counts      uint16
	blockIn     time.Time
	lastRequest time.Time
	sync.Mutex
}

type users struct {
	usersMap map[string]*user
	sync.Mutex
}

var limiterUsers = &users{
	usersMap: make(map[string]*user),
}

func timeout() {
	for {
		select {
		case <-duration.C:
			// deleta os usuarios que atinjiram o tempo limite em memoria
			limiterUsers.Lock()
			for userIp, userInfo := range limiterUsers.usersMap {
				if time.Since(userInfo.lastRequest) > durationInMemory {
					delete(limiterUsers.usersMap, userIp)
				}
			}
			limiterUsers.Unlock()
		case <-renew.C:
			fmt.Println(limiterUsers)
			// caso o tempo de block tenha excedido desbloqueia os usuarios e reseta suas requisições
			for _, userInfo := range limiterUsers.usersMap {
				// caso o tempo de block tenha excedido(ou nao esteja bloqueado), desbloqueia o usuario e reseta suas requisições
				if time.Since(userInfo.blockIn) > durationBan {
					userInfo.counts = 0
					userInfo.block = false
				}
			}
		}
	}
}

func myLimiter(c *fiber.Ctx) error {
	ip := c.IP()

	// ao fazer testes com racecondition percebi a necessidade de utilizar mutex
	limiterUsers.Lock()
	reqUser, ok := limiterUsers.usersMap[ip]
	if !ok {
		// insere um novo usuario no map
		limiterUsers.usersMap[ip] = &user{ip: ip, counts: 1}
		// cria uma goroutine para este usuario
		limiterUsers.Unlock()
		// retorna o proximo middleware/handler
		return c.Next()
	}
	limiterUsers.Unlock()

	reqUser.Lock()
	defer reqUser.Unlock()
	// reseta o tempo em que o usuario ficara em memoria
	reqUser.lastRequest = time.Now()

	// caso o usuario esteja bloqueado
	if reqUser.block {
		// verifica se ainda passou o tempo de "bloqueamento"
		if time.Since(reqUser.blockIn) < durationBan {
			return fiber.ErrTooManyRequests // retorna 429
		} else {
			reqUser.block = false
			reqUser.counts = 0
		}
	}
	// Adiciona uma requisição para o usuario
	reqUser.counts++
	if reqUser.counts >= limitRequest {
		reqUser.block = true
		reqUser.blockIn = time.Now()
	}
	return c.Next()
}

func New(config config.Config) fiber.Handler {
	limitRequest = config.LimitRequest
	renewIn = config.RenewIn
	durationInMemory = config.DurationInMemory
	durationBan = config.DurationBan

	renew = time.NewTicker(renewIn)
	duration = time.NewTicker(durationInMemory)
	go timeout()

	return myLimiter
}
