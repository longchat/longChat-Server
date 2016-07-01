package api

import (
	"time"

	"github.com/kataras/iris"
	"github.com/kataras/iris/config"
	"github.com/kataras/iris/sessions/providers/redis"
	"github.com/longchat/longChat-Server/idService/generator"
	"github.com/longchat/longChat-Server/storageService/storage"
)

func Iint(framework *iris.Framework, idGen *generator.IdGenerator, store *storage.Storage) {
	framework.Config.Render.Rest.Gzip = true
	framework.Config.Render.Template.Gzip = true
	framework.Config.Sessions = config.Sessions{
		Provider:   "redis",
		Cookie:     "longchatSess",
		Expires:    config.CookieExpireNever,
		GcDuration: time.Duration(2) * time.Hour,
	}
	redis.Config.Network = "tcp"
	redis.Config.Addr = "127.0.0.1:6379"
	redis.Config.Prefix = "Sess"
	redis.Config.Password = "123456"

	ua := UserApi{idGen: idGen, store: store}
	ua.RegisterRoute(framework)
	au := AuthApi{store: store}
	au.RegisterRoute(framework)
}
