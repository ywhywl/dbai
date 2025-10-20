package logic

import (
	"context"

	"db_ai/service/aitools/rpc/internal/svc"
	"db_ai/service/aitools/rpc/types/aitools"

	"github.com/zeromicro/go-zero/core/logx"
)

type SayHelloLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSayHelloLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SayHelloLogic {
	return &SayHelloLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SayHelloLogic) SayHello(in *aitools.SayHelloReq) (*aitools.SayHelloResp, error) {
	// todo: add your logic here and delete this line

	return &aitools.SayHelloResp{Pong: "pong"}, nil
}
