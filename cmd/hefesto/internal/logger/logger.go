// Package logger provides a simple leveled logger for Hefesto CLI.
//
// Output goes to stderr by default so it never interferes with TUI or
// command output on stdout. When the --verbose flag is set the level is
// lowered to Debug, otherwise Info is the minimum.
//
// If the logger is never initialized every call is a no-op, so existing
// tests continue to pass without modification.
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level represents a logging severity.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelNone // silences everything
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "?????"
	}
}

// -------------------------------------------------------------------
// Global logger (package-level singleton)
// -------------------------------------------------------------------

var (
	globalMu  sync.RWMutex
	globalLog *Logger
)

// Logger writes structured, leveled log messages.
type Logger struct {
	mu     sync.Mutex
	level  Level
	out    io.Writer
	file   *os.File
	closed bool
}

// Init creates the global logger.
//
//   - verbose=true  → Debug level
//   - verbose=false → Info level
//
// It also tries to open ~/.config/hefesto/hefesto.log for persistent
// debugging.  If the file cannot be opened the logger still works —
// it just writes to stderr.
func Init(verbose bool) {
	lvl := LevelInfo
	if verbose {
		lvl = LevelDebug
	}

	l := &Logger{
		level: lvl,
		out:   os.Stderr,
	}

	// Try to set up log file (best-effort).
	homeDir, err := os.UserHomeDir()
	if err == nil {
		logDir := filepath.Join(homeDir, ".config", "hefesto")
		_ = os.MkdirAll(logDir, 0755) //nolint:gosec // G301: log directory under user home, standard permissions acceptable
		logPath := filepath.Join(logDir, "hefesto.log")
		if f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil { //nolint:gosec // G304: logPath built from UserHomeDir, not user input
			l.file = f
			l.out = io.MultiWriter(os.Stderr, f)
		}
	}

	globalMu.Lock()
	// Close any previous logger.
	if globalLog != nil {
		globalLog.close()
	}
	globalLog = l
	globalMu.Unlock()
}

// Close closes the global logger and any open log file.
func Close() {
	globalMu.Lock()
	defer globalMu.Unlock()
	if globalLog != nil {
		globalLog.close()
		globalLog = nil
	}
}

// close is the unsynchronized version.
func (l *Logger) close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return
	}
	l.closed = true
	if l.file != nil {
		_ = l.file.Close()
		l.file = nil
	}
}

// -------------------------------------------------------------------
// Package-level convenience functions
// -------------------------------------------------------------------

func Debug(msg string, args ...any) {
	logf(LevelDebug, msg, args...)
}

func Info(msg string, args ...any) {
	logf(LevelInfo, msg, args...)
}

func Warn(msg string, args ...any) {
	logf(LevelWarn, msg, args...)
}

func Error(msg string, args ...any) {
	logf(LevelError, msg, args...)
}

// logf writes a log line at the given level. If the global logger has not
// been initialized, it silently returns — this keeps existing tests working.
func logf(lvl Level, msg string, args ...any) {
	globalMu.RLock()
	l := globalLog
	globalMu.RUnlock()

	if l == nil {
		return // no-op when not initialized
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if lvl < l.level || l.closed {
		return
	}

	ts := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] %s %s", lvl, ts, fmt.Sprintf(msg, args...))

	// Append optional key=value pairs would go here in the future.
	_, _ = fmt.Fprintln(l.out, line)
}
