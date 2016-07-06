package message

import (
	"fmt"
	"log"
	"time"

	"github.com/kataras/iris"
	iconfig "github.com/kataras/iris/config"
	"github.com/kataras/iris/sessions/providers/redis"
	"github.com/kataras/iris/websocket"
	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/common/protoc"
	"github.com/longchat/longChat-Server/common/util"

	"google.golang.org/grpc"
)

type Messenger struct {
	client *protoc.RouterClient
	conn   *grpc.ClientConn
}

func (m *Messenger) Close() {
	if m.conn != nil {
		m.conn.Close()
	}
}
func (m *Messenger) Init(framework *iris.Framework) {
	privateToken, err := config.GetConfigString(consts.PrivateToken)
	if err != nil {
		log.Fatalf(consts.ErrGetConfigFailed(consts.PrivateToken, err))
	}
	routerAddr, err := config.GetConfigString(consts.RouterServiceAddress)
	if err != nil {
		log.Fatalf(consts.ErrGetConfigFailed(consts.RouterServiceAddress, err))
	}
	m.conn, err = grpc.Dial(routerAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("dial to router(%s) failed!err:=%v", err)
	}
	c := protoc.NewRouterClient(m.conn)
	m.client = &c

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

	/*framework.UseFunc(func(c *iris.Context) {
		fmt.Println("connId:", c.ConnID(), c.Session().GetString("UserName"))
		if c.Session().GetString("UserName") == "" {
			c.JSON(http.StatusUnauthorized, dto.PasswordNotMatchErrRsp())
			return
		}
		c.Next()
	})*/
	framework.Config.Websocket.Endpoint = "/websocket"

	ws := framework.Websocket
	ws.OnConnection(func(c websocket.Connection) {
		var uid int = util.RandomInt(0, 100)
		c.OnMessage(func(data []byte) {
			message := string(data)
			id, expire, valid := util.DecodeToken(message, privateToken)
			if valid && expire > time.Now().UnixNano() {
				fmt.Println("ok!", id)
				return
			}
			fmt.Println(uid)
			c.EmitMessage([]byte("Me: " + message))
		})
		c.OnDisconnect(func() {
		})
	})
}
