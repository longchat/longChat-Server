package generator

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/common/log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type IdGenerator struct {
	conn       *grpc.ClientConn
	rpcEnabled bool
	client     IdGeneratorClient

	counters [100]int64
	step     int64
}

func (is *IdGenerator) Init(rpcEnabled bool) error {
	is.rpcEnabled = rpcEnabled
	if rpcEnabled {
		address, err := config.GetConfigString(consts.IdServiceAddress)
		if err != nil {
			return errors.New(consts.ErrGetConfigFailed(consts.IdServiceAddress, err))
		}
		is.conn, err = grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			return errors.New(consts.ErrGetConfigFailed(consts.IdServiceAddress, err))
		}
		is.client = NewIdGeneratorClient(is.conn)
	}

	startIdx, err := config.GetConfigInt64(consts.IdServiceStartIdx)
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.IdServiceStartIdx, err))
	}
	for i := range is.counters {
		is.counters[i] = startIdx
	}

	step, err := config.GetConfigInt64(consts.IdServiceStep)
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.IdServiceStep, err))
	}
	is.step = step
	return nil
}

func (is *IdGenerator) Close() {
	if is.conn != nil {
		is.conn.Close()
	}
}

func (is *IdGenerator) Generate(idType GenerateReq_IdType) (int64, error) {
	if is.rpcEnabled {
		reply, err := is.client.Generate(context.Background(), &GenerateReq{idType})
		if err != nil {
			log.ERROR.Printf(consts.ErrRPCCallFailed("idService", "Generate", err))
			return 0, err
		}
		return reply.Id, nil
	} else {
		return is.generate(idType, &(is.counters[idType]), is.step), nil
	}
}

func (is *IdGenerator) generate(idType GenerateReq_IdType, idx *int64, step int64) int64 {
	//共19位,前13位是时间戳，中间4位是计数器，后2位是类型Id
	return (int64(time.Now().UnixNano())/1000000)*1000000 + (atomic.AddInt64(idx, is.step)%10000)*100 + int64(idType)
}
