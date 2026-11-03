package xiao

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// global 是此后创建的Context所用的zap，运行期可被替换
var global atomic.Pointer[zap.SugaredLogger]

func init() {
	global.Store(zap.NewNop().Sugar())
	if _, err := UseDefaultLogger(); err != nil {
		panic(err)
	}
}

// ReplaceLogger 替换此后创建的Context所用的zap，已有的Context不受影响，nil视同zap.NewNop()；
// 返回的函数用于恢复替换前的logger
func ReplaceLogger(logger *zap.Logger) func() {
	if logger == nil {
		logger = zap.NewNop()
	}
	var prev = global.Swap(logger.Sugar())
	return func() { ReplaceLogger(prev.Desugar()) }
}

// UseDefaultLogger 采用默认的全局logger：info级别、console格式、输出到stderr，返回值同ReplaceLogger
func UseDefaultLogger() (func(), error) {
	return UseSimpleLogger("info", "stderr", "console", false)
}

// UseDevelopLogger 以debug级别将日志写入当前目录下名为log的文件，适用于开发、测试与简单工具，返回值同ReplaceLogger
func UseDevelopLogger() (func(), error) {
	return UseSimpleLogger("debug", "log", "console", false)
}

// UseSimpleLogger 以NewSimpleLogger创建的logger替换全局logger，返回值同ReplaceLogger
func UseSimpleLogger(level, outpath, encoding string, disableCaller bool) (func(), error) {
	var logger, err = NewSimpleLogger(level, outpath, encoding, disableCaller)
	if err != nil {
		return nil, err
	}
	return ReplaceLogger(logger), nil
}

// NewSimpleLogger 按本库的默认风格创建zap.Logger，outpath为文件路径或file://地址时自动创建所在目录
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
		return nil, errors.New("unexpected log level " + level)
	}

	if dir := filepath.Dir(sinkFilePath(outpath)); dir != "." && dir != ".." && dir != "/" {
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
		ErrorOutputPaths: nil, // zap自身的内部错误，置空即丢弃
		InitialFields:    nil,
	}
	return zcfg.Build()
}

// sinkFilePath 按zap打开输出的规则取出文件路径，非文件类输出返回空
func sinkFilePath(outpath string) string {
	if filepath.IsAbs(outpath) {
		return outpath
	}
	var u, err = url.Parse(outpath)
	if err != nil || (u.Scheme != "" && u.Scheme != "file") {
		return ""
	}
	return u.Path
}
