package message

import (
	messagepb "github.com/longchat/longChat-Server/common/protoc"
)

type hubCenter struct {
	wp *workerPool
	//优先user上线，再group上线
	userMap  map[int64]*conn
	groupMap map[int64]map[uint32]*conn
}

type online struct {
	wsConn *conn
	online *messagepb.OnlineReq
}

type removeConn struct {
	wsConn *conn
}

var (
	msgCh    chan *messagepb.MessageReq
	onlineCh chan online
	rmConnCh chan removeConn
)

func startHub() {
	hub := hubCenter{
		userMap:  make(map[int64]*conn, 512),
		groupMap: make(map[int64]map[uint32]*conn, 128),
		wp:       newWorkerPool(),
	}
	go hub.hub()
	go hub.controller()
}

func (hub *hubCenter) hub() {
	for {
		select {
		case msg := <-msgCh:
			l := len(msgCh)
			if l > 0 {
				for i := 0; i < l; i++ {
					msg2 := <-msgCh
					msg.Messages = append(msg.Messages, msg2.Messages...)
				}
			}
			hub.processMessage(msg)
		case online := <-onlineCh:
			hub.handleOnline(online)
		case rm := <-rmConnCh:
		}
	}
}

func (hub *hubCenter) controller() {

}

func (hub *hubCenter) handleOnline(req online) {
	for i := range req.online.Items {
		data := req.online.Items[i]
		if data.IsGroup {

		} else {
			user, isok := hub.userMap[data.Id]
			if data.IsOnline {
				user = req.wsConn
				hub.userMap[data.Id] = user
			} else if isok {
				delete(hub.userMap, data.Id)
			}
		}
	}
}

func (hub *hubCenter) processMessage(msg *messagepb.MessageReq) {
	var jobs map[uint32]job
	jobs = make(map[uint32]job)
	for i := range msg.Messages {
		data := msg.Messages[i]
		if data.IsGroupMessage {
			var exceptConnId uint32
			if !IsLeafServer {
				userFrom, isok := hub.userMap[data.From]
				if isok {
					exceptConnId = userFrom.Id
				}
			}

			group, isok := hub.groupMap[data.To]
			if isok {
				for k, v := range group {
					if !IsLeafServer {
						if v.Id == exceptConnId {
							continue
						}
					}
					ajob, isok := jobs[k]
					if isok {
						ajob.message.Messages = append(ajob.message.Messages, data)
					} else {
						msgReq := messagepb.MessageReq{Messages: []*messagepb.MessageReq_Message{data}}
						ajob = job{wsConn: v, message: &msgReq}
					}
					jobs[k] = ajob
				}
			}
		} else {
			userConn, isok := hub.userMap[data.To]
			if isok {
				ajob, isok := jobs[userConn.Id]
				if isok {
					ajob.message.Messages = append(ajob.message.Messages, data)
				} else {
					msgReq := messagepb.MessageReq{Messages: []*messagepb.MessageReq_Message{data}}
					ajob = job{wsConn: userConn, message: &msgReq}
				}
				jobs[userConn.Id] = ajob
			}
		}
	}
	for _, v := range jobs {
		var aworker *worker
		hub.wp.lock.Lock()
		if len(hub.wp.idle) > 0 {
			aworker = hub.wp.idle[len(hub.wp.idle)-1]
			hub.wp.idle = hub.wp.idle[:len(hub.wp.idle)-1]
			hub.wp.lock.Unlock()
		} else {
			hub.wp.lock.Unlock()
			aworker = hub.wp.pool.Get().(*worker)
		}
		aworker.ch <- v
	}
}
