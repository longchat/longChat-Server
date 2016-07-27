package message

import (
	"sync"

	"github.com/gorilla/websocket"
)

type conn struct {
	Id       uint32
	userIds  []int64
	groupIds []int64
	ws       *websocket.Conn
	wLock    sync.Mutex
}

func (c *conn) readPump() {

}
