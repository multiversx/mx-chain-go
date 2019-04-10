package logger_test

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/stretchr/testify/assert"
)

func TestDebug(t *testing.T) {
	t.Parallel()
	var str bytes.Buffer
	log := logger.NewElrondLogger()
	log.SetLevel(logger.LogDebug)
	log.SetOutput(&str)
	log.Debug("abc")
	logString := str.String()
	assert.True(t, strings.Contains(logString, `"level":"debug"`))
	assert.True(t, strings.Contains(logString, `"msg":"abc"`))
}

func TestInfo(t *testing.T) {
	t.Parallel()
	var str bytes.Buffer
	log := logger.NewElrondLogger()
	log.SetLevel(logger.LogDebug)
	log.SetOutput(&str)
	log.Info("abc")
	logString := str.String()
	assert.True(t, strings.Contains(logString, `"level":"info"`))
	assert.True(t, strings.Contains(logString, `"msg":"abc"`))
}

func TestWarn(t *testing.T) {
	t.Parallel()
	var str bytes.Buffer
	log := logger.NewElrondLogger()
	log.SetOutput(&str)
	log.Warn("abc")
	logString := str.String()
	assert.True(t, strings.Contains(logString, `"level":"warning"`))
	assert.True(t, strings.Contains(logString, `"msg":"abc"`))
}

func TestError(t *testing.T) {
	t.Parallel()
	var str bytes.Buffer
	log := logger.NewElrondLogger()
	log.SetOutput(&str)
	log.Error("abc")
	logString := str.String()
	assert.True(t, strings.Contains(logString, `"level":"error"`))
	assert.True(t, strings.Contains(logString, `"msg":"abc"`))
}

func TestPanic(t *testing.T) {
	t.Parallel()
	var str bytes.Buffer
	log := logger.NewElrondLogger()
	log.SetOutput(&str)
	swallowPanicLog(t, "abc", "TestPanic should have panic", log)

	logString := str.String()
	assert.True(t, strings.Contains(logString, `"level":"panic"`))
	assert.True(t, strings.Contains(logString, `"msg":"abc"`))
}

func TestSetLevel(t *testing.T) {
	t.Parallel()
	var str bytes.Buffer
	log := logger.NewElrondLogger()
	log.SetOutput(&str)
	log.SetLevel(logger.LogDebug)
	log.Debug("abc")
	assert.True(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()

	log.SetLevel(logger.LogInfo)
	log.Debug("abc")
	assert.True(t, len(str.String()) == 0)
	str.Reset()
	log.Info("abc")
	assert.True(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()

	log.SetLevel(logger.LogWarning)
	log.Info("abc")
	assert.True(t, len(str.String()) == 0)
	str.Reset()
	log.Warn("abc")
	assert.True(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()

	log.SetLevel(logger.LogError)
	log.Warn("abc")
	assert.True(t, len(str.String()) == 0)
	str.Reset()
	log.Error("abc")
	assert.True(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()

	log.SetLevel(logger.LogPanic)
	log.Error("abc")
	assert.True(t, len(str.String()) == 0)
	str.Reset()

	swallowPanicLog(t, "abc", "TestSetLevel should have panic", log)
	assert.True(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()

	log.SetLevel("this should go on the default case")
	log.Warn("abc")
	assert.True(t, len(str.String()) == 0)
	str.Reset()
	log.Error("abc")
	assert.True(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()
}

func TestWithFile(t *testing.T) {
	t.Parallel()
	log := logger.DefaultLogger()
	log.Warn("This test should pass if the file was opened in the correct mode")
}

func TestConcurrencyWithFileWriter(t *testing.T) {
	t.Parallel()
	log := logger.DefaultLogger()

	wg := sync.WaitGroup{}
	wg.Add(999)

	for i := 1; i < 1000; i++ {
		go func(index int) {
			log.Warn("I will error if I'll run into concurrency issues", index)
			wg.Done()
		}(i)
	}

	wg.Add(1)
	var str bytes.Buffer
	go func() {
		log.ApplyOptions(logger.WithFile(&str))
		wg.Done()
	}()

	wg.Wait()
}

func TestWithOptions(t *testing.T) {
	var str bytes.Buffer
	log := logger.NewElrondLogger(logger.WithStackTraceDepth(1), logger.WithFile(&str))
	assert.Equal(t, log.StackTraceDepth(), 1, "WithStackTraceDepth does not set the correct option")
	assert.Equal(t, log.File(), &str)
}

func TestLazyFileWriter_WriteSomeLinesBeforeProvidingFile(t *testing.T) {
	var str bytes.Buffer
	expectedString := "expectedString"
	log := logger.NewElrondLogger(logger.WithStackTraceDepth(1))
	log.Warn(expectedString)
	log.ApplyOptions(logger.WithFile(&str))
	log.Warn("this is a warning")
	assert.Contains(t, str.String(), expectedString)
}

func swallowPanicLog(t *testing.T, logMsg string, panicMsg string, log *logger.Logger) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf(panicMsg)
		}
	}()
	log.Panic(logMsg)
}
