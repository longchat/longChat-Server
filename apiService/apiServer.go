package main

import (
	"flag"
	slog "log"

	"github.com/kataras/iris/v12"
	"github.com/longchat/longChat-Server/apiService/api"
	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/graphService/graph"
	"github.com/longchat/longChat-Server/idService/generator"
	"github.com/longchat/longChat-Server/storageService/storage"
)

func main() {
	pconfig := flag.String("config", "../config.cfg", "config file")
	psection := flag.String("section", "dev", "section of config file to apply")
	flag.Parse()
	config.InitConfig(pconfig, psection)
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
	store, err := storage.NewStorage()
	if err != nil {
		slog.Fatalf("init DB failed!err:=%v\n", err)
	}
	defer store.Close()

	neoUrl, err := config.GetConfigString(consts.Neo4JDbUrl)
	if err != nil {
		slog.Fatalln(consts.ErrGetConfigFailed(consts.Neo4JDbUrl, err))
	}
	err = graph.NewDb(neoUrl)
	if err != nil {
		slog.Fatalf("init graph service failed: %v", err)
	}
	defer graph.FinDb()
	framework := iris.New()
	api.Iint(framework, &idGen, store)
	framework.Run(iris.Addr(addr))
}
