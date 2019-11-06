package logger_test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go/logger"
)

func TestLogger_ExampleCreateLoggerAndOutputSimpleMessages(t *testing.T) {
	t.Parallel()

	//the following instruction might be done inside a var declaration, once on each package
	// or in the init func of the package
	log := logger.GetOrCreate("test_logger1")
	//manual set of the log lev is required here for demonstration purposes
	log.SetLevel(logger.LogTrace)

	log.Trace("a trace message")
	log.Debug("a debug message")
	log.Info("an information message")
	log.Warn("a warning message")
	log.Error("an error message")
}

func TestLogger_ExampleMessagesWithArguments(t *testing.T) {
	t.Parallel()

	log := logger.GetOrCreate("test_logger2")
	log.SetLevel(logger.LogInfo)

	log.Info("message1", "an-int", 45, "a-string", "string")
	log.Info("message2", "a-map", map[string]int{"key1": 0, "key2": 1})
	log.Info("message3", "a-slice", []int{1, 2, 3, 4, 5})
	log.Info("message4", "nil", nil)
	hash := generateHash()
	log.Info("message5", "short-hash", logger.ConvertHash(hash), "long-hash", hex.EncodeToString(hash))
}

func generateHash() []byte {
	buff := make([]byte, 32)
	_, _ = rand.Reader.Read(buff)
	return buff
}
