package message

import (
	"fmt"

	"github.com/longchat/longChat-Server/common/protoc"
)

type Session struct {
	userId     int64
	groupIds   []int64
	conns      []*Conn
	command    chan interface{}
	send       chan *protoc.MessageReq
	maxConnNum int
}

type connAdd struct {
	conn *Conn
}

type connDel struct {
	conn *Conn
}

func (s *Session) closeSess() {
	close(s.send)
	close(s.command)
}

func (s *Session) startSession(conn *Conn, uid int64, gids []int64) {
	s.conns = append(s.conns, conn)
	s.userId = uid
	s.groupIds = gids
	s.command = make(chan interface{}, 2)
	s.send = make(chan *protoc.MessageReq, 200)
	s.maxConnNum = 1
	go conn.handleConn(s.command)
	go s.handleSession()
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
				//max conn num
				if len(s.conns) >= 4 {
					continue
				}
				s.conns = append(s.conns, value.conn)
				go value.conn.handleConn(s.command)
			case connDel:
				s.DelConn(value)
			}
		case message, isok := <-s.send:
			if !isok {
				return
			}
			for i := range s.conns {
				s.conns[i].send <- message
			}
		}
	}
}

func (s *Session) DelConn(del connDel) {
	del.conn.closeConn()
	var i int
	for i = range s.conns {
		if s.conns[i].id == del.conn.id {
			break
		}
	}
	clen := len(s.conns)
	if i == clen-1 && s.conns[i].id != del.conn.id {
		panic(fmt.Sprintf("session's conns(%v) doesn't contain conn(%d)", s.conns, del.conn.id))
	} else {
		if clen == 1 {
			command <- sessionDel{session: s, connsCount: s.maxConnNum}
			s.conns = nil
			fmt.Println("del session")
		} else {
			copy(s.conns[i:clen-1], s.conns[i+1:clen])
			s.conns = s.conns[:clen-1]
			fmt.Println(s.conns)

		}
	}
}
