package smartContract

import (
	"bytes"
	"errors"
	"math"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const DummyScAddress = "00000000000000000500fabd9501b7e5353de57a4e319857c2fb99089770720a"

func createMockArgumentsForSCQuery() ArgsNewSCQueryService {
	return ArgsNewSCQueryService{
		VmContainer:              &mock.VMContainerMock{},
		EconomicsFee:             &mock.FeeHandlerStub{},
		BlockChainHook:           &testscommon.BlockChainHookStub{},
		BlockChain:               &testscommon.ChainHandlerStub{},
		WasmVMChangeLocker:       &sync.RWMutex{},
		Bootstrapper:             &mock.BootstrapperStub{},
		AllowExternalQueriesChan: common.GetClosedUnbufferedChannel(),
	}
}

func TestNewSCQueryService_NilVmShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForSCQuery()
	args.VmContainer = nil
	target, err := NewSCQueryService(args)

	assert.Nil(t, target)
	assert.Equal(t, process.ErrNoVM, err)
}

func TestNewSCQueryService_NilFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForSCQuery()
	args.EconomicsFee = nil
	target, err := NewSCQueryService(args)

	assert.Nil(t, target)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewSCQueryService_NilBLockChainShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForSCQuery()
	args.BlockChain = nil
	target, err := NewSCQueryService(args)

	assert.Nil(t, target)
	assert.Equal(t, process.ErrNilBlockChain, err)
}

func TestNewSCQueryService_NilBLockChainHookShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForSCQuery()
	args.BlockChainHook = nil
	target, err := NewSCQueryService(args)

	assert.Nil(t, target)
	assert.Equal(t, process.ErrNilBlockChainHook, err)
}

func TestNewSCQueryService_NilWasmVMLockerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForSCQuery()
	args.WasmVMChangeLocker = nil
	target, err := NewSCQueryService(args)

	assert.Nil(t, target)
	assert.Equal(t, process.ErrNilLocker, err)
}

func TestNewSCQueryService_NilBootstrapperShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForSCQuery()
	args.Bootstrapper = nil
	target, err := NewSCQueryService(args)

	assert.Nil(t, target)
	assert.Equal(t, process.ErrNilBootstrapper, err)
}

func TestNewSCQueryService_ShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForSCQuery()
	target, err := NewSCQueryService(args)

	assert.NotNil(t, target)
	assert.Nil(t, err)
	assert.False(t, target.IsInterfaceNil())
}

func TestExecuteQuery_GetNilAddressShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForSCQuery()
	target, _ := NewSCQueryService(args)

	query := process.SCQuery{
		ScAddress: nil,
		FuncName:  "function",
		Arguments: [][]byte{},
	}

	output, err := target.ExecuteQuery(&query)

	assert.Nil(t, output)
	assert.Equal(t, process.ErrNilScAddress, err)
}

func TestExecuteQuery_EmptyFunctionShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForSCQuery()
	target, _ := NewSCQueryService(args)

	query := process.SCQuery{
		ScAddress: []byte{0},
		FuncName:  "",
		Arguments: [][]byte{},
	}

	output, err := target.ExecuteQuery(&query)

	assert.Nil(t, output)
	assert.Equal(t, process.ErrEmptyFunctionName, err)
}

func TestExecuteQuery_ShouldPerformActionsInRegardsToAllowanceChannel(t *testing.T) {
	t.Parallel()

	chanAllowedQueries := make(chan struct{})
	args := createMockArgumentsForSCQuery()
	args.AllowExternalQueriesChan = chanAllowedQueries
	target, _ := NewSCQueryService(args)

	query := process.SCQuery{
		ScAddress: []byte(DummyScAddress),
		FuncName:  "func",
		Arguments: [][]byte{},
	}

	output, err := target.ExecuteQuery(&query)
	assert.Equal(t, process.ErrQueriesNotAllowedYet, err)
	assert.Nil(t, output)

	close(chanAllowedQueries)
	_, err = target.ExecuteQuery(&query)
	assert.NoError(t, err)
}

