package outport

import (
	"errors"
	"fmt"
	"sync"
	atomicGo "sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/outport/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const counterPositionInLogMessage = 5

func createSaveBlockArgs() *outportcore.OutportBlockWithHeaderAndBody {
	return &outportcore.OutportBlockWithHeaderAndBody{
		OutportBlock: &outportcore.OutportBlock{},
		HeaderDataWithBody: &outportcore.HeaderDataWithBody{
			Body:       &block.Body{},
			Header:     &block.HeaderV2{},
			HeaderHash: []byte("hash"),
		},
	}
}

func TestNewOutport(t *testing.T) {
	t.Parallel()

	t.Run("invalid retrial time should error", func(t *testing.T) {
		outportHandler, err := NewOutport(0, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})

		assert.True(t, errors.Is(err, ErrInvalidRetrialInterval))
		assert.True(t, check.IfNil(outportHandler))
	})
	t.Run("should work", func(t *testing.T) {
		outportHandler, err := NewOutport(minimumRetrialInterval, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})

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
		SaveAccountsCalled: func(accounts *outportcore.Accounts) error {
			numCalled1++
			if numCalled1 < 10 {
				return expectedError
			}

			return nil
		},
	}
	driver2 := &mock.DriverStub{
		SaveAccountsCalled: func(accounts *outportcore.Accounts) error {
			numCalled2++
			return nil
		},
	}
	outportHandler, _ := NewOutport(minimumRetrialInterval, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})
	numLogDebugCalled := uint32(0)
	outportHandler.logHandler = func(logLevel logger.LogLevel, message string, args ...interface{}) {
		if logLevel == logger.LogError {
			assert.Fail(t, "should have not called log error")
		}
		if logLevel == logger.LogDebug {
			atomicGo.AddUint32(&numLogDebugCalled, 1)
		}
	}

	outportHandler.SaveAccounts(&outportcore.Accounts{})
	time.Sleep(time.Second)
	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	outportHandler.SaveAccounts(&outportcore.Accounts{})
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
		SaveBlockCalled: func(args *outportcore.OutportBlock) error {
			numCalled1++
			if numCalled1 < 10 {
				return expectedError
			}

			return nil
		},
	}
	driver2 := &mock.DriverStub{
		SaveBlockCalled: func(args *outportcore.OutportBlock) error {
			numCalled2++
			return nil
		},
	}
	outportHandler, _ := NewOutport(minimumRetrialInterval, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})
	numLogDebugCalled := uint32(0)
	outportHandler.logHandler = func(logLevel logger.LogLevel, message string, args ...interface{}) {
		if logLevel == logger.LogError {
			assert.Fail(t, "should have not called log error")
		}
		if logLevel == logger.LogDebug {
			atomicGo.AddUint32(&numLogDebugCalled, 1)
		}
	}

	args := createSaveBlockArgs()
	_ = outportHandler.SaveBlock(args)
	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	_ = outportHandler.SaveBlock(args)
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
		SaveRoundsInfoCalled: func(roundsInfos *outportcore.RoundsInfo) error {
			numCalled1++
			if numCalled1 < 10 {
				return expectedError
			}

			return nil
		},
	}
	driver2 := &mock.DriverStub{
		SaveRoundsInfoCalled: func(roundsInfos *outportcore.RoundsInfo) error {
			numCalled2++
			return nil
		},
	}
	outportHandler, _ := NewOutport(minimumRetrialInterval, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})
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
		SaveValidatorsPubKeysCalled: func(validatorsRating *outportcore.ValidatorsPubKeys) error {
			numCalled1++
			if numCalled1 < 10 {
				return expectedError
			}

			return nil
		},
	}
	driver2 := &mock.DriverStub{
		SaveValidatorsPubKeysCalled: func(validatorsRating *outportcore.ValidatorsPubKeys) error {
			numCalled2++
			return nil
		},
	}
	outportHandler, _ := NewOutport(minimumRetrialInterval, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})
	numLogDebugCalled := uint32(0)
	outportHandler.logHandler = func(logLevel logger.LogLevel, message string, args ...interface{}) {
		if logLevel == logger.LogError {
			assert.Fail(t, "should have not called log error")
		}
		if logLevel == logger.LogDebug {
			atomicGo.AddUint32(&numLogDebugCalled, 1)
		}
	}

	outportHandler.SaveValidatorsPubKeys(&outportcore.ValidatorsPubKeys{})
	time.Sleep(time.Second)

	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	outportHandler.SaveValidatorsPubKeys(&outportcore.ValidatorsPubKeys{})
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
		SaveValidatorsRatingCalled: func(validatorsRating *outportcore.ValidatorsRating) error {
			numCalled1++
			if numCalled1 < 10 {
				return expectedError
			}

			return nil
		},
	}
	driver2 := &mock.DriverStub{
		SaveValidatorsRatingCalled: func(validatorsRating *outportcore.ValidatorsRating) error {
			numCalled2++
			return nil
		},
	}
	outportHandler, _ := NewOutport(minimumRetrialInterval, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})
	numLogDebugCalled := uint32(0)
	outportHandler.logHandler = func(logLevel logger.LogLevel, message string, args ...interface{}) {
		if logLevel == logger.LogError {
			assert.Fail(t, "should have not called log error")
		}
		if logLevel == logger.LogDebug {
			atomicGo.AddUint32(&numLogDebugCalled, 1)
		}
	}

	outportHandler.SaveValidatorsRating(&outportcore.ValidatorsRating{})
	time.Sleep(time.Second)

	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	outportHandler.SaveValidatorsRating(&outportcore.ValidatorsRating{})
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
		RevertIndexedBlockCalled: func(blockData *outportcore.BlockData) error {
			numCalled1++
			if numCalled1 < 10 {
				return expectedError
			}

			return nil
		},
	}
	driver2 := &mock.DriverStub{
		RevertIndexedBlockCalled: func(blockData *outportcore.BlockData) error {
			numCalled2++
			return nil
		},
	}
	outportHandler, _ := NewOutport(minimumRetrialInterval, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})
	numLogDebugCalled := uint32(0)
	outportHandler.logHandler = func(logLevel logger.LogLevel, message string, args ...interface{}) {
		if logLevel == logger.LogError {
			assert.Fail(t, "should have not called log error")
		}
		if logLevel == logger.LogDebug {
			atomicGo.AddUint32(&numLogDebugCalled, 1)
		}
	}

	args := createSaveBlockArgs()
	_ = outportHandler.RevertIndexedBlock(args.HeaderDataWithBody)
	time.Sleep(time.Second)

	_ = outportHandler.SubscribeDriver(driver1)
	_ = outportHandler.SubscribeDriver(driver2)

	_ = outportHandler.RevertIndexedBlock(args.HeaderDataWithBody)
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
		FinalizedBlockCalled: func(finalizedBlock *outportcore.FinalizedBlock) error {
			numCalled1++
			if numCalled1 < 10 {
				return expectedError
			}

			return nil
		},
	}
	driver2 := &mock.DriverStub{
		FinalizedBlockCalled: func(finalizedBlock *outportcore.FinalizedBlock) error {
			numCalled2++
			return nil
		},
	}
	outportHandler, _ := NewOutport(minimumRetrialInterval, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})
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
		outportHandler, _ := NewOutport(minimumRetrialInterval, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})

		require.False(t, outportHandler.HasDrivers())

		err := outportHandler.SubscribeDriver(nil)
		require.Equal(t, ErrNilDriver, err)
		require.False(t, outportHandler.HasDrivers())
	})
	t.Run("should work", func(t *testing.T) {
		outportHandler, _ := NewOutport(minimumRetrialInterval, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})

		require.False(t, outportHandler.HasDrivers())

		err := outportHandler.SubscribeDriver(&mock.DriverStub{})
		require.Nil(t, err)
		require.True(t, outportHandler.HasDrivers())
	})
}

