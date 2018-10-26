package logger_test

import (
	"bytes"
	"elrond/elrond-go-sandbox/logger"
	"strings"
	"testing"

	"gotest.tools/assert"
)

func init() {
	logger.EL.SetLevel(logger.LvlDebugString)
}

func TestDebug(t *testing.T) {
	var str bytes.Buffer
	logger.EL.SetOutput(&str)
	logger.EL.Debug("abc")
	logString := str.String()
	assert.Assert(t, strings.Contains(logString, `"fields.level":"DEBUG"`))
	assert.Assert(t, strings.Contains(logString, `"msg":"abc"`))
}

func TestInfo(t *testing.T) {
	var str bytes.Buffer
	logger.EL.SetOutput(&str)
	logger.EL.Info("abc")
	logString := str.String()
	assert.Assert(t, strings.Contains(logString, `"fields.level":"INFO"`))
	assert.Assert(t, strings.Contains(logString, `"msg":"abc"`))
}

func TestWarn(t *testing.T) {
	var str bytes.Buffer
	logger.EL.SetOutput(&str)
	logger.EL.Warn("abc")
	logString := str.String()
	assert.Assert(t, strings.Contains(logString, `"fields.level":"WARNING"`))
	assert.Assert(t, strings.Contains(logString, `"msg":"abc"`))
}

func TestError(t *testing.T) {
	var str bytes.Buffer
	logger.EL.SetOutput(&str)
	logger.EL.Error("abc")
	logString := str.String()
	assert.Assert(t, strings.Contains(logString, `"fields.level":"ERROR"`))
	assert.Assert(t, strings.Contains(logString, `"msg":"abc"`))
}

func TestPanic(t *testing.T) {
	var str bytes.Buffer
	logger.EL.SetOutput(&str)
	swallowPanicLog(t, "abc", "TestPanic should have panic")

	logString := str.String()
	assert.Assert(t, strings.Contains(logString, `"fields.level":"PANIC"`))
	assert.Assert(t, strings.Contains(logString, `"msg":"abc"`))
}

func TestSetLevel(t *testing.T) {
	var str bytes.Buffer
	logger.EL.SetOutput(&str)

	logger.EL.SetLevel(logger.LvlDebugString)
	logger.EL.Debug("abc")
	assert.Assert(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()

	logger.EL.SetLevel(logger.LvlInfoString)
	logger.EL.Debug("abc")
	assert.Assert(t, len(str.String()) == 0)
	str.Reset()
	logger.EL.Info("abc")
	assert.Assert(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()

	logger.EL.SetLevel(logger.LvlWarningString)
	logger.EL.Info("abc")
	assert.Assert(t, len(str.String()) == 0)
	str.Reset()
	logger.EL.Warn("abc")
	assert.Assert(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()

	logger.EL.SetLevel(logger.LvlErrorString)
	logger.EL.Warn("abc")
	assert.Assert(t, len(str.String()) == 0)
	str.Reset()
	logger.EL.Error("abc")
	assert.Assert(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()

	logger.EL.SetLevel(logger.LvlPanicString)
	logger.EL.Error("abc")
	assert.Assert(t, len(str.String()) == 0)
	str.Reset()

	swallowPanicLog(t, "abc", "TestSetLevel should have panic")
	assert.Assert(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()

	logger.EL.SetLevel("this should go on the default case")
	logger.EL.Warn("abc")
	assert.Assert(t, len(str.String()) == 0)
	str.Reset()
	logger.EL.Error("abc")
	assert.Assert(t, strings.Contains(str.String(), `"msg":"abc"`))
	str.Reset()
}

func TestNoOutputFile(t *testing.T) {
	var str bytes.Buffer
	el := logger.NewElrondLogger(nil)
	el.SetLevel(logger.LvlWarningString)
	el.SetOutput(&str)
	el.Warn("abc")
	assert.Assert(t, strings.Contains(str.String(), `"msg":"abc"`))
}

func swallowPanicLog(t *testing.T, logMsg string, panicMsg string) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf(panicMsg)
		}
	}()
	logger.EL.Panic(logMsg)
}
