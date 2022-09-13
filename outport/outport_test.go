package outport

import (
	"errors"
	"sync"
	atomicGo "sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/outport/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const counterPositionInLogMessage = 5

func TestNewOutport(t *testing.T) {
	t.Parallel()

	t.Run("invalid retrial time should error", func(t *testing.T) {
		outportHandler, err := NewOutport(0)

		assert.True(t, errors.Is(err, ErrInvalidRetrialInterval))
		assert.True(t, check.IfNil(outportHandler))
	})
	t.Run("should work", func(t *testing.T) {
		outportHandler, err := NewOutport(minimumRetrialInterval)

		assert.Nil(t, err)
		assert.False(t, check.IfNil(outportHandler))
	})
}

func TestOutport_SaveAccounts(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	numCalled1 := 0
	numCalled2 := 0
	driver1 := &mock.DriverStub{
		SaveAccountsCalled: func(blockTimestamp uint64, acc []data.UserAccountHandler) error {
			numCalled1++
			if numCalled1 < 10 {
				return expectedError
			}

			return nil
		},
	}
	driver2 := &mock.DriverStub{
		SaveAccountsCalled: func(blockTimestamp uint64, acc []data.UserAccountHandler) error {
			numCalled2++
			return nil
		},
	}
	outportHandler, _ := NewOutport(minimumRetrialInterval)
	numLogDebugCalled := uint32(0)
	outportHandler.logHandler = func(logLevel logger.LogLevel, message string, args ...interface{}) {
		if logLevel == logger.LogError {
			assert.Fail(t, "should have not called log error")
		}
		if logLevel == logger.LogDebug {
			atomicGo.AddUint32(&numLogDebugCalled, 1)
		}
	}

	outportHandler.SaveAccounts(0, []data.UserAccountHandler{})
	time.Sleep(time.Second)
	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	outportHandler.SaveAccounts(0, []data.UserAccountHandler{})
	time.Sleep(time.Second)

	assert.Equal(t, 10, numCalled1)
	assert.Equal(t, 1, numCalled2)
	assert.Equal(t, uint32(4), atomicGo.LoadUint32(&numLogDebugCalled))
}

func TestOutport_SaveBlock(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	numCalled1 := 0
	numCalled2 := 0
	driver1 := &mock.DriverStub{
		SaveBlockCalled: func(args *indexer.ArgsSaveBlockData) error {
			numCalled1++
			if numCalled1 < 10 {
				return expectedError
			}

			return nil
		},
	}
	driver2 := &mock.DriverStub{
		SaveBlockCalled: func(args *indexer.ArgsSaveBlockData) error {
			numCalled2++
			return nil
		},
	}
	outportHandler, _ := NewOutport(minimumRetrialInterval)
	numLogDebugCalled := uint32(0)
	outportHandler.logHandler = func(logLevel logger.LogLevel, message string, args ...interface{}) {
		if logLevel == logger.LogError {
			assert.Fail(t, "should have not called log error")
		}
		if logLevel == logger.LogDebug {
			atomicGo.AddUint32(&numLogDebugCalled, 1)
		}
	}

	outportHandler.SaveBlock(nil)
	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	outportHandler.SaveBlock(nil)
	time.Sleep(time.Second)

	assert.Equal(t, 10, numCalled1)
	assert.Equal(t, 1, numCalled2)
	assert.Equal(t, uint32(4), atomicGo.LoadUint32(&numLogDebugCalled))
}

func TestOutport_SaveRoundsInfo(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	numCalled1 := 0
	numCalled2 := 0
	driver1 := &mock.DriverStub{
		SaveRoundsInfoCalled: func(roundsInfos []*indexer.RoundInfo) error {
			numCalled1++
			if numCalled1 < 10 {
				return expectedError
			}

			return nil
		},
	}
	driver2 := &mock.DriverStub{
		SaveRoundsInfoCalled: func(roundsInfos []*indexer.RoundInfo) error {
			numCalled2++
			return nil
		},
	}
	outportHandler, _ := NewOutport(minimumRetrialInterval)
	numLogDebugCalled := uint32(0)
	outportHandler.logHandler = func(logLevel logger.LogLevel, message string, args ...interface{}) {
		if logLevel == logger.LogError {
			assert.Fail(t, "should have not called log error")
		}
		if logLevel == logger.LogDebug {
			atomicGo.AddUint32(&numLogDebugCalled, 1)
		}
	}

	outportHandler.SaveRoundsInfo(nil)
	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	outportHandler.SaveRoundsInfo(nil)

	time.Sleep(time.Second)
	assert.Equal(t, 10, numCalled1)
	assert.Equal(t, 1, numCalled2)
	assert.Equal(t, uint32(4), atomicGo.LoadUint32(&numLogDebugCalled))
}

func TestOutport_SaveValidatorsPubKeys(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	numCalled1 := 0
	numCalled2 := 0
	driver1 := &mock.DriverStub{
		SaveValidatorsPubKeysCalled: func(validatorsPubKeys map[uint32][][]byte, epoch uint32) error {
			numCalled1++
			if numCalled1 < 10 {
				return expectedError
			}

			return nil
		},
	}
	driver2 := &mock.DriverStub{
		SaveValidatorsPubKeysCalled: func(validatorsPubKeys map[uint32][][]byte, epoch uint32) error {
			numCalled2++
			return nil
		},
	}
	outportHandler, _ := NewOutport(minimumRetrialInterval)
	numLogDebugCalled := uint32(0)
	outportHandler.logHandler = func(logLevel logger.LogLevel, message string, args ...interface{}) {
		if logLevel == logger.LogError {
			assert.Fail(t, "should have not called log error")
		}
		if logLevel == logger.LogDebug {
			atomicGo.AddUint32(&numLogDebugCalled, 1)
		}
	}

	outportHandler.SaveValidatorsPubKeys(nil, 0)
	time.Sleep(time.Second)

	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	outportHandler.SaveValidatorsPubKeys(nil, 0)
	time.Sleep(time.Second)

	assert.Equal(t, 10, numCalled1)
	assert.Equal(t, 1, numCalled2)
	assert.Equal(t, uint32(4), atomicGo.LoadUint32(&numLogDebugCalled))
}

func TestOutport_SaveValidatorsRating(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	numCalled1 := 0
	numCalled2 := 0
	driver1 := &mock.DriverStub{
		SaveValidatorsRatingCalled: func(indexID string, infoRating []*indexer.ValidatorRatingInfo) error {
			numCalled1++
			if numCalled1 < 10 {
				return expectedError
			}

			return nil
		},
	}
	driver2 := &mock.DriverStub{
		SaveValidatorsRatingCalled: func(indexID string, infoRating []*indexer.ValidatorRatingInfo) error {
			numCalled2++
			return nil
		},
	}
	outportHandler, _ := NewOutport(minimumRetrialInterval)
	numLogDebugCalled := uint32(0)
	outportHandler.logHandler = func(logLevel logger.LogLevel, message string, args ...interface{}) {
		if logLevel == logger.LogError {
			assert.Fail(t, "should have not called log error")
		}
		if logLevel == logger.LogDebug {
			atomicGo.AddUint32(&numLogDebugCalled, 1)
		}
	}

	outportHandler.SaveValidatorsRating("", nil)
	time.Sleep(time.Second)

	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	outportHandler.SaveValidatorsRating("", nil)
	time.Sleep(time.Second)

	assert.Equal(t, 10, numCalled1)
	assert.Equal(t, 1, numCalled2)
	assert.Equal(t, uint32(4), atomicGo.LoadUint32(&numLogDebugCalled))
}

func TestOutport_RevertIndexedBlock(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	numCalled1 := 0
	numCalled2 := 0
	driver1 := &mock.DriverStub{
		RevertBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
			numCalled1++
			if numCalled1 < 10 {
				return expectedError
			}

			return nil
		},
	}
	driver2 := &mock.DriverStub{
		RevertBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
			numCalled2++
			return nil
		},
	}
	outportHandler, _ := NewOutport(minimumRetrialInterval)
	numLogDebugCalled := uint32(0)
	outportHandler.logHandler = func(logLevel logger.LogLevel, message string, args ...interface{}) {
		if logLevel == logger.LogError {
			assert.Fail(t, "should have not called log error")
		}
		if logLevel == logger.LogDebug {
			atomicGo.AddUint32(&numLogDebugCalled, 1)
		}
	}

	outportHandler.RevertIndexedBlock(nil, nil)
	time.Sleep(time.Second)

	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	outportHandler.RevertIndexedBlock(nil, nil)
	time.Sleep(time.Second)

	assert.Equal(t, 10, numCalled1)
	assert.Equal(t, 1, numCalled2)
	assert.Equal(t, uint32(4), atomicGo.LoadUint32(&numLogDebugCalled))
}

