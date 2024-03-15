package forking

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/mock"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createGasScheduleNotifierArgs() ArgsNewGasScheduleNotifier {
	return ArgsNewGasScheduleNotifier{
		GasScheduleConfig: config.GasScheduleConfig{
			GasScheduleByEpochs: []config.GasScheduleByEpochs{
				{
					StartEpoch: 0,
					FileName:   "gasScheduleV1.toml",
				},
				{
					StartEpoch: 2,
					FileName:   common.LatestGasScheduleFileName,
				},
			}},
		ConfigDir:          "../../cmd/node/config/gasSchedules",
		EpochNotifier:      NewGenericEpochNotifier(),
		WasmVMChangeLocker: &sync.RWMutex{},
	}
}

func TestNewGasScheduleNotifierConstructorErrors(t *testing.T) {
	t.Parallel()

	args := createGasScheduleNotifierArgs()
	args.GasScheduleConfig = config.GasScheduleConfig{}
	g, err := NewGasScheduleNotifier(args)
	assert.Equal(t, err, core.ErrInvalidGasScheduleConfig)
	assert.Nil(t, g)

	args = createGasScheduleNotifierArgs()
	args.EpochNotifier = nil
	g, err = NewGasScheduleNotifier(args)
	assert.Equal(t, err, core.ErrNilEpochStartNotifier)
	assert.Nil(t, g)

	args = createGasScheduleNotifierArgs()
	args.ConfigDir = ""
	g, err = NewGasScheduleNotifier(args)
	assert.NotNil(t, err)
	assert.Nil(t, g)

	args = createGasScheduleNotifierArgs()
	args.WasmVMChangeLocker = nil
	g, err = NewGasScheduleNotifier(args)
	assert.Equal(t, err, common.ErrNilWasmChangeLocker)
	assert.Nil(t, g)
}

func TestNewGasScheduleNotifier(t *testing.T) {
	t.Parallel()

	args := createGasScheduleNotifierArgs()
	g, err := NewGasScheduleNotifier(args)
	assert.Nil(t, err)
	assert.NotNil(t, g)
}

func TestGasScheduleNotifier_RegisterNotifyHandlerNilHandlerShouldNotAdd(t *testing.T) {
	t.Parallel()

	args := createGasScheduleNotifierArgs()
	g, err := NewGasScheduleNotifier(args)
	assert.Nil(t, err)

	g.RegisterNotifyHandler(nil)
	assert.Equal(t, 0, len(g.handlers))
}

func TestGasScheduleNotifier_RegisterNotifyHandlerShouldWork(t *testing.T) {
	t.Parallel()

	args := createGasScheduleNotifierArgs()
	g, err := NewGasScheduleNotifier(args)
	assert.Nil(t, err)

	initialConfirmation := false
	handler := &mock.GasScheduleSubscribeHandlerStub{
		GasScheduleChangeCalled: func(_ map[string]map[string]uint64) {
			initialConfirmation = true
		},
	}

	g.RegisterNotifyHandler(handler)
	assert.Equal(t, 1, len(g.handlers))
	assert.True(t, g.handlers[0] == handler) //pointer testing
	assert.True(t, initialConfirmation)
}

func TestGasScheduleNotifier_UnregisterAllShouldWork(t *testing.T) {
	t.Parallel()

	args := createGasScheduleNotifierArgs()
	g, err := NewGasScheduleNotifier(args)
	assert.Nil(t, err)
	g.RegisterNotifyHandler(&mock.GasScheduleSubscribeHandlerStub{})
	g.RegisterNotifyHandler(&mock.GasScheduleSubscribeHandlerStub{})

	assert.Equal(t, 2, len(g.handlers))

	g.UnRegisterAll()

	assert.Equal(t, 0, len(g.handlers))
}

