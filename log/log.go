package log

import (
	"Three_kingdoms_SLG/global"
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var DefaultLog *zap.Logger

func InitT1() {

	maxSize := global.Config.LogServer.Maxsize
	fileDir := global.Config.LogServer.File_die
	maxBackups := global.Config.LogServer.Max_backups
	maxAge := global.Config.LogServer.Max_age
	compress := global.Config.LogServer.Compress
	sa := strings.Split(filepath.Base(os.Args[0]), ".")
	fileName := sa[0] + ".log"
	fmt.Println(fileName)
	hook := lumberjack.Logger{
		Filename:   path.Join(fileDir, fileName), // 日志文件路径
		MaxSize:    maxSize,                      // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: maxBackups,                   // 日志文件最多保存多少个备份
		MaxAge:     maxAge,                       // 文件最多保存多少天
		Compress:   compress,                     // 是否压缩
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}

	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zap.InfoLevel)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook)),
		atomicLevel,
	)

	caller := zap.AddCaller()
	development := zap.Development()
	DefaultLog = zap.New(core, caller, development)

}
