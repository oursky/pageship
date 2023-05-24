package app

import "go.uber.org/zap"

type zapLogger struct{ *zap.Logger }

func (l zapLogger) Debug(format string, args ...any) {
	l.Logger.Sugar().Debugf(format, args...)
}

func (l zapLogger) Info(format string, args ...any) {
	l.Logger.Sugar().Infof(format, args...)
}

func (l zapLogger) Error(msg string, err error) {
	l.Logger.Error(msg, zap.Error(err))
}
