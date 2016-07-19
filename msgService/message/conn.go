package message

import (
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
	send    chan interface{}
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
		} else if data.GetHeader() == DataTypeUserMessage {
			var msg DataUserMessage
			err := data.GetData(&msg)
			if err != nil {
				fmt.Println("unmarshall data(%s) failed!", string(message))
				break
			}
			to, err := strconv.ParseInt(msg.To, 10, 64)
			if err != nil {
				continue
			}
			msgChan <- &protoc.MessageReq{Id: time.Now().UnixNano(), From: c.session.userId, To: to, Content: msg.Content, Type: msg.Type}
		}
	}
	close(leave)
}

func (c *Conn) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func packData(data interface{}) []byte {
	var err error
	var mbytes []byte
	var md MessageData
	switch value := data.(type) {
	case *protoc.MessageReq:
		var groupMsg DataGroupMessage = DataGroupMessage{
			Id:      fmt.Sprintf("%d", value.Id),
			From:    fmt.Sprintf("%d", value.From),
			GroupId: fmt.Sprintf("%d", value.GroupId),
			Content: value.Content,
			Type:    value.Type,
		}
		mbytes, err = md.Serialize(DataTypeGroupMessage, &groupMsg)
		if err != nil {
			fmt.Println("marshal data failed!err:=%v", *value)
		}
	case *getGroupMembers:
		var groupMember DataGroupMembers = DataGroupMembers{
			GroupId: value.groupId,
			UserIds: value.userIds,
		}
		mbytes, err = md.Serialize(DataTypeGroupMemberList, &groupMember)
		if err != nil {
			fmt.Println("marshal data failed!err:=%v", *value)
		}
	}
	return mbytes
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
			mbytes := packData(message)
			w.Write(mbytes)
			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				message = <-c.send
				mbytes := packData(message)
				w.Write(mbytes)
			}

			if err := w.Close(); err != nil {
				return
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
