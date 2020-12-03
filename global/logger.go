package global

import "github.com/sirupsen/logrus"

// 日志输出
func Logger(message string, source string)  {
	logrus.SetLevel(logrus.TraceLevel)

	logrus.Trace(message + " " + source)
}
