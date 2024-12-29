package xiao

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	_L = zap.NewNop()
	_S = _L.Sugar()
)

func init() {
	if _, err := UseDefaultLogger(); err != nil {
		panic(err)
	}
}

// ReplaceLogger 用给定的zap.Logger替换context内部的默认全局zap.Logger和zap.SugaredLogger
func ReplaceLogger(logger *zap.Logger) func() {
	var prev = _L
	_L = logger
	_S = logger.Sugar()
	return func() { ReplaceLogger(prev) }
}

// UseDefaultLogger 使用预定义的简单zap.Logger替换全局默认zap.Logger。
func UseDefaultLogger() (func(), error) {
	var logger, err = NewSimpleLogger("info", "stderr", "console", false)
	if err != nil {
		return nil, err
	}
	return ReplaceLogger(logger), nil
}

// UseDevlopLogger 使用预定义的简单zap.Logger替换全局默认zap.Logger，适合于开发、测试、写简单的工具时用。
func UseDevelopLogger() (func(), error) {
	var logger, err = NewSimpleLogger("debug", "log", "console", false)
	if err != nil {
		return nil, err
	}
	return ReplaceLogger(logger), nil
}

// UseSimpleLogger 使用简单的默认风格logger替换掉全局的zap.Logger
func UseSimpleLogger(level, outpath, encoding string, disableCaller bool) (func(), error) {
	var logger, err = NewSimpleLogger(level, outpath, encoding, disableCaller)
	if err != nil {
		return nil, err
	}
	return ReplaceLogger(logger), nil
}

// NewSimpleLogger 生成并返回一个简单的默认风格的zap.Logger
func NewSimpleLogger(level, outpath, encoding string, disableCaller bool) (*zap.Logger, error) {
	var zlevel zapcore.Level
	switch strings.ToLower(level) {
	case "debug", "dbg":
		zlevel = zap.DebugLevel
	case "info", "inf":
		zlevel = zap.InfoLevel
	case "warning", "warn":
		zlevel = zap.WarnLevel
	case "error", "err":
		zlevel = zap.ErrorLevel
	case "panic":
		zlevel = zap.PanicLevel
	case "fatal":
		zlevel = zap.FatalLevel
	default:
		return nil, errors.New("Unexpected log level " + level)
	}

	if dir := filepath.Dir(outpath); dir != "." && dir != ".." && dir != "/" {
		if _, e := os.Stat(dir); errors.Is(e, os.ErrNotExist) {
			if e := os.MkdirAll(dir, 0755); e != nil {
				return nil, e
			}
		}
	}

	var zcfg = zap.Config{
		Level:             zap.NewAtomicLevelAt(zlevel),
		Development:       false,
		DisableCaller:     disableCaller,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          encoding,

		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "M",
			LevelKey:       "L",
			TimeKey:        "T",
			NameKey:        "N",
			CallerKey:      "C",
			FunctionKey:    "",
			StacktraceKey:  "",
			SkipLineEnding: false,
			LineEnding:     "\n",

			EncodeLevel: zapcore.LowercaseLevelEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
			},
			EncodeDuration: zapcore.NanosDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,

			NewReflectedEncoder: nil,
			ConsoleSeparator:    "",
		},

		OutputPaths:      []string{outpath},
		ErrorOutputPaths: nil, // only zap internal error
		InitialFields:    nil,
	}
	return zcfg.Build()
}
