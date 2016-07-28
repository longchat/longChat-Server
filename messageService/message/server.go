package message

import (
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
	IsLeafServer bool
	cookieName   string
	redisPrefix  string
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func startServer() {
	msgCh = make(chan *messagepb.MessageReq, 256)
	if IsLeafServer {
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
	}

}

func (s *Server) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	if IsLeafServer {
		s.serveLeaf(w, r)
	} else {
		s.serveNode(w, r)
	}
}

func (s *Server) serveNode(w http.ResponseWriter, r *http.Request) {
	wsConn, err := s.getWsConn(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}
	wsConn.readPump()
	rmConnCh <- removeConn{wsConn}
	s.releaseWsConn(wsConn)
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
	olItems := make([]*messagepb.OnlineReq_Item, 0, 4)
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
	req := &messagepb.OnlineReq{olItems}
	wsConn, err := s.getWsConn(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}
	onlineCh <- online{wsConn, req}
	wsConn.readPump()
	for i := range req.Items {
		req.Items[i].IsOnline = false
	}
	onlineCh <- online{wsConn, req}
	s.releaseWsConn(wsConn)
}

func (s *Server) getWsConn(w http.ResponseWriter, r *http.Request) (*conn, error) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	wsConn := s.connPool.Get().(*conn)
	wsConn.ws = ws
	wsConn.Id = atomic.AddUint32(&s.maxConnId, 1)
	wsConn.wLock.Lock()
	wsConn.state = ConnStateWorking
	wsConn.wLock.Unlock()
	return wsConn, nil
}

func (s *Server) releaseWsConn(wsConn *conn) {
	wsConn.wLock.Lock()
	wsConn.state = ConnStateIdle
	wsConn.wLock.Unlock()
	wsConn.ws.Close()
	s.connPool.Put(wsConn)
}
