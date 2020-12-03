package global

import (
	"compress/gzip"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func PanicError(err error, source string)  {
	if err == nil {
		return
	}

	msg := "发生错误：" + err.Error() + ">>>>>>来源 " + source
	logrus.Fatal(msg)
}

func GetBodyString(response *http.Response) string {
	var s []byte
	if response.Header.Get("Content-Encoding") == "gzip" {
		reader, _ := gzip.NewReader(response.Body)
		s, _= ioutil.ReadAll(reader)
		defer reader.Close()
	}else {
		s, _= ioutil.ReadAll(response.Body)
	}

	return (string(s))
}

func IndexExcelRow(index int)string{
	var Letters = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	// index := 10000
	result := Letters[index %26]
	index = index / 26
	for index > 0 {
		index = index - 1
		result = Letters[index %26] + result
		index = index / 26
	}
	return result
}
