package router

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/common/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Router struct {
	conn       *grpc.ClientConn
	rpcEnabled bool
	client     routerClient

	gLock  *sync.RWMutex
	sLock  *sync.RWMutex
	groups map[int64]groupworker
	sChans map[net.IP]chan message
}

type groupworker struct {
	senders []net.IP
	mChan   chan message
}

type message struct {
	id      int64
	from    int64
	to      string
	content string
	msgtype string
}

func (r *Router) Init(rpcEnabled bool) error {
	r.rpcEnabled = rpcEnabled
	if rpcEnabled {
		address, err := config.GetConfigString(consts.RouterServiceAddress)
		if err != nil {
			return errors.New(consts.ErrGetConfigFailed(consts.RouterServiceAddress, err))
		}
		r.conn, err = grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			return errors.New(consts.ErrGetConfigFailed(consts.IdServiceAddress, err))
		}
		r.client = NewRouterClient(r.conn)
	}
	r.gLock = new(sync.RWMutex)
	r.sLock = new(sync.RWMutex)
	return nil
}

func (r *Router) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
}

func (r *Router) groupFetcher(gId int64, gCh chan message) {
	for msg := range gCh {
		r.gLock.RLock()
		group := r.groups[gId]
		r.gLock.RUnlock()
		for i := range group.senders {
			r.sLock.RLock()
			sCh := r.sChans[group.senders[i]]
			r.sLock.RUnlock()
			sCh <- message
		}
	}
}

func (r *Router) sender(addr string, sCh chan message) {
	for msg := range sCh {

	}
}

func (r *Router) messageIn(req *MessageInReq) {
	r.gLock.RLock()
	group, isok := r.groups[req.GroupId]
	if !isok {
		r.gLock.RUnlock()
		return
	} else {
		mCh = group.mChan
		r.gLock.RUnlock()
		mCh <- message{from: req.From, to: req.To, content: req.Content, msgtype: req.Type}
	}
}

func (r *Router) groupUp(req *GroupUpReq) {
	ip := net.ParseIP(req.ServerAddr)
	r.gLock.Lock()
	groupWork, isok := r.groups[req.GroupId]
	if !isok {
		msgChan := make(chan message)
		senders := make([]net.IP, 1)
		senders[0] = net.ParseIP(ip)
		groupWork = groupworker{senders: senders, mChan: msgChan}
		r.groups[req.GroupId] = groupWork
		r.gLock.Unlock()
		go r.groupFetcher(req.GroupId, msgChan)
	} else {
		for i := range groupWork.senders {
			if groupWork[i] == ip {
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
	_, isok := r.sChans[ip]
	if isok {
		r.sLock.Unlock()
		return
	}
	r.sChans[ip] = make(chan message, 1000)
	r.sLock.Unlock()
	go sender(req.ServerAddr, sChan)

}
