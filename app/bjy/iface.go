package bjy

import (
	"github.com/tidwall/gjson"
)

// Websocket 客户端接入实现
type Client interface {
	Send(tframe string) error
	Receive()
	ReceiveHandle(b []byte)
	RegisterRHandle(f func(gjson.Result))
	SetClient(host string, o_host string)
	Stop()
}

type Queue interface {
	Serve()
	Start()
	Push(v interface{}) int64
	Remove(id int64)
	Pop() int64
	Get(int64) string
}
