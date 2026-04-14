package logger

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	Init("debug")
	if level != LevelDebug {
		t.Errorf("expected LevelDebug, got %d", level)
	}

	Init("info")
	if level != LevelInfo {
		t.Errorf("expected LevelInfo, got %d", level)
	}

	Init("warn")
	if level != LevelWarn {
		t.Errorf("expected LevelWarn, got %d", level)
	}

	Init("error")
	if level != LevelError {
		t.Errorf("expected LevelError, got %d", level)
	}

	Init("unknown")
	if level != LevelInfo {
		t.Errorf("expected LevelInfo for unknown, got %d", level)
	}

	Init("DEBUG")
	if level != LevelDebug {
		t.Errorf("expected LevelDebug for uppercase, got %d", level)
	}
}

func TestDebugf_AtDebugLevel(t *testing.T) {
	Init("debug")
	var buf bytes.Buffer
	debugL = log.New(&buf, "[DEBUG] ", 0)

	Debugf("hello %s", "world")
	if !strings.Contains(buf.String(), "hello world") {
		t.Errorf("expected 'hello world' in output, got %s", buf.String())
	}
}

func TestDebugf_AtInfoLevel(t *testing.T) {
	Init("info")
	var buf bytes.Buffer
	debugL = log.New(&buf, "[DEBUG] ", 0)

	Debugf("should not appear")
	if buf.Len() != 0 {
		t.Errorf("expected no output at info level, got %s", buf.String())
	}
}

func TestInfof(t *testing.T) {
	Init("info")
	var buf bytes.Buffer
	infoL = log.New(&buf, "[INFO] ", 0)

	Infof("info msg %d", 42)
	if !strings.Contains(buf.String(), "info msg 42") {
		t.Errorf("expected 'info msg 42', got %s", buf.String())
	}
}

func TestWarnf(t *testing.T) {
	Init("warn")
	var buf bytes.Buffer
	warnL = log.New(&buf, "[WARN] ", 0)

	Warnf("warn msg")
	if !strings.Contains(buf.String(), "warn msg") {
		t.Errorf("expected 'warn msg', got %s", buf.String())
	}
}

func TestErrorf(t *testing.T) {
	Init("error")
	var buf bytes.Buffer
	errorL = log.New(&buf, "[ERROR] ", 0)

	Errorf("error msg")
	if !strings.Contains(buf.String(), "error msg") {
		t.Errorf("expected 'error msg', got %s", buf.String())
	}
}

func TestWarnf_Suppressed(t *testing.T) {
	Init("error")
	var buf bytes.Buffer
	warnL = log.New(&buf, "[WARN] ", 0)

	Warnf("should not appear")
	if buf.Len() != 0 {
		t.Errorf("expected no output, got %s", buf.String())
	}
}