func TestExecuteQuery_AllowanceChannelShouldWorkUnderConcurrentRequests(t *testing.T) {
	t.Parallel()

	chanAllowedQueries := make(chan struct{})
	args := createMockArgumentsForSCQuery()
	args.AllowExternalQueriesChan = chanAllowedQueries
	target, _ := NewSCQueryService(args)

	query := process.SCQuery{
		ScAddress: []byte(DummyScAddress),
		FuncName:  "func",
		Arguments: [][]byte{},
	}

	defer func() {
		r := recover()
		assert.Nil(t, r)
	}()

	numTries := 200
	wg := sync.WaitGroup{}
	wg.Add(numTries)

	go func() {
		time.Sleep(50 * time.Millisecond)
		close(chanAllowedQueries)
	}()

	for i := 0; i < numTries; i++ {
		go func(idx int) {
			select {
			case <-chanAllowedQueries:
				_, err := target.ExecuteQuery(&query)
				assert.NoError(t, err)
			default:
				output, err := target.ExecuteQuery(&query)
				assert.Equal(t, process.ErrQueriesNotAllowedYet, err)
				assert.Nil(t, output)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestExecuteQuery_ShouldReceiveQueryCorrectly(t *testing.T) {
	t.Parallel()

	funcName := "function"
	scAddress := []byte(DummyScAddress)
	args := []*big.Int{big.NewInt(42), big.NewInt(43)}
	runWasCalled := false

	mockVM := &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
			runWasCalled = true
			assert.Equal(t, int64(42), big.NewInt(0).SetBytes(input.Arguments[0]).Int64())
			assert.Equal(t, int64(43), big.NewInt(0).SetBytes(input.Arguments[1]).Int64())
			assert.Equal(t, scAddress, input.CallerAddr)
			assert.Equal(t, funcName, input.Function)

			return &vmcommon.VMOutput{
				ReturnCode: vmcommon.Ok,
			}, nil
		},
	}
	argsNewSCQuery := createMockArgumentsForSCQuery()
	argsNewSCQuery.VmContainer = &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return mockVM, nil
		},
	}
	argsNewSCQuery.EconomicsFee = &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return uint64(math.MaxUint64)
		},
	}

	target, _ := NewSCQueryService(argsNewSCQuery)

	dataArgs := make([][]byte, len(args))
	for i, arg := range args {
		dataArgs[i] = append(dataArgs[i], arg.Bytes()...)
	}
	query := process.SCQuery{
		ScAddress: scAddress,
		FuncName:  funcName,
		Arguments: dataArgs,
	}

	_, _ = target.ExecuteQuery(&query)
	assert.True(t, runWasCalled)
}

func TestExecuteQuery_ReturnsCorrectly(t *testing.T) {
	t.Parallel()

	d := [][]byte{[]byte("90"), []byte("91")}

	mockVM := &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
			return &vmcommon.VMOutput{
				ReturnCode: vmcommon.Ok,
				ReturnData: d,
			}, nil
		},
	}

	argsNewSCQuery := createMockArgumentsForSCQuery()
	argsNewSCQuery.VmContainer = &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return mockVM, nil
		},
	}
	argsNewSCQuery.EconomicsFee = &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return uint64(math.MaxUint64)
		},
	}

	target, _ := NewSCQueryService(argsNewSCQuery)

	query := process.SCQuery{
		ScAddress: []byte(DummyScAddress),
		FuncName:  "function",
		Arguments: [][]byte{},
	}

	vmOutput, err := target.ExecuteQuery(&query)

	assert.Nil(t, err)
	assert.Equal(t, d[0], vmOutput.ReturnData[0])
	assert.Equal(t, d[1], vmOutput.ReturnData[1])
}

