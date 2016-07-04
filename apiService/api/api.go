package api

import (
	"log"
	"time"

	"github.com/kataras/iris"
	iconfig "github.com/kataras/iris/config"
	"github.com/kataras/iris/sessions/providers/redis"
	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/idService/generator"
	"github.com/longchat/longChat-Server/storageService/storage"
)

func Iint(framework *iris.Framework, idGen *generator.IdGenerator, store *storage.Storage) {
	framework.Config.Render.Rest.Gzip = true
	framework.Config.Render.Template.Gzip = true

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

	ua := UserApi{idGen: idGen, store: store}
	ua.RegisterRoute(framework)
	au := AuthApi{store: store}
	au.RegisterRoute(framework)
	ga := GroupApi{idGen: idGen, store: store}
	ga.RegisterRoute(framework)

	staicPath, err := config.GetConfigString(consts.ApiServiceStaticPath)
	if err != nil {
		log.Fatalf(consts.ErrGetConfigFailed(consts.ApiServiceAddress, err))
	}
	framework.Static("/static", staicPath, 1)
}
