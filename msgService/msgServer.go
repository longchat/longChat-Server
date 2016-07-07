package main

import (
	slog "log"
	"net/http"

	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/msgService/message"
	"github.com/longchat/longChat-Server/storageService/storage"
)

func main() {
	config.InitConfig()
	addr, err := config.GetConfigString(consts.MsgServiceAddress)
	if err != nil {
		slog.Fatalln(consts.ErrGetConfigFailed(consts.MsgServiceAddress, err))
	}
	store := storage.Storage{}
	err = store.Init()
	defer store.Close()
	if err != nil {
		slog.Fatalf("init DB failed!err:=%v\n", err)
	}

	m := message.Messenger{}
	m.Init(&store)
	defer m.Close()

	http.HandleFunc("/websocket", m.ServeWs)
	http.ListenAndServe(addr, nil)
}
