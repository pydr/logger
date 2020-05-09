package logger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type (
	logConfig struct {
		Desc      string
		Level     string
		Stdout    bool
		Encoding  string
		AddCaller bool
		Color     bool
		FilesOut  bool
		LogsPath  []*logFilePath
	}

	logFilePath struct {
		Level string
		Hook  *lumberjack.Logger
	}
)

var Logger *zap.Logger
var err error

func init() {
	Logger, err = New("./logs.json")
	if err != nil {
		fmt.Print("init logger failed: miss default config file 'logs.json' in current program root directory， or use New() to manual initialization")
		os.Exit(1)
	}
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func New(configPath string) (*zap.Logger, error) {
	logcfg := logConfig{
		Desc:      "development",
		Level:     "error",
		Stdout:    true,
		Encoding:  "console",
		AddCaller: true,
		Color:     true,
		FilesOut:  false,
		LogsPath: []*logFilePath{{
			Level: "error",
			Hook: &lumberjack.Logger{
				Filename:   "./logs/zlog.log", // Filename is the file to write logs to.
				MaxSize:    1024,              // megabytes
				MaxAge:     7,                 // days
				MaxBackups: 3,                 // the maximum number of old log files to retain.
			},
		},
		},
	}

	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(file, &logcfg); err != nil {
		return nil, err
	}

	// Output should also go to standard out.
	consoleDebugging := zapcore.Lock(os.Stdout)

	var encoderConfig zapcore.EncoderConfig
	var fileEncoder, consoleEncoder zapcore.Encoder
	if strings.EqualFold(logcfg.Desc, "production") {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else if strings.EqualFold(logcfg.Desc, "development") {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		return nil, errors.Errorf("'%s' in the configuration file for desc is an invalid value, it could be 'development' or 'production'", logcfg.Desc)
	}

	encoderConfig.EncodeTime = TimeEncoder
	if strings.EqualFold(logcfg.Encoding, "json") {
		fileEncoder = zapcore.NewJSONEncoder(encoderConfig)
		consoleEncoder = zapcore.NewJSONEncoder(encoderConfig)
	} else if strings.EqualFold(logcfg.Encoding, "console") {
		fileEncoder = zapcore.NewConsoleEncoder(encoderConfig)
		if logcfg.Color {
			encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		consoleEncoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		return nil, errors.Errorf("'%s' in the configuration file for desc is an invalid value, it could be 'development' or 'production'", logcfg.Encoding)
	}

	var cores []zapcore.Core
	if !logcfg.Stdout && !logcfg.FilesOut || logcfg.Stdout {
		cores = append(cores, zapcore.NewCore(consoleEncoder, consoleDebugging, getLevel(logcfg.Level)))
	}
	if logcfg.FilesOut {
		if len(logcfg.LogsPath) > 0 {
			for i := 0; i < len(logcfg.LogsPath); i++ {
				cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.AddSync(logcfg.LogsPath[i].Hook), getLevel(logcfg.LogsPath[i].Level)))
			}
		}
	}
	core := zapcore.NewTee(cores...)
	// From a zapcore.Core to construct a Logger.

	var logger *zap.Logger
	if logcfg.AddCaller {
		logger = zap.New(core, zap.AddCaller())
	} else {
		logger = zap.New(core)
	}

	return logger, nil
}

func getLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "panic", "dpanic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	case "error":
		return zapcore.ErrorLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "info":
		return zapcore.InfoLevel
	case "debug":
		return zapcore.DebugLevel
	default:
		return zapcore.DebugLevel
	}
}
