package main

import (
	slog "log"
	"net"

	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/common/protoc"
	"github.com/longchat/longChat-Server/routerService/router"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type routerServer struct {
	router *router.Router
}

func (r *routerServer) Message(ctx context.Context, in *protoc.MessageReq) (*protoc.BaseRsp, error) {
	r.router.MessageIn(in)
	return &protoc.BaseRsp{}, nil
}

func (r *routerServer) GroupUp(ctx context.Context, in *protoc.GroupUpReq) (*protoc.BaseRsp, error) {
	r.router.GroupUp(in)
	return &protoc.BaseRsp{}, nil
}

func (r *routerServer) ServerUp(ctx context.Context, in *protoc.ServerUpReq) (*protoc.BaseRsp, error) {
	r.router.ServerUp(in)
	return &protoc.BaseRsp{}, nil
}

func main() {
	config.InitConfig()
	routerSer := routerServer{router: &router.Router{}}
	routerSer.router.Init()

	addr, err := config.GetConfigString(consts.RouterServiceAddress)
	if err != nil {
		slog.Fatalln(consts.ErrGetConfigFailed(consts.RouterServiceAddress, err))
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Fatalf("failed to listen on %s!err:=%v", addr, err)
	}
	s := grpc.NewServer()
	protoc.RegisterRouterServer(s, &routerSer)
	s.Serve(lis)
}
