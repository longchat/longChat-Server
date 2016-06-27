package service

import (
	slog "log"
	"longChat-Server/common/consts"
	"longChat-Server/common/util/config"
	"longChat-Server/common/util/log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var rpcEnabled = true
var client IdServiceClient
var conn *grpc.ClientConn

func InitIdService(isRPCEnabled bool) {
	rpcEnabled = isRPCEnabled
	initCounters()
	if rpcEnabled {
		address, err := config.GetConfigString(consts.IdServiceAddress)
		if err != nil {
			slog.Fatalln(consts.ErrGetConfigFailed(consts.IdServiceAddress, err))
		}
		conn, err = grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			slog.Fatalln(consts.ErrDialRemoteServiceFailed(address, err))
		}
		client = NewIdServiceClient(conn)
	}
}

func CloseIdService() {
	if rpcEnabled {
		conn.Close()
	}
}

func Generate(idType GenerateReq_IdType) (int64, error) {
	if rpcEnabled {
		reply, err := client.Generate(context.Background(), &GenerateReq{idType})
		if err != nil {
			log.ERROR.Printf(consts.ErrRPCCallFailed("idService", "Generate", err))
			return 0, err
		}
		println(reply.Id)
		return reply.Id, nil
	} else {
		return generate(idType), nil
	}
}
