package logger

import (
	"testing"
)

func TestInit(t *testing.T) {
	err := Init(false)
	if err != nil {
		t.Errorf("Init failed: %v", err)
	}
	if log == nil {
		t.Error("Logger should be initialized")
	}
}

func TestInit_Debug(t *testing.T) {
	err := Init(true)
	if err != nil {
		t.Errorf("Init failed: %v", err)
	}
}

func TestDebug(t *testing.T) {
	Init(false)
	Debug("test message")
}

func TestInfo(t *testing.T) {
	Init(false)
	Info("test info message")
}

func TestWarn(t *testing.T) {
	Init(false)
	Warn("test warning message")
}

func TestError(t *testing.T) {
	Init(false)
	Error("test error message")
}

func TestWith(t *testing.T) {
	Init(false)
	l := With()
	if l == nil {
		t.Error("With should return a logger")
	}
}

func TestSync(t *testing.T) {
	Init(false)
	Sync()
}