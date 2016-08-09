package api

import (
	"log"
	"time"

	"github.com/iris-contrib/sessiondb/redis"
	"github.com/iris-contrib/sessiondb/redis/service"
	"github.com/kataras/iris"
	iconfig "github.com/kataras/iris/config"
	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/idService/generator"
	"github.com/longchat/longChat-Server/storageService/storage"
)

func Iint(framework *iris.Framework, idGen *generator.IdGenerator, store *storage.Storage) {
	framework.Config.Gzip = true

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
		Cookie:     cookie,
		GcDuration: time.Duration(2) * time.Hour,
	}

	db := redis.New(service.Config{Network: service.DefaultRedisNetwork,
		Addr:          redisAddr,
		Password:      redisPsw,
		Database:      "0",
		MaxIdle:       4,
		MaxActive:     4,
		IdleTimeout:   service.DefaultRedisIdleTimeout,
		Prefix:        redisPrefix,
		MaxAgeSeconds: service.DefaultRedisMaxAgeSeconds}) // optionally configure the bridge between your redis server

	framework.UseSessionDB(db)

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
