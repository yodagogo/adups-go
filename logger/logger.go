package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//Logger xx
type Logger *zap.SugaredLogger

//Config logger config
type Config struct {
	Env         string //运行环境
	LogPath     string //log保存路径默认当前路径下的logs
	MaxSize     int    //log大小
	MaxAge      int    //log保存天数
	ServiceName string //服务名称
	Level       string //日志基本
	Format      string //日志格式json,file
	Mode        string //1:detail(默认打印完整的log格式),2:data(仅打印数据部分)
	log         *zap.SugaredLogger
}

var levelMap = map[string]zapcore.Level{
	"debug":   zapcore.DebugLevel,
	"info":    zapcore.InfoLevel,
	"warning": zapcore.WarnLevel,
	"error":   zapcore.ErrorLevel,
	"fatal":   zapcore.FatalLevel,
	"panic":   zapcore.PanicLevel,
}

//BuildConfig 初始化logger
func (lg *Config) BuildConfig() Logger {
	var logger *zap.Logger
	if len(lg.LogPath) == 0 {
		lg.LogPath = "logs"
	}

	atoLevel := zap.NewAtomicLevel()

	//默认Info级别
	atoLevel.SetLevel(zap.InfoLevel)

	lv, ok := levelMap[lg.Level]
	if ok {
		atoLevel.SetLevel(lv)
	}
	if _, err := os.Stat(lg.LogPath); !os.IsExist(err) {
		os.Mkdir(lg.LogPath, 0755)
	}
	var fileName = fmt.Sprintf("%s/%s", lg.LogPath, time.Now().Format("20060102"))
	if len(lg.ServiceName) != 0 {
		fileName = fmt.Sprintf("%s/%s", lg.LogPath, lg.ServiceName)
	}

	InfoLogFile := fileName + "_info"
	ErrLogFile := fileName + "_err"
	// 实现两个判断日志等级的interface  可以自定义级别展示
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.WarnLevel && lvl >= lv
	})
	errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel && lvl >= lv
	})

	infoWriter := getWriter(InfoLogFile, lg.MaxAge)
	errorWriter := getWriter(ErrLogFile, lg.MaxAge)

	//cfg默认是detail 打印完整的log信息
	var cfg = zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		CallerKey:      "file",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     timeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	//仅打印log 数据部分
	if lg.Mode == "data" {
		cfg = zapcore.EncoderConfig{
			MessageKey:     "msg",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     timeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
	}
	var core zapcore.Core
	//默认json格式
	core = zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(cfg), zapcore.AddSync(infoWriter), infoLevel),
		zapcore.NewCore(zapcore.NewJSONEncoder(cfg), zapcore.AddSync(errorWriter), errorLevel),
	)
	if lg.Format == "file" {

		core = zapcore.NewTee(
			zapcore.NewCore(zapcore.NewConsoleEncoder(cfg), zapcore.AddSync(infoWriter), infoLevel),
			zapcore.NewCore(zapcore.NewConsoleEncoder(cfg), zapcore.AddSync(errorWriter), errorLevel),
		)
	}

	logger = zap.New(core, zap.AddCaller())

	if len(lg.ServiceName) != 0 {
		logger = zap.New(core, zap.AddCaller(), zap.Fields(zap.String("serviceName", lg.ServiceName)))
	}
	lg.log = logger.Sugar()
	return logger.Sugar()
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}
func getWriter(filename string, maxAge int) io.Writer {
	// 生成rotatelogs的Logger 实际生成的文件名 demo.logservice.YYmmddHH

	hook, err := rotatelogs.New(
		filename+".%Y%m%d.log", // 没有使用go风格反人类的format格式
		//rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Hour*24*time.Duration(maxAge)), // 保存天数
		rotatelogs.WithRotationTime(time.Hour*24),                 //切割频率 24小时
	)
	if err != nil {
		log.Println("日志启动异常")
		panic(err)
	}
	return hook
}