func TestExecuteQuery_GasProvidedShouldBeApplied(t *testing.T) {
	t.Parallel()

	t.Run("no gas defined, should use max uint64", func(t *testing.T) {
		t.Parallel()

		runSCWasCalled := false
		mockVM := &mock.VMExecutionHandlerStub{
			RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
				require.Equal(t, uint64(math.MaxUint64), input.GasProvided)
				runSCWasCalled = true
				return &vmcommon.VMOutput{}, nil
			},
		}
		argsNewSCQuery := createMockArgumentsForSCQuery()
		argsNewSCQuery.VmContainer = &mock.VMContainerMock{
			GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
				return mockVM, nil
			},
		}
		argsNewSCQuery.EconomicsFee = &mock.FeeHandlerStub{
			MaxGasLimitPerBlockCalled: func() uint64 {
				return uint64(math.MaxUint64)
			},
		}

		target, _ := NewSCQueryService(argsNewSCQuery)

		query := process.SCQuery{
			ScAddress: []byte(DummyScAddress),
			FuncName:  "function",
			Arguments: [][]byte{},
		}

		_, err := target.ExecuteQuery(&query)
		require.Nil(t, err)
		require.True(t, runSCWasCalled)
	})

	t.Run("custom gas defined, should use it", func(t *testing.T) {
		t.Parallel()

		maxGasLimit := uint64(1_500_000_000)
		runSCWasCalled := false
		mockVM := &mock.VMExecutionHandlerStub{
			RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
				require.Equal(t, maxGasLimit, input.GasProvided)
				runSCWasCalled = true
				return &vmcommon.VMOutput{}, nil
			},
		}
		argsNewSCQuery := createMockArgumentsForSCQuery()
		argsNewSCQuery.VmContainer = &mock.VMContainerMock{
			GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
				return mockVM, nil
			},
		}
		argsNewSCQuery.EconomicsFee = &mock.FeeHandlerStub{
			MaxGasLimitPerBlockCalled: func() uint64 {
				return uint64(math.MaxUint64)
			},
		}

		argsNewSCQuery.MaxGasLimitPerQuery = maxGasLimit

		target, _ := NewSCQueryService(argsNewSCQuery)

		query := process.SCQuery{
			ScAddress: []byte(DummyScAddress),
			FuncName:  "function",
			Arguments: [][]byte{},
		}

		_, err := target.ExecuteQuery(&query)
		require.Nil(t, err)
		require.True(t, runSCWasCalled)
	})
}

func TestExecuteQuery_WhenNotOkCodeShouldNotErr(t *testing.T) {
	t.Parallel()

	mockVM := &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
			return &vmcommon.VMOutput{
				ReturnCode:    vmcommon.OutOfGas,
				ReturnMessage: "add more gas",
			}, nil
		},
	}
	argsNewSCQuery := createMockArgumentsForSCQuery()
	argsNewSCQuery.VmContainer = &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return mockVM, nil
		},
	}
	argsNewSCQuery.EconomicsFee = &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return uint64(math.MaxUint64)
		},
	}

	target, _ := NewSCQueryService(argsNewSCQuery)

	query := process.SCQuery{
		ScAddress: []byte(DummyScAddress),
		FuncName:  "function",
		Arguments: [][]byte{},
	}

	returnedData, err := target.ExecuteQuery(&query)

	assert.Nil(t, err)
	assert.NotNil(t, returnedData)
	assert.Contains(t, returnedData.ReturnMessage, "add more gas")
}

