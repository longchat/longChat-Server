package message

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/common/protoc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

//若该用户在本地则session不为nil
type userState struct {
	serverAddress string
	gids          []int64
	session       *Session
}

type userAdd struct {
	userId        int64
	gids          []int64
	serverAddress string
}
type userRemove struct {
	userId int64
}

type serverRemove struct {
	serverAddress string
}

type wsAdd struct {
	ws       *websocket.Conn
	userId   int64
	groupIds []int64
}

type sessionDel struct {
	session    *Session
	connsCount int
}

type getGroupMembers struct {
	groupId int64
	connId  int
	session *Session
	userIds []string
}

type chatGroup struct {
	users       map[int64]*Session
	isPublished bool
}

//若该用户在本地则session不为nil
var groups map[int64]chatGroup
var users map[int64]userState
var command chan interface{}

var msgServer map[string]chan *protoc.MessageReq
var msgChan chan *protoc.MessageReq

func controller() {
	connCount := 0
	msgServer = make(map[string]chan *protoc.MessageReq, 8)
	users = make(map[int64]userState, 1000)
	groups = make(map[int64]chatGroup, 100)
	command = make(chan interface{}, 100)
	msgChan = make(chan *protoc.MessageReq, 800)
	for {
		select {
		case item := <-command:
			switch value := item.(type) {
			case wsAdd:
				connCount++
				addWebsocket(value)
			case sessionDel:
				connCount = connCount - value.connsCount
				removeSession(value)
			case userAdd:
				addUser(value)
			case serverRemove:
				delete(msgServer, value.serverAddress)
			case userRemove:
				removeUser(value)
			case getGroupMembers:
				groupUsers, isok := groups[value.groupId]
				if isok {
					ids := make([]string, 0, 100)
					for k, _ := range groupUsers.users {
						ids = append(ids, fmt.Sprintf("%d", k))
					}
					value.userIds = ids
					value.session.send <- &value
				}
			}
		case message := <-msgChan:
			processMessage(message)
		}
	}
}

func serverMessagePump(msgCh chan *protoc.MessageReq, serverAddr string) {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.ERROR.Printf("dial to MsgServer(%s) failed!err:=%s", serverAddr, err)
		command <- serverRemove{serverAddr}
		return
	}
	defer conn.Close()
	client := protoc.NewMessageClient(conn)
	for msg := range msgCh {
		reply, err := client.Message(context.Background(), msg)
		if err != nil || reply.StatusCode != 0 {
			log.ERROR.Printf("send message to MsgServer failed!err:=%s", err)
			command <- serverRemove{serverAddr}
			return
		}
	}
}

func processMessage(msg *protoc.MessageReq) {
	if msg.GroupId > 0 {
		usersess, isok := groups[msg.GroupId]
		if isok {
			for _, v := range usersess.users {
				if v != nil {
					v.send <- msg
				}
			}
		}
	} else {
		userState, isok := users[msg.To]
		if isok {
			if userState.session == nil {
				serverCh, isok := msgServer[userState.serverAddress]
				fmt.Println("userstate:", userState, "serverch:", serverCh)

				if isok {
					serverCh <- msg
				} else {
					delete(users, msg.To)
				}
			} else {
				userState.session.send <- msg
			}
		}
	}
}

func removeSession(s sessionDel) {
	s.session.closeSess()
	for _, data := range s.session.groupIds {
		group, _ := groups[data]
		delete(group.users, s.session.userId)
		if len(group.users) == 0 {
			delete(groups, data)
		} else {
			groups[data] = group
		}
	}
	delete(users, s.session.userId)
}

func removeUser(ru userRemove) {
	state, isok := users[ru.userId]
	if !isok {
		return
	}
	for _, data := range state.gids {
		groupUsers, isok := groups[data]
		if isok {
			delete(groupUsers.users, ru.userId)
		}
	}
	delete(users, ru.userId)
}

func addUser(usa userAdd) {
	ch, isok := msgServer[usa.serverAddress]
	if !isok {
		ch = make(chan *protoc.MessageReq, 256)
		msgServer[usa.serverAddress] = ch
		go serverMessagePump(ch, usa.serverAddress)
	}

	state, isok := users[usa.userId]
	if !isok {
		state.serverAddress = usa.serverAddress
		state.gids = usa.gids
		for _, data := range usa.gids {
			groupUsers, isok := groups[data]
			if !isok {
				groupUsers.users = make(map[int64]*Session, 100)
			} else {
				groupUsers.users[usa.userId] = nil
			}
			groups[data] = groupUsers
		}
	} else {
		if state.session != nil {
			fmt.Printf("conflict userState session:%v serveraddr:%s", state.session, usa.serverAddress)
		}
		state.gids = usa.gids
		state.serverAddress = usa.serverAddress
	}
	users[usa.userId] = state
}

func addWebsocket(s wsAdd) {
	state, isok := users[s.userId]
	if !isok {
		state.session = new(Session)
		state.session.startSession(s.ws, s.userId, s.groupIds)
		users[s.userId] = state
	} else {
		state.session.command <- connAdd{ws: s.ws}
	}
	if !isok {
		for _, data := range s.groupIds {
			group, isok := groups[data]
			if !isok {
				group.users = make(map[int64]*Session, 20)
				reply, err := router.GroupUp(context.Background(), &protoc.GroupUpReq{GroupId: data, ServerAddr: msgAddr})
				if err != nil || reply.StatusCode != 0 {
					fmt.Printf("group up to router failed!err:=%v,err:=%v", err, reply.Err)
				}
			} else if !group.isPublished {
				reply, err := router.GroupUp(context.Background(), &protoc.GroupUpReq{GroupId: data, ServerAddr: msgAddr})
				if err != nil || reply.StatusCode != 0 {
					fmt.Printf("group up to router failed!err:=%v,err:=%v", err, reply.Err)
				}
			}
			group.users[s.userId] = state.session
			groups[data] = group
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
