package main

import (
	"fmt"
	"os"
	"task_client/app/bjy"
)

func main() {
	args := os.Args
	if len(args) <= 2 {
		fmt.Println("参数异常, 请键入 课程ID 和 课程链接")
		return
	}

	var class_id = args[1]
	var class_url = args[2]
	bus := bjy.Newbusiness()
	bus.Handle(class_url, class_id)
}
