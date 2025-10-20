package common

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

// Executor 远程命令执行器
type Executor struct {
	host        string
	port        int
	username    string
	password    string
	privateKey  string
	client      *ssh.Client
	config      *ssh.ClientConfig
	connTimeout time.Duration
	cmdTimeout  time.Duration
}

// Config 配置选项
type Config struct {
	Host        string
	Port        int
	Username    string
	Password    string
	PrivateKey  string        // 私钥文件路径
	ConnTimeout time.Duration // 连接超时时间
	CmdTimeout  time.Duration // 命令执行超时时间
}

// NewExecutorWithConfig 使用配置创建新的远程执行器
func NewExecutorWithConfig(cfg Config) *Executor {
	return &Executor{
		host:        cfg.Host,
		port:        cfg.Port,
		username:    cfg.Username,
		password:    cfg.Password,
		privateKey:  cfg.PrivateKey,
		connTimeout: cfg.ConnTimeout,
		cmdTimeout:  cfg.CmdTimeout,
	}
}

// NewExecutor 创建一个新的远程执行器(兼容旧版本)
func NewExecutor(host string, port int, username, password string) *Executor {
	return &Executor{
		host:     host,
		port:     port,
		username: username,
		password: password,
	}
}

// Connect 连接到远程主机
func (e *Executor) Connect() error {
	if e.client != nil {
		// 检查现有连接是否仍然有效
		_, _, err := e.client.SendRequest("keepalive@openssh.com", true, nil)
		if err == nil {
			return nil // 连接仍然有效
		}
		e.client.Close()
	}

	// 设置认证方法
	authMethods := make([]ssh.AuthMethod, 0)

	// 添加密码认证
	if e.password != "" {
		authMethods = append(authMethods, ssh.Password(e.password))
	}

	// 添加公钥认证
	if e.privateKey != "" {
		signer, err := e.parsePrivateKey()
		if err != nil {
			return fmt.Errorf("failed to parse private key: %v", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return errors.New("no authentication method provided")
	}

	// 设置默认超时
	if e.connTimeout == 0 {
		e.connTimeout = 10 * time.Second
	}

	e.config = &ssh.ClientConfig{
		User:            e.username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         e.connTimeout,
	}

	// 设置连接重试逻辑
	var lastErr error
	maxRetries := 3
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", e.host, e.port), e.config)
		if err == nil {
			e.client = client
			return nil
		}

		lastErr = err
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2 // 指数退避
		}
	}

	return fmt.Errorf("failed to connect after %d attempts: %v", maxRetries, lastErr)
}

// parsePrivateKey 解析私钥文件
func (e *Executor) parsePrivateKey() (ssh.Signer, error) {
	keyPath := e.privateKey
	if keyPath == "" {
		return nil, errors.New("private key path is empty")
	}

	// 展开路径中的 ~
	if keyPath[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %v", err)
		}
		keyPath = filepath.Join(home, keyPath[2:])
	}

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %v", err)
	}

	// 尝试解析不带密码的私钥
	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err == nil {
		return signer, nil
	}

	// 如果私钥有密码，提示需要密码(这里简化处理，实际应用中可以从配置获取密码)
	return nil, fmt.Errorf("private key may be encrypted, password support not implemented: %v", err)
}

// IsConnected 检查是否已连接
func (e *Executor) IsConnected() bool {
	if e.client == nil {
		return false
	}

	_, _, err := e.client.SendRequest("keepalive@openssh.com", true, nil)
	return err == nil
}

// Execute 执行远程命令（使用默认超时）
func (e *Executor) Execute(command string) (string, int, error) {
	// 设置默认超时30秒
	defaultTimeout := 30 * time.Second
	if e.cmdTimeout > 0 {
		defaultTimeout = e.cmdTimeout
	}
	return e.ExecuteWithTimeout(command, defaultTimeout)
}

// ExecuteWithTimeout 执行远程命令（带超时控制）
func (e *Executor) ExecuteWithTimeout(command string, timeout time.Duration) (string, int, error) {
	if timeout <= 0 {
		timeout = 30 * time.Second // 确保有默认超时
	}

	if !e.IsConnected() {
		if err := e.Connect(); err != nil {
			return "", -1, fmt.Errorf("reconnect failed: %v", err)
		}
	}

	session, err := e.client.NewSession()
	if err != nil {
		return "", -1, fmt.Errorf("session creation failed: %v", err)
	}

	// 设置输出
	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	// 执行命令
	errChan := make(chan error, 1)
	go func() {
		errChan <- session.Run(command)
	}()

	// 设置超时
	select {
	case err := <-errChan:
		output := stdoutBuf.String()
		if stderrBuf.Len() > 0 {
			output += "\n" + stderrBuf.String()
		}

		if err != nil {
			if exitErr, ok := err.(*ssh.ExitError); ok {
				return output, exitErr.ExitStatus(), err
			}
			return "", -1, err
		}
		return output, 0, nil

	case <-time.After(timeout):
		session.Close()
		output := stdoutBuf.String()
		if stderrBuf.Len() > 0 {
			output += "\n" + stderrBuf.String()
		}
		return output, -1, fmt.Errorf("command timed out after %v", timeout)
	}
}

// Close 关闭SSH连接
func (e *Executor) Close() error {
	if e.client != nil {
		return e.client.Close()
	}
	return nil
}
