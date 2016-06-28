package consts

import (
	"fmt"
)

const (
	IdServiceAddress  = "service.id.address"
	IdServiceStartIdx = "service.id.startidx"
	IdServiceStep     = "service.id.step"

	StorageServiceDbName = "service.storage.db.name"
	StorageServiceDbAddr = "service.storage.db.addr"
	StorageServiceDbUser = "service.storage.db.user"
	StorageServiceDbPsw  = "service.storage.db.psw"
)

func ErrGetConfigFailed(configName string, err error) string {
	return fmt.Sprintf("get config(%s) failed!err:=%s", configName, err.Error())
}

func ErrDialRemoteServiceFailed(addr string, err error) string {
	return fmt.Sprintf("dial to server(%s) failed!err:=%s", addr, err.Error())
}

func ErrRPCCallFailed(service string, function string, err error) string {
	return fmt.Sprintf("rpc call(%s,%s) failed!err:=%s", service, function, err.Error())
}
