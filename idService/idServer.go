package main

import (
	slog "log"
	"net"

	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/idService/generator"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// server is used to implement helloworld.GreeterServer.
type generatorServer struct {
	idGenerator *generator.IdGenerator
}

// SayHello implements helloworld.GreeterServer
func (s *generatorServer) Generate(ctx context.Context, in *generator.GenerateReq) (*generator.GenerateRsp, error) {
	id, _ := s.idGenerator.Generate(in.Type)
	return &generator.GenerateRsp{Id: id}, nil
}

func main() {
	config.InitConfig()

	server := generatorServer{}
	server.idGenerator = &generator.IdGenerator{}
	err := server.idGenerator.Init(false)
	defer server.idGenerator.Close()
	if err != nil {
		slog.Fatalf("init idGenerator failed!err:=%v", err)
	}

	idAddress, err := config.GetConfigString(consts.IdServiceAddress)
	if err != nil {
		slog.Fatalln(consts.ErrGetConfigFailed(consts.IdServiceAddress, err))
	}
	lis, err := net.Listen("tcp", idAddress)
	if err != nil {
		slog.Fatalf("failed to listen on %s!err:=%v", idAddress, err)
	}
	s := grpc.NewServer()
	generator.RegisterIdGeneratorServer(s, &server)
	s.Serve(lis)
}
