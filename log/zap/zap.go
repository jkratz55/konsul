package zap

import (
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Wrapper is a type that wraps a zap Logger and adapts it to a hclog.Logger
type Wrapper struct {
	logger *zap.Logger
	name   string
}

// Wrap accepts a zap Logger and wraps it to adapt to a hclog.Logger.
//
// A nil logger will cause a panic.
func Wrap(logger *zap.Logger) hclog.Logger {
	if logger == nil {
		panic("cannot wrap nil zap.Logger")
	}
	return Wrapper{
		logger: logger.WithOptions(zap.AddCallerSkip(1)),
		name:   "",
	}
}

func (w Wrapper) Log(level hclog.Level, msg string, args ...any) {
	switch level {
	// Zap doesn't have a Trace level so it gets mapped to Debug
	case hclog.NoLevel, hclog.Trace, hclog.Debug:
		w.Debug(msg, args...)
	case hclog.Info:
		w.Info(msg, args...)
	case hclog.Warn:
		w.Warn(msg, args...)
	case hclog.Error:
		w.Error(msg, args...)
	}
}

func (w Wrapper) Trace(msg string, args ...any) {
	// Zap doesn't have a Trace level, Debug is the closest level
	w.logger.Debug(msg, convertArgsToZapFields(args...)...)
}

func (w Wrapper) Debug(msg string, args ...any) {
	w.logger.Debug(msg, convertArgsToZapFields(args...)...)
}

func (w Wrapper) Info(msg string, args ...any) {
	w.logger.Info(msg, convertArgsToZapFields(args...)...)
}

func (w Wrapper) Warn(msg string, args ...any) {
	w.logger.Warn(msg, convertArgsToZapFields(args...)...)
}

func (w Wrapper) Error(msg string, args ...any) {
	w.logger.Error(msg, convertArgsToZapFields(args...)...)
}

func (w Wrapper) IsTrace() bool {
	// Zap doesn't have a trace level, always return false
	return false
}

func (w Wrapper) IsDebug() bool {
	return w.logger.Level() == zap.DebugLevel
}

func (w Wrapper) IsInfo() bool {
	return w.logger.Level() == zap.InfoLevel
}

func (w Wrapper) IsWarn() bool {
	return w.logger.Level() == zap.WarnLevel
}

func (w Wrapper) IsError() bool {
	return w.logger.Level() == zap.ErrorLevel
}

func (w Wrapper) ImpliedArgs() []any {
	w.logger.Warn("ImpliedArgs in not implemented... this will always return nil")
	return nil
}

func (w Wrapper) With(args ...any) hclog.Logger {
	return Wrapper{
		logger: w.logger.With(convertArgsToZapFields(args...)...),
		name:   w.name,
	}
}

func (w Wrapper) Name() string {
	return w.name
}

func (w Wrapper) Named(name string) hclog.Logger {
	var newName string
	if w.name != "" {
		newName = fmt.Sprintf("%s.%s", w.name, name)
	} else {
		newName = name
	}
	return Wrapper{
		logger: w.logger.Named(newName),
		name:   newName,
	}
}

func (w Wrapper) ResetNamed(name string) hclog.Logger {
	return Wrapper{
		logger: w.logger.Named(name),
		name:   name,
	}
}

func (w Wrapper) SetLevel(level hclog.Level) {
	w.logger.Warn("SetLevel on Wrapper is a no-op")
}

func (w Wrapper) GetLevel() hclog.Level {
	switch w.logger.Level() {
	case zap.DebugLevel:
		return hclog.Debug
	case zap.InfoLevel:
		return hclog.Info
	case zap.WarnLevel:
		return hclog.Warn
	case zap.ErrorLevel, zap.FatalLevel, zap.DPanicLevel, zap.PanicLevel:
		// hclog doesn't have concept of Panic and Fatal levels so it gets mapped
		// to error level
		return hclog.Error
	default:
		return hclog.NoLevel
	}
}

func (w Wrapper) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(w.StandardWriter(opts), "", log.LstdFlags)
}

func (w Wrapper) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return hclog.DefaultOutput
}

func convertArgsToZapFields(args ...any) []zapcore.Field {
	fields := make([]zapcore.Field, 0)
	for i := len(args); i > 0; i -= 2 {
		left := i - 2
		if left < 0 {
			left = 0
		}

		items := args[left:i]
		switch l := len(items); l {
		case 2:
			k, ok := items[0].(string)
			if ok {
				fields = append(fields, zap.Any(k, items[1]))
			} else {
				fields = append(fields, zap.Any(fmt.Sprintf("arg%d", i-1), items[1]))
				fields = append(fields, zap.Any(fmt.Sprintf("arg%d", left), items[0]))
			}
		case 1:
			fields = append(fields, zap.Any(fmt.Sprintf("arg%d", left), items[0]))
		}
	}

	return fields
}
