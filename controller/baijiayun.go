package controller

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"runtime"
	"task_client/service/baijiayun"
	"task_client/utils/logger"
)

type BaijiayunController struct{}

func (t *BaijiayunController) Post(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	queue := baijiayun.New()
	var p map[string]string
	err := json.NewDecoder(r.Body).Decode(&p)

	logger.PanicError(err, "encode", false)
	logger.Info("request", r.PostForm.Encode())

	data := map[string]interface{}{
		"class_id": p["class_id"],
		"url":      p["url"],
	}
	id := queue.Push(data)
	data["id"] = id

	response(w, r, data)
}
func (t *BaijiayunController) GoroutineList(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	goroutineList := runtime.NumGoroutine()
	data := map[string]int{
		"count": goroutineList,
	}
	response(w, r, data)
}
