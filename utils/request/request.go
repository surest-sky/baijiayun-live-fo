package request

import (
	"bytes"
	"encoding/json"
	"github.com/spf13/cast"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"task_client/utils/logger"
)

var (
	FormData = "writer.FormDataContentType()"
	FormWFU  = "x-www-form-urlencoded"
	FormJson = "application/json"
)

type Reuqest struct {
	Ctype   string
	Method  string
	Url     string
	Data    interface{}
	Payload io.Reader
}

var request *Reuqest
var req *http.Request
var err error
var writer *multipart.Writer

func NewRequest(r *Reuqest) string {
	request = r
	client := &http.Client{}
	switch request.Ctype {
	case FormData:
		FormDataReq()
		r.Ctype = writer.FormDataContentType()
		break
	case FormJson:
		data, _ := json.Marshal(r.Data)
		request.Payload = strings.NewReader(string(data))
		break
	}
	logger.Info("requesting", request)
	req, err = http.NewRequest(request.Method, request.Url, request.Payload)
	req.Header.Set("Content-Type", r.Ctype)

	res, err := client.Do(req)
	logger.PanicError(err, "requested", false)
	logger.Info("requested", request)
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	logger.PanicError(err, "requested readed", false)
	logger.Info("requested readed", map[string]interface{}{
		"请求结果": string(body),
		"请求参数": request,
	})

	return string(body)
}

func FormDataReq() {
	payload := &bytes.Buffer{}
	writer = multipart.NewWriter(payload)
	items := request.Data.(map[string]interface{})
	for key, item := range items {
		_ = writer.WriteField(key, cast.ToString(item))
	}
	err = writer.Close()
	logger.PanicError(err, "formData", true)
	request.Payload = payload
}