func TestExecuteQuery_ShouldCallRunScSequentially(t *testing.T) {
	t.Parallel()

	running := int32(0)

	mockVM := &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
			atomic.AddInt32(&running, 1)
			time.Sleep(time.Millisecond)

			val := atomic.LoadInt32(&running)
			assert.Equal(t, int32(1), val)

			atomic.AddInt32(&running, -1)

			return &vmcommon.VMOutput{
				ReturnCode: vmcommon.Ok,
			}, nil
		},
	}

	argsNewSCQuery := createMockArgumentsForSCQuery()
	argsNewSCQuery.VmContainer = &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return mockVM, nil
		},
	}
	argsNewSCQuery.EconomicsFee = &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return uint64(math.MaxUint64)
		},
	}
	target, _ := NewSCQueryService(argsNewSCQuery)

	noOfGoRoutines := 50
	wg := sync.WaitGroup{}
	wg.Add(noOfGoRoutines)
	for i := 0; i < noOfGoRoutines; i++ {
		go func() {
			query := process.SCQuery{
				ScAddress: []byte(DummyScAddress),
				FuncName:  "function",
				Arguments: [][]byte{},
			}

			_, _ = target.ExecuteQuery(&query)
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestSCQueryService_ExecuteQueryShouldNotIncludeCallerAddressAndValue(t *testing.T) {
	t.Parallel()

	callerAddressAndCallValueAreNotSet := false
	mockVM := &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
			if input.CallValue.Cmp(big.NewInt(0)) == 0 && bytes.Equal(input.CallerAddr, input.RecipientAddr) {
				callerAddressAndCallValueAreNotSet = true
			}
			return &vmcommon.VMOutput{
				ReturnCode: vmcommon.Ok,
				ReturnData: [][]byte{[]byte("ok")},
			}, nil
		},
	}

	argsNewSCQuery := createMockArgumentsForSCQuery()
	argsNewSCQuery.VmContainer = &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return mockVM, nil
		},
	}
	argsNewSCQuery.EconomicsFee = &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return uint64(math.MaxUint64)
		},
	}
	target, _ := NewSCQueryService(argsNewSCQuery)

	query := process.SCQuery{
		ScAddress: []byte(DummyScAddress),
		FuncName:  "function",
		Arguments: [][]byte{},
	}

	_, err := target.ExecuteQuery(&query)
	require.NoError(t, err)
	require.True(t, callerAddressAndCallValueAreNotSet)
}

func TestSCQueryService_ExecuteQueryShouldIncludeCallerAddressAndValue(t *testing.T) {
	t.Parallel()

	expectedCallerAddr := []byte("caller addr")
	expectedValue := big.NewInt(37)
	callerAddressAndCallValueAreSet := false
	mockVM := &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
			if input.CallValue.Cmp(expectedValue) == 0 && bytes.Equal(input.CallerAddr, expectedCallerAddr) {
				callerAddressAndCallValueAreSet = true
			}
			return &vmcommon.VMOutput{
				ReturnCode: vmcommon.Ok,
				ReturnData: [][]byte{[]byte("ok")},
			}, nil
		},
	}

	argsNewSCQuery := createMockArgumentsForSCQuery()
	argsNewSCQuery.VmContainer = &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return mockVM, nil
		},
	}
	argsNewSCQuery.EconomicsFee = &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return uint64(math.MaxUint64)
		},
	}
	target, _ := NewSCQueryService(argsNewSCQuery)

	query := process.SCQuery{
		ScAddress:  []byte(DummyScAddress),
		FuncName:   "function",
		CallerAddr: expectedCallerAddr,
		CallValue:  expectedValue,
		Arguments:  [][]byte{},
	}

	_, err := target.ExecuteQuery(&query)
	require.NoError(t, err)
	require.True(t, callerAddressAndCallValueAreSet)
}

func TestSCQueryService_ShouldFailIfNodeIsNotSynced(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForSCQuery()
	args.Bootstrapper = &mock.BootstrapperStub{
		GetNodeStateCalled: func() common.NodeState {
			return common.NsNotSynchronized
		},
	}

	qs, _ := NewSCQueryService(args)

	res, err := qs.ExecuteQuery(&process.SCQuery{
		ShouldBeSynced: true,
		ScAddress:      []byte(DummyScAddress),
		FuncName:       "function",
	})
	require.Nil(t, res)
	require.Equal(t, process.ErrNodeIsNotSynced, err)
}

