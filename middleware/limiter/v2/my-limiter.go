package v2

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

/*
Esta versão consome menos memoria pois utiliza apenas uma goroutine para deletar ou para resetar as requisições do usuario
Todavia, supondo um limite de 20 requisições por minuto e o usuario faça 19 requisições em apenas segundo e o contador resete
o usuario podera fazer mais 20 requisições novamente, o que somaria 39 requisições em menos de um minuto no pior dos casos
*/

const (
	// limite de requisições
	limitRequest uint8 = 2
	// renova a cada minuto
	renewIn = time.Minute * 1
	// tempo em memoria
	durationInMemory = time.Minute * 5
	// duração do timeout no usuario
	durationTimeout = time.Minute * 5
)

var (
	// usado para renovar as requisiçoes do usuario
	renew = time.NewTicker(renewIn)
	// usado para gerenciar o tempo do usuario em memoria
	duration = time.NewTicker(durationInMemory)
)

type user struct {
	ip    string
	block bool
	// quantidade de requisições
	counts      uint8
	blockIn     time.Time
	lastRequest time.Time
	sync.Mutex
}

type users struct {
	usersMap map[string]*user
	sync.Mutex
}

var limiter = &users{
	usersMap: make(map[string]*user),
}

func timeout() {
	for {
		select {
		case <-duration.C:
			fmt.Println("Deletando usuarios...")
			// deleta os usuarios que atinjiram o tempo limite em memoria
			for userIp, userInfo := range limiter.usersMap {
				if time.Since(userInfo.lastRequest) > durationInMemory {
					delete(limiter.usersMap, userIp)
				}
			}
			fmt.Println(limiter)
		case <-renew.C:
			fmt.Println("Resetando usuarios...")
			// caso o tempo de block tenha excedido desbloqueia os usuarios e reseta suas requisições
			for _, userInfo := range limiter.usersMap {
				// caso o tempo de block tenha excedido(ou nao esteja bloqueado), desbloqueia o usuario e reseta suas requisições
				if time.Since(userInfo.blockIn) > durationTimeout {
					userInfo.counts = 0
					userInfo.block = false
				}
			}
		}
	}
}

func init() {
	go timeout()
}
func MyLimiter(c *fiber.Ctx) error {
	ip := c.IP()

	// ao fazer testes com racecondition percebi a necessidade de utilizar mutex
	limiter.Lock()
	reqUser, ok := limiter.usersMap[ip]
	if !ok {
		// insere um novo usuario no map
		limiter.usersMap[ip] = &user{ip: ip, counts: 1}
		// cria uma goroutine para este ussuario
		limiter.Unlock()
		// retorna o proximo middleware/handler
		return c.Next()
	}
	limiter.Unlock()

	reqUser.Lock()
	defer reqUser.Unlock()
	// reseta o tempo em que o usuario ficara em memoria
	reqUser.lastRequest = time.Now()

	// caso o usuario esteja bloqueado
	if reqUser.block {
		// verifica se ainda passou o tempo de "bloqueamento"
		if time.Since(reqUser.blockIn) < durationTimeout {
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
