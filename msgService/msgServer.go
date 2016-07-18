package main

import (
	slog "log"
	"net"
	"net/http"

	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/common/protoc"
	"github.com/longchat/longChat-Server/msgService/message"
	"github.com/longchat/longChat-Server/storageService/storage"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type msgServer struct {
	msg *message.Messenger
}

func (ms *msgServer) Message(ctx context.Context, in *protoc.MessageReq) (*protoc.BaseRsp, error) {
	ms.msg.MessageIn(in)
	return &protoc.BaseRsp{}, nil
}

func (ms *msgServer) UserUp(ctx context.Context, in *protoc.UserUpReq) (*protoc.BaseRsp, error) {
	ms.msg.UserUp(in)
	return &protoc.BaseRsp{}, nil
}

func (ms *msgServer) UserDown(ctx context.Context, in *protoc.UserDownReq) (*protoc.BaseRsp, error) {
	ms.msg.UserDown(in)
	return &protoc.BaseRsp{}, nil
}

func main() {
	config.InitConfig()
	addr, err := config.GetConfigString(consts.MsgServiceAddress)
	if err != nil {
		slog.Fatalln(consts.ErrGetConfigFailed(consts.MsgServiceAddress, err))
	}
	store := storage.Storage{}
	err = store.Init()
	defer store.Close()
	if err != nil {
		slog.Fatalf("init DB failed!err:=%v\n", err)
	}

	m := message.Messenger{}
	go serveGrpc(&m)
	m.Init(&store)

	defer m.Close()
	http.HandleFunc("/websocket", m.ServeWs)
	http.ListenAndServe(addr, nil)
}

func serveGrpc(m *message.Messenger) {
	backaddr, err := config.GetConfigString(consts.MsgServiceBackendAddress)
	if err != nil {
		slog.Fatalln(consts.ErrGetConfigFailed(consts.MsgServiceAddress, err))
	}
	lis, err := net.Listen("tcp", backaddr)
	if err != nil {
		slog.Fatalf("failed to listen on %s!err:=%v", backaddr, err)
	}

	mServer := msgServer{msg: m}
	s := grpc.NewServer()
	protoc.RegisterMessageServer(s, &mServer)
	s.Serve(lis)
}