func TestSCQueryService_ShouldWorkIfNodeIsSynced(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForSCQuery()
	args.Bootstrapper = &mock.BootstrapperStub{
		GetNodeStateCalled: func() common.NodeState {
			return common.NsSynchronized
		},
	}

	qs, _ := NewSCQueryService(args)

	res, err := qs.ExecuteQuery(&process.SCQuery{
		ShouldBeSynced: true,
		ScAddress:      []byte(DummyScAddress),
		FuncName:       "function",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestSCQueryService_ShouldFailIfStateChanged(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForSCQuery()

	rootHashCalled := false
	args.BlockChain = &testscommon.ChainHandlerStub{
		GetCurrentBlockRootHashCalled: func() []byte {
			if !rootHashCalled {
				rootHashCalled = true
				return []byte("first root hash")
			}

			return []byte("second root hash")
		},
	}

	qs, _ := NewSCQueryService(args)

	res, err := qs.ExecuteQuery(&process.SCQuery{
		SameScState: true,
		ScAddress:   []byte(DummyScAddress),
		FuncName:    "function",
	})
	require.Nil(t, res)
	require.True(t, errors.Is(err, process.ErrStateChangedWhileExecutingVmQuery))
}

func TestSCQueryService_ShouldWorkIfStateDidntChange(t *testing.T) {
	t.Parallel()

	args := createMockArgumentsForSCQuery()

	args.BlockChain = &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				RootHash: []byte("same root hash"),
			}

		},
	}

	qs, _ := NewSCQueryService(args)

	res, err := qs.ExecuteQuery(&process.SCQuery{
		SameScState: true,
		ScAddress:   []byte(DummyScAddress),
		FuncName:    "function",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestSCQueryService_ComputeTxCostScCall(t *testing.T) {
	t.Parallel()

	consumedGas := uint64(10000)
	mockVM := &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
			return &vmcommon.VMOutput{
				GasRemaining: uint64(math.MaxUint64) - consumedGas,
				ReturnCode:   vmcommon.Ok,
			}, nil
		},
	}

	argsNewSCQuery := createMockArgumentsForSCQuery()
	argsNewSCQuery.VmContainer = &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return mockVM, nil
		},
	}
	argsNewSCQuery.EconomicsFee = &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return uint64(math.MaxUint64)
		},
	}
	target, _ := NewSCQueryService(argsNewSCQuery)

	tx := &transaction.Transaction{
		RcvAddr: []byte(DummyScAddress),
		Data:    []byte("increment"),
	}
	cost, err := target.ComputeScCallGasLimit(tx)
	require.Nil(t, err)
	require.Equal(t, consumedGas, cost)
}

func TestSCQueryService_ComputeScCallGasLimitRetCodeNotOK(t *testing.T) {
	t.Parallel()

	message := "function not found"
	consumedGas := uint64(10000)
	mockVM := &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
			return &vmcommon.VMOutput{
				GasRemaining:  uint64(math.MaxUint64) - consumedGas,
				ReturnCode:    vmcommon.FunctionNotFound,
				ReturnMessage: message,
			}, nil
		},
	}

	argsNewSCQuery := createMockArgumentsForSCQuery()
	argsNewSCQuery.VmContainer = &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return mockVM, nil
		},
	}
	argsNewSCQuery.EconomicsFee = &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return uint64(math.MaxUint64)
		},
	}
	target, _ := NewSCQueryService(argsNewSCQuery)

	tx := &transaction.Transaction{
		RcvAddr: []byte(DummyScAddress),
		Data:    []byte("incrementaaaa"),
	}
	_, err := target.ComputeScCallGasLimit(tx)
	require.Equal(t, errors.New(message), err)
}

func TestNewSCQueryService_CloseShouldWork(t *testing.T) {
	t.Parallel()

	closeCalled := false
	argsNewSCQueryService := ArgsNewSCQueryService{
		VmContainer: &mock.VMContainerMock{
			CloseCalled: func() error {
				closeCalled = true
				return nil
			},
		},
		EconomicsFee:             &mock.FeeHandlerStub{},
		BlockChainHook:           &testscommon.BlockChainHookStub{},
		BlockChain:               &testscommon.ChainHandlerStub{},
		WasmVMChangeLocker:       &sync.RWMutex{},
		Bootstrapper:             &mock.BootstrapperStub{},
		AllowExternalQueriesChan: common.GetClosedUnbufferedChannel(),
	}

	target, _ := NewSCQueryService(argsNewSCQueryService)

	err := target.Close()
	assert.Nil(t, err)
	assert.True(t, closeCalled)
}
