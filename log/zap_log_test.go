package log

import (
	"testing"
)

func TestInitLogger(t *testing.T) {
	InitLogger(FileName("西永一哥"), StdOutput(true), Fields(String("test", "app")))
	Error("上街", String("一起", "吃饭"))
	Info("上街", String("一起", "吃面"))
}
