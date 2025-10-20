package logic

import (
	"context"
	"time"

	"db_ai/common"
	"db_ai/service/aitools/rpc/internal/svc"
	"db_ai/service/aitools/rpc/types/aitools"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoteCommandLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoteCommandLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoteCommandLogic {
	return &RemoteCommandLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RemoteCommandLogic) RemoteCommand(in *aitools.RemoteCommandReq) (*aitools.RemoteCommandResp, error) {
	l.Logger.Infof("接收到远程命令请求 - 主机: %s:%d, 命令: %s", in.Host, in.Port, in.Commands)

	executor := common.NewExecutorWithConfig(common.Config{
		Host:        in.Host,
		Port:        int(in.Port),
		Username:    l.svcCtx.Config.RemoteCommand.TestUsername,
		Password:    l.svcCtx.Config.RemoteCommand.TestPassword,
		PrivateKey:  l.svcCtx.Config.RemoteCommand.TestPrivateKey,
		ConnTimeout: time.Duration(l.svcCtx.Config.RemoteCommand.ConnTimeout) * time.Second,
		CmdTimeout:  time.Duration(l.svcCtx.Config.RemoteCommand.CmdTimeout) * time.Second,
	})
	defer executor.Close()

	l.Logger.Infof("开始执行SSH命令 - 主机: %s:%d", in.Host, in.Port)
	startTime := time.Now()

	output, status, err := executor.Execute(in.Commands)

	duration := time.Since(startTime)

	if err != nil {
		l.Logger.Errorf("SSH命令执行失败 - 主机: %s:%d, 耗时: %v, 状态码: %d, 错误: %v",
			in.Host, in.Port, duration, status, err)
		return &aitools.RemoteCommandResp{
			Output: output,
			Status: int32(status),
		}, err
	}

	l.Logger.Infof("SSH命令执行成功 - 主机: %s:%d, 耗时: %v, 状态码: %d, 输出长度: %d",
		in.Host, in.Port, duration, status, len(output))
	l.Logger.Debugf("命令输出内容: %s", output)

	return &aitools.RemoteCommandResp{
		Output: output,
		Status: int32(status),
	}, err
}
