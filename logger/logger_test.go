package logger_test

import (
	"elrond/elrond-go-sandbox/logger"
	"testing"
)

func TestInvalidArgument(t *testing.T) {
	logger.InvalidArgument("test", "4234242342342")
}
