package message

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	messagepb "github.com/longchat/longChat-Server/common/protoc"
	"github.com/longchat/longChat-Server/common/util"
)

const (
	MessageTypeMessage int = 1
	MessageTypeOnline  int = 2
	MessageTypeGroup   int = 3
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

func (wsConn *conn) readPump(userId int64) {
	for {
		_, msg, err := wsConn.ws.ReadMessage()
		if err != nil {
			break
		}
		if int(msg[0]) == MessageTypeMessage {
			msg = msg[1:]
			var messageReq messagepb.MessageReq
			err := proto.Unmarshal(msg, &messageReq)
			if err != nil || len(messageReq.Messages) == 0 {
				continue
			}
			if userId != 0 {
				for i := range messageReq.Messages {
					messageReq.Messages[i].From = util.Int2Bytes(userId)
				}
			}
			msgCh <- message{wsConn: wsConn, messageReq: messageReq}
		} else if int(msg[0]) == MessageTypeOnline {
			msg = msg[1:]
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
	bytes, err := proto.Marshal(pb)
	if err != nil {
		return err
	}
	rb := make([]byte, len(bytes)+1)
	rb[0] = byte(messageType)
	copy(rb[1:], bytes)
	id := wsConn.Id
	wsConn.wLock.Lock()
	defer wsConn.wLock.Unlock()
	if wsConn.state == ConnStateIdle || wsConn.Id != id {
		return err
	}
	wsConn.ws.SetWriteDeadline(time.Now().Add(time.Second * 2))
	writer, err := wsConn.ws.NextWriter(websocket.BinaryMessage)
	if err != nil {
		wsConn.state = ConnStateIdle
		return err
	}
	_, err = writer.Write(rb)
	if err != nil {
		wsConn.state = ConnStateIdle
		return err
	}
	err = writer.Close()
	if err != nil {
		wsConn.state = ConnStateIdle
		return err
	}
	return nil
}