func TestGasScheduleNotifier_CheckEpochSameEpochShouldNotCall(t *testing.T) {
	t.Parallel()

	args := createGasScheduleNotifierArgs()
	g, err := NewGasScheduleNotifier(args)
	assert.Nil(t, err)
	numCalls := uint32(0)
	g.RegisterNotifyHandler(&mock.GasScheduleSubscribeHandlerStub{
		GasScheduleChangeCalled: func(_ map[string]map[string]uint64) {
			atomic.AddUint32(&numCalls, 1)
		},
	})

	g.EpochConfirmed(0, 0)
	g.EpochConfirmed(0, 0)

	assert.Equal(t, uint32(1), atomic.LoadUint32(&numCalls))
}

func TestGasScheduleNotifier_CheckEpochShouldCall(t *testing.T) {
	t.Parallel()

	args := createGasScheduleNotifierArgs()
	g, err := NewGasScheduleNotifier(args)
	assert.Nil(t, err)
	newEpoch := uint32(839843)
	numCalled := uint32(0)
	g.RegisterNotifyHandler(&mock.GasScheduleSubscribeHandlerStub{
		GasScheduleChangeCalled: func(gasMap map[string]map[string]uint64) {
			atomic.AddUint32(&numCalled, 1)

			if numCalled == 2 {
				assert.Equal(t, gasMap["BaseOperationCost"]["AoTPreparePerByte"], uint64(100))
			} else {
				assert.Equal(t, gasMap["BaseOperationCost"]["AoTPreparePerByte"], uint64(50))
			}
		},
	})

	g.EpochConfirmed(newEpoch, 0)

	assert.Equal(t, uint32(2), atomic.LoadUint32(&numCalled))
	assert.Equal(t, newEpoch, g.currentEpoch)
	assert.Equal(t, uint64(100), g.LatestGasSchedule()["BaseOperationCost"]["AoTPreparePerByte"])
	assert.Equal(t, uint64(100), g.LatestGasScheduleCopy()["BaseOperationCost"]["AoTPreparePerByte"])
}

func TestGasScheduleNotifier_CheckEpochInSyncShouldWork(t *testing.T) {
	t.Parallel()

	args := createGasScheduleNotifierArgs()
	g, err := NewGasScheduleNotifier(args)
	assert.Nil(t, err)
	newEpoch := uint32(839843)

	handlerWait := time.Second
	numCalls := uint32(0)

	handler := &mock.GasScheduleSubscribeHandlerStub{
		GasScheduleChangeCalled: func(gasMap map[string]map[string]uint64) {
			time.Sleep(handlerWait)
			atomic.AddUint32(&numCalls, 1)
		},
	}
	g.RegisterNotifyHandler(handler)

	start := time.Now()
	g.EpochConfirmed(newEpoch, 0)
	end := time.Now()

	assert.Equal(t, uint32(2), atomic.LoadUint32(&numCalls))
	assert.True(t, end.Sub(start) >= handlerWait)
}

func TestGasScheduleNotifier_EpochConfirmedShouldNotCauseDeadlock(t *testing.T) {
	t.Parallel()

	for i := 0; i < 100; i++ {
		testGasScheduleNotifierDeadlock(t)
	}
}

func testGasScheduleNotifierDeadlock(t *testing.T) {
	args := createGasScheduleNotifierArgs()
	g, _ := NewGasScheduleNotifier(args)

	chFinish := make(chan struct{})
	go func() {
		time.Sleep(time.Millisecond * 10)

		args.WasmVMChangeLocker.Lock()
		_ = g.LatestGasSchedule()
		args.WasmVMChangeLocker.Unlock()

		close(chFinish)
	}()

	go func() {
		time.Sleep(time.Millisecond * 10)

		g.EpochConfirmed(2, 0)
	}()

	select {
	case <-chFinish:
	case <-time.After(time.Second):
		require.Fail(t, "deadlock detected in EpochConfirmed function")
	}
}

func TestGasScheduleNotifier_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var g *gasScheduleNotifier
	require.True(t, g.IsInterfaceNil())

	g, _ = NewGasScheduleNotifier(createGasScheduleNotifierArgs())
	require.False(t, g.IsInterfaceNil())
}
