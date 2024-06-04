package transactionsfee

import (
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/defaults"
	"github.com/stretchr/testify/require"
)

func TestNewBuiltInFunctionsCost(t *testing.T) {
	t.Parallel()

	builtInCostHandler, err := NewBuiltInFunctionsCost(nil)
	require.Nil(t, builtInCostHandler)
	require.Equal(t, err, process.ErrNilGasSchedule)
}

func TestBuiltInFunctionsCost_GetESDTTransferBuiltInCost(t *testing.T) {
	t.Parallel()

	builtInCostHandler, err := NewBuiltInFunctionsCost(testscommon.NewGasScheduleNotifierMock(defaults.FillGasMapInternal(map[string]map[string]uint64{}, 100)))
	require.Nil(t, err)

	require.Equal(t, uint64(100), builtInCostHandler.GetESDTTransferBuiltInCost())
}

func TestBuiltInFunctionsCost_Concurrent(t *testing.T) {
	gs := defaults.FillGasMapInternal(map[string]map[string]uint64{}, 100)
	builtInCostHandler, err := NewBuiltInFunctionsCost(testscommon.NewGasScheduleNotifierMock(gs))
	require.Nil(t, err)

	wg := sync.WaitGroup{}
	numCalls := 100
	wg.Add(numCalls)
	for i := 0; i < numCalls; i++ {
		go func(idx int) {
			switch idx % 2 {
			case 0:
				builtInCostHandler.GasScheduleChange(gs)
			case 1:
				builtInCostHandler.GetESDTTransferBuiltInCost()
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}
