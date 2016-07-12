package message

import (
	"encoding/json"
	"fmt"
	"log"
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
	pingPeriod = ((pongWait * 9) / 10) + 6

	// Maximum message size allowed from peer.
	maxMessageSize = 1024
)

// Conn is an middleman between the websocket connection and the hub.
type Conn struct {
	id int
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan *protoc.MessageReq
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Conn) readPump(uid int64, gids []int64) {
	defer func() {
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		data := MessageData{Data: message}
		if data.GetHeader() == DataTypeGroupMessage {
			var gMsg DataGroupMessage
			err := data.GetData(&gMsg)
			if err != nil {
				fmt.Println("unmarshall data(%s) failed!", string(message))
				c.ws.Close()
				close(c.send)
				break
			}
			gId, _ := strconv.ParseInt(gMsg.GroupId, 10, 64)
			reply, err := router.Message(context.Background(), &protoc.MessageReq{From: uid, GroupId: gId, Content: gMsg.Content, Type: gMsg.Type})
			if err != nil || reply.StatusCode != 0 {
				fmt.Println("post message to server failed!err:=%v,err:=%v", err, reply.Err)
			}
		}
	}
	command <- connDel{conn: c, userId: uid, groupId: gids}
}

// write writes a message with the given message type and payload.
func (c *Conn) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Conn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
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
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
