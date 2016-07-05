package router

import (
	"net"
	"sync"
	"time"

	"github.com/longchat/longChat-Server/common/channel"
	"github.com/longchat/longChat-Server/common/channel/protoc"

	"google.golang.org/grpc"
)

func convertIpTo4Byte(ip net.IP) [4]byte {
	var b [4]byte
	b[0] = ip[0]
	b[1] = ip[1]
	b[2] = ip[2]
	b[3] = ip[3]
	return b
}

type Router struct {
	conn       *grpc.ClientConn
	rpcEnabled bool
	client     routerClient

	gLock  *sync.RWMutex
	sLock  *sync.RWMutex
	groups map[int64]groupworker
	sChans map[[4]byte]chan protoc.Message
}

type groupworker struct {
	senders []net.IP
	mChan   chan protoc.Message
}

func (r *Router) Init(rpcEnabled bool) error {
	r.gLock = new(sync.RWMutex)
	r.sLock = new(sync.RWMutex)
	return nil
}

func (r *Router) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
}

func (r *Router) groupFetcher(gId int64, gCh chan protoc.Message) {
	for msg := range gCh {
		r.gLock.RLock()
		group := r.groups[gId]
		r.gLock.RUnlock()
		msg.Id = time.Now().UnixNano()
		for i := range group.senders {
			//r.sLock.RLock()
			//sCh := r.sChans[convertIpTo4Byte(group.senders[i])]
			//r.sLock.RUnlock()
			//sCh <- msg
			channel.GetMsgChannel(group.senders[i]) <- msg
		}
	}
}

func (r *Router) messageIn(req *MessageInReq) {
	r.gLock.RLock()
	group, isok := r.groups[req.GroupId]
	if !isok {
		r.gLock.RUnlock()
		return
	} else {
		mCh := group.mChan
		r.gLock.RUnlock()
		mCh <- protoc.Message{From: req.From, To: req.To, Content: req.Content, Type: req.Type}
	}
}

func (r *Router) groupUp(req *GroupUpReq) {
	ip := net.ParseIP(req.ServerAddr)
	r.gLock.Lock()
	groupWork, isok := r.groups[req.GroupId]
	if !isok {
		msgChan := make(chan protoc.Message)
		senders := make([]net.IP, 1)
		senders[0] = ip
		groupWork = groupworker{senders: senders, mChan: msgChan}
		r.groups[req.GroupId] = groupWork
		r.gLock.Unlock()
		go r.groupFetcher(req.GroupId, msgChan)
	} else {
		for i := range groupWork.senders {
			if convertIpTo4Byte(groupWork.senders[i]) == convertIpTo4Byte(ip) {
				r.gLock.Unlock()
				return
			}
		}
		groupWork.senders = append(groupWork.senders, ip)
		r.groups[req.GroupId] = groupWork
		r.gLock.Unlock()
	}

}

func (r *Router) serverUp(req *ServerUpReq) {
	ip := net.ParseIP(req.ServerAddr)
	r.sLock.Lock()
	_, isok := r.sChans[convertIpTo4Byte(ip)]
	if isok {
		r.sLock.Unlock()
		return
	}
	r.sChans[convertIpTo4Byte(ip)] = make(chan protoc.Message, 1000)
	r.sLock.Unlock()
}
