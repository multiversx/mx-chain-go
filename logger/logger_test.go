package logger_test

import (
	"sync/atomic"

	"testing"

	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/logger/mock"
	"github.com/stretchr/testify/assert"
)

func generateTestLogOutputSubject() (logger.LogOutputHandler, *int32) {
	numCalls := int32(0)
	los := logger.NewLogOutputSubject()
	_ = los.AddObserver(
		&mock.WriterStub{
			WriteCalled: func(p []byte) (n int, err error) {
				atomic.AddInt32(&numCalls, 1)
				return 0, nil
			},
		},
		&mock.FormatterStub{
			OutputCalled: func(line logger.LogLineHandler) []byte {
				return nil
			},
		},
	)

	return los, &numCalls
}

//------- Trace

func TestLogger_TraceShouldNotCallIfLogLevelIsHigher(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogDebug, los)

	log.Trace("test")

	assert.Equal(t, int32(0), atomic.LoadInt32(numCalls))
}

func TestLogger_TraceShouldCallIfLogLevelIsEqual(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogTrace, los)

	log.Trace("test")

	assert.Equal(t, int32(1), atomic.LoadInt32(numCalls))
}

//------- Debug

func TestLogger_DebugShouldNotCallIfLogLevelIsHigher(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogInfo, los)

	log.Debug("test")

	assert.Equal(t, int32(0), atomic.LoadInt32(numCalls))
}

func TestLogger_DebugShouldCallIfLogLevelIsEqual(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogDebug, los)

	log.Debug("test")

	assert.Equal(t, int32(1), atomic.LoadInt32(numCalls))
}

func TestLogger_DebugShouldCallIfLogLevelIsLower(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogTrace, los)

	log.Debug("test")

	assert.Equal(t, int32(1), atomic.LoadInt32(numCalls))
}

//------- Info

func TestLogger_InfoShouldNotCallIfLogLevelIsHigher(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogWarning, los)

	log.Info("test")

	assert.Equal(t, int32(0), atomic.LoadInt32(numCalls))
}

func TestLogger_InfoShouldCallIfLogLevelIsEqual(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogInfo, los)

	log.Info("test")

	assert.Equal(t, int32(1), atomic.LoadInt32(numCalls))
}

func TestLogger_InfoShouldCallIfLogLevelIsLower(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogDebug, los)

	log.Info("test")

	assert.Equal(t, int32(1), atomic.LoadInt32(numCalls))
}

//------- Warn

func TestLogger_WarnShouldNotCallIfLogLevelIsHigher(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogError, los)

	log.Warn("test")

	assert.Equal(t, int32(0), atomic.LoadInt32(numCalls))
}

func TestLogger_WarnShouldCallIfLogLevelIsEqual(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogWarning, los)

	log.Warn("test")

	assert.Equal(t, int32(1), atomic.LoadInt32(numCalls))
}

func TestLogger_WarnShouldCallIfLogLevelIsLower(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogInfo, los)

	log.Warn("test")

	assert.Equal(t, int32(1), atomic.LoadInt32(numCalls))
}

//------- Error

func TestLogger_ErrorShouldNotCallIfLogLevelIsHigher(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogNone, los)

	log.Error("test")

	assert.Equal(t, int32(0), atomic.LoadInt32(numCalls))
}

func TestLogger_ErrorShouldCallIfLogLevelIsEqual(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogError, los)

	log.Error("test")

	assert.Equal(t, int32(1), atomic.LoadInt32(numCalls))
}

func TestLogger_ErrorShouldCallIfLogLevelIsLower(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogWarning, los)

	log.Error("test")

	assert.Equal(t, int32(1), atomic.LoadInt32(numCalls))
}

//------- LogIfError

func TestLogger_LogIfErrorShouldNotCallIfErrorIsNil(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogError, los)

	log.LogIfError(nil)

	assert.Equal(t, int32(0), atomic.LoadInt32(numCalls))
}

func TestLogger_LogIfErrorShouldCallIfErrorIsNotNil(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogError, los)

	log.LogIfError(logger.ErrNilFormatter)

	assert.Equal(t, int32(1), atomic.LoadInt32(numCalls))
}

//------- SetLevel

func TestLogger_SetLevelShouldWork(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogInfo, los)

	log.Debug("test")
	assert.Equal(t, int32(0), atomic.LoadInt32(numCalls))

	log.SetLevel(logger.LogDebug)

	log.Debug("test")
	assert.Equal(t, int32(1), atomic.LoadInt32(numCalls))
}

//------- Log

func TestLogger_LogNilShouldNotCallWrite(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogError, los)

	log.Log(nil)

	assert.Equal(t, int32(0), atomic.LoadInt32(numCalls))
}

func TestLogger_LogShouldWork(t *testing.T) {
	t.Parallel()

	los, numCalls := generateTestLogOutputSubject()
	log := logger.NewLogger("test", logger.LogError, los)

	log.Log(&logger.LogLine{})

	assert.Equal(t, int32(1), atomic.LoadInt32(numCalls))
}
