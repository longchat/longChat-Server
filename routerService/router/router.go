package router

import (
	"errors"
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
	gLock      *sync.RWMutex
	groupAddrs map[int64][]string
	groupChan  map[int64]chan *message
	senderChan map[string]chan *message
}

type message struct {
	id      int64
	from    string
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
	return nil
}

func (r *Router) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
}

func (r *Router) messageIn(req *MessageInReq) {
	ch := groupChan[req.Groupid]
	if ch == nil {
		r.gLock.Lock()
		ch = groupChan[req.Groupid]
		if ch == nil {

		}
		r.gLock.Unlock()
	}
}

func (r *Router) groupUp(req *GroupUpReq) {
	r.gLock.Lock()
	addrs, isok := r.groups[req.GroupId]
	if !isok {
		addrs := make([]string, 1)
		addrs[0] = req.ServerAddr
		groupChan[req.GroupId] = make(chan *message)
	} else {
		for i := range addrs {
			if addrs[i] == addr {
				return
			}
		}
		addrs = append(addrs, req.ServerAddr)
	}
	r.groups[id] = addrs
	r.gLock.Unlock()
}
