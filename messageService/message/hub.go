package message

import (
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	messagepb "github.com/longchat/longChat-Server/common/protoc"
)

const (
	TypeMessage = 0
	TypeUser    = 1
	TypeGroup   = 2
)

type worker struct {
	lastUseTs int64
	ch        chan job
}

type workerPool struct {
	idle []worker
	lock sync.Mutex
}

type job struct {
	conn    *conn
	message *messagepb.MessageReq
}

var (
	userMap  map[int64]*conn
	groupMap map[int64]map[uint32]*conn
)

var msgCh chan *messagepb.MessageReq

func hub() {
	var wp = workerPool{}
	for {
		select {
		case msg := <-msgCh:
			if len(msgCh) > 0 {
				l := len(msgCh)
				for i := 0; i < l; i++ {
					msg2 := <-msgCh
					msg.Messages = append(msg.Messages, msg2.Messages...)
				}
				var jobs map[uint32]job
				jobs = make(map[uint32]job)
				for i := range msg.Messages {
					data := msg.Messages[i]
					if data.IsGroupMessage {
						group, isok := groupMap[data.To]
						if isok {
							for k, v := range group {
								ajob, isok := jobs[k]
								if isok {
									ajob.message.Messages = append(ajob.message.Messages, data)
								} else {
									msgReq := messagepb.MessageReq{Messages: []*messagepb.MessageReq_Message{data}}
									ajob = job{conn: v, message: &msgReq}
								}
								jobs[k] = ajob
							}
						}
					} else {
						userConn, isok := userMap[data.To]
						if isok {
							ajob, isok := jobs[userConn.Id]
							if isok {
								ajob.message.Messages = append(ajob.message.Messages, data)
							} else {
								msgReq := messagepb.MessageReq{Messages: []*messagepb.MessageReq_Message{data}}
								ajob = job{conn: userConn, message: &msgReq}
							}
							jobs[userConn.Id] = ajob
						}
					}
				}
				for _, v := range jobs {
					var aworker worker
					wp.lock.Lock()
					if len(wp.idle) > 0 {
						aworker = wp.idle[len(wp.idle)-1]
						wp.idle = wp.idle[:len(wp.idle)-1]
					} else {
						//todo:new worker
					}
					wp.lock.Unlock()
					aworker.ch <- v
				}
			}

		}
	}
}

func controller() {

}

func (wp *workerPool) writePump(worker *worker) {
	for {
		worker.lastUseTs = time.Now().UnixNano()
		wp.lock.Lock()
		wp.idle = append(wp.idle, *worker)
		wp.lock.Unlock()
		select {
		case info, isok := <-worker.ch:
			if !isok {
				break
			}
			writer, err := info.conn.ws.NextWriter(TypeMessage)
			bytes, err := proto.Marshal(info.message)
			if err != nil {
				continue
			}
			id := info.conn.Id
			info.conn.wLock.Lock()
			if info.conn.Id != id {
				info.conn.wLock.Unlock()
				continue
			}
			_, err = writer.Write(bytes)
			if err != nil {
				info.conn.wLock.Unlock()
				continue
			}
			err = writer.Close()
			if err != nil {
				info.conn.wLock.Unlock()
				continue
			}
		}
	}
}
