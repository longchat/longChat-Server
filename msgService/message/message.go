package message

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kataras/iris"
	iconfig "github.com/kataras/iris/config"
	"github.com/kataras/iris/sessions/providers/redis"
	"github.com/kataras/iris/websocket"
	"github.com/longchat/longChat-Server/apiService/api/dto"
	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
)

func Init(framework *iris.Framework) {
	redisAddr, err := config.GetConfigString(consts.RedisAddress)
	if err != nil {
		log.Fatalf(consts.ErrGetConfigFailed(consts.RedisAddress, err))
	}
	redisPsw, err := config.GetConfigString(consts.RedisPassword)
	if err != nil {
		log.Fatalf(consts.ErrGetConfigFailed(consts.RedisPassword, err))
	}
	redisPrefix, err := config.GetConfigString(consts.SessionPrefix)
	if err != nil {
		log.Fatalf(consts.ErrGetConfigFailed(consts.SessionPrefix, err))
	}
	cookie, err := config.GetConfigString(consts.SessionCookieName)
	if err != nil {
		log.Fatalf(consts.ErrGetConfigFailed(consts.SessionCookieName, err))
	}
	framework.Config.Sessions = iconfig.Sessions{
		Provider:   "redis",
		Cookie:     cookie,
		Expires:    iconfig.CookieExpireNever,
		GcDuration: time.Duration(2) * time.Hour,
	}
	redis.Config.Network = "tcp"
	redis.Config.Addr = redisAddr
	redis.Config.Prefix = redisPrefix
	redis.Config.Password = redisPsw

	framework.UseFunc(func(c *iris.Context) {
		fmt.Println(c.Session().ID(), c.Session().GetString("UserName"))
		if c.Session().GetString("UserName") == "" {
			c.JSON(http.StatusUnauthorized, dto.PasswordNotMatchErrRsp())
			return
		}
		c.Next()
	})
	framework.Config.Websocket.Endpoint = "/websocket"

	ws := framework.Websocket
	ws.OnConnection(func(c websocket.Connection) {

		c.OnMessage(func(data []byte) {
			message := string(data)
			c.To(websocket.Broadcast).EmitMessage([]byte("Message from: " + c.ID() + "-> " + message))
			c.EmitMessage([]byte("Me: " + message))
		})

		c.OnDisconnect(func() {
		})
	})
}
