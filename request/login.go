package request

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-redis/redis"
	"net/http"
	netUrl "net/url"
	"os"
	"strings"
	"talk/global"
	"time"
)

const classUrl = "https://www.talk915.com/teacher/teascheduleClass.action"
const loginUrl  = "https://www.talk915.com/user/loginAction!login.action"
var client = &http.Client{}

type ResponseResult struct {
	REDIRECTURL string
	RESULTMESSAGE string
	RESULTCODE string
}


func Login()  {
	setSessionId()
	loginForm()
}

/**
 * 获取session ID
 */
func setSessionId() string {
	sessionid, _:= global.REDIS_CLIENT.Get(loginUrl).Result()
	fmt.Println("<<< session id :  " + sessionid)
	if len(sessionid) == 0 {
		reqest, _ := http.NewRequest("GET", loginUrl, nil)
		reqest.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36")
		resp, err := client.Do(reqest)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("<<< 开始获取session ID")

		cookie := resp.Header["Set-Cookie"][0]
		r := strings.Split(cookie, ";")[0]
		r = strings.Split(r, "=")[1]
		sessionid = r;
		global.REDIS_CLIENT.Set(loginUrl, sessionid, time.Minute * 10)
	}
	fmt.Println("<<< 获取session ID成功:  " + sessionid)
	fmt.Println("<<< 开始正式登陆")

	return sessionid
}

/**
 * 登陆
 */
func loginForm()  {
	sessionid, _ := global.REDIS_CLIENT.Get(loginUrl).Result()
	data := netUrl.Values{}
	data["username"] = []string{"ta-peng"}
	data["password"] = []string{"Sophie123456"}
	str := data.Encode()

	fmt.Println("<<< 开始登录 :" + sessionid)
	fmt.Println(strings.NewReader(str))

	reqest, _ := http.NewRequest("POST", loginUrl, strings.NewReader(str))
	reqest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cookie := &http.Cookie{
		Name:       "JSESSIONID",
		Value:      sessionid,
	}
	reqest.AddCookie(cookie)
	reqest.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36")
	reqest.Header.Add("Host", "https://www.talk915.com/")

	resp, err := client.Do(reqest)
	if err != nil {
		fmt.Println("<<< 登陆失败" + err.Error())
		return
	}


	result := global.GetBodyString(resp)
	// json 解析
	var ResponseResult *ResponseResult
	_ = json.Unmarshal([]byte(result), &ResponseResult)

	if ResponseResult.RESULTCODE == "0" {
		fmt.Println("<<< 登陆成功")
		fmt.Println(ResponseResult)
		return
	}

	fmt.Println("<<< 登陆失败 xxxxx")

	fmt.Println(global.GetBodyString(resp))
	os.Exit(-1)

}

