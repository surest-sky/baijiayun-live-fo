package bjy

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	netUrl "net/url"
	"os"
	"regexp"
	"strings"
	"task_client/utils/logger"
	"task_client/utils/request"
	"time"
)

type BjyH struct {
	Token       string
	Classid     string
	XClassid    string
	RoomUrl     string
	ReqClient   Client
	AuthClient  Client
	LoginClient Client
	AuthParam
	AuthUserParam
	LoginParams
	Room Room
}

type AuthParam struct {
	SdkAppid   string `json:"sdkAppid"`
	Identifier string `json:"identifier"`
	UserSig    string `json:"userSig"`
}

type AuthUserParam struct {
	Data     []int  `json:"data"`
	Openid   string `json:"openid"`
	TagKey   string `json:"tag_key"`
	Tinyid   string `json:"tinyid"`
	Socketid string `json:"socketid"`
	Ua       string `json:"ua"`
	Version  string `json:"version"`
}

type LoginParams struct {
	Token     string                 `json:"token"`
	StartTime string                 `json:"start_time"`
	EndTime   string                 `json:"end_time"`
	UserID    string                 `json:"user_id"`
	Ts        string                 `json:"ts"`
	Number    string                 `json:"number"`
	User      map[string]interface{} `json:"user"`
	WsHost    string                 `json:"ws_host"`
}

type Room struct {
	Count   int64  `json:"count"`
	ClassId string `json:"class_id"`
}

func (t *BjyH) Serve(url string, x_class_id string) {
	// 解析链接
	// 从源代码中解析data数据
	t.ParseP(url, x_class_id)

	// 设置连接 && 注册函数
	t.SetReqClient()

	// 发送一条ws信息 获取鉴权数据
	t.MessageToken()

	// 设置鉴权链接
	//s.SetSdkClient()

	select {}
}

// 处理 Token
func (t *BjyH) EventToken(result gjson.Result) {
	token := result.Get("webrtc_info.token").String()
	if len(token) > 0 {
		t.AuthParam = AuthParam{
			SdkAppid:   result.Get("webrtc_info.appId").String(),
			Identifier: result.Get("id").String(),
			UserSig:    token,
		}
		u := result.Get("user").Value().(map[string]interface{})
		u["webrtc_info"] = result.Get("webrtc_info").Value().(map[string]interface{})
		t.LoginParams.User = u
		t.LoginParams.Number = result.Get("user.number").String()
		t.LoginParams.UserID = result.Get("id").String()
		t.LoginParams.WsHost = fmt.Sprintf("%s:%s", result.Get("room_server.url").String(), result.Get("room_server.port").String())

		t.SetLoginClient()
		t.GetUserActive()
	}
}

// 处理需要鉴权的参数
func (t *BjyH) EventAuthP(result gjson.Result) {
	cmdInt := result.Get("cmd").Int()
	fmt.Println("----- 我收到消息了 EventAuthP------", cmdInt)

	if cmdInt == 19 {
		t.AuthUserParam = AuthUserParam{
			Data:     []int{},
			Openid:   result.Get("content.openid").String(),
			TagKey:   "on_quality_report",
			Tinyid:   result.Get("content.tinyid").String(),
			Socketid: result.Get("content.socketid").String(),
			Ua:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.82 Safari/537.36 Edg/93.0.961.52",
			Version:  "4.11.1",
		}
		fmt.Println("AuthUserParam Successed !", t.AuthUserParam)

		// 发送获取用户列表消息
		t.MessageUserList()
	}
}

// 处理用户列表
func (t *BjyH) EventUserList(result gjson.Result) {
	cmdInt := result.Get("cmd").Int()
	if cmdInt == 23 {
		fmt.Println("Userlist Successed ! \nToken: ", result.String())
	}
	fmt.Println("----- 我收到消息了 EventUserList------", cmdInt)
}

// 处理用户列表
func (t *BjyH) EventLoginUserList(result gjson.Result) {
	s := result.Get("message_type").String()
	if s == "user_count_change" {
		t.Room = Room{
			Count:   result.Get("accumulative_user_count").Int(),
			ClassId: t.XClassid,
		}
		t.Complete()
	}
}

