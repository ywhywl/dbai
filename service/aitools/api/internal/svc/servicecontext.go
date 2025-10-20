package svc

import (
	"db_ai/service/aitools/api/internal/config"
	"db_ai/service/aitools/rpc/aitoolsrpc"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	AitoolsRpc aitoolsrpc.AiToolsRpc
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		AitoolsRpc: aitoolsrpc.NewAiToolsRpc(zrpc.MustNewClient(c.AitoolsRpc)),
	}
}
