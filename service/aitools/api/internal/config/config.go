package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// Remove the Timeout field from the Config struct below, since rest.RestConf already includes a Timeout field.

type Config struct {
	rest.RestConf
	AitoolsRpc         zrpc.RpcClientConf
	PidFile            string
	BackupCheckCommand string
}
