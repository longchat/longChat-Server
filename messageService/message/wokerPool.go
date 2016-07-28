package message

import (
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	messagepb "github.com/longchat/longChat-Server/common/protoc"
)

const (
	MaxWorkerCount = 64 * 1024
)

type job struct {
	wsConn  *conn
	message *messagepb.MessageReq
}

type worker struct {
	lastUseTs int64
	ch        chan job
}

type workerPool struct {
	idle        []*worker
	lock        sync.Mutex
	pool        sync.Pool
	workerCount int32
}

func newWorkerPool() *workerPool {
	wp := &workerPool{
		idle: make([]*worker, 0, 128),
	}
	wp.pool = sync.Pool{
		New: func() interface{} {
			w := new(worker)
			w.ch = make(chan job)
			go wp.working(w)
			return w
		},
	}
	return wp
}

func (wp *workerPool) releaseWorker(worker *worker) {
	worker.lastUseTs = time.Now().UnixNano()
	wp.lock.Lock()
	wp.idle = append(wp.idle, worker)
	wp.lock.Unlock()
}

func (wp *workerPool) working(worker *worker) {
	for {
		select {
		case info, isok := <-worker.ch:
			if !isok {
				break
			}
			writer, err := info.wsConn.ws.NextWriter(MessageTypeMessage)
			bytes, err := proto.Marshal(info.message)
			if err != nil {
				wp.releaseWorker(worker)
				continue
			}
			id := info.wsConn.Id
			info.wsConn.wLock.Lock()
			if info.wsConn.state == ConnStateIdle || info.wsConn.Id != id {
				info.wsConn.wLock.Unlock()
				wp.releaseWorker(worker)
				continue
			}
			_, err = writer.Write(bytes)
			if err != nil {
				info.wsConn.wLock.Unlock()
				wp.releaseWorker(worker)
				continue
			}
			err = writer.Close()
			if err != nil {
				info.wsConn.wLock.Unlock()
				wp.releaseWorker(worker)
				continue
			}
		}
	}

}
