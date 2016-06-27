package main

import (
	"flag"
	slog "log"
	"longChat-Server/common/consts"
	"longChat-Server/common/util/config"
	idService "longChat-Server/idService/service"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) Generate(ctx context.Context, in *idService.GenerateReq) (*idService.GenerateRsp, error) {
	id, _ := idService.Generate(in.Type)
	return &idService.GenerateRsp{Id: id}, nil
}

func main() {
	pconfig := flag.String("config", "../config.cfg", "config file")
	psection := flag.String("section", "dev", "section of config file to apply")
	flag.Parse()
	env := map[string]string{
		"config":  *pconfig,
		"section": *psection,
	}
	config.InitConfigEnv(env)
	err := config.LoadConfigFile()
	if err != nil {
		slog.Fatalf("LoadConfigFile from %s failed. err=%v\n", *pconfig, err)
	}
	idAddress, err := config.GetConfigString(consts.IdServiceAddress)
	if err != nil {
		slog.Fatalln(consts.ErrGetConfigFailed(consts.IdServiceAddress, err))
	}
	idService.InitIdService(false)

	lis, err := net.Listen("tcp", idAddress)
	if err != nil {
		slog.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	idService.RegisterIdServiceServer(s, &server{})
	s.Serve(lis)
}
