package core

import (
	"github.com/chuccp/shareExplorer/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
)

func initLogger(path string, level string) (*zap.Logger, error) {
	outLevel := zapcore.InfoLevel
	if strings.EqualFold(level, "debug") {
		outLevel = zapcore.DebugLevel
	}
	if strings.EqualFold(level, "error") {
		outLevel = zapcore.ErrorLevel
	}
	writeFileCore, err := getFileLogWriter(path, outLevel)
	if err != nil {
		return nil, err
	}
	core := zapcore.NewTee(writeFileCore, getStdoutLogWriter())
	return zap.New(core, zap.AddCaller()), nil
}
func getFileLogWriter(path string, level zapcore.Level) (zapcore.Core, error) {

	logger := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, zapcore.AddSync(logger), level)
	return core, nil
}
func getStdoutLogWriter() zapcore.Core {
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, os.Stdout, zapcore.DebugLevel)
	return core
}
func getEncoder() zapcore.Encoder {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout(util.TimestampFormat)
	return zapcore.NewJSONEncoder(config)
}
