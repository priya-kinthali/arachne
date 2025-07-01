package main

import (
	"log"
	"os"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging functionality
type Logger struct {
	level LogLevel
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

// NewLogger creates a new logger with the specified level
func NewLogger(level string) *Logger {
	logLevel := parseLogLevel(level)

	flags := log.Ldate | log.Ltime | log.Lmicroseconds

	return &Logger{
		level: logLevel,
		debug: log.New(os.Stdout, "🔍 DEBUG ", flags),
		info:  log.New(os.Stdout, "ℹ️  INFO  ", flags),
		warn:  log.New(os.Stderr, "⚠️  WARN  ", flags),
		error: log.New(os.Stderr, "❌ ERROR ", flags),
	}
}

// parseLogLevel converts string to LogLevel
func parseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.debug.Printf(format, v...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= INFO {
		l.info.Printf(format, v...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= WARN {
		l.warn.Printf(format, v...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.error.Printf(format, v...)
	}
}

// LogRequest logs HTTP request details
func (l *Logger) LogRequest(method, url string, start time.Time) {
	duration := time.Since(start)
	l.Debug("HTTP %s %s completed in %v", method, url, duration)
}

// LogRetry logs retry attempts
func (l *Logger) LogRetry(url string, attempt int, err error) {
	l.Warn("Retry %d for %s: %v", attempt, url, err)
}

// LogSuccess logs successful scraping
func (l *Logger) LogSuccess(url string, status int, size int, duration time.Duration) {
	l.Info("✅ Scraped %s (Status: %d, Size: %d bytes, Duration: %v)", url, status, size, duration)
}

// LogFailure logs failed scraping
func (l *Logger) LogFailure(url string, err error) {
	l.Error("❌ Failed to scrape %s: %v", url, err)
}
