package main

import (
	slog "log"

	"github.com/kataras/iris"
	"github.com/longchat/longChat-Server/apiService/api"
	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/idService/generator"
	"github.com/longchat/longChat-Server/storageService/storage"
)

func main() {
	config.InitConfig()
	accPath, err := config.GetConfigString(consts.AccessLogPath)
	if err != nil {
		slog.Fatalf(consts.ErrGetConfigFailed(consts.AccessLogPath, err))
	}
	errPath, err := config.GetConfigString(consts.ErrorLogPath)
	if err != nil {
		slog.Fatalf(consts.ErrGetConfigFailed(consts.ErrorLogPath, err))
	}
	err = log.InitLogger(errPath, accPath, 1024, 5*1024)
	if err != nil {
		slog.Fatalf("init log failed!err:=%v\n", err)
	}
	defer log.FiniLogger()

	addr, err := config.GetConfigString(consts.ApiServiceAddress)
	if err != nil {
		slog.Fatalln(consts.ErrGetConfigFailed(consts.ApiServiceAddress, err))
	}
	idGen := generator.IdGenerator{}
	err = idGen.Init(true)
	defer idGen.Close()
	if err != nil {
		slog.Fatalf("init IdGenerator failed!err:=%v\n", err)
	}
	store := storage.Storage{}
	err = store.Init()
	defer store.Close()
	if err != nil {
		slog.Fatalf("init DB failed!err:=%v\n", err)
	}

	framework := iris.New()
	api.Iint(framework, &idGen, &store)
	framework.Listen(addr)

}
