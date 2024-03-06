# My limiter
Sim, eu sei que o framework fiber tem um limiter
```
import "github.com/gofiber/fiber/v2/middleware/limiter"

...limiter.New()
```
A diferença entre o meu é o padrão do fiber é que o meu limiter adiciona uma penalidade para o usuario que ultrapassa o limite de requisições em determinado intervalo. 

### V1
Na versão v1 o limiter utiliza apenas uma goroutine para gerenciar todos usuarios, pois quando o temporizador for acionado deletara todos os usuarios que não estiveram ativo até determinado intervalo, e tambem um "ticker" para resetar as requisições de cada usuario que tenha passado o tempo de ban ou que não tenha sido bloqueado

### V2
Na versão v2 o limiter utiliza goroutines para gerenciar cada usuario, onde para cada usuario, tera um temporizador para gerenciar a exclusão do usuario na memoria e tambem um "ticker" para resetar as requisições dado o intervalo de tempo.

### Anotações
Caso o sistema tenha alto trafego recomendo que se utilize v1
