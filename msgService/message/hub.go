package message

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/longchat/longChat-Server/common/protoc"
	"golang.org/x/net/context"
)

var groupUser map[int64]map[int64]*Session
var userSession map[int64]*Session
var command chan interface{}

var msgChan chan *protoc.MessageReq

type wsAdd struct {
	ws       *websocket.Conn
	userId   int64
	groupIds []int64
}

type sessionDel struct {
	session    *Session
	connsCount int
}

func controller() {
	connId := 0
	connCount := 0
	userSession = make(map[int64]*Session, 500)
	groupUser = make(map[int64]map[int64]*Session, 100)
	command = make(chan interface{})
	msgChan = make(chan *protoc.MessageReq, 800)
	for {
		select {
		case item := <-command:
			switch value := item.(type) {
			case wsAdd:
				connId++
				connCount++
				addWebsocket(value, connId)
			case sessionDel:
				connCount = connCount - value.connsCount
				removeSession(value)
			}
		case message := <-msgChan:
			processMessage(message)
		}
	}
}

func processMessage(msg *protoc.MessageReq) {
	if msg.GroupId > 0 {
		usersess, isok := groupUser[msg.GroupId]
		if isok {
			for _, v := range usersess {
				v.send <- msg
			}
		}
	}
}

func removeSession(s sessionDel) {
	s.session.closeSess()
	for _, data := range s.session.groupIds {
		group, _ := groupUser[data]
		delete(group, s.session.userId)
		if len(group) == 0 {
			delete(groupUser, data)
		} else {
			groupUser[data] = group
		}
	}
	delete(userSession, s.session.userId)
	fmt.Println("removed:", groupUser)

}

func addWebsocket(s wsAdd, id int) {
	session, isok := userSession[s.userId]
	if !isok {
		session = new(Session)
		conn := Conn{id: id, session: session, ws: s.ws, send: make(chan *protoc.MessageReq, 128)}
		session.startSession(&conn, s.userId, s.groupIds)
		userSession[s.userId] = session
	} else {
		conn := Conn{id: id, session: session, ws: s.ws, send: make(chan *protoc.MessageReq, 128)}
		session.command <- connAdd{conn: &conn}
	}
	if !isok {
		for _, data := range s.groupIds {
			group, isok := groupUser[data]
			if !isok {
				group = make(map[int64]*Session, 20)
				reply, err := router.GroupUp(context.Background(), &protoc.GroupUpReq{GroupId: data, ServerAddr: msgAddr})
				if err != nil || reply.StatusCode != 0 {
					fmt.Printf("group up to router failed!err:=%v,err:=%v", err, reply.Err)
				}
			}
			group[s.userId] = session
			groupUser[data] = group
		}
	}
}

var storeChan chan *protoc.MessageReq
var readChan chan userRead

type userRead struct {
	userId    int64
	groupRead map[int64]int64
}

func msgPersist() {
	userReadMap := make(map[int64]map[int64]int64, 200)
	for {
		select {
		case read := <-readChan:
			markRead(userReadMap, read)
		}
	}
}

func markRead(readMap map[int64]map[int64]int64, read userRead) {
	/*userGroups, isok := readMap[read.userId]
	if isok {
		userGroups = read.groupRead
	} else {
		for k, v := range read.groupRead {
		}
	}*/
}