func TestOutport_FinalizedBlock(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	numCalled1 := 0
	numCalled2 := 0
	driver1 := &mock.DriverStub{
		FinalizedBlockCalled: func(headerHash []byte) error {
			numCalled1++
			if numCalled1 < 10 {
				return expectedError
			}

			return nil
		},
	}
	driver2 := &mock.DriverStub{
		FinalizedBlockCalled: func(headerHash []byte) error {
			numCalled2++
			return nil
		},
	}
	outportHandler, _ := NewOutport(minimumRetrialInterval)
	numLogDebugCalled := uint32(0)
	outportHandler.logHandler = func(logLevel logger.LogLevel, message string, args ...interface{}) {
		if logLevel == logger.LogError {
			assert.Fail(t, "should have not called log error")
		}
		if logLevel == logger.LogDebug {
			atomicGo.AddUint32(&numLogDebugCalled, 1)
		}
	}

	outportHandler.FinalizedBlock(nil)
	time.Sleep(time.Second)

	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	outportHandler.FinalizedBlock(nil)
	time.Sleep(time.Second)

	assert.Equal(t, 10, numCalled1)
	assert.Equal(t, 1, numCalled2)
	assert.Equal(t, uint32(4), atomicGo.LoadUint32(&numLogDebugCalled))
}

