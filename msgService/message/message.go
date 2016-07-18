package message

import (
	"fmt"
	slog "log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/common/protoc"
	"github.com/longchat/longChat-Server/storageService/storage"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var msgAddr string
var cookieName string
var redisPrefix string

var router protoc.RouterClient
var conn *grpc.ClientConn
var store *storage.Storage

type Messenger struct {
	upgrader websocket.Upgrader
}

func (m *Messenger) Close() {
	if conn != nil {
		conn.Close()
	}
}
func (m *Messenger) Init(db *storage.Storage) {
	routerAddr, err := config.GetConfigString(consts.RouterServiceAddress)
	if err != nil {
		slog.Fatalf(consts.ErrGetConfigFailed(consts.RouterServiceAddress, err))
	}
	msgAddr, err = config.GetConfigString(consts.MsgServiceBackendAddress)
	if err != nil {
		slog.Fatalf(consts.ErrGetConfigFailed(consts.MsgServiceBackendAddress, err))
	}
	cookieName, err = config.GetConfigString(consts.SessionCookieName)
	if err != nil {
		slog.Fatalf(consts.ErrGetConfigFailed(consts.SessionCookieName, err))
	}
	m.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	store = db

	conn, err = grpc.Dial(routerAddr, grpc.WithInsecure())
	if err != nil {
		slog.Fatalf("dial to router(%s) failed!err:=%v", err)
	}
	c := protoc.NewRouterClient(conn)
	router = c
	rsp, err := router.ServerUp(context.Background(), &protoc.ServerUpReq{ServerAddr: msgAddr})
	if err != nil || rsp.StatusCode != 0 {
		slog.Fatalf("send ServerUp to router failed!err1:=%v,err2:=%v", err, rsp.Err)
	}
	go controller()
}

func (m *Messenger) UserUp(req *protoc.UserUpReq) {
	u, err := store.GetUserById(req.UserId)
	if err != nil {
		slog.Println(err)
		return
	}
	var gids []int64 = make([]int64, 0, 5)
	for i := range u.JoinedGroups {
		gids = append(gids, u.JoinedGroups[i].Id)
	}
	command <- userAdd{serverAddress: req.ServerAddr, userId: req.UserId, gids: gids}
}

func (m *Messenger) UserDown(req *protoc.UserDownReq) {
	command <- userRemove{userId: req.UserId}
}

func (m *Messenger) MessageIn(req *protoc.MessageReq) {
	msgChan <- req
}

func (m *Messenger) ServeWs(w http.ResponseWriter, r *http.Request) {
	cookieRaw, err := r.Cookie(cookieName)
	if err != nil {
		fmt.Println("cookie invalid")
	}
	if cookieRaw == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
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

	rmap, err := store.GetSessionValue(sId)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	id, isok := rmap["Id"]
	if !isok || id.(int64) <= 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	ws, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Println(err)
		return
	}
	u, err := store.GetUserById(id.(int64))
	if err != nil {
		slog.Println(err)
		return
	}
	var gids []int64 = make([]int64, 0, 5)
	for i := range u.JoinedGroups {
		gids = append(gids, u.JoinedGroups[i].Id)
	}
	command <- wsAdd{ws: ws, userId: u.Id, groupIds: gids}
	reply, err := router.UserUp(context.Background(), &protoc.UserUpReq{UserId: u.Id, ServerAddr: msgAddr})
	if err != nil || reply.StatusCode != 0 {
		log.ERROR.Printf("userup(%d) to router failed!\n", u.Id)
		return
	}
}
