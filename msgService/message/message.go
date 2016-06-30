package message

import (
	"fmt"

	"github.com/kataras/iris"
	"github.com/kataras/iris/websocket"
)

func Init(framework *iris.Framework) {
	framework.Config.Websocket.Endpoint = "/websocket"

	ws := framework.Websocket
	ws.OnConnection(func(c websocket.Connection) {

		c.OnMessage(func(data []byte) {
			message := string(data)
			c.To(websocket.Broadcast).EmitMessage([]byte("Message from: " + c.ID() + "-> " + message))
			c.EmitMessage([]byte("Me: " + message))
		})

		c.OnDisconnect(func() {
			fmt.Printf("\nConnection with ID: %s has been disconnected!", c.ID())
		})
	})
}
