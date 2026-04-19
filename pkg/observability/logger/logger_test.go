package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildOutputCreatesDefaultLogPath(t *testing.T) {
	tempDir := t.TempDir()
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(currentDir)
	})

	output, logPath := buildOutput(10)
	if output == nil {
		t.Fatal("expected output writer")
	}
	if logPath != "logs/app.log" {
		t.Fatalf("expected default log path logs/app.log, got %q", logPath)
	}
}

func TestNormalizeMaxBytesUsesDefaultForNonPositiveValues(t *testing.T) {
	if got := normalizeMaxBytes(0); got != defaultMaxLogBytes {
		t.Fatalf("expected default bytes for zero, got %d", got)
	}
	if got := normalizeMaxBytes(-1); got != defaultMaxLogBytes {
		t.Fatalf("expected default bytes for negative, got %d", got)
	}
	if got := normalizeMaxBytes(5); got != 5*1024*1024 {
		t.Fatalf("expected 5MB in bytes, got %d", got)
	}
}

func TestRotatingFileWriterRotatesWhenFileGetsTooLarge(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "logs", "app.log")

	writer, err := newRotatingFileWriter(logPath, 10)
	if err != nil {
		t.Fatalf("new rotating writer: %v", err)
	}

	if _, err := writer.Write(bytes.Repeat([]byte("a"), 8)); err != nil {
		t.Fatalf("write first chunk: %v", err)
	}
	if _, err := writer.Write(bytes.Repeat([]byte("b"), 8)); err != nil {
		t.Fatalf("write second chunk: %v", err)
	}

	current, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read current log: %v", err)
	}
	if string(current) != "bbbbbbbb" {
		t.Fatalf("expected current log to contain rotated write, got %q", string(current))
	}

	backup, err := os.ReadFile(logPath + ".1")
	if err != nil {
		t.Fatalf("read backup log: %v", err)
	}
	if string(backup) != "aaaaaaaa" {
		t.Fatalf("expected backup log to contain original write, got %q", string(backup))
	}
}
