package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	// AiToolsRpc    zrpc.RpcClientConf
	RemoteCommand struct {
		TestUsername   string
		TestPassword   string
		TestPrivateKey string
		ConnTimeout    int
		CmdTimeout     int
	}
}
