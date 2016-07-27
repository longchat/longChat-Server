package message

import (
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	messagepb "github.com/longchat/longChat-Server/common/protoc"
)

type conn struct {
	Id   uint32
	ws   *websocket.Conn
	send chan interface{}
}

func (c *conn) readPump() {
	uLock.RLock()
	uLock.RUnlock()
}

func (c *conn) writePump() {
	for {
		select {
		case w, isok := <-c.send:
			writer, err := w.conn.ws.NextWriter(MessageReq)
			bytes, err := proto.Marshal(w.message)
			if err != nil {
				continue
			}
			writer.Write(bytes)
			writer.Close()
		}
	}
}