func TestOutport_SubscribeDriver(t *testing.T) {
	t.Parallel()

	t.Run("nil driver should error", func(t *testing.T) {
		outportHandler, _ := NewOutport(minimumRetrialInterval)

		require.False(t, outportHandler.HasDrivers())

		err := outportHandler.SubscribeDriver(nil)
		require.Equal(t, ErrNilDriver, err)
		require.False(t, outportHandler.HasDrivers())
	})
	t.Run("should work", func(t *testing.T) {
		outportHandler, _ := NewOutport(minimumRetrialInterval)

		require.False(t, outportHandler.HasDrivers())

		err := outportHandler.SubscribeDriver(&mock.DriverStub{})
		require.Nil(t, err)
		require.True(t, outportHandler.HasDrivers())
	})
}

func TestOutport_Close(t *testing.T) {
	t.Parallel()

	outportHandler, _ := NewOutport(minimumRetrialInterval)

	localErr := errors.New("local err")
	driver1 := &mock.DriverStub{
		CloseCalled: func() error {
			return localErr
		},
	}
	driver2 := &mock.DriverStub{
		CloseCalled: func() error {
			return nil
		},
	}

	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	err := outportHandler.Close()
	require.Equal(t, localErr, err)
}

func TestOutport_CloseWhileDriverIsStuckInContinuousErrors(t *testing.T) {
	t.Parallel()

	outportHandler, _ := NewOutport(minimumRetrialInterval)

	localErr := errors.New("driver stuck in error")
	driver1 := &mock.DriverStub{
		SaveBlockCalled: func(args *indexer.ArgsSaveBlockData) error {
			return localErr
		},
		RevertBlockCalled: func(header data.HeaderHandler, body data.BodyHandler) error {
			return localErr
		},
		SaveRoundsInfoCalled: func(roundsInfos []*indexer.RoundInfo) error {
			return localErr
		},
		SaveValidatorsPubKeysCalled: func(validatorsPubKeys map[uint32][][]byte, epoch uint32) error {
			return localErr
		},
		SaveValidatorsRatingCalled: func(indexID string, infoRating []*indexer.ValidatorRatingInfo) error {
			return localErr
		},
		SaveAccountsCalled: func(timestamp uint64, acc []data.UserAccountHandler) error {
			return localErr
		},
		FinalizedBlockCalled: func(headerHash []byte) error {
			return localErr
		},
		CloseCalled: func() error {
			return nil
		},
	}

	_ = outportHandler.SubscribeDriver(driver1)

	wg := &sync.WaitGroup{}
	wg.Add(9)
	go func() {
		outportHandler.SaveAccounts(0, nil)
		wg.Done()
	}()
	go func() {
		outportHandler.SaveBlock(nil)
		wg.Done()
	}()
	go func() {
		outportHandler.RevertIndexedBlock(nil, nil)
		wg.Done()
	}()
	go func() {
		outportHandler.SaveRoundsInfo(nil)
		wg.Done()
	}()
	go func() {
		outportHandler.SaveValidatorsPubKeys(nil, 0)
		wg.Done()
	}()
	go func() {
		outportHandler.SaveValidatorsRating("", nil)
		wg.Done()
	}()
	go func() {
		outportHandler.SaveAccounts(0, nil)
		wg.Done()
	}()
	go func() {
		outportHandler.FinalizedBlock(nil)
		wg.Done()
	}()
	go func() {
		_ = outportHandler.Close()
		wg.Done()
	}()

	chDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(chDone)
	}()

	select {
	case <-chDone:
	case <-time.After(time.Second):
		require.Fail(t, "unable to close all drivers because of a stuck driver")
	}
}

