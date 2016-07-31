package message

import (
	"fmt"
	slog "log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	messagepb "github.com/longchat/longChat-Server/common/protoc"
	"github.com/longchat/longChat-Server/storageService/storage"
)

type Server struct {
	store     *storage.Storage
	connPool  sync.Pool
	maxConnId uint32
}

var (
	hasParentServer bool
	isLeafServer    bool
	cookieName      string
	redisPrefix     string
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func StartServer(store *storage.Storage, addr string, parentAddr string, isLeaf bool) {
	if parentAddr != "" {
		hasParentServer = true
	}
	isLeafServer = isLeaf
	msgCh = make(chan message, 256)
	if isLeafServer {
		onlineCh = make(chan online, 128)
	} else {
		rmConnCh = make(chan removeConn)
		onlineCh = make(chan online, 32)
	}
	s := Server{
		connPool: sync.Pool{
			New: func() interface{} {
				return &conn{}
			},
		},
		store: store,
	}
	if hasParentServer {
		s.connectParentAndStartHub(parentAddr)
	} else {
		go startHub(nil)
	}
	http.HandleFunc("/websocket", s.serveWebSocket)
	http.ListenAndServe(addr, nil)
}

func (s *Server) connectParentAndStartHub(addr string) {
	rawC, err := net.Dial("tcp4", addr)
	if err != nil {
		slog.Fatalln("tcp dial to parent server failed", err)
	}
	urlA := url.URL{
		Scheme: "ws",
		Host:   addr,
		Path:   "websocket",
	}
	header := http.Header(make(map[string][]string))
	ws, _, err := websocket.NewClient(rawC, &urlA, header, 4096, 4096)
	if err != nil {
		slog.Fatalln("can't create websocket connect to parent server!", err)
	}
	wsConn := s.getWsConn(ws)
	defer s.releaseWsConn(wsConn)
	go startHub(wsConn)
	wsConn.readPump(0)
}

func (s *Server) serveWebSocket(w http.ResponseWriter, r *http.Request) {
	if isLeafServer {
		s.serveLeaf(w, r)
	} else {
		s.serveNode(w, r)
	}
}

func (s *Server) serveNode(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}
	wsConn := s.getWsConn(ws)
	defer s.releaseWsConn(wsConn)
	wsConn.readPump(0)
	rmConnCh <- removeConn{wsConn}
}

func (s *Server) serveLeaf(w http.ResponseWriter, r *http.Request) {
	cookieRaw, err := r.Cookie(cookieName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("no cookie"))
		return
	}
	c1, err := url.QueryUnescape(cookieRaw.Value)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid session cookie"))
		return
	}
	sId, err := url.QueryUnescape(c1)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid session cookie"))
		return
	}
	rmap, err := s.store.GetSessionValue(sId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}
	uid, isok := rmap["Id"]
	if !isok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid session"))
		return
	}
	userId := uid.(int64)
	user, err := s.store.GetUserById(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}
	olItems := make([]*messagepb.OnlineReq_Item, 1, 4)
	olItems[0] = &messagepb.OnlineReq_Item{
		Id:       user.Id,
		IsOnline: true,
		IsGroup:  false,
	}
	for i := range user.JoinedGroups {
		olItems = append(olItems, &messagepb.OnlineReq_Item{
			Id:       user.JoinedGroups[i].Id,
			IsOnline: true,
			IsGroup:  true,
		})
	}
	req := messagepb.OnlineReq{olItems}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("upgrader  failed!", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}
	wsConn := s.getWsConn(ws)
	defer s.releaseWsConn(wsConn)
	onlineCh <- online{wsConn, req}
	wsConn.readPump(userId)
	for i := range req.Items {
		req.Items[i].IsOnline = false
	}
	onlineCh <- online{wsConn, req}
}

func (s *Server) getWsConn(ws *websocket.Conn) *conn {
	wsConn := s.connPool.Get().(*conn)
	wsConn.ws = ws
	wsConn.Id = atomic.AddUint32(&s.maxConnId, 1)
	wsConn.wLock.Lock()
	wsConn.state = ConnStateWorking
	wsConn.wLock.Unlock()
	return wsConn
}

func (s *Server) releaseWsConn(wsConn *conn) {
	wsConn.wLock.Lock()
	wsConn.state = ConnStateIdle
	wsConn.wLock.Unlock()
	wsConn.ws.Close()
	s.connPool.Put(wsConn)
}
