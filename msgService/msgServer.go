package main

import (
	slog "log"

	"github.com/kataras/iris"
	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/msgService/message"
)

func main() {
	config.InitConfig()
	addr, err := config.GetConfigString(consts.MsgServiceAddress)
	if err != nil {
		slog.Fatalln(consts.ErrGetConfigFailed(consts.MsgServiceAddress, err))
	}
	framework := iris.New()
	m := message.Messenger{}
	m.Init(framework)
	defer m.Close()
	framework.Listen(addr)
}
