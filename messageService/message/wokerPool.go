package message

import (
	"fmt"
	"sync"
	"time"

	messagepb "github.com/longchat/longChat-Server/common/protoc"
)

const (
	MaxWorkerCount               = 64 * 1024
	MaxIdleTime    time.Duration = time.Second * 5
)

type job struct {
	wsConn    *conn
	message   messagepb.MessageReq
	onlineReq messagepb.OnlineReq
}

type worker struct {
	lastUseTs int64
	ch        chan job
}

type workerPool struct {
	idle        []*worker
	idleLock    sync.Mutex
	pool        sync.Pool
	workerCount int32
}

func newWorkerPool() *workerPool {
	wp := &workerPool{
		idle: make([]*worker, 0, 128),
	}
	wp.pool = sync.Pool{
		New: func() interface{} {
			return nil
		},
	}
	return wp
}

func (wp *workerPool) idleCleaner() {
	for {
		time.Sleep(MaxIdleTime)
		var i int
		var releases []*worker
		nowTs := time.Now().UnixNano()
		wp.idleLock.Lock()
		for i = range wp.idle {
			if wp.idle[i].lastUseTs >= nowTs-int64(MaxIdleTime) {
				break
			}
		}
		releases = wp.idle[:i]
		wp.idle = wp.idle[i:]
		wp.idleLock.Unlock()
		for j := range releases {
			//pass an empty job{} to stop the goroutine
			releases[j].ch <- job{}
			wp.pool.Put(releases[j])
		}
	}
}

func (wp *workerPool) getWorkers(workers *[]*worker, jobLen int) {
	var workerSelf []*worker
	wp.idleLock.Lock()
	remain := len(wp.idle) - jobLen
	if remain >= 0 {
		workerSelf = wp.idle[remain:]
		wp.idle = wp.idle[:remain]
	} else {
		workerSelf = wp.idle[:]
		wp.idle = wp.idle[:0]
	}
	copy(*workers, workerSelf)
	wp.idleLock.Unlock()
	if remain < 0 {
		for i := 0; i < (-remain); i++ {
			aworker := wp.pool.Get()
			if aworker == nil {
				w := new(worker)
				w.ch = make(chan job)
				aworker = w
			}
			go wp.worker(aworker.(*worker))
			(*workers)[jobLen+remain+i] = aworker.(*worker)
		}
	}
}

func (wp *workerPool) releaseWorker(worker *worker) {
	wp.idleLock.Lock()
	worker.lastUseTs = time.Now().UnixNano()
	wp.idle = append(wp.idle, worker)
	wp.idleLock.Unlock()
}

func (wp *workerPool) worker(worker *worker) {
	for info := range worker.ch {
		if info.wsConn == nil {
			break
		}
		if len(info.message.Messages) > 0 {
			info.wsConn.writeAndFlush(MessageTypeMessage, &info.message)
		} else if len(info.onlineReq.Items) > 0 {
			info.wsConn.writeAndFlush(MessageTypeOnline, &info.onlineReq)
		}
		if len(worker.ch) != 0 {
			panic("worker.ch's length must be zero")
		}
		wp.releaseWorker(worker)
	}
}
