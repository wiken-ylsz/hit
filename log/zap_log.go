package log

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

var zlog *zap.Logger

type Options struct {
	FileName    string // log file name  带路径
	AppName     string // field of appName
	SaveToLocal bool   // 是否保存至本地
	StdOutput   bool   // 是否打印标准输出
	MaxSize     int    // 单个文件大小, 单位M
	MaxAge      int    // 文件保存的天数
	MaxBackups  int    // 保存的日志文件的个数
	Fields []zap.Field
}

type OptionFunc func(options *Options)

func FileName(fileName string) OptionFunc {
	return func(options *Options) {
		options.AppName = fileName
	}
}

func AppName(appName string) OptionFunc {
	return func(options *Options) {
		options.AppName  = appName
	}
}

func SaveToLocal(b bool) OptionFunc {
	return func(options *Options) {
		options.SaveToLocal = b
	}
}

func StdOutput(b bool) OptionFunc  {
	return func(options *Options) {
		options.StdOutput = b
	}
}

func MaxSize(size int) OptionFunc  {
	return func(options *Options) {
		options.MaxSize = size
	}
}

func MaxAge(days int) OptionFunc  {
	return func(options *Options) {
		options.MaxAge = days
	}
}

func MaxBackups(maxBackups int) OptionFunc  {
	return func(options *Options) {
		options.MaxBackups = maxBackups
	}
}

func Fields(fields ...zap.Field) OptionFunc {
	return func(options *Options) {
		options.Fields = fields
	}
}


// InitLogger 初始化日志,
func InitLogger(options ...OptionFunc) {
	// 此处的配置是从我的项目配置文件读取的，读者可以根据自己的情况来设置
	op := &Options{
		FileName:    "./log.log",
		SaveToLocal: true,
		StdOutput:   false,
		MaxSize:     100,
		MaxAge:      30,
		MaxBackups:  30,
	}
	for _, o := range options {
		o(op)
	}

	hook := lumberjack.Logger{
		Filename:   op.FileName, // 日志文件路径
		MaxSize:    op.MaxSize,     // 每个日志文件保存的大小 单位:M
		MaxAge:     op.MaxAge,      // 文件最多保存多少天
		MaxBackups: op.MaxBackups,      // 日志文件最多保存多少个备份
		Compress:   false,   // 是否压缩
	}
	encoderConfig := zapcore.EncoderConfig{
		MessageKey: "msg",
		LevelKey:   "level",
		TimeKey:    "time",
		NameKey:    "logger",
		CallerKey:      "file",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 短路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zap.DebugLevel)
	var writes = make([]zapcore.WriteSyncer, 0)
	//var writes = []zapcore.WriteSyncer{zapcore.AddSync(&hook)}
	if op.SaveToLocal {
		writes = append(writes, zapcore.AddSync(&hook))
	}
	// 如果是开发环境，同时在控制台上也输出
	if op.StdOutput {
		writes = append(writes, zapcore.AddSync(os.Stdout))
	}
	if len(writes) == 0 {
		fmt.Println("没有设置任何日志输出,请知晓!")
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(writes...),
		atomicLevel,
	)

	// 开启开发模式，堆栈跟踪
	caller := zap.AddCaller()
	// 开启文件及行号
	development := zap.Development()

	// 设置初始化字段
	if len(op.Fields) != 0 {
		//field := zap.Fields(zap.String("appName", op.AppName))
		if op.AppName != "" {
			op.Fields = append(op.Fields, zap.String("app_name", op.AppName))
		}
		field := zap.Fields(op.Fields...)
		zlog = zap.New(core, caller, development, field)
		return
	}
	// 构造日志
	zlog = zap.New(core, caller, development)
}

func String(key, val string) zap.Field {
	return zap.String(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

func Float64(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}

func Duration(key string, val time.Duration) zap.Field {
	return zap.Duration(key, val)
}
func Time(key string, val time.Time) zap.Field  {
	return zap.Time(key, val)
}

func Binary(key string, val []byte) zap.Field  {
	return zap.Binary(key, val)
}

func Bool(key string, val bool) zap.Field  {
	return  zap.Bool(key, val)
}
func Any(key string,val interface{}) zap.Field {
	return zap.Any(key, val)
}

func ZapError(err error) zap.Field {
	return zap.Error(err)
}

func Array(key string, val zapcore.ArrayMarshaler) zap.Field {
	return zap.Array(key, val)
}

func Info(msg string, fields ...zap.Field) {
	zlog.Info(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	zlog.Debug(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	zlog.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	zlog.Fatal(msg, fields...)
}

func Panic(msg string, fields ...zap.Field)  {
	zlog.Panic(msg, fields...)
}

