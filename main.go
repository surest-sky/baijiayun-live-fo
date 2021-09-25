package main

import (
	"task_client/service/router"
	"task_client/service/task"
)

func main() {
	go task.BjServe()
	router.RouterServe()
}
