package service

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type BjyH struct {
	Token      string
	Classid    string
	RoomUrl    string
	ReqClient  Client
	AuthClient Client
	Params     struct {
		SdkAppid   string `json:"sdkAppid"`
		Identifier string `json:"identifier"`
		UserSig    string `json:"userSig"`
	}
}

// 处理token
func (t BjyH) ProcessToken(result gjson.Result) {
	token := result.Get("webrtc_info.token").String()
	if len(token) > 0 {
		t.Params.SdkAppid = result.Get("webrtc_info.appId").String()
		t.Params.Identifier = result.Get("id").String()
		t.Params.UserSig = token
		fmt.Println("Token Successed ! \n Token: ", token)
	}

	fmt.Println("----- 我收到消息了 ------")
}

// 处理用户列表
func (t BjyH) ProcessUserList(result gjson.Result) {
	fmt.Println("----- 我收到消息了 ------")
}

// 获取待处理的函数
func (t BjyH) getProcessF() []func(gjson.Result) {
	return []func(gjson.Result){
		t.ProcessToken,
		t.ProcessUserList,
	}
}

// 设置req 客户端
// example: "https://e83301793.at.baijiayun.com"
func (t BjyH) SetReqClient() {
	var (
		ws_host     = "wss://pro-signal.baijiayun.com/"
		origin_host = t.RoomUrl
	)

	client, err := GetHClient(ws_host, origin_host)
	if err != nil {
		t.logger(err, "Req Connect : ")
		return
	}

	t.ReqClient = client
	var funcs = t.getProcessF()
	for _, f := range funcs {
		t.ReqClient.RegisterRHandle(f)
	}
}

// 设置 Token
func (t BjyH) SetToken() {
	str := map[string]interface{}{
		"message_type":      "server_info_req",
		"class_id":          t.Classid,
		"class_type":        4,
		"webrtc_type":       3,
		"update_token":      0,
		"link_capability":   0,
		"user_type":         2,
		"end_type":          0,
		"enrolled_students": 0,
		"free_of_charge":    false,
		"special_customer":  "",
		"udp_foreign_proxy": 0,
		"tcp_foreign_proxy": 0,
		"ms_config": map[string]interface{}{
			"live_stream_cdn_list": []string{"ws"},
			"assign_lan_up":        0,
		},
		"user": map[string]interface{}{
			"number":              "373653",
			"group":               0,
			"type":                2,
			"name":                "Peng",
			"actual_name":         "Peng",
			"avatar":              "https://img.baijiayun.com/0bjcloud/live/avatar/v2/helperv4_1.png",
			"status":              0,
			"end_type":            0,
			"is_backdoor":         0,
			"is_record":           0,
			"webrtc_support":      1,
			"is_audition":         false,
			"audition_duration":   0,
			"replace_user_number": "",
			"ext_info":            "",
		},
	}
	s, _ := json.Marshal(str)
	r := string(s)

	if err := t.ReqClient.Send(r); err != nil {
		t.logger(err)
		return
	}
}

// 从 Html 中获取 data json
func (t BjyH) setData(url string) {
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		t.logger(err)
		return
	}

	req.Header.Add("cache-control", "max-age=0")
	req.Header.Add("sec-ch-ua", "\"Microsoft Edge\";v=\"93\", \" Not;A Brand\";v=\"99\", \"Chromium\";v=\"93\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"macOS\"")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.82 Safari/537.36 Edg/93.0.961.52")
	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	res, err := client.Do(req)
	if err != nil {
		t.logger(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.logger(err)
		return
	}

	r := string(body)

	// 解析出 Data
	reg, err := regexp.Compile("var data = {(.*)?};")
	if reg == nil {
		t.logger(err)
		return
	}
	result := reg.FindAllStringSubmatch(r, -1)
	if len(result) == 0 {
		return
	}

	data := result[0][0]
	fmt.Println("data: ", data)
	data = strings.ReplaceAll(data, "var data = ", "")
	data = strings.ReplaceAll(data, ";", "")

	t.Classid = gjson.Parse(data).Get("class.id").String()
}

// 解析链接url
func (t BjyH) ParseP(url string) {
	var urlReg = `(https:\/\/.*).com/`
	var roomId = `(\d+)&`

	reg, _ := regexp.Compile(urlReg)
	t.RoomUrl = reg.FindString(url)

	reg, _ = regexp.Compile(roomId)
	r := reg.FindString(url)
	t.Classid = strings.ReplaceAll("&", "", r)
}

func (t BjyH) logger(e error, p ...string) {
	if e != nil {
		return
	}
	var prefix string
	if len(p) == 0 {
		prefix = "Error :"
	} else {
		prefix = p[0]
	}

	fmt.Println(prefix, e.Error())
}
