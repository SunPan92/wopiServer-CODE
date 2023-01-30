package middleware

import (
	"wopi-server/config"
	"wopi-server/g"

	"github.com/kataras/iris/v12"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	HEADER_CONTENT_TYPE = "Content-Type"
	CONTENT_TYPE_JSON   = "application/json"
	CONTENT_TYPE_FORM   = "application/x-www-form-urlencoded"
	DURATION_NS         = 1
	DURATION_US         = 1000 * DURATION_NS
	DURATION_MS         = 1000 * DURATION_US
	DURATION_S          = 1000 * DURATION_MS
	DURATION_MINU       = 60 * DURATION_S
	DURATION_HOUR       = 60 * DURATION_MINU
	DURATION_DAY        = 24 * DURATION_HOUR
)

// InitLogger 初始化日志配置
// logPath 日志文件路径
// logLevel 日志级别 debug/info/warn/error
// maxSize 单个文件大小,MB
// maxBackups 保存的文件个数
// maxAge 保存的天数， 没有的话不删除
// compress 压缩
// jsonFormat 是否输出为json格式
// showLine 显示代码行
// logInConsole 是否同时输出到控制台
func InitLogger(cfg *config.LogConfig) {
	hook := lumberjack.Logger{
		Filename:   cfg.LogPath,    // 日志文件路径
		MaxSize:    cfg.MaxSize,    // MB
		MaxBackups: cfg.MaxBackups, // 最多保留30个备份
		Compress:   cfg.Compress,   // 是否压缩 disabled by default
	}
	if cfg.MaxAge > 0 {
		hook.MaxAge = cfg.MaxAge // days
	}
	var syncer zapcore.WriteSyncer
	if cfg.LogInConsole {
		syncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook))
	} else {
		syncer = zapcore.AddSync(&hook)
	}
	// encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "Log",
		CallerKey:      "line",
		MessageKey:     "msg",
		StacktraceKey:  "trace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,                          // 小写编码器
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"), // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,                         //
		EncodeCaller:   zapcore.ShortCallerEncoder,                             // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	var encoder zapcore.Encoder
	if cfg.JsonFormat {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 设置日志级别,debug可以打印出info,debug,warn；info级别可以打印warn，info；warn只能打印warn
	// debug->info->warn->error
	var level zapcore.Level
	switch cfg.LogLevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	core := zapcore.NewCore(
		encoder,
		syncer,
		level,
	)

	g.Log = zap.New(core, zap.Development(), zap.AddStacktrace(zap.ErrorLevel))
	if cfg.ShowLine {
		g.Log = g.Log.WithOptions(zap.AddCaller())
	}
	//使用zap替换全局日志
	zap.ReplaceGlobals(g.Log)
}

// IrisLogger 接收iris框架默认的日志
func IrisLogger() iris.Handler {
	return func(ctx iris.Context) {
		start := time.Now()
		path := ctx.Request().URL.Path
		query := ctx.Request().URL.Query()
		body := getRequestBody(ctx.Request())
		g.Log.Info(
			fmt.Sprintf("%d, %v, %s", ctx.GetStatusCode(), ctx.Request().Method, path),
			zap.Any("query", query),
			zap.String("body", body),
			zap.String("ip", ctx.RemoteAddr()),
		)
		ctx.Next()
		cost := time.Since(start)
		costNs := cost.Nanoseconds()
		var duration string
		var value float64
		if costNs < DURATION_US {
			duration = "纳秒"
			value = float64(costNs)
		} else if costNs < DURATION_MS {
			duration = "微妙"
			value = float64(costNs/DURATION_NS) / 1000
		} else if costNs < DURATION_S {
			duration = "毫秒"
			value = float64(costNs/DURATION_US) / 1000
		} else {
			duration = "秒"
			value = float64(costNs/DURATION_MS) / 1000
		}
		g.Log.Info(
			fmt.Sprintf("%d, %v, %s, cost(%s): %v", ctx.GetStatusCode(), ctx.Request().Method, path, duration, value),
			//zap.String("query", query),
			//zap.String("body", body),
			//zap.String("ip", ctx.RemoteAddr()),
		)
	}
}

//打印错误日志
func logError(stack bool, path string, request string, err interface{}, stackMsg string) {
	if stack {
		zap.L().Error("错误信息:\n",
			zap.String("path", path),
			zap.String("request", request),
			zap.Any("error", err),
			zap.String("stack", stackMsg),
		)
	} else {
		zap.L().Error("错误信息:\n",
			zap.String("path", path),
			zap.String("request", request),
			zap.Any("error", err),
		)
	}
}

// format stack info
func getStackErrorMsg() string {
	stacktrace := ""
	for i := 1; ; i++ {
		_, f, l, got := runtime.Caller(i)
		if !got {
			break
		}
		stacktrace += fmt.Sprintf("%s:%d\n", f, l)
	}
	return stacktrace
}

// get the body of request
func getRequestBody(r *http.Request) string {
	// get the body of request when content type is json or form
	body := ""
	contentType := r.Header.Get(HEADER_CONTENT_TYPE)
	if strings.Contains(contentType, CONTENT_TYPE_JSON) || strings.Contains(contentType, CONTENT_TYPE_FORM) {
		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			g.Log.Error("read request body error", zap.Error(err))
		} else {
			body = string(buf)
		}
		r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	}
	return body
}

func ErrorString(err interface{}) string {
	if err == nil {
		return ""
	} else {
		errI := zap.Any("error", err).Interface
		return fmt.Sprintf("%v", errI)
	}
}
