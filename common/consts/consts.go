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

	ApiServiceAddress = "service.api.address"

	ErrorLogPath  = "log.error.path"
	AccessLogPath = "log.access.path"

	TlsEnable = "security.tls.enable"
)

func ErrGetConfigFailed(configName string, err error) string {
	return fmt.Sprintf("get config(%s) failed!err:=%s\n", configName, err.Error())
}

func ErrDialRemoteServiceFailed(addr string, err error) string {
	return fmt.Sprintf("dial to server(%s) failed!err:=%s\n", addr, err.Error())
}

func ErrRPCCallFailed(service string, function string, err error) string {
	return fmt.Sprintf("rpc call(%s,%s) failed!err:=%s\n", service, function, err.Error())
}
