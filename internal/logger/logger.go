package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"go-web-bot/internal/ctime"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

var current = InfoLevel
var std = log.New(os.Stdout, "", 0)

func Init(level string, toFile bool, file string) {
	current = parse(level)
	if toFile {
		_ = os.MkdirAll(filepath.Dir(file), 0o755)
		f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err == nil {
			std.SetOutput(io.MultiWriter(os.Stdout, f))
		}
	}
}
func parse(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "warning", "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}
func logf(l Level, name, format string, args ...any) {
	if l < current {
		return
	}
	std.Printf("%s [%s] %s", ctime.FormatNow(), name, fmt.Sprintf(format, args...))
}
func Debug(format string, args ...any)   { logf(DebugLevel, "DEBUG", format, args...) }
func Info(format string, args ...any)    { logf(InfoLevel, "INFO", format, args...) }
func Warning(format string, args ...any) { logf(WarnLevel, "WARN", format, args...) }
func Error(format string, args ...any)   { logf(ErrorLevel, "ERROR", format, args...) }
func Fatal(format string, args ...any)   { logf(ErrorLevel, "FATAL", format, args...); os.Exit(1) }
