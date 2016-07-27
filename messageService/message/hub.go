package message

import (
	"sync"
)

const (
	MessageReq = 0
	UserReq    = 1
	GroupReq   = 2
)

type worker struct {
	isWorking  bool
	lastWorkTs int64
	writeCh    chan writeInfo
}

type workerPool struct {
	workers     []worker
	writeChPool sync.Pool
}

type connInfo struct {
	userIds  []int64
	groupIds []int64
	conn     conn
}

type writeInfo struct {
	conn    *conn
	message *messagepb.MessageReq
}

var (
	uLock    *sync.RWMutex
	cLock    *sync.RWMutex
	gLock    *sync.RWMutex
	userMap  map[int64]uint32
	connMap  map[uint32]ConnInfo
	groupMap map[int64]map[uint32]struct{}
)

func controller() {

}
