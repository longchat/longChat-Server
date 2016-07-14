package message

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/longchat/longChat-Server/common/protoc"
	"golang.org/x/net/context"
)

var (
	newline = []byte{'\r', '\n'}
	space   = []byte{' '}
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 2 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = ((pongWait * 9) / 10) + 5

	// Maximum message size allowed from peer.
	maxMessageSize = 1024
)

type Conn struct {
	id      int
	ws      *websocket.Conn
	send    chan *protoc.MessageReq
	session *Session
}

func (c *Conn) closeConn() {
	close(c.send)
	c.ws.Close()
}

func (c *Conn) handleConn(command chan interface{}) {
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	leave := make(chan bool)
	go c.writePump(command, leave)
	for {
		_, message, err := c.ws.ReadMessage()
		fmt.Println("read msg(%s)err:=%v", string(message), err)
		if err != nil {
			break
		}
		data := MessageData{Data: message}
		if data.GetHeader() == DataTypeGroupMessage {
			var gMsg DataGroupMessage
			err := data.GetData(&gMsg)
			if err != nil {
				fmt.Println("unmarshall data(%s) failed!", string(message))
				break
			}
			gId, _ := strconv.ParseInt(gMsg.GroupId, 10, 64)
			reply, err := router.Message(context.Background(), &protoc.MessageReq{From: c.session.userId, GroupId: gId, Content: gMsg.Content, Type: gMsg.Type})
			if err != nil || reply.StatusCode != 0 {
				fmt.Println("post message to server failed!err:=%v,err:=%v", err, reply.Err)
				break
			}
		}
	}
	close(leave)
}

func (c *Conn) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *Conn) writePump(command chan interface{}, leave chan bool) {
	//key为groupid，value为lastreadid
	groupReadMap := make(map[int64]int64, 6)
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		command <- groupReadMark{groupRead: groupReadMap}
		command <- connDel{connId: c.id}
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := c.ws.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			var groupMsg DataGroupMessage = DataGroupMessage{
				Id:      fmt.Sprintf("%d", message.Id),
				From:    fmt.Sprintf("%d", message.From),
				GroupId: fmt.Sprintf("%d", message.GroupId),
				Content: message.Content,
				Type:    message.Type,
			}
			mbytes, err := json.Marshal(&groupMsg)
			if err != nil {
				fmt.Println("marshal data failed!err:=%v", *message)
			}
			w.Write(mbytes)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				message = <-c.send
				groupMsg = DataGroupMessage{
					Id:      fmt.Sprintf("%d", message.Id),
					From:    fmt.Sprintf("%d", message.From),
					GroupId: fmt.Sprintf("%d", message.GroupId),
					Content: message.Content,
					Type:    message.Type,
				}
				mbytes, err := json.Marshal(&groupMsg)
				if err != nil {
					fmt.Println("marshal data failed!err:=%v", *message)
				}
				w.Write(mbytes)
			}

			if err := w.Close(); err != nil {
				return
			}
			if message.GroupId > 0 {
				groupReadMap[message.GroupId] = message.Id
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
			command <- groupReadMark{groupRead: groupReadMap}
		case <-leave:
			return
		}
	}
}
