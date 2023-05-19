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
