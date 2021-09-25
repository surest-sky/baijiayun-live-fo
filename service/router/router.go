package router

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func RouterServe() {
	router := httprouter.New()

	UseApi(router)

	fmt.Println("访问: :", "http://127.0.0.1:8085")
	log.Fatal(http.ListenAndServe(":8085", router))
}
