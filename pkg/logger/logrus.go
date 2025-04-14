// Package logger ...
package logger

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

// Option ...
type Option struct {
}

var isDev bool

func init() {
	isDev = os.Getenv("IS_DEV") == "true"
}

// Setup ...
func Setup(option ...*Option) {
	mw := &lumberjack.Logger{
		Filename:   "logs/app.log", // 日志文件路径
		MaxSize:    50,             // 文件最大大小（MB）
		MaxBackups: 10,             // 保留旧文件的最大数量
		MaxAge:     30,             // 保留旧文件的最大天数
		Compress:   false,          // 是否压缩旧文件
	}

	logrus.SetReportCaller(true)
	if isDev {
		logrus.SetFormatter(&logrus.TextFormatter{TimestampFormat: time.DateTime})
		logrus.SetOutput(mw)
		logrus.AddHook(&consoleHook{logger: logrus.StandardLogger()})
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.DateTime})
	}
}

type consoleHook struct {
	logger *logrus.Logger
	//handler *lumberjack.Logger
}

func (h *consoleHook) Fire(entry *logrus.Entry) error {
	// 将日志输出到控制台
	s, err := entry.String()
	if err != nil {
		return err
	}
	//h.logger.Println(s)

	fmt.Println(s)
	return nil
}

func (h *consoleHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}
}