func TestOutport_Close(t *testing.T) {
	t.Parallel()

	outportHandler, _ := NewOutport(minimumRetrialInterval, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})

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

	outportHandler, _ := NewOutport(minimumRetrialInterval, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})

	localErr := errors.New("driver stuck in error")
	driver1 := &mock.DriverStub{
		SaveBlockCalled: func(args *outportcore.OutportBlock) error {
			return localErr
		},
		RevertIndexedBlockCalled: func(blockData *outportcore.BlockData) error {
			return localErr
		},
		SaveRoundsInfoCalled: func(roundsInfos *outportcore.RoundsInfo) error {
			return localErr
		},
		SaveValidatorsPubKeysCalled: func(validatorsPubKeys *outportcore.ValidatorsPubKeys) error {
			return localErr
		},
		SaveValidatorsRatingCalled: func(validatorsRating *outportcore.ValidatorsRating) error {
			return localErr
		},
		SaveAccountsCalled: func(accounts *outportcore.Accounts) error {
			return localErr
		},
		FinalizedBlockCalled: func(finalizedBlock *outportcore.FinalizedBlock) error {
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
		outportHandler.SaveAccounts(nil)
		wg.Done()
	}()
	go func() {
		_ = outportHandler.SaveBlock(nil)
		wg.Done()
	}()
	go func() {
		_ = outportHandler.RevertIndexedBlock(nil)
		wg.Done()
	}()
	go func() {
		outportHandler.SaveRoundsInfo(nil)
		wg.Done()
	}()
	go func() {
		outportHandler.SaveValidatorsPubKeys(nil)
		wg.Done()
	}()
	go func() {
		outportHandler.SaveValidatorsRating(nil)
		wg.Done()
	}()
	go func() {
		outportHandler.SaveAccounts(nil)
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
	outportHandler, _ := NewOutport(minimumRetrialInterval, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})
	outportHandler.messageCounter = currentCounter
	outportHandler.timeForDriverCall = time.Second
	logErrorCalled := atomic.Flag{}
	numLogDebugCalled := uint32(0)
	outportHandler.logHandler = func(logLevel logger.LogLevel, message string, args ...interface{}) {
		if logLevel == logger.LogWarning {
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
		SaveBlockCalled: func(args *outportcore.OutportBlock) error {
			time.Sleep(time.Second * 5)
			return nil
		},
	})

	args := createSaveBlockArgs()
	_ = outportHandler.SaveBlock(args)

	assert.True(t, logErrorCalled.IsSet())
	assert.Equal(t, uint32(1), atomicGo.LoadUint32(&numLogDebugCalled))
}

