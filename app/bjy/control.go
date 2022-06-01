package bjy

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	netUrl "net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// 课程所需必要素
type ClassRx struct {
	ClassId  string
	ClassUrl string
	RoomId   string
}

// 登录 ws 所需参数
type LoginParams struct {
	Token     string
	StartTime string
	EndTime   string
	UserID    string
	Ts        string
	User      User `json:"user"`
	Number    string
	WsHost    string
}

//
type AuthParam struct {
	SdkAppid   string
	Identifier string
	UserSig    string
}

type Server struct {
	Ip        string `json:"ip"`
	KcpServer string `json:"kcp_server"`
	Port      int64  `json:"port"`
	Url       string `json:"url"`
}

// 房间的情况
type Room struct {
	Count   int
	RoomUrl string
	RoomId  int64
}

const (
	TimeDuration  = time.Second * 2
	WsConnectHost = "wss://pro-video-ms.baijiayun.com/"
)

var (
	HtmlHeader = map[string]string{
		"cache-control":             "max-age=0",
		"sec-ch-ua-mobile":          "\"Microsoft Edge\";v=\"93\", \" Not;A Brand\";v=\"99\", \"Chromium\";v=\"93\"",
		"sec-ch-ua":                 "?0",
		"sec-ch-ua-platform":        "macOS",
		"upgrade-insecure-requests": "1",
		"user-agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.82 Safari/537.36 Edg/93.0.961.52",
		"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
	}
	user = User{}
	wg   = sync.WaitGroup{}
)

// 实例
type BjyServer struct {
	ClassRx     ClassRx
	LoginParams LoginParams
	AuthParam   AuthParam
	Room        Room
	ReqClient   Client
	AuthClient  Client
	LoginClient Client
}

func Newbusiness() *BjyServer {
	return &BjyServer{}
}

func (s *BjyServer) Handle(url string, id string) {
	wg.Add(1)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	s.ClassRx.ClassUrl = url
	s.ClassRx.ClassId = id
	s.ParseProcessor()
	s.ParseHtmlProcessor()
	s.SetReqClient()
	s.MessageToken()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		fmt.Println(-1)
	}
}

// URL 参数解析
func (s *BjyServer) ParseProcessor() {
	up, err := netUrl.Parse(s.ClassRx.ClassUrl)
	if err != nil {
		Echo("解析链接出错: " + err.Error())
		return
	}

	m, err := netUrl.ParseQuery(up.RawQuery)
	if err != nil {
		Echo("解析Query出错: " + err.Error())
		return
	}

	s.Room.RoomUrl = fmt.Sprintf("https://%s", up.Host)
	s.Room.RoomId = cast.ToInt64(m.Get("room_id"))

	s.ClassRx.RoomId = m.Get("room_id")
}

// 解析对应网站的 HTML 元素
func (s *BjyServer) ParseHtmlProcessor() {
	const method = "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, s.ClassRx.ClassUrl, nil)
	if err != nil {
		Echo("获取请求实例失败: " + err.Error())
		return
	}

	for key, value := range HtmlHeader {
		req.Header.Add(key, value)
	}

	res, err := client.Do(req)
	if err != nil {
		Echo("请求数据失败: " + err.Error())
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Echo("读取数据资源失败: " + err.Error())
		return
	}

	r := string(body)
	reg, err := regexp.Compile("var data = {(.*)?};")
	if err != nil {
		Echo("读取数据资源失败: " + err.Error())
		return
	}

	result := reg.FindAllStringSubmatch(r, -1)
	if len(result) == 0 {
		Echo("读取数据资源为空")
		return
	}

	data := result[0][0]
	data = strings.ReplaceAll(data, "var data = ", "")
	data = strings.ReplaceAll(data, ";", "")
	rs := gjson.Parse(data)
	params := LoginParams{
		Token:     rs.Get("token").String(),
		StartTime: rs.Get("startTimeTs").String(),
		EndTime:   rs.Get("endTimeTs").String(),
		UserID:    "",
		Ts:        cast.ToString(time.Now().Unix()) + "000",
	}
	s.LoginParams = params
	s.Room.Count = 0
	Echo("解析Html完成")
}

// 设置 ws 请求客户端
func (s *BjyServer) SetReqClient() {
	var client = GetWebSocketClient()
	client.SetClient(WsConnectHost, s.Room.RoomUrl)

	// 给客户端注册一些函数
	s.ReqClient = client
	s.ReqClient.RegisterRHandle(s.EventAuth)

	// 开始处理
	go func() {
		s.ReqClient.Receive()
	}()
}

// 设置登录后的客户端: 需要使用 websocket 进行登录，登录后再进行登录另外一个websocket
// 设置 ws 请求客户端
func (s *BjyServer) SetLoginClient() {
	var client = GetWebSocketClient()
	client.SetClient(s.LoginParams.WsHost, s.Room.RoomUrl)

	// 给客户端注册一些函数
	s.LoginClient = client
	//var funcs = s.getEventFs()
	//for _, f := range funcs {
	//	s.LoginClient.RegisterRHandle(f)
	//}
	s.LoginClient.RegisterRHandle(s.EventUser)

	// 开始处理
	go func() {
		s.LoginClient.Receive()
	}()
	Echo("登录端成功 -- successed !")
}

