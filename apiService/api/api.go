package api

import (
	"log"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"github.com/kataras/iris/v12/sessions/sessiondb/redis"
	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/idService/generator"
	"github.com/longchat/longChat-Server/storageService/storage"
)

func Iint(framework *iris.Application, idGen *generator.IdGenerator, store *storage.Storage) {
	framework.Use(iris.Gzip)

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

	db := redis.New(redis.Config{
		Network:   "tcp",
		Addr:      redisAddr,
		Timeout:   time.Duration(30) * time.Second,
		MaxActive: 4,
		Password:  redisPsw,
		Database:  "0",
		Prefix:    redisPrefix,
		Delim:     "-",
		Driver:    redis.Redigo(), // redis.Radix() can be used instead.
	})

	sessManager := sessions.New(sessions.Config{
		Cookie:  cookie,
		Expires: 2 * time.Hour,
	})
	sessManager.UseDatabase(db)
	framework.Use(sessManager.Handler())

	addrStr, err := config.GetConfigString(consts.LeafMsgServiceAddress)
	if err != nil {
		log.Fatalf(consts.ErrGetConfigFailed(consts.SessionCookieName, err))
	}
	addrs := strings.Split(addrStr, ",")
	ua := UserApi{idGen: idGen, store: store, serverAddrs: addrs}
	ua.RegisterRoute(framework)
	au := AuthApi{store: store}
	au.RegisterRoute(framework)
	ga := GroupApi{idGen: idGen, store: store}
	ga.RegisterRoute(framework)

	staicPath, err := config.GetConfigString(consts.ApiServiceStaticPath)
	if err != nil {
		log.Fatalf(consts.ErrGetConfigFailed(consts.ApiServiceAddress, err))
	}
	framework.HandleDir("/static", staicPath)
}