func TestOutport_SaveBlockDriverIsNotStuck(t *testing.T) {
	t.Parallel()

	currentCounter := uint64(778)
	outportHandler, _ := NewOutport(minimumRetrialInterval, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})
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
		SaveBlockCalled: func(args *outportcore.OutportBlock) error {
			return nil
		},
	})

	args := createSaveBlockArgs()
	_ = outportHandler.SaveBlock(args)
	time.Sleep(time.Second)

	assert.Equal(t, uint32(2), atomicGo.LoadUint32(&numLogDebugCalled))
}

func TestOutport_SettingsRequestAndReceive(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	t.Run("RegisterHandlerForSettingsRequest errors, should not add the driver", func(t *testing.T) {
		t.Parallel()

		driver := &mock.DriverStub{
			RegisterHandlerCalled: func(handlerFunction func() error, _ string) error {
				return errExpected
			},
		}

		outportHandler, _ := NewOutport(time.Second, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})
		err := outportHandler.SubscribeDriver(driver)
		assert.Equal(t, errExpected, err)
		require.False(t, outportHandler.HasDrivers())
	})
	t.Run("SetCurrentSettings errors, should not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, fmt.Sprintf("should have not failed %v", r))
			}
		}()

		currentSettingsCalled := false
		var callback func() error
		driver := &mock.DriverStub{
			RegisterHandlerCalled: func(handlerFunction func() error, _ string) error {
				callback = handlerFunction

				return nil
			},
			SetCurrentSettingsCalled: func(config outportcore.OutportConfig) error {
				currentSettingsCalled = true
				return errExpected
			},
		}

		outportHandler, _ := NewOutport(time.Second, outportcore.OutportConfig{}, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})
		err := outportHandler.SubscribeDriver(driver)
		assert.Nil(t, err)

		assert.False(t, currentSettingsCalled)

		err = callback()
		assert.Equal(t, errExpected, err)
		assert.True(t, currentSettingsCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		var driverRequestHandler func() error
		receivedOutportConfig := outportcore.OutportConfig{}
		driver := &mock.DriverStub{
			RegisterHandlerCalled: func(handlerFunction func() error, _ string) error {
				driverRequestHandler = handlerFunction

				return nil
			},
			SetCurrentSettingsCalled: func(config outportcore.OutportConfig) error {
				receivedOutportConfig = config
				return nil
			},
		}

		providedConfig := outportcore.OutportConfig{
			IsInImportDBMode: true,
		}
		outportHandler, _ := NewOutport(time.Second, providedConfig, &enableEpochsHandlerMock.EnableEpochsHandlerStub{}, &testscommon.EnableRoundsHandlerStub{})
		err := outportHandler.SubscribeDriver(driver)
		assert.Nil(t, err)
		assert.True(t, outportHandler.HasDrivers())

		assert.NotNil(t, driverRequestHandler) // the RegisterHandlerForSettingsRequest should have been called, handler set

		// the expected config should be empty as the handler should not call the driver's SetCurrentSettings automatically at subscribe time
		assert.Equal(t, outportcore.OutportConfig{}, receivedOutportConfig)

		// driver calls the handler because it wants the config
		err = driverRequestHandler()
		assert.Nil(t, err)
		assert.Equal(t, providedConfig, receivedOutportConfig)
	})
}
