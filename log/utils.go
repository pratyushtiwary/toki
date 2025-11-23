package log

import "github.com/fatih/color"

func Log(format string, args ...interface{}) {
	color.Blue(format, args...)
}

func Error(format string, args ...interface{}) {
	color.Red(format, args...)
}

func Info(format string, args ...interface{}) {
	color.Green(format, args...)
}

func Warn(format string, args ...interface{}) {
	color.Yellow(format, args...)
}
