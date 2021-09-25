package task

import (
	"fmt"
	"task_client/service/baijiayun"
	"time"
)

func BjServe() {
	queue := baijiayun.New()
	var bus *baijiayun.BjyH
	var id int64
	for {
		id = queue.Pop()
		if id > 0 {
			fmt.Println("id: ", id)

			bus = baijiayun.Newbusiness()
			bus.Handle(id)
		}
		time.Sleep(time.Second / 2)
	}
}
