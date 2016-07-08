package message

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/common/protoc"
	"github.com/longchat/longChat-Server/storageService/storage"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var cookieName string
var redisPrefix string

type Messenger struct {
	router   protoc.RouterClient
	conn     *grpc.ClientConn
	store    *storage.Storage
	upgrader websocket.Upgrader
}

func (m *Messenger) Close() {
	if m.conn != nil {
		m.conn.Close()
	}
}
func (m *Messenger) Init(store *storage.Storage) {
	routerAddr, err := config.GetConfigString(consts.RouterServiceAddress)
	if err != nil {
		log.Fatalf(consts.ErrGetConfigFailed(consts.RouterServiceAddress, err))
	}
	msgAddr, err := config.GetConfigString(consts.MsgServiceBackendAddress)
	if err != nil {
		log.Fatalf(consts.ErrGetConfigFailed(consts.MsgServiceBackendAddress, err))
	}
	cookieName, err = config.GetConfigString(consts.SessionCookieName)
	if err != nil {
		log.Fatalf(consts.ErrGetConfigFailed(consts.SessionCookieName, err))
	}
	m.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	m.store = store

	m.conn, err = grpc.Dial(routerAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("dial to router(%s) failed!err:=%v", err)
	}
	c := protoc.NewRouterClient(m.conn)
	m.router = c
	rsp, err := m.router.ServerUp(context.Background(), &protoc.ServerUpReq{ServerAddr: msgAddr})
	if err != nil || rsp.StatusCode != 0 {
		log.Fatalf("send ServerUp to router failed!err1:=%v,err2:=%v", err, rsp.Err)
	}
}

func (m *Messenger) MessageIn(req *protoc.MessageReq) {
	fmt.Println("message in:", *req)
}

func (m *Messenger) ServeWs(w http.ResponseWriter, r *http.Request) {
	cookieRaw, err := r.Cookie(cookieName)
	if err != nil {
		fmt.Println("cookie invalid")
	}
	if len(cookieRaw.Value) < 8 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	c1, err := url.QueryUnescape(cookieRaw.Value)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	sId, err := url.QueryUnescape(c1)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rmap, err := m.store.GetSessionValue(sId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	id, isok := rmap["Id"]
	if !isok || id.(int64) <= 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	fmt.Println(id)
	ws, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	conn := &Conn{send: make(chan []byte, 256), ws: ws}
	u, err := m.store.GetUserById(id.(int64))
	if err != nil {
		log.Println(err)
		return
	}
	var gids []int64 = make([]int64, 0, 5)
	for i := range u.JoinedGroups {
		gids = append(gids, u.JoinedGroups[i].Id)
	}
	command <- connAdd{conn: conn, userId: u.Id, groupId: gids}
	go conn.writePump()
	conn.readPump(u.Id, gids)
}
