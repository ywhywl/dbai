package aitools

import (
	"context"
	"strconv"
	"strings"

	"db_ai/service/aitools/api/internal/svc"
	"db_ai/service/aitools/api/internal/types"

	"db_ai/service/aitools/rpc/aitoolsrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetbackuplogbyhostsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetbackuplogbyhostsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetbackuplogbyhostsLogic {
	return &GetbackuplogbyhostsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetbackuplogbyhostsLogic) Getbackuplogbyhosts(req *types.RemoteCommandReq) (resp *types.RemoteCommandResp, err error) {
	// Initialize response object
	resp = &types.RemoteCommandResp{}

	l.Logger.Infof("开始备份日志检查，主机字符串: %s", req.HostIps)

	// Split comma-separated host IPs
	hostList := strings.Split(req.HostIps, ",")
	l.Logger.Infof("解析后的主机列表: %v", hostList)

	for _, hostPort := range hostList {
		// Trim whitespace from each host
		hostPort = strings.TrimSpace(hostPort)
		if hostPort == "" {
			continue // Skip empty entries
		}

		l.Logger.Infof("处理主机: %s", hostPort)

		var host, port string
		var portInt int

		// Parse host and port correctly, add default port 22 if not specified
		parts := strings.Split(hostPort, ":")
		l.Logger.Debugf("字符串分割结果: %v, 长度: %d", parts, len(parts))

		if len(parts) == 1 {
			// No port specified, use default port 22
			host = parts[0]
			port = "22"
			portInt = 22
			l.Logger.Infof("主机 %s 未指定端口，使用默认端口 22", host)
		} else if len(parts) == 2 {
			// Port specified
			host = parts[0]
			port = parts[1]
			var err error
			portInt, err = strconv.Atoi(port)
			if err != nil {
				l.Logger.Errorf("端口格式无效: %s，错误: %v", port, err)
				continue // Skip invalid port
			}
			l.Logger.Infof("主机 %s 指定端口: %d", host, portInt)
		} else {
			l.Logger.Errorf("主机格式无效: %s，跳过处理", hostPort)
			continue // Skip invalid format
		}

		l.Logger.Infof("执行远程命令 - 主机: %s:%d, 命令: %s", host, portInt, l.svcCtx.Config.BackupCheckCommand)

		AitoolsResp, err := l.svcCtx.AitoolsRpc.RemoteCommand(l.ctx, &aitoolsrpc.RemoteCommandReq{
			Host:     host,
			Port:     int32(portInt),
			Commands: l.svcCtx.Config.BackupCheckCommand,
		})
		if err != nil {
			l.Logger.Errorf("远程命令执行失败 - 主机: %s:%d, 错误: %v", host, portInt, err)
			return nil, err
		}

		l.Logger.Infof("远程命令执行成功 - 主机: %s:%d, 输出长度: %d", host, portInt, len(AitoolsResp.Output))
		l.Logger.Debugf("命令输出内容: %s", AitoolsResp.Output)

		resp.Successful = true
		resp.Data = types.RemoteCommandRespData{
			HostIp: host,
			Output: AitoolsResp.Output,
		}
	}

	l.Logger.Info("备份日志检查完成")
	return
}
