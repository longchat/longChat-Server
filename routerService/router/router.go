package router

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/common/protoc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func convertIpTo4Byte(ip net.IP) [4]byte {
	return [4]byte{ip[0], ip[1], ip[2], ip[3]}
}

type Router struct {
	gLock  *sync.RWMutex
	sLock  *sync.RWMutex
	groups map[int64]groupworker
	sChans map[string]chan interface{}
}

type groupworker struct {
	senders []string
	mChan   chan *protoc.MessageReq
}

func (r *Router) Init() error {
	r.gLock = new(sync.RWMutex)
	r.sLock = new(sync.RWMutex)
	r.groups = make(map[int64]groupworker, 200)
	r.sChans = make(map[string]chan interface{}, 256)
	return nil
}

func (r *Router) sender(addr string, sCh chan interface{}) error {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.ERROR.Printf("dial to MsgServer(%s) failed!err:=%s", addr, err)
		return err
	}
	defer conn.Close()
	client := protoc.NewMessageClient(conn)
	for data := range sCh {
		var reply *protoc.BaseRsp
		var err error
		switch value := data.(type) {
		case *protoc.MessageReq:
			reply, err = client.Message(context.Background(), value)
		case *protoc.UserUpReq:
			reply, err = client.UserUp(context.Background(), value)
		case *protoc.UserDownReq:
			reply, err = client.UserDown(context.Background(), value)
		}
		if err != nil || reply.StatusCode != 0 {
			if err == nil {
				err = errors.New(reply.Err)
			}
			log.ERROR.Printf("send message to MsgServer failed!err:=%s", err)
			return err
		}
	}
	return nil
}

func (r *Router) groupFetcher(gId int64, gCh chan *protoc.MessageReq) {
	for msg := range gCh {
		r.gLock.RLock()
		group := r.groups[gId]
		r.gLock.RUnlock()
		msg.Id = time.Now().UnixNano()
		for i := range group.senders {
			r.sLock.RLock()
			sCh := r.sChans[group.senders[i]]
			r.sLock.RUnlock()
			sCh <- msg
		}
	}
}

func (r *Router) MessageIn(req *protoc.MessageReq) {
	r.gLock.RLock()
	group, isok := r.groups[req.GroupId]
	if !isok {
		r.gLock.RUnlock()
		return
	} else {
		mCh := group.mChan
		r.gLock.RUnlock()
		mCh <- req
	}
}

func (r *Router) UserUp(req *protoc.UserUpReq) {
	r.sLock.RLock()
	for k, v := range r.sChans {
		if k != req.ServerAddr {
			v <- req
		}
	}
	r.sLock.RUnlock()
}

func (r *Router) UserDown(req *protoc.UserDownReq) {
	r.sLock.RLock()
	for k, v := range r.sChans {
		if k != req.ServerAddr {
			v <- req
		}
	}
	r.sLock.RUnlock()
}

func (r *Router) GroupUp(req *protoc.GroupUpReq) {
	ip := req.ServerAddr
	r.gLock.Lock()
	groupWork, isok := r.groups[req.GroupId]
	if !isok {
		msgChan := make(chan *protoc.MessageReq)
		senders := make([]string, 1)
		senders[0] = ip
		groupWork = groupworker{senders: senders, mChan: msgChan}
		r.groups[req.GroupId] = groupWork
		r.gLock.Unlock()
		go r.groupFetcher(req.GroupId, msgChan)
	} else {
		for i := range groupWork.senders {
			if groupWork.senders[i] == ip {
				r.gLock.Unlock()
				return
			}
		}
		groupWork.senders = append(groupWork.senders, ip)
		r.groups[req.GroupId] = groupWork
		r.gLock.Unlock()
	}
}

func (r *Router) ServerUp(req *protoc.ServerUpReq) {
	r.sLock.Lock()
	_, isok := r.sChans[req.ServerAddr]
	if isok {
		r.sLock.Unlock()
		return
	}
	chans := make(chan interface{}, 1000)
	r.sChans[req.ServerAddr] = chans
	r.sLock.Unlock()
	go r.sender(req.ServerAddr, chans)
}
