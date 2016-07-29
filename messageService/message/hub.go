package message

import (
	"fmt"
	"time"

	messagepb "github.com/longchat/longChat-Server/common/protoc"
)

const (
	ForceFlushMessageCount uint32        = 256
	JobFlushInterval       time.Duration = time.Millisecond * 5
)

type hubCenter struct {
	parentJob job
	wp        *workerPool
	//优先user上线，再group上线
	userMap  map[int64]*conn
	groupMap map[int64]map[uint32]*conn

	jobs         map[uint32]job
	messageCount uint32
}

type message struct {
	messageReq messagepb.MessageReq
	wsConn     *conn
}

type online struct {
	wsConn    *conn
	onlineReq messagepb.OnlineReq
}

type removeConn struct {
	wsConn *conn
}

var (
	msgCh    chan message
	onlineCh chan online
	rmConnCh chan removeConn
)

func startHub(parentConn *conn) {
	hub := hubCenter{
		parentJob: job{
			wsConn: parentConn,
		},
		userMap:  make(map[int64]*conn, 1024),
		groupMap: make(map[int64]map[uint32]*conn, 128),
		wp:       newWorkerPool(),
		jobs:     make(map[uint32]job, 128),
	}
	go hub.hub()
}

func (hub *hubCenter) hub() {
	tickCh := time.Tick(JobFlushInterval)
	for {
		select {
		case msg := <-msgCh:
			hub.processMessage(msg)
		case online := <-onlineCh:
			hub.handleOnline(online)
		case rm := <-rmConnCh:
			hub.removeConn(rm)
		case _ = <-tickCh:
			hub.dispatchJobs()
		}
	}
}

func (hub *hubCenter) removeConn(rmConn removeConn) {
	for k, v := range hub.userMap {
		if v.Id == rmConn.wsConn.Id {
			delete(hub.userMap, k)
			if hasParentServer {
				hub.parentJob.onlineReq.Items = append(hub.parentJob.onlineReq.Items, &messagepb.OnlineReq_Item{
					Id:       k,
					IsGroup:  false,
					IsOnline: false,
				})
				hub.messageCount++
			}
		}
	}
	for k, v := range hub.groupMap {
		for k2, v2 := range v {
			if v2.Id == rmConn.wsConn.Id {
				delete(v, k2)
				if len(v) == 0 {
					delete(hub.groupMap, k)
					if hasParentServer {
						hub.parentJob.onlineReq.Items = append(hub.parentJob.onlineReq.Items, &messagepb.OnlineReq_Item{
							Id:       k,
							IsGroup:  true,
							IsOnline: false,
						})
						hub.messageCount++
					}
				} else {
					hub.groupMap[k] = v
				}
				break
			}
		}
	}
	if hub.messageCount >= ForceFlushMessageCount {
		hub.dispatchJobs()
	}
}

func (hub *hubCenter) handleOnline(req online) {
	for i := range req.onlineReq.Items {
		data := req.onlineReq.Items[i]
		if data.IsGroup {
			group, isok := hub.groupMap[data.Id]
			if isok {
				conns, isok := group[req.wsConn.Id]
				if data.IsOnline {
					conns = req.wsConn
					group[req.wsConn.Id] = conns
				} else if isok {
					delete(group, req.wsConn.Id)
					if len(group) == 0 {
						delete(hub.groupMap, data.Id)
						if hasParentServer {
							if req.wsConn.Id == hub.parentJob.wsConn.Id {
								panic(fmt.Sprintf("onlineReq can't come from parent server"))
							}
							hub.parentJob.onlineReq.Items = append(hub.parentJob.onlineReq.Items, data)
							hub.messageCount++
						}
					}
				}
			} else if data.IsOnline {
				group = make(map[uint32]*conn, 10)
				group[req.wsConn.Id] = req.wsConn
				if hasParentServer {
					if req.wsConn.Id == hub.parentJob.wsConn.Id {
						panic(fmt.Sprintf("onlineReq can't come from parent server"))
					}
					hub.parentJob.onlineReq.Items = append(hub.parentJob.onlineReq.Items, data)
					hub.messageCount++
				}
			}
			if len(group) > 0 {
				hub.groupMap[data.Id] = group
			}
		} else {
			user, isok := hub.userMap[data.Id]
			if data.IsOnline {
				user = req.wsConn
				hub.userMap[data.Id] = user
			} else if isok {
				delete(hub.userMap, data.Id)
			}
			if hasParentServer {
				if req.wsConn.Id == hub.parentJob.wsConn.Id {
					panic(fmt.Sprintf("onlineReq can't come from parent server"))
				}
				hub.parentJob.onlineReq.Items = append(hub.parentJob.onlineReq.Items, data)
				hub.messageCount++
			}
		}
	}
	if hub.messageCount >= ForceFlushMessageCount {
		hub.dispatchJobs()
	}
}

