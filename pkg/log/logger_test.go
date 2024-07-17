package log

import (
	"fmt"
	"go.uber.org/zap/zapcore"
	"testing"
)

func TestLog(t *testing.T) {
	logger := NewLogger()
	logger.SetLogName("yiEr")
	logger.SetLevel(zapcore.InfoLevel)
	logger.SetOutputPath(fmt.Sprintf("./%s.log", logger.logName))
	logger.Infof("hello,debug,%s", "邱一二")
}
