package request

import (
	"encoding/json"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/PuerkitoBio/goquery"
	"math"
	"net/http"
	netUrl "net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"talk/config"
	"talk/global"
	"time"
)

type student struct {
	Name  string
	Age   string
	No    string
	Level string
}

type class struct {
	CLASS_ID         string
	Time             string
	Tool             string
	Teahcher_Name    string
	Lesson_Plan      string
	Student          string
	Flowers          string
	Feedback         string
	Org_Class        string
	Experience_Class string
	First_Class      string
	AssessClass      string
	VipClass         string
	Status           string
}

var dir = config.GetString("EXCEL_DIR")
var filename = fmt.Sprintf("%s/%s", dir, time.Now().Format("2006-01-02"))

/** 初始化登录 */
func initLogin() {
	if s, _ := global.REDIS_CLIENT.Get(loginUrl).Result(); len(s) == 0 {
		Login()
	}
	s, _ := global.REDIS_CLIENT.Get(loginUrl).Result()

	if len(s) == 0 {
		Login()
	}
}

/** 处理课程，分发 */
func HandleClass() {
	panic(filename)
	// 初始化登录
	initLogin()

	// 分发异步脚本
	synClass()
}

/** 开始处理 */
func synClass() {
	// 第一次跑html
	var start_at = time.Now().Format("2006-01-02")
	var end_at = time.Now().Format("2006-01-02")
	s := dowloadClass(start_at, end_at, 1)

	// 获取html
	html, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	global.PanicError(err, "解析html")

	// 获取当前页有多少条，设置channel缓冲池
	total := html.Find(".pt_10 .red").Text()
	t, _ := strconv.ParseInt(total, 10, 32)
	pageCount := int(math.Ceil(float64(t / 10)))
	classChan := make(map[int]chan string, 50)
	// 执行脚本
	fmt.Println("开始执行脚本")

	// 分发协程脚本去执行文件解析和下载
	for i := 1; i < pageCount; i++ {
		//filename := "./json/" + strconv.Itoa(i) + ".json"
		classChan[i] = make(chan string)
		go parseHtml(dowloadClass(start_at, end_at, i), i, classChan[i])
	}

	// 消费channel数据，准备写入excel
	var classes4 []map[string]string
	var classes5 []map[string]string
	var classes6 []map[string]string
	var classes []map[string]string
	for _, class := range classChan {
		var tempClass []map[string]string
		msg := <-class
		_ = json.Unmarshal([]byte(msg), &tempClass)
		for _, v := range tempClass {
			if v["Status"] == "4" {
				classes4 = append(classes4, v)
				continue
			}
			if v["Status"] == "5" {
				classes5 = append(classes5, v)
				continue
			}
			if v["Status"] == "6" {
				classes6 = append(classes6, v)
				continue
			}
			classes = append(classes, v)
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go importExcel(classes4, filename+"-4.xls", &wg)
	wg.Add(1)
	go importExcel(classes5, filename+"-5.xls", &wg)
	wg.Add(1)
	go importExcel(classes6, filename+"-6.xls", &wg)
	wg.Add(1)
	go importExcel(classes, filename+"-1.xls", &wg)
	wg.Wait()

	fmt.Println("success!")
}

/** 执行读取文件 */
func dowloadClass(startDate string, endDate string, c int) string {
	currentPage := strconv.Itoa(c)
	sessionid, _ := global.REDIS_CLIENT.Get(loginUrl).Result()
	html, _ := global.REDIS_CLIENT.Get(classUrl).Result()
	if len(html) == 0 {
		data := netUrl.Values{
			"down":               []string{""},
			"page.currentPage":   []string{currentPage},
			"page.pageCount":     []string{"249"},
			"records":            []string{"5000"},
			"date1":              []string{startDate},
			"date2":              []string{endDate},
			"datebookCourseType": []string{"0"},
			"userid":             []string{"-1"},
			"courseType":         []string{"-1"},
			"teacherStatus":      []string{"1"},
			"uncomitted":         []string{"0"},
			"status":             []string{"-1"},
		}
		str := data.Encode()

		request, _ := http.NewRequest("POST", classUrl, strings.NewReader(str))
		request.AddCookie(&http.Cookie{
			Name:  "JSESSIONID",
			Value: sessionid,
		})
		request.Header.Add("Host", "www.talk915.com")
		request.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.149 Safari/537.36")
		request.Header.Add("Accept-Encoding", "gzip, deflate, br")
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		response, err := client.Do(request)
		//defer resp.Body.Close()

		if err != nil {
			fmt.Println("<<< dowloadClass 失败" + err.Error())
			html = ""
		}
		html := global.GetBodyString(response)
		return html

		global.REDIS_CLIENT.Set(classUrl, html, time.Minute*20)
		return ""
	}
	return html
}

/** html 解析 */
func parseHtml(s string, current int, classChan chan string) {
	fmt.Println("current: " + strconv.Itoa(current))
	html, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	global.PanicError(err, "解析html")

	row := make([]class, 0)
	html.Find("#tableCourseList tr").Each(func(i int, node *goquery.Selection) {
		isVip := "0"
		if len(node.Find("td").Eq(0).Find("label").Text()) > 0 {
			isVip = "1"
		}

		// 首行不读取
		if i == 0 {
			return
		}

		temp := class{
			CLASS_ID:         node.Find("td").Eq(0).Find("p").Text(),
			Teahcher_Name:    node.Find("td").Eq(1).Find("p").Text(),
			VipClass:         isVip,
			Time:             node.Find("td").Eq(2).Text(),
			Tool:             trim(node.Find("td").Eq(3).Text()),
			Lesson_Plan:      trim(node.Find("td").Eq(4).Text()),
			Student:          getStudent(node.Find("td").Eq(5)),
			Feedback:         "",
			Org_Class:        "",
			Experience_Class: "",
			First_Class:      "",
			AssessClass:      "",
			Status:           getStatus(trim(node.Find("td").Eq(8).Find("p").Text())),
		}
		row = append(row, temp)
	})

	//filename := "./json/" + strconv.Itoa(current) + ".json"
	// 取消文件写入
	result, _ := json.Marshal(row)
	//f, err := os.Create(filename)
	//_, _ = io.Copy(f, strings.NewReader(string(result)))

	classChan <- string(result)

	fmt.Println("解析完成>>>>>")
}

/** 自定义获取学生资料 */
func getStudent(node *goquery.Selection) string {
	student := student{
		Name:  strings.ReplaceAll(node.Find("p").Eq(0).Text(), "Name : ", ""),
		Age:   strings.ReplaceAll(node.Find("p").Eq(1).Text(), "Age : ", ""),
		No:    strings.ReplaceAll(node.Find("p").Eq(2).Text(), "No : ", ""),
		Level: strings.ReplaceAll(node.Find("p").Eq(3).Text(), "Level : ", ""),
	}
	s, _ := json.Marshal(student)
	return string(s)
}

// 获取状态
func getStatus(statusTest string) string {
	if statusTest == "Absent with notice" {
		return "6"
	}

	if statusTest == "Absent without noitce" {
		return "4"
	}

	if statusTest == "Canceled" {
		return "5"
	}
	return "1"
}

/** 格式化一些乱七八糟的字符串 */
func trim(text string) string {
	text = strings.ReplaceAll(text, " ", "")
	text = strings.ReplaceAll(text, "", "")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", "")
	text = strings.TrimSpace(text)
	return text
}

/** 导入excel */
func importExcel(classes []map[string]string, filename string, wg *sync.WaitGroup) {
	// 读取excel
	f := initExcel()
	sheet := "Sheet1"
	for index, class := range classes {
		item := mapToSlice(class)
		for k, v := range item {
			excelRow := global.IndexExcelRow(k) + strconv.Itoa(index+2)
			f.SetCellValue(sheet, excelRow, v)
		}
	}

	if err := f.SaveAs(filename); err != nil {
		fmt.Println(err)
	}

	wg.Done()
}

/** 初始化excel文件 */
func initExcel() *excelize.File {
	f := excelize.NewFile()
	f.SetActiveSheet(1)
	sheet := "Sheet1"
	classHeader := [14]string{"Class ID", "Time", "Tool", "Teacher Name", "Lesson Plan", "Student", "Flowers", "Feedback", "Org Class", "Experience Class", "First Class", "Assess Class", "Vip Class", "Status"}
	for k, v := range classHeader {
		excelRow := global.IndexExcelRow(k)
		f.SetCellValue(sheet, excelRow+"1", v)
		f.SetColWidth(sheet, excelRow, excelRow, 20)
	}
	return f
}

/** map函数转slice */
func mapToSlice(m map[string]string) []interface{} {
	class := &class{
		CLASS_ID:         "",
		Time:             "",
		Tool:             "",
		Teahcher_Name:    "",
		Lesson_Plan:      "",
		Student:          "",
		Flowers:          "",
		Feedback:         "",
		Org_Class:        "",
		Experience_Class: "",
		First_Class:      "",
		AssessClass:      "",
		VipClass:         "",
	}
	var typeInfo = reflect.TypeOf(*class)

	num := typeInfo.NumField()
	var keys []string
	for i := 0; i < num; i++ {
		keys = append(keys, typeInfo.Field(i).Name)
	}

	s := make([]interface{}, 0, len(m))
	for _, v := range keys {
		s = append(s, m[string(v)])
	}
	return s
}
