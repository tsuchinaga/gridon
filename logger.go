package gridon

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

var (
	loggerSingleton    ILogger
	loggerSingletonMtx sync.Mutex
)

// getLogger - ロガーの取得
func getLogger() (ILogger, error) {
	loggerSingletonMtx.Lock()
	defer loggerSingletonMtx.Unlock()

	if loggerSingleton == nil {
		logger := &logger{}

		notice, err := os.OpenFile("notice.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		logger.notice = log.New(io.MultiWriter(os.Stdout, notice), "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

		warning, err := os.OpenFile("warning.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		logger.warning = log.New(io.MultiWriter(os.Stderr, warning), "", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

		loggerSingleton = logger
	}

	return loggerSingleton, nil
}

// ILogger - ロガーのインターフェース
type ILogger interface {
	Notice(v ...interface{})
	Warning(v ...interface{})
}

// logger - ロガー
type logger struct {
	notice  *log.Logger
	warning *log.Logger
}

func (l *logger) Notice(v ...interface{}) {
	_ = l.notice.Output(2, fmt.Sprintln(v...))
}

func (l *logger) Warning(v ...interface{}) {
	_ = l.warning.Output(2, fmt.Sprintln(v...))
}
