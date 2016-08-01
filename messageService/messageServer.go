package main

import (
	slog "log"

	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/messageService/message"
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
	addr, err := config.GetConfigString(consts.MsgServiceAddress)
	if err != nil {
		slog.Fatalln(consts.ErrGetConfigFailed(consts.MsgServiceAddress, err))
	}
	parentAddr, err := config.GetConfigString(consts.ParentServiceAddress)
	if err != nil {
		slog.Fatalln(consts.ErrGetConfigFailed(consts.ParentServiceAddress, err))
	}
	isLeaf, err := config.GetConfigBool(consts.IsLeafServer)
	if err != nil {
		slog.Fatalln(consts.ErrGetConfigFailed(consts.IsLeafServer, err))
	}
	store, err := storage.NewStorage()
	if err != nil {
		slog.Fatalln("init store failed!", err)
	}
	defer store.Close()
	if err != nil {
		slog.Fatalf("init DB failed!err:=%v\n", err)
	}
	message.StartServer(store, addr, parentAddr, isLeaf)
}
