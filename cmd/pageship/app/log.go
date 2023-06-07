package app

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ErrCancelled = errors.New("cancelled")

func Debug(format string, args ...any) {
	if debugMode {
		color.HiBlack(" DEBUG   "+format, args...)
	}
}

func Info(format string, args ...any) {
	color.White("  INFO   "+format, args...)
}

func Warn(format string, args ...any) {
	color.Yellow("  WARN   "+format, args...)
}

func Error(format string, args ...any) {
	color.Red(" ERROR   "+format, args...)
}

var zapLogger = zap.New(&zapCore{level: zapcore.DebugLevel})

type zapCore struct {
	level  zapcore.Level
	fields []zapcore.Field
}

func (c *zapCore) Enabled(l zapcore.Level) bool {
	return c.level.Enabled(l)
}

func (c *zapCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *zapCore) Sync() error { return nil }

func (c *zapCore) With(fields []zapcore.Field) zapcore.Core {
	return &zapCore{level: c.level, fields: append(c.fields, fields...)}
}

func (c *zapCore) Write(e zapcore.Entry, fields []zapcore.Field) error {
	fields = append(c.fields, fields...)

	var log func(format string, args ...any)
	switch {
	case e.Level >= zapcore.ErrorLevel:
		log = Error
	case e.Level >= zapcore.WarnLevel:
		log = Warn
	case e.Level >= zapcore.InfoLevel:
		log = Info
	default:
		log = Debug
	}

	var buf strings.Builder
	if e.LoggerName != "" {
		buf.WriteString(e.LoggerName)
		buf.WriteString(": ")
	}
	buf.WriteString(e.Message)

	if len(fields) > 0 {
		buf.WriteString("   ")
		enc := zapcore.NewMapObjectEncoder()
		for _, f := range fields {
			f.AddTo(enc)
		}
		var writeFields func(key string, m map[string]any)
		writeFields = func(key string, m map[string]any) {
			for k, v := range m {
				if mm, ok := v.(map[string]any); ok {
					writeFields(key+k+".", mm)
					continue
				}
				buf.WriteRune(' ')
				buf.WriteString(key + k)
				buf.WriteString("=")
				buf.WriteString(fmt.Sprint(v))
			}
		}
		writeFields("", enc.Fields)
	}

	log("%s", buf.String())
	return nil
}
