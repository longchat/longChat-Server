package api

import (
	"log"
	"strings"
	"time"

	"github.com/kataras/iris"
	"github.com/kataras/iris/sessions"
	"github.com/kataras/iris/sessions/sessiondb/redis"
	"github.com/kataras/iris/sessions/sessiondb/redis/service"

	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/idService/generator"
	"github.com/longchat/longChat-Server/storageService/storage"
)

func Init(framework *iris.Application, idGen *generator.IdGenerator, store *storage.Storage) {
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
	sess := sessions.New(sessions.Config{
		Cookie:  cookie,
		Expires: time.Duration(2) * time.Hour,
	})

	db := redis.New(service.Config{Network: service.DefaultRedisNetwork,
		Addr:        redisAddr,
		Password:    redisPsw,
		Database:    "0",
		MaxIdle:     4,
		MaxActive:   4,
		IdleTimeout: service.DefaultRedisIdleTimeout,
		Prefix:      redisPrefix})

	sess.UseDatabase(db)

	framework.Use(iris.Gzip)
	framework.StaticWeb("/static", staicPath)

	addrStr, err := config.GetConfigString(consts.LeafMsgServiceAddress)
	if err != nil {
		log.Fatalf(consts.ErrGetConfigFailed(consts.SessionCookieName, err))
	}
	addrs := strings.Split(addrStr, ",")
	ua := UserApi{idGen: idGen, store: store, serverAddrs: addrs}
	ua.RegisterRoute(framework)
	au := AuthApi{store: store}
	au.RegisterRoute(framework, sess)
	ga := GroupApi{idGen: idGen, store: store}
	ga.RegisterRoute(framework, sess)

	staicPath, err := config.GetConfigString(consts.ApiServiceStaticPath)
	if err != nil {
		log.Fatalf(consts.ErrGetConfigFailed(consts.ApiServiceAddress, err))
	}

}
