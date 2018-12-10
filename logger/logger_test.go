package logger_test

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"gotest.tools/assert"
)

func TestDebug(t *testing.T) {
	t.Parallel()
	var str bytes.Buffer
	log := logger.NewElrondLogger()
	log.SetLevel(logger.LogDebug)
	log.SetOutput(&str)
	log.Debug("abc")
	logString := str.String()
	assert.Assert(t, strings.Contains(logString, `"level":"debug"`))
	assert.Assert(t, strings.Contains(logString, `"msg":"abc"`))
}

func TestInfo(t *testing.T) {
	t.Parallel()
	var str bytes.Buffer
	log := logger.NewElrondLogger()
	log.SetLevel(logger.LogDebug)
	log.SetOutput(&str)
	log.Info("abc")
	logString := str.String()
	assert.Assert(t, strings.Contains(logString, `"level":"info"`))
	assert.Assert(t, strings.Contains(logString, `"msg":"abc"`))
}

func TestWarn(t *testing.T) {
	t.Parallel()
	var str bytes.Buffer
	log := logger.NewElrondLogger()
	log.SetOutput(&str)
	log.Warn("abc")
	logString := str.String()
	assert.Assert(t, strings.Contains(logString, `"level":"warning"`))
	assert.Assert(t, strings.Contains(logString, `"msg":"abc"`))
}

func TestError(t *testing.T) {
	t.Parallel()
	var str bytes.Buffer
	log := logger.NewElrondLogger()
	log.SetOutput(&str)
	log.Error("abc")
	logString := str.String()
	assert.Assert(t, strings.Contains(logString, `"level":"error"`))
	assert.Assert(t, strings.Contains(logString, `"msg":"abc"`))
}

func TestPanic(t *testing.T) {
	t.Parallel()
	var str bytes.Buffer
	log := logger.NewElrondLogger()
	log.SetOutput(&str)
	swallowPanicLog(t, "abc", "TestPanic should have panic", log)

	logString := str.String()
	assert.Assert(t, strings.Contains(logString, `"level":"panic"`))
	assert.Assert(t, strings.Contains(logString, `"msg":"abc"`))
}

func TestSetLevel(t *testing.T) {
	t.Parallel()
	var str bytes.Buffer
	log := logger.NewElrondLogger()
	log.SetOutput(&str)
	log.SetLevel(logger.LogDebug)
	log.Debug("abc")
	assert.Assert(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()

	log.SetLevel(logger.LogInfo)
	log.Debug("abc")
	assert.Assert(t, len(str.String()) == 0)
	str.Reset()
	log.Info("abc")
	assert.Assert(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()

	log.SetLevel(logger.LogWarning)
	log.Info("abc")
	assert.Assert(t, len(str.String()) == 0)
	str.Reset()
	log.Warn("abc")
	assert.Assert(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()

	log.SetLevel(logger.LogError)
	log.Warn("abc")
	assert.Assert(t, len(str.String()) == 0)
	str.Reset()
	log.Error("abc")
	assert.Assert(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()

	log.SetLevel(logger.LogPanic)
	log.Error("abc")
	assert.Assert(t, len(str.String()) == 0)
	str.Reset()

	swallowPanicLog(t, "abc", "TestSetLevel should have panic", log)
	assert.Assert(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()

	log.SetLevel("this should go on the default case")
	log.Warn("abc")
	assert.Assert(t, len(str.String()) == 0)
	str.Reset()
	log.Error("abc")
	assert.Assert(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()
}

func TestWithFile(t *testing.T) {
	t.Parallel()
	log := logger.NewDefaultLogger()
	log.Warn("This test should pass if the file was opened in the correct mode")
}

func TestConcurrencyWithFileWriter(t *testing.T) {
	t.Parallel()
	log := logger.NewDefaultLogger()

	wg := sync.WaitGroup{}
	wg.Add(999)

	for i := 1; i < 1000; i++ {
		go func(index int) {
			log.Warn("I will error if I'll run into concurrency issues", index)
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestWithOptions(t *testing.T) {
	var str bytes.Buffer
	log := logger.NewElrondLogger(logger.WithStackTraceDepth(1), logger.WithFile(&str))
	assert.Equal(t, log.StackTraceDepth(), 1, "WithStackTraceDepth does not set the correct option")
	assert.Equal(t, log.File(), &str)
}

func swallowPanicLog(t *testing.T, logMsg string, panicMsg string, log *logger.Logger) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf(panicMsg)
		}
	}()
	log.Panic(logMsg)
}