// 消息通知
func (t *BjyH) EventNotice(result gjson.Result) {
	msg := map[string]interface{}{
		"class_id": t.Classid,
		"room_url": t.RoomUrl,
		"message":  result.Value().(map[string]interface{}),
	}
	logger.Info("WS", msg)
}

// 获取待处理的函数
func (t *BjyH) getEventFs() []func(gjson.Result) {
	return []func(gjson.Result){
		t.EventLoginUserList,
		t.EventToken,
		t.EventNotice,
	}
}

// 设置req 客户端
// example: "https://e83301793.at.baijiayun.com"
func (t *BjyH) SetReqClient() {
	var (
		ws_host     = "wss://pro-signal.baijiayun.com/"
		origin_host = t.RoomUrl
	)

	server := GetServer()
	server.SetClient(ws_host, origin_host)
	if err != nil {
		t.logger(err, "Req Connect : ")
		return
	}

	t.ReqClient = server
	var funcs = t.getEventFs()
	for _, f := range funcs {
		t.ReqClient.RegisterRHandle(f)
	}

	go func() {
		t.ReqClient.Receive()
	}()

	fmt.Println("-- init -- req client -- successed !")
}

// 设置req 客户端
// example: "https://e83301793.at.baijiayun.com"
func (t *BjyH) SetLoginClient() {
	var (
		ws_host     = t.LoginParams.WsHost
		origin_host = t.RoomUrl
	)

	server := GetServer()
	server.SetClient(ws_host, origin_host)
	if err != nil {
		t.logger(err, "Req Connect : ")
		return
	}

	t.LoginClient = server
	var funcs = t.getEventFs()
	for _, f := range funcs {
		t.LoginClient.RegisterRHandle(f)
	}

	go func() {
		t.LoginClient.Receive()
	}()

	fmt.Println("-- init -- req client -- successed !")
}

// 设置 sdk 客户端
// example: "https://e83301793.at.baijiayun.com"
func (t *BjyH) SetSdkClient() {
	var (
		ws_host     = fmt.Sprintf("wss://qcloud.rtc.qq.com/?sdkAppid=%s&identifier=%s&userSig=%s", t.AuthParam.SdkAppid, t.AuthParam.Identifier, t.AuthParam.UserSig)
		origin_host = t.RoomUrl
	)

	fmt.Println("ws_host", ws_host)

	server := GetServer()
	server.SetClient(ws_host, origin_host)
	if err != nil {
		t.logger(err, "Req Connect : ")
		return
	}

	t.AuthClient = server
	var funcs = t.getEventFs()
	for _, f := range funcs {
		t.AuthClient.RegisterRHandle(f)
	}

	go func() {
		t.AuthClient.Receive()
	}()

	fmt.Println("-- init -- sdk client -- successed !")
}