func TestOutport_SaveBlockDriverStuck(t *testing.T) {
	t.Parallel()

	currentCounter := uint64(778)
	outportHandler, _ := NewOutport(minimumRetrialInterval)
	outportHandler.messageCounter = currentCounter
	outportHandler.timeForDriverCall = time.Second
	logErrorCalled := atomic.Flag{}
	numLogDebugCalled := uint32(0)
	outportHandler.logHandler = func(logLevel logger.LogLevel, message string, args ...interface{}) {
		if logLevel == logger.LogError {
			logErrorCalled.SetValue(true)
			assert.Equal(t, "outport.monitorCompletionOnDriver took too long", message)
			assert.Equal(t, currentCounter+1, args[counterPositionInLogMessage])
		}
		if logLevel == logger.LogDebug {
			atomicGo.AddUint32(&numLogDebugCalled, 1)
			assert.Equal(t, currentCounter+1, args[counterPositionInLogMessage])
		}
	}

	_ = outportHandler.SubscribeDriver(&mock.DriverStub{
		SaveBlockCalled: func(args *indexer.ArgsSaveBlockData) error {
			time.Sleep(time.Second * 5)
			return nil
		},
	})

	outportHandler.SaveBlock(nil)

	assert.True(t, logErrorCalled.IsSet())
	assert.Equal(t, uint32(1), atomicGo.LoadUint32(&numLogDebugCalled))
}

func TestOutport_SaveBlockDriverIsNotStuck(t *testing.T) {
	t.Parallel()

	currentCounter := uint64(778)
	outportHandler, _ := NewOutport(minimumRetrialInterval)
	outportHandler.messageCounter = currentCounter
	outportHandler.timeForDriverCall = time.Second
	numLogDebugCalled := uint32(0)
	outportHandler.logHandler = func(logLevel logger.LogLevel, message string, args ...interface{}) {
		if logLevel == logger.LogError {
			assert.Fail(t, "should have not called log error")
		}
		if logLevel == logger.LogDebug {
			if atomicGo.LoadUint32(&numLogDebugCalled) == 0 {
				assert.Equal(t, "outport.monitorCompletionOnDriver starting", message)
				assert.Equal(t, currentCounter+1, args[counterPositionInLogMessage])
			}
			if atomicGo.LoadUint32(&numLogDebugCalled) == 1 {
				assert.Equal(t, "outport.monitorCompletionOnDriver ended", message)
				assert.Equal(t, currentCounter+1, args[counterPositionInLogMessage])
			}

			atomicGo.AddUint32(&numLogDebugCalled, 1)
		}
	}

	_ = outportHandler.SubscribeDriver(&mock.DriverStub{
		SaveBlockCalled: func(args *indexer.ArgsSaveBlockData) error {
			return nil
		},
	})

	outportHandler.SaveBlock(nil)
	time.Sleep(time.Second)

	assert.Equal(t, uint32(2), atomicGo.LoadUint32(&numLogDebugCalled))
}
