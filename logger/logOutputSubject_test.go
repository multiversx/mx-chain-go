package logger_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/logger/mock"
	"github.com/stretchr/testify/assert"
)

//------- AddObserver

func TestLogOutputSubject_AddObserverNilWriterShouldError(t *testing.T) {
	t.Parallel()

	los := logger.NewLogOutputSubject()

	err := los.AddObserver(nil, &mock.FormatterStub{})

	assert.Equal(t, logger.ErrNilWriter, err)
}

func TestLogOutputSubject_AddObserverNilFormatterShouldError(t *testing.T) {
	t.Parallel()

	los := logger.NewLogOutputSubject()

	err := los.AddObserver(&mock.WriterStub{}, nil)

	assert.Equal(t, logger.ErrNilFormatter, err)
}

func TestLogOutputSubject_AddObserverShouldWork(t *testing.T) {
	t.Parallel()

	los := logger.NewLogOutputSubject()

	err := los.AddObserver(&mock.WriterStub{}, &mock.FormatterStub{})
	writers, formatters := los.Observers()

	assert.Nil(t, err)
	assert.Equal(t, 1, len(writers))
	assert.Equal(t, 1, len(formatters))
}

//------- Output

func TestLogOutputSubject_OutputNoObserversShouldDoNothing(t *testing.T) {
	t.Parallel()

	los := logger.NewLogOutputSubject()

	los.Output(nil)
}

func TestLogOutputSubject_OutputShouldCallFormatterAndWriter(t *testing.T) {
	t.Parallel()

	var formatterCalled = int32(0)
	var writerCalled = int32(0)
	los := logger.NewLogOutputSubject()
	_ = los.AddObserver(
		&mock.WriterStub{
			WriteCalled: func(p []byte) (n int, err error) {
				atomic.AddInt32(&writerCalled, 1)
				return 0, nil
			},
		},
		&mock.FormatterStub{
			OutputCalled: func(line *logger.LogLine) []byte {
				atomic.AddInt32(&formatterCalled, 1)
				return nil
			},
		},
	)

	los.Output(nil)

	assert.Equal(t, int32(1), atomic.LoadInt32(&writerCalled))
	assert.Equal(t, int32(1), atomic.LoadInt32(&formatterCalled))
}

func TestLogOutputSubject_OutputCalledConcurrentShouldWork(t *testing.T) {
	t.Parallel()

	var formatterCalled = int32(0)
	var writerCalled = int32(0)
	los := logger.NewLogOutputSubject()
	_ = los.AddObserver(
		&mock.WriterStub{
			WriteCalled: func(p []byte) (n int, err error) {
				atomic.AddInt32(&writerCalled, 1)
				return 0, nil
			},
		},
		&mock.FormatterStub{
			OutputCalled: func(line *logger.LogLine) []byte {
				atomic.AddInt32(&formatterCalled, 1)
				return nil
			},
		},
	)

	numCalls := 1000
	wg := &sync.WaitGroup{}
	wg.Add(numCalls)
	for i := 0; i < numCalls; i++ {
		go func() {
			time.Sleep(time.Millisecond)
			los.Output(nil)
			wg.Done()
		}()
	}

	wg.Wait()

	assert.Equal(t, int32(numCalls), atomic.LoadInt32(&writerCalled))
	assert.Equal(t, int32(numCalls), atomic.LoadInt32(&formatterCalled))
}