// 设置 Token
func (t *BjyH) MessageToken() {
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

// 获取用户列表
func (t *BjyH) MessageUserList() {
	var s []byte
	pingP := map[string]interface{}{
		"data":    "",
		"openid":  t.AuthUserParam.Openid,
		"tag_key": "ping",
		"tinyid":  t.AuthUserParam.Tinyid,
		"version": t.AuthUserParam.Version,
	}
	s, _ = json.Marshal(pingP)
	t.AuthClient.Send(string(s))

	cRoom := `{"tag_key":"on_create_room","data":{"openid":"22516010","tinyid":"144115261421665962","peerconnectionport":"","useStrRoomid":1,"roomid":"21092398878340","sdkAppID":"1400310688","socketid":"socketid_copy","userSig":"userSig_token","privMapEncrypt":"","privMap":"","relayip":"11.177.123.23","dataport":9000,"stunport":8800,"checkSigSeq":"65537","pstnBizType":0,"pstnPhoneNumber":null,"recordId":null,"role":"user","jsSdkVersion":"5101","sdpSemantics":"unified-plan","browserVersion":"NotSupportedBrowser","closeLocalMedia":true,"trtcscene":2,"trtcrole":20,"isAuxUser":0,"autoSubscribe":false},"openid":"22516010","tinyid":"144115261421665962","version":"4.11.1","ua":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.82 Safari/537.36 Edg/93.0.961.52","report":{"AbilityOption":{"GeneralLimit":{"CPULimit":{"uint32_CPU_num":"8","str_CPU_name":"MacIntel","uint32_CPU_maxfreq":"0","model":"","uint32_total_memory":"0"},"uint32_terminal_type":"12","uint32_device_type":"0","str_os_verion":"Mac","uint32_link_type":"1","str_client_version":"4.11.1","uint32_net_type":"0","ua":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.82 Safari/537.36 Edg/93.0.961.52","version":""},"AVLimit":{"EncVideoCodec":"H264","EncVideoWidth":0,"EncVideoHeight":0,"EncVideoBr":"0","EncVideoFps":0,"EncAudioCodec":"opus","EncAudioFS":0,"EncAudioCh":0,"EncAudioBr":"0"}}}}`
	cRoom = strings.ReplaceAll(cRoom, "22516010", t.AuthUserParam.Openid)
	cRoom = strings.ReplaceAll(cRoom, "144115261421665962", t.AuthUserParam.Tinyid)
	cRoom = strings.ReplaceAll(cRoom, "userSig_token", t.AuthParam.UserSig)
	cRoom = strings.ReplaceAll(cRoom, "socketid_copy", t.AuthUserParam.Socketid)
	t.AuthClient.Send(cRoom)

	var um = t.AuthUserParam
	s, _ = json.Marshal(um)
	if err := t.AuthClient.Send(string(s)); err != nil {
		t.logger(err)
		return
	}

	//sRoom := `{"tag_key":"on_quality_report","data":{"WebRTCQualityReq":{"uint64_begine_utime":1632382436572,"uint64_end_utime":1632382438574,"uint32_real_num":173,"uint32_delay":0,"uint32_CPU_curfreq":0,"uint32_total_send_bps":0,"uint32_total_recv_bps":271640,"AudioReportState":{"uint32_audio_enc_pkg_br":0,"uint32_audio_real_recv_pkg":100,"uint32_audio_flow":300,"uint32_audio_real_recv_br":1200,"uint32_audio_delay":0,"uint32_audio_jitter":0,"uint32_microphone_status":0,"sentAudioLevel":0,"sentAudioEnergy":0,"AudioDecState":[{"uint32_audio_delay":0,"uint32_audio_jitter":0.005,"uint32_audio_real_recv_pkg":100,"uint32_audio_flow":300,"uint32_audio_real_recv_br":1200,"uint64_sender_uin":"144115225021193563","userId":"22581530","packetsLost":0,"totalPacketsLost":0,"audioLevel":0,"audioEnergy":0}]},"VideoReportState":{"uint32_video_delay":0,"uint32_video_snd_br":0,"uint32_video_total_real_recv_pkg":73,"uint32_video_rcv_br":270440,"uint32_send_total_pkg":0,"VideoEncState":[{"uint32_enc_width":0,"uint32_enc_height":0,"uint32_capture_fps":0,"uint32_enc_fps":0}],"VideoDecState":[{"uint32_video_recv_fps":14.5,"uint32_video_recv_br":270440,"uint32_video_real_recv_pkg":73,"uint32_dec_height":240,"uint32_dec_width":320,"uint32_video_jitter":0,"uint64_sender_uin":"144115225021193563","userId":"22581530","packetsLost":0,"totalPacketsLost":0,"uint32_video_strtype":0,"int32_video_freeze_ms":0},{"uint32_video_recv_fps":0,"uint32_video_recv_br":0,"uint32_video_real_recv_pkg":0,"uint32_dec_height":0,"uint32_dec_width":0,"uint32_video_jitter":0,"uint64_sender_uin":"144115225021193563","userId":"22581530","packetsLost":0,"totalPacketsLost":0,"uint32_video_strtype":2,"int32_video_freeze_ms":0}]},"RTTReportState":{"uint32_delay":0,"RTTDecState":[{"uint32_delay":7,"uint64_sender_uin":"144115225021193563"}]}},"eventList":[],"sdkAppId":1400310688,"tinyid":"144115263213178308","roomid":"21092394683932","socketid":"a207240c-7c7c-45bf-aef9-a067d294bcf0","clientip":"61.141.253.161","serverip":"14.22.4.168","cpunumber":8,"cpudevice":"MacIntel","devicename":"MacIntel","ostype":"MacIntel","mode":""},"openid":"22587260","tinyid":"144115263213178308","version":"4.11.1","ua":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.82 Safari/537.36 Edg/93.0.961.52"}`
	//sRoom = strings.ReplaceAll(sRoom, "1400310688", t.AuthParam.Identifier)
	//sRoom = strings.ReplaceAll(sRoom, "144115263213178308", t.AuthUserParam.Tinyid)
	//sRoom = strings.ReplaceAll(sRoom, "a207240c-7c7c-45bf-aef9-a067d294bcf0", t.AuthUserParam.Socketid)
	//sRoom = strings.ReplaceAll(sRoom, "22587260", t.AuthUserParam.Openid)
	//t.AuthClient.Send(sRoom)
}

// 从 Html 中获取 data json
func (t *BjyH) setData(url string) {
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
	data = strings.ReplaceAll(data, "var data = ", "")
	data = strings.ReplaceAll(data, ";", "")
	rs := gjson.Parse(data)

	t.Classid = rs.Get("class.id").String()
	t.LoginParams = LoginParams{
		Token:     rs.Get("token").String(),
		StartTime: rs.Get("startTimeTs").String(),
		EndTime:   rs.Get("endTimeTs").String(),
		UserID:    "",
		Ts:        cast.ToString(time.Now().Unix()) + "000",
	}
	fmt.Println(rs.Get("class.id").String())
	fmt.Println(rs)

	t.Room = Room{
		Count:   0,
		ClassId: t.XClassid,
	}
}

// 解析链接url
func (t *BjyH) ParseP(url string, x_class_id string) {
	up, err := netUrl.Parse(url)
	t.logger(err)

	m, _ := netUrl.ParseQuery(up.RawQuery)

	t.RoomUrl = fmt.Sprintf("https://%s", up.Host)
	t.Classid = m.Get("room_id")
	t.XClassid = x_class_id

	t.setData(url)
}

func (t BjyH) GetUserActive() {
	// 登录
	ms := map[string]interface{}{
		"message_type": "login_req",
		"speak_state":  0,
		"token":        t.LoginParams.Token,
		"user":         t.LoginParams.User,
		"start_time":   t.LoginParams.StartTime,
		"end_time":     t.LoginParams.EndTime,
		"support": map[string]interface{}{
			"points_decoder":              2,
			"doodle_version":              1,
			"protocol_version":            1,
			"link_type_consistency":       0,
			"teacher_preferred_link_type": 1,
		},
		"get_cache_group": true,
		"class_id":        t.Classid,
		"user_id":         t.LoginParams.UserID,
		"ts":              t.LoginParams.Ts,
		"signal_send_by": map[string]interface{}{
			"id":       t.LoginParams.UserID,
			"number":   t.LoginParams.Number,
			"type":     2,
			"group":    0,
			"end_type": 0,
		},
	}
	s, err := json.Marshal(ms)
	t.logger(err)

	err = t.LoginClient.Send(string(s))
	t.logger(err)

	// 获取活跃用户
	ms = map[string]interface{}{
		"message_type": "user_active_req",
		"class_id":     t.Classid,
		"user_id":      t.UserID,
		"ts":           t.Ts,
	}
	s, err = json.Marshal(ms)
	t.logger(err)

	err = t.LoginClient.Send(string(s))
	t.logger(err)
}

// 周期完成
func (t *BjyH) Complete() {
	// 触发通知
	r := &request.Reuqest{
		Ctype:  request.FormJson,
		Method: "POST",
		Url:    "https://trmk.teachingrecord.com/api/late_class/" + t.Classid,
		Data:   t.Room,
	}
	request.NewRequest(r)
	t.stop()
}

// 结束
func (t *BjyH) stop() {
	// 关闭已开启的连接
	t.LoginClient.Stop()
	//t.AuthClient.Stop()
	t.ReqClient.Stop()

	// Stop
	logger.Info("client all closed", nil)

	// 关闭当前任务 & 退出当前进程
	os.Exit(0)
}

func (t *BjyH) logger(e error, p ...string) {
	if e == nil {
		return
	}
	var prefix string
	if len(p) == 0 {
		prefix = "Error :"
	} else {
		prefix = p[0]
	}

	fmt.Println("Error", err.Error())
	logger.Error(prefix, err.Error())
}
