package publisher

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is used for publisher to print log.
type Logger interface {
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})

	Info(args ...interface{})
	Infof(template string, args ...interface{})

	Warn(args ...interface{})
	Warnf(template string, args ...interface{})

	Error(args ...interface{})
	Errorf(template string, args ...interface{})

	Panic(args ...interface{})
	Panicf(template string, args ...interface{})
}

var logger Logger = newZapLogger()

// SetLogger set an external logger.
func SetLogger(l Logger) {
	logger = l
}

func newZapLogger() *zap.SugaredLogger {
	var cfg *zap.Config
	if os.Getenv("DEBUG") == "1" {
		cfg = &zap.Config{
			Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
			Development: true,
			Encoding:    "console",
			EncoderConfig: zapcore.EncoderConfig{
				MessageKey:       "M",
				LevelKey:         "L",
				TimeKey:          "T",
				NameKey:          "N",
				CallerKey:        "C",
				FunctionKey:      zapcore.OmitKey,
				StacktraceKey:    "S",
				LineEnding:       zapcore.DefaultLineEnding,
				EncodeLevel:      zapcore.CapitalLevelEncoder,
				EncodeTime:       zapcore.ISO8601TimeEncoder,
				EncodeDuration:   zapcore.StringDurationEncoder,
				EncodeCaller:     zapcore.ShortCallerEncoder,
				EncodeName:       zapcore.FullNameEncoder,
				ConsoleSeparator: "  ",
			},
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}
	} else {
		cfg = &zap.Config{
			Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
			Development: false,
			Sampling: &zap.SamplingConfig{
				Initial:    100,
				Thereafter: 100,
			},
			Encoding: "console",
			EncoderConfig: zapcore.EncoderConfig{
				MessageKey:       "M",
				LevelKey:         "L",
				TimeKey:          "T",
				NameKey:          "N",
				CallerKey:        "C",
				FunctionKey:      zapcore.OmitKey,
				StacktraceKey:    "S",
				LineEnding:       zapcore.DefaultLineEnding,
				EncodeLevel:      zapcore.CapitalLevelEncoder,
				EncodeTime:       zapcore.ISO8601TimeEncoder,
				EncodeDuration:   zapcore.StringDurationEncoder,
				EncodeCaller:     zapcore.ShortCallerEncoder,
				EncodeName:       zapcore.FullNameEncoder,
				ConsoleSeparator: "  ",
			},
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}
	}
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return l.Sugar()
}
