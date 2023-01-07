package zerolog

import (
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
	"github.com/rs/zerolog"
)

// Wrapper is a type that wraps a zerolog Logger implementing the hclog.Logger
// interface.
type Wrapper struct {
	logger zerolog.Logger
	name   string
}

// Wrap wraps a zerolog Logger and returns a wrapper that implements the
// hclog.Logger interface.
func Wrap(logger zerolog.Logger) hclog.Logger {
	return Wrapper{
		logger: logger,
		name:   "",
	}
}

func (w Wrapper) Log(level hclog.Level, msg string, args ...interface{}) {
	switch level {
	case hclog.NoLevel, hclog.Trace:
		w.Trace(msg, args...)
	case hclog.Debug:
		w.Debug(msg, args...)
	case hclog.Info:
		w.Info(msg, args...)
	case hclog.Warn:
		w.Warn(msg, args...)
	case hclog.Error:
		w.Error(msg, args...)
	}
}

func (w Wrapper) Trace(msg string, args ...interface{}) {
	event := w.logger.Trace().Fields(args)
	if w.name != "" {
		event.Str("logger", w.name)
	}
	event.Msg(msg)
}

func (w Wrapper) Debug(msg string, args ...interface{}) {
	event := w.logger.Debug().Fields(args)
	if w.name != "" {
		event.Str("logger", w.name)
	}
	event.Msg(msg)
}

func (w Wrapper) Info(msg string, args ...interface{}) {
	event := w.logger.Info().Fields(args)
	if w.name != "" {
		event.Str("logger", w.name)
	}
	event.Msg(msg)
}

func (w Wrapper) Warn(msg string, args ...interface{}) {
	event := w.logger.Warn().Fields(args)
	if w.name != "" {
		event.Str("logger", w.name)
	}
	event.Msg(msg)
}

func (w Wrapper) Error(msg string, args ...interface{}) {
	event := w.logger.Error().Fields(args)
	if w.name != "" {
		event.Str("logger", w.name)
	}
	event.Msg(msg)
}

func (w Wrapper) IsTrace() bool {
	return w.logger.GetLevel() == zerolog.TraceLevel
}

func (w Wrapper) IsDebug() bool {
	return w.logger.GetLevel() == zerolog.DebugLevel
}

func (w Wrapper) IsInfo() bool {
	return w.logger.GetLevel() == zerolog.InfoLevel
}

func (w Wrapper) IsWarn() bool {
	return w.logger.GetLevel() == zerolog.WarnLevel
}

func (w Wrapper) IsError() bool {
	return w.logger.GetLevel() == zerolog.ErrorLevel
}

func (w Wrapper) ImpliedArgs() []interface{} {
	return []interface{}{}
}

func (w Wrapper) With(args ...interface{}) hclog.Logger {
	return Wrapper{
		logger: w.logger.With().Fields(args).Logger(),
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
		logger: w.logger,
		name:   newName,
	}
}

func (w Wrapper) ResetNamed(name string) hclog.Logger {
	return Wrapper{
		logger: w.logger,
		name:   name,
	}
}

func (w Wrapper) SetLevel(level hclog.Level) {
	// nop
}

func (w Wrapper) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(w.StandardWriter(opts), "", log.LstdFlags)
}

func (w Wrapper) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return hclog.DefaultOutput
}
