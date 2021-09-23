package bjy

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
	SetClient(host string, o_host string)
}

type HandleClient struct {
	C       *websocket.Conn
	Handler []func(mapE gjson.Result)
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
	if _, err := c.C.Write([]byte(tframe)); err != nil {
		return err
	}

	return nil
}

// 接受帧数据
func (c *HandleClient) Receive() {
	var msg = make([]byte, 5000)
	var n int
	if n, err = c.C.Read(msg); err != nil {
		fmt.Println("Received Error: ", err.Error())
	}

	c.ReceiveHandle(msg[:n])
}

// 接受消息处理器
func (c *HandleClient) ReceiveHandle(b []byte) {
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
func (c *HandleClient) RegisterRHandle(f func(gjson.Result)) {
	c.Handler = append(c.Handler, f)
}