func (hub *hubCenter) processMessage(msg message) {
	for i := range msg.messageReq.Messages {
		data := msg.messageReq.Messages[i]
		if data.IsGroupMessage {
			var exceptConnId uint32
			if !isLeafServer {
				userFromConn, isok := hub.userMap[data.From]
				if isok {
					exceptConnId = userFromConn.Id
				}
			}
			group, isok := hub.groupMap[data.To]
			if isok {
				for k, v := range group {
					if !isLeafServer {
						if v.Id == exceptConnId {
							continue
						}
					}
					ajob, isok := hub.jobs[k]
					if isok {
						ajob.message.Messages = append(ajob.message.Messages, data)
					} else {
						msgReq := messagepb.MessageReq{Messages: []*messagepb.MessageReq_Message{data}}
						ajob = job{wsConn: v, message: msgReq}
					}
					hub.jobs[k] = ajob
					hub.messageCount++
				}
			}
			if hasParentServer && msg.wsConn.Id != hub.parentJob.wsConn.Id {
				hub.parentJob.message.Messages = append(hub.parentJob.message.Messages, data)
				hub.messageCount++
			}
		} else {
			userConn, isok := hub.userMap[data.To]
			if isok {
				ajob, isok := hub.jobs[userConn.Id]
				if isok {
					ajob.message.Messages = append(ajob.message.Messages, data)

				} else {
					msgReq := messagepb.MessageReq{Messages: []*messagepb.MessageReq_Message{data}}
					ajob = job{wsConn: userConn, message: msgReq}
				}
				hub.jobs[userConn.Id] = ajob
				hub.messageCount++
			} else if hasParentServer && msg.wsConn.Id != hub.parentJob.wsConn.Id {
				hub.parentJob.message.Messages = append(hub.parentJob.message.Messages, data)
				hub.messageCount++
			}
		}
	}
	if hub.messageCount >= ForceFlushMessageCount {
		hub.dispatchJobs()
	}
}

func (hub *hubCenter) dispatchJobs() {
	needJobCount := len(hub.jobs)
	parentJobCount := 0
	if hasParentServer && (len(hub.parentJob.message.Messages) > 0 || len(hub.parentJob.onlineReq.Items) > 0) {
		parentJobCount++
	}
	var workers []*worker
	hub.wp.getWorkers(&workers, needJobCount+parentJobCount)
	if len(workers) != needJobCount {
		panic(fmt.Sprintf("the number of workers(%d) doesn't equal the nubmer of jobs(%d)", len(workers), needJobCount))
	}
	var i int
	for _, v := range hub.jobs {
		workers[i].ch <- v
		i++
	}
	workers[i].ch <- hub.parentJob
	if parentJobCount > 0 {
		hub.parentJob.message.Messages = make([]*messagepb.MessageReq_Message, 0, 50)
		hub.parentJob.onlineReq.Items = make([]*messagepb.OnlineReq_Item, 0, 10)
	}
	hub.jobs = make(map[uint32]job, 128)
}
