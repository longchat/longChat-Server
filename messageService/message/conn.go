package message

import (
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	messagepb "github.com/longchat/longChat-Server/common/protoc"
)

const (
	MessageTypeMessage int = 0
	MessageTypeOnline  int = 1
)

type connState uint8

const (
	ConnStateIdle    connState = 0
	ConnStateWorking connState = 1
)

type conn struct {
	Id    uint32
	ws    *websocket.Conn
	wLock sync.Mutex
	state connState
}

func (wsConn *conn) readPump() {
	for {
		msgType, msg, err := wsConn.ws.ReadMessage()
		if err != nil {
			break
		}
		if msgType == MessageTypeMessage {
			var messageReq messagepb.MessageReq
			err := proto.Unmarshal(msg, &messageReq)
			if err != nil {
				continue
			}
			msgCh <- message{wsConn: wsConn, messageReq: messageReq}
		} else if msgType == MessageTypeOnline {
			var onlineReq messagepb.OnlineReq
			err := proto.Unmarshal(msg, &onlineReq)
			if err != nil {
				continue
			}
			onlineCh <- online{wsConn: wsConn, onlineReq: onlineReq}
		}
	}
}

func (wsConn *conn) writeAndFlush(messageType int, pb proto.Message) error {
	writer, err := wsConn.ws.NextWriter(MessageTypeMessage)
	bytes, err := proto.Marshal(pb)
	if err != nil {
		return err
	}
	id := wsConn.Id
	wsConn.wLock.Lock()
	defer wsConn.wLock.Unlock()
	if wsConn.state == ConnStateIdle || wsConn.Id != id {
		return err
	}
	_, err = writer.Write(bytes)
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}
	return nil
}
