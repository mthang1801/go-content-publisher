package logger

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	defaultOutput  io.Writer = os.Stdout
	defaultLogFile string
)

const defaultMaxLogBytes int64 = 10 * 1024 * 1024

func New(level string, maxSizeMB int) *slog.Logger {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	defaultOutput, defaultLogFile = buildOutput(maxSizeMB)
	log.SetOutput(defaultOutput)

	return slog.New(slog.NewTextHandler(defaultOutput, &slog.HandlerOptions{Level: logLevel}))
}

func Output() io.Writer {
	return defaultOutput
}

func LogFilePath() string {
	return defaultLogFile
}

func buildOutput(maxSizeMB int) (io.Writer, string) {
	logPath := filepath.Join("logs", "app.log")
	writer, err := newRotatingFileWriter(logPath, normalizeMaxBytes(maxSizeMB))
	if err != nil {
		return os.Stdout, ""
	}
	return io.MultiWriter(os.Stdout, writer), logPath
}

func normalizeMaxBytes(maxSizeMB int) int64 {
	if maxSizeMB <= 0 {
		return defaultMaxLogBytes
	}
	return int64(maxSizeMB) * 1024 * 1024
}

type rotatingFileWriter struct {
	mu       sync.Mutex
	path     string
	maxBytes int64
	file     *os.File
	size     int64
}

func newRotatingFileWriter(path string, maxBytes int64) (*rotatingFileWriter, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("maxBytes must be positive")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}

	return &rotatingFileWriter{
		path:     path,
		maxBytes: maxBytes,
		file:     file,
		size:     info.Size(),
	}, nil
}

func (w *rotatingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return 0, os.ErrClosed
	}

	if w.size > 0 && w.size+int64(len(p)) > w.maxBytes {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

func (w *rotatingFileWriter) rotate() error {
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return err
		}
		w.file = nil
	}

	backupPath := w.path + ".1"
	_ = os.Remove(backupPath)
	if err := os.Rename(w.path, backupPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	file, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	w.file = file
	w.size = 0
	return nil
}
