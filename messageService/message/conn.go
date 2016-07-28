package message

import (
	"sync"

	"github.com/gorilla/websocket"
)

const (
	MessageTypeMessage int = 0
	MessageTypeUser    int = 1
	MessageTypeGroup   int = 2
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

func (c *conn) readPump() {

}
