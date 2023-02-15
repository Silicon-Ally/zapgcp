// Package zapgcp is a simple package for configuring a zap (go.uber.org/zap)
// logger for use with Google Cloud Platform's Cloud Logging infrastructure.
// The production config is based on this example:
// https://github.com/uber-go/zap/issues/1095#issuecomment-1149455643
package zapgcp

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	// Local determines whether to use a local development config, which is tuned
	// for readability in a console, or the production config, which is meant to be
	// consumed by GCP Cloud Logging.
	Local bool

	// MinLogLevel sets the lowest level to actually output logs for.
	MinLogLevel zapcore.Level

	Options []zap.Option
}

func (cfg *Config) ToZapConfig() (zap.Config, []zap.Option) {
	if cfg.Local {
		zCfg := zap.NewDevelopmentConfig()
		zCfg.Level = zap.NewAtomicLevelAt(cfg.MinLogLevel)
		return zCfg, cfg.Options
	}

	zCfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(cfg.MinLogLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "severity",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    encodeLevel,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.MillisDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	opts := []zap.Option{
		zap.AddStacktrace(zap.DPanicLevel),
	}
	opts = append(opts, cfg.Options...)
	return zCfg, opts
}

// New returns a zap Logger configured for the environment specified in `cfg`.
func New(cfg *Config) (*zap.Logger, error) {
	zCfg, opts := cfg.ToZapConfig()
	return zCfg.Build(opts...)
}

func encodeLevel(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendString("DEBUG")
	case zapcore.InfoLevel:
		enc.AppendString("INFO")
	case zapcore.WarnLevel:
		enc.AppendString("WARNING")
	case zapcore.ErrorLevel:
		enc.AppendString("ERROR")
	case zapcore.DPanicLevel:
		enc.AppendString("CRITICAL")
	case zapcore.PanicLevel:
		enc.AppendString("ALERT")
	case zapcore.FatalLevel:
		enc.AppendString("EMERGENCY")
	}
}
