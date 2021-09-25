package main

import "task_client/service/baijiayun"

func main() {
	//go task.BjServe()
	//router.RouterServe()

	bus := baijiayun.Newbusiness()
	bus.Handle(123456)
}
