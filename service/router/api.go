package router

import (
	"github.com/julienschmidt/httprouter"
	"task_client/controller"
)

func UseApi(router *httprouter.Router) {
	baijiayun := new(controller.BaijiayunController)
	router.POST("/post_baijiayun", baijiayun.Post)
	router.POST("/list_goroutine", baijiayun.GoroutineList)
}