// 用户获取层面的简体
func (s *BjyServer) EventUser(result gjson.Result) {
	messageType := result.Get("message_type").String()
	processing := true
	if messageType == "user_active_res" {
		userList := result.Get("user_list").Value().([]interface{})
		s.Room.Count = len(userList)
		Echo("获取到结果: ", s.Room.Count)
		Echo("当前直播间用户：", userList)
		processing = false
	}
	message := result.String()
	Echo("消息回复: ", message)

	if processing == false {
		Echo("已得到结果: 离开频道 ==========", s.Room.Count)
		fmt.Println(s.Room.Count)
		wg.Done()
	}
}

// 授权阶段的数据捕获
func (s *BjyServer) EventAuth(result gjson.Result) {
	serverStr := result.Get("room_server").String()
	if len(serverStr) > 0 {
		server := Server{}
		_ = json.Unmarshal([]byte(serverStr), &server)
		s.LoginParams.WsHost = fmt.Sprintf("%s:%d", server.Url, server.Port)
		_user := result.Get("user").Value().(map[string]interface{})
		_user["webrtc_info"] = result.Get("webrtc_info").Value().(map[string]interface{})
		_userByte, _ := json.Marshal(_user)
		_ = json.Unmarshal(_userByte, &user)
		s.LoginParams.User = user
		s.LoginParams.Number = result.Get("user.number").String()
		s.LoginParams.UserID = result.Get("id").String()

		// 链接 Login 服务
		s.SetLoginClient()
		s.SendUserActive()
	}
}

// 设置 Token
func (t *BjyServer) MessageToken() {
	str := map[string]interface{}{
		"message_type":      "server_info_req",
		"class_id":          cast.ToString(t.Room.RoomId),
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
		Echo("客户端峰发送消息失败: ", err.Error())
		return
	}
	Echo("客户端峰发送消息成功 !", r)
}

// 获取待处理的函数
func (t *BjyServer) getEventFs() []func(gjson.Result) {
	return []func(gjson.Result){
		t.EventLoginUserList,
		t.EventToken,
		t.EventNotice,
	}
}

// 处理用户列表
func (t *BjyServer) EventLoginUserList(result gjson.Result) {
	s := result.Get("message_type").String()
	if s == "user_count_change" {
		t.Room = Room{
			//Count: result.Get("accumulative_user_count").Int(),
		}
	}
}

// 消息通知
func (s *BjyServer) EventNotice(result gjson.Result) {
	msg := map[string]interface{}{
		"class_id": s.ClassRx.ClassId,
		"room_url": s.Room.RoomUrl,
		"message":  result.Value(),
	}
	Echo(msg)
	Echo(result.String())
}

// 处理 Token
func (t *BjyServer) EventToken(result gjson.Result) {
	token := result.Get("webrtc_info.token").String()
	Echo("token", result)
	if len(token) > 0 {
		t.AuthParam = AuthParam{
			SdkAppid:   result.Get("webrtc_info.appId").String(),
			Identifier: result.Get("id").String(),
			UserSig:    token,
		}
		u := result.Get("user").Value().(map[string]interface{})
		u["webrtc_info"] = result.Get("webrtc_info").Value().(map[string]interface{})
		//t.LoginParams.User = u
		t.LoginParams.Number = result.Get("user.number").String()
		t.LoginParams.UserID = result.Get("id").String()
		t.LoginParams.WsHost = fmt.Sprintf("%s:%s", result.Get("room_server.url").String(), result.Get("room_server.port").String())

		t.SetLoginClient()
		t.SendUserActive()
	}
}

// 告诉客户端我来了
func (t *BjyServer) SendUserActive() {
	// 登录
	t.LoginParams.User.WebrtcInfo.WebrtcExt.Resolution = CResolution
	t.LoginParams.User.WebrtcInfo.AppId = 1400310688
	ms := map[string]interface{}{
		"message_type": "login_req",
		"speak_state":  0,
		"token":        t.LoginParams.Token,
		"user":         t.LoginParams.User,
		"start_time":   time.Now().Unix(),
		"end_time":     time.Now().Unix(),
		"support": map[string]interface{}{
			"points_decoder":              2,
			"doodle_version":              1,
			"protocol_version":            1,
			"link_type_consistency":       0,
			"teacher_preferred_link_type": 1,
		},
		"get_cache_group": true,
		"class_id":        cast.ToString(t.Room.RoomId),
		"user_id":         t.LoginParams.UserID,
		"ts":              time.Now().Unix(),
		"signal_send_by": map[string]interface{}{
			"id":       t.LoginParams.UserID,
			"number":   t.LoginParams.Number,
			"type":     2,
			"group":    0,
			"end_type": 0,
		},
	}
	s, err := json.Marshal(ms)
	if err != nil {
		Echo("解析结果失败: ", err.Error())
		return
	}

	err = t.LoginClient.Send(string(s))
	if err != nil {
		Echo("客户端发送消息失败: ", err.Error())
		return
	}
	Echo("客户端发送消息成功 等待回复 !")

	// 获取活跃用户
	ms = map[string]interface{}{
		"message_type": "user_active_req",
		"class_id":     cast.ToString(t.Room.RoomId),
		"user_id":      t.LoginParams.UserID,
		"ts":           t.LoginParams.Ts,
		"signal_send_by": map[string]interface{}{
			"end_type": 0,
			"group":    0,
			"id":       t.LoginParams.UserID,
			"number":   "373653",
			"type":     2,
		},
	}
	s, _ = json.Marshal(ms)
	err = t.LoginClient.Send(string(s))
	if err != nil {
		Echo("客户端发送消息失败: ", err.Error())
		return
	}

	Echo("客户端发送消息成功 等待回复 !", t.LoginParams.WsHost)
}

func Echo(args ...interface{}) {
	//return
	fmt.Println(args)
}
