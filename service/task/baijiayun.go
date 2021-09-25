package task

import (
	"task_client/service/baijiayun"
	"time"
)

func BjServe() {
	queue := baijiayun.New()
	var bus *baijiayun.BjyH
	var id int64

	// 无线循环
	for {
		id = queue.Pop()
		if id > 0 {
			bus = baijiayun.Newbusiness()
			bus.Handle(id)
		}
		time.Sleep(time.Second / 2)
	}
}
