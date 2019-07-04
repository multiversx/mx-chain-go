package logger_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/logger"
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

func TestErrorWithoutFileRoll(t *testing.T) {
	t.Parallel()
	var str bytes.Buffer
	log := logger.NewElrondLogger()
	log.SetOutput(&str)
	log.ErrorWithoutFileRoll("abc")
	logString := str.String()
	assert.True(t, strings.Contains(logString, `"level":"error"`))
	assert.True(t, strings.Contains(logString, `"msg":"abc"`))
}

func TestLogIfError(t *testing.T) {
	t.Parallel()
	var str bytes.Buffer
	log := logger.NewElrondLogger()
	log.SetOutput(&str)
	log.LogIfError(errors.New("error"))
	logString := str.String()
	assert.True(t, strings.Contains(logString, `"level":"error"`))
	assert.True(t, strings.Contains(logString, `"msg":"error"`))
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
	log := logger.NewElrondLogger()
	log.Warn("This test should pass if the file was opened in the correct mode")
}

func TestConcurrencyWithFileWriter(t *testing.T) {
	t.Parallel()
	log := logger.NewElrondLogger()

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

func TestWithFileRotation(t *testing.T) {
	log := logger.NewElrondLogger()

	err := log.ApplyOptions(logger.WithFileRotation("", "logs", "log"))
	assert.Nil(t, err)

	file := log.File()
	assert.NotNil(t, file)
}

func TestWithStderrRedirect(t *testing.T) {
	log := logger.NewElrondLogger()

	err := log.ApplyOptions(logger.WithStderrRedirect())
	assert.Nil(t, err)
}

func TestHeadline(t *testing.T) {
	log := logger.NewElrondLogger()
	message := "One blockchain to rule them all"
	delimiter := "///"
	timestamp := "-"

	formatedMessage := log.Headline(message, delimiter, timestamp)
	assert.Contains(t, formatedMessage, message)
	assert.Contains(t, formatedMessage, delimiter)
	assert.Contains(t, formatedMessage, timestamp)
}

func TestRedirectStderr(t *testing.T) {
	file, _ := core.CreateFile("", "logs", "log")
	err := logger.RedirectStderr(file)
	assert.Nil(t, err)

	message := "redirect ok"
	os.Stderr.WriteString(message)

	data, err := ioutil.ReadFile(file.Name())
	assert.Nil(t, err)
	assert.Contains(t, string(data), message)
}

func TestRedirectStderrWithNilFile(t *testing.T) {
	err := logger.RedirectStderr(nil)
	assert.NotNil(t, err)
}

func TestRollFiles(t *testing.T) {
	t.Parallel()
	log := logger.NewElrondLogger()
	mockTime := time.Date(2019, 1, 1, 1, 1, 1, 1, time.Local)

	err := log.ApplyOptions(logger.WithFileRotation("", "logs", "log"))
	assert.Nil(t, err)

	for i := 0; i < logger.NrOfFilesToRemember()*2; i++ {
		log.SetCreationTime(mockTime)
		log.RollFiles()
	}

	assert.Equal(t, logger.NrOfFilesToRemember(), log.GetNrOfLogFiles())
}

func swallowPanicLog(t *testing.T, logMsg string, panicMsg string, log *logger.Logger) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf(panicMsg)
		}
	}()
	log.Panic(logMsg)
}
