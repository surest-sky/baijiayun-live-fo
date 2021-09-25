package baijiayun

import (
	"context"
	"fmt"
	"github.com/tidwall/gjson"
	"golang.org/x/net/websocket"
	"task_client/utils/logger"
)

type HandleClient struct {
	C       *websocket.Conn
	Handler []func(context.Context, gjson.Result)
}

var (
	err error
)

func GetServer() Client {
	return &HandleClient{}
}

// 发送帧数据
func (c *HandleClient) SetClient(host string, o_host string) {
	client, err := websocket.Dial(host, "", o_host)
	if err != nil {
		fmt.Println("Error :", err.Error())
	}

	c.C = client
}

// 设置服务器
func (c *HandleClient) Send(tframe string) error {
	logger.Info("Send: ", tframe)
	if _, err := c.C.Write([]byte(tframe)); err != nil {
		return err
	}

	return nil
}

// 接受帧数据
func (c *HandleClient) Receive(ctx context.Context) {
	var msg = make([]byte, 7000)
	var n int
	for {
		if n, err = c.C.Read(msg); err != nil {
			logger.Error("client closed", err.Error())
			return
		}

		c.ReceiveHandle(ctx, msg[:n])
	}
}

// 接受消息处理器
func (c *HandleClient) ReceiveHandle(ctx context.Context, b []byte) {
	if len(b) == 0 {
		fmt.Println("receive string empty")
		return
	}

	for _, handle := range c.Handler {
		res := gjson.Parse(string(b))
		handle(ctx, res)
	}
}

// 注册接受消息处理器
func (c *HandleClient) RegisterRHandle(f func(context.Context, gjson.Result)) {
	c.Handler = append(c.Handler, f)
}

// 结束
func (c *HandleClient) Stop() {
	if c.C.IsClientConn() {
		_ = c.C.Close()
	}
}
