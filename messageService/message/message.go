package message

import (
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/longchat/longChat-Server/storageService/storage"
)

var (
	IsLeafServer bool
	cookieName   string
	redisPrefix  string
	store        *storage.Storage
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	var userId int64
	cookieRaw, err := r.Cookie(cookieName)
	if err == nil && cookieRaw != nil {
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
		rmap, err := store.GetSessionValue(sId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
			return
		}
		userId, isok := rmap["Id"]
		if !isok {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("invalid session"))
			return
		}
	} else if IsLeafServer {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("no cookie"))
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}

}
