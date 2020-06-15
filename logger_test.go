package logger

import (
	"testing"

	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	logger, err := New("./logs.jsonss")
	if err != nil {
		t.Fatalf("init logger failed: %v", err)
	}

	logger.Debug("This is a DEBUG message")
	logger.Info("This is an INFO message")
	logger.Info("This is an INFO message with fields", zap.String("region", "us-west"), zap.Int("id", 2))
	logger.Warn("This is a WARN message")
	logger.Error("This is an ERROR message")
	logger.Fatal("This is an Fatal")
	logger.Panic("This is an Panic")
}
