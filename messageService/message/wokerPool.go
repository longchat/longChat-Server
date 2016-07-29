package message

import (
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
			w := new(worker)
			w.ch = make(chan job)
			return w
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
			wp.pool.Put(releases[i])
		}
	}
}

func (wp *workerPool) getWorkers(workers *[]*worker, jobLen int) {
	wp.idleLock.Lock()
	fetchLen := len(wp.idle) - jobLen
	if fetchLen >= 0 {
		*workers = wp.idle[fetchLen:]
		wp.idle = wp.idle[:fetchLen]
	} else {
		*workers = wp.idle[:]
		wp.idle = wp.idle[:0]
	}
	wp.idleLock.Unlock()
	if fetchLen < 0 {
		for i := 0; i < (-fetchLen); i++ {
			aworker := wp.pool.Get().(*worker)
			go wp.worker(aworker)
			*workers = append(*workers, aworker)
		}
	}
}

func (wp *workerPool) releaseWorker(worker *worker) {
	worker.lastUseTs = time.Now().UnixNano()
	wp.idleLock.Lock()
	wp.idle = append(wp.idle, worker)
	wp.idleLock.Unlock()
}

func (wp *workerPool) worker(worker *worker) {
	for info := range worker.ch {
		if info.wsConn == nil {
			break
		}
		var err error
		if len(info.message.Messages) > 0 {
			err = info.wsConn.writeAndFlush(MessageTypeMessage, &info.message)
		} else if len(info.onlineReq.Items) > 0 {
			err = info.wsConn.writeAndFlush(MessageTypeOnline, &info.onlineReq)
		}
		if err != nil {
			wp.releaseWorker(worker)
			continue
		}
	}
}
