package message

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/longchat/longChat-Server/storageService/storage"

	"github.com/gorilla/websocket"
	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/common/protoc"
	"google.golang.org/grpc"
)

var cookieName string
var redisPrefix string

type Messenger struct {
	router   *protoc.RouterClient
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
	m.conn, err = grpc.Dial(routerAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("dial to router(%s) failed!err:=%v", err)
	}
	c := protoc.NewRouterClient(m.conn)
	m.router = &c

	cookieName, err = config.GetConfigString(consts.SessionCookieName)
	if err != nil {
		log.Fatalf(consts.ErrGetConfigFailed(consts.SessionCookieName, err))
	}
	fmt.Println(cookieName, redisPrefix)
	m.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	m.store = store
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
	fmt.Println(sId)

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
	go conn.writePump()
	conn.readPump()
}
