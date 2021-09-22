package service

import (
	"fmt"
	"github.com/tidwall/gjson"
	"golang.org/x/net/websocket"
)

type Client interface {
	Send(tframe string) error
	Receive()
	ReceiveHandle(b []byte)
	RegisterRHandle(f func(gjson.Result))
	SetC(*websocket.Conn)
}

type HandleClient struct {
	C       *websocket.Conn
	Handler []func(mapE gjson.Result)
}

var (
	err error
)

func GetHClient(ws_host string, origin_host string) (Client, error) {
	client := &HandleClient{}

	c, err := websocket.Dial(ws_host, "", origin_host)
	client.SetC(c)
	if err != nil {
		return client, err
	}

	return client, nil
}

// 发送帧数据
func (c HandleClient) SetC(conn *websocket.Conn) {
	c.C = conn
}

// 设置服务器
func (c HandleClient) Send(tframe string) error {
	if _, err := c.C.Write([]byte(tframe)); err != nil {
		return err
	}

	return nil
}

// 接受帧数据
func (c HandleClient) Receive() {
	var msg = make([]byte, 5000)
	var n int
	if n, err = c.C.Read(msg); err != nil {
		fmt.Println("Received Error: ", err.Error())
	}

	c.ReceiveHandle(msg[:n])
}

// 接受消息处理器
func (c HandleClient) ReceiveHandle(b []byte) {
	if len(b) == 0 {
		fmt.Println("receive string empty")
		return
	}

	for _, handle := range c.Handler {
		mapE := gjson.Parse(string(b))
		handle(mapE)
	}
}

// 注册接受消息处理器
func (c HandleClient) RegisterRHandle(f func(gjson.Result)) {
	c.Handler = append(c.Handler, f)
}
