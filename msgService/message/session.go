package message

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/common/protoc"
	"golang.org/x/net/context"
)

type Session struct {
	userId        int64
	groupIds      []int64
	conns         []Conn
	command       chan interface{}
	send          chan interface{}
	maxConnNum    int
	activeGroupId int64
}

type groupReadMark struct {
	groupRead map[int64]int64
}

type connAdd struct {
	ws *websocket.Conn
}

type connDel struct {
	connId int
}

func (s *Session) closeSess() {
	close(s.send)
	close(s.command)
}

func (s *Session) startSession(ws *websocket.Conn, uid int64, gids []int64) {
	conn := Conn{id: 0, session: s, ws: ws, send: make(chan interface{}, 128)}

	s.conns = append(s.conns, conn)
	s.userId = uid
	s.groupIds = gids
	s.command = make(chan interface{}, 4)
	s.send = make(chan interface{}, 200)
	s.maxConnNum = 1
	s.activeGroupId = 1
	go conn.handleConn(s.command)
	go s.handleSession()
	command <- getGroupMembers{connId: 0, session: s, groupId: s.activeGroupId}
}

func (s *Session) handleSession() {
	for {
		select {
		case item, isok := <-s.command:
			if !isok {
				return
			}
			switch value := item.(type) {
			case connAdd:
				s.maxConnNum++
				if len(s.conns) >= 4 {
					continue
				}
				conn := Conn{id: len(s.conns), session: s, ws: value.ws, send: make(chan interface{}, 128)}
				s.conns = append(s.conns, conn)
				go conn.handleConn(s.command)
				command <- getGroupMembers{connId: conn.id, session: s, groupId: s.activeGroupId}
			case connDel:
				if s.DelConn(value) {
					return
				}
			case groupReadMark:
				//todo:sotre group read mark in db
			}
		case message, isok := <-s.send:
			if !isok {
				return
			}
			switch value := message.(type) {
			case *getGroupMembers:
				s.conns[value.connId].send <- value
			case *protoc.MessageReq:
				for i := range s.conns {
					s.conns[i].send <- value
				}
			}
			//todo:if s.conns.send block here and the conn want to close it self by send command to session then it will be deadlock groutine

		}
	}
}

func (s *Session) DelConn(del connDel) bool {
	clen := len(s.conns)
	if del.connId >= clen {
		panic(fmt.Sprintf("session's conns(%v) doesn't contain conn(%d)", s.conns, del.connId))
	} else {
		s.conns[del.connId].closeConn()
		if clen == 1 {
			command <- sessionDel{session: s, connsCount: s.maxConnNum}
			s.conns = nil
			reply, err := router.UserDown(context.Background(), &protoc.UserDownReq{ServerAddr: msgAddr, UserId: s.userId})
			if err != nil || reply.StatusCode != 0 {
				log.ERROR.Printf("userdown(%d) to router failed!\n", s.userId)
				return true
			}
			return true
		} else {
			copy(s.conns[del.connId:clen-1], s.conns[del.connId+1:clen])
			s.conns = s.conns[:clen-1]
		}
	}
	return false
}
