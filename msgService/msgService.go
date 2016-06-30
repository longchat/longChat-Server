package main

import (
	"github.com/longchat/longChat-Server/msgService/message"

	"github.com/kataras/iris"
)

func main() {

	framework := iris.New()
	message.Init(framework)
	framework.Listen(":8082")
}
