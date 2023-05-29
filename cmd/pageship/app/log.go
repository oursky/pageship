package app

import "github.com/fatih/color"

func Debug(format string, args ...any) {
	if debugMode {
		color.HiBlack(" DEBUG   "+format, args...)
	}
}

func Info(format string, args ...any) {
	color.White("  INFO   "+format, args...)
}

func Error(format string, args ...any) {
	color.Red(" ERROR   "+format, args...)
}

type logger struct{}

func (logger) Debug(format string, args ...any) {
	Debug(format, args...)
}

func (logger) Info(format string, args ...any) {
	Info(format, args...)
}

func (logger) Error(msg string, err error) {
	if err == nil {
		Error(msg)
	} else {
		Error(msg+": %s", err)
	}
}
