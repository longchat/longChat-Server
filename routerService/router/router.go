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
	sChans map[[4]byte]chan protoc.MessageReq
}

type groupworker struct {
	senders []net.IP
	mChan   chan protoc.MessageReq
}

func (r *Router) Init() error {
	r.gLock = new(sync.RWMutex)
	r.sLock = new(sync.RWMutex)
	return nil
}

func (r *Router) sender(addr string, sCh chan protoc.MessageReq) error {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.ERROR.Printf("dial to MsgServer(%s) failed!err:=%s", addr, err)
		return err
	}
	defer conn.Close()
	client := protoc.NewMessageClient(conn)

	for data := range sCh {
		reply, err := client.Message(context.Background(), &data)
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

func (r *Router) groupFetcher(gId int64, gCh chan protoc.MessageReq) {
	for msg := range gCh {
		r.gLock.RLock()
		group := r.groups[gId]
		r.gLock.RUnlock()
		msg.Id = time.Now().UnixNano()
		for i := range group.senders {
			r.sLock.RLock()
			sCh := r.sChans[convertIpTo4Byte(group.senders[i])]
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
		mCh <- protoc.MessageReq{From: req.From, To: req.To, Content: req.Content, Type: req.Type}
	}
}

func (r *Router) GroupUp(req *protoc.GroupUpReq) {
	ip := net.ParseIP(req.ServerAddr)
	r.gLock.Lock()
	groupWork, isok := r.groups[req.GroupId]
	if !isok {
		msgChan := make(chan protoc.MessageReq)
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

func (r *Router) ServerUp(req *protoc.ServerUpReq) {
	ip := net.ParseIP(req.ServerAddr)
	r.sLock.Lock()
	_, isok := r.sChans[convertIpTo4Byte(ip)]
	if isok {
		r.sLock.Unlock()
		return
	}
	chans := make(chan protoc.MessageReq, 1000)
	r.sChans[convertIpTo4Byte(ip)] = chans
	r.sLock.Unlock()
	go r.sender(req.ServerAddr, chans)
}
