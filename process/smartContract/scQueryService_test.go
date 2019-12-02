package smartContract

import (
	"math"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

const DummyScAddress = "00000000000000000500fabd9501b7e5353de57a4e319857c2fb99089770720a"

func TestNewSCQueryService_NilVmShouldErr(t *testing.T) {
	t.Parallel()

	target, err := NewSCQueryService(nil, uint64(math.MaxUint64))

	assert.Nil(t, target)
	assert.Equal(t, process.ErrNoVM, err)
}

func TestNewSCQueryService_ShouldWork(t *testing.T) {
	t.Parallel()

	target, err := NewSCQueryService(&mock.VMContainerMock{}, uint64(math.MaxUint64))

	assert.NotNil(t, target)
	assert.Nil(t, err)
}

func TestExecuteQuery_GetNilAddressShouldErr(t *testing.T) {
	t.Parallel()

	target, _ := NewSCQueryService(&mock.VMContainerMock{}, uint64(math.MaxUint64))

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

	target, _ := NewSCQueryService(&mock.VMContainerMock{}, uint64(math.MaxUint64))

	query := process.SCQuery{
		ScAddress: []byte{0},
		FuncName:  "",
		Arguments: [][]byte{},
	}

	output, err := target.ExecuteQuery(&query)

	assert.Nil(t, output)
	assert.Equal(t, process.ErrEmptyFunctionName, err)
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

	target, _ := NewSCQueryService(
		&mock.VMContainerMock{
			GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
				return mockVM, nil
			},
		},
		uint64(math.MaxUint64),
	)

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

	data := [][]byte{[]byte("90"), []byte("91")}

	mockVM := &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
			return &vmcommon.VMOutput{
				ReturnCode: vmcommon.Ok,
				ReturnData: data,
			}, nil
		},
	}

	target, _ := NewSCQueryService(
		&mock.VMContainerMock{
			GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
				return mockVM, nil
			},
		},
		uint64(math.MaxUint64),
	)

	query := process.SCQuery{
		ScAddress: []byte(DummyScAddress),
		FuncName:  "function",
		Arguments: [][]byte{},
	}

	vmOutput, err := target.ExecuteQuery(&query)

	assert.Nil(t, err)
	assert.Equal(t, data[0], vmOutput.ReturnData[0])
	assert.Equal(t, data[1], vmOutput.ReturnData[1])
}

func TestExecuteQuery_WhenNotOkCodeShouldErr(t *testing.T) {
	t.Parallel()

	mockVM := &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
			return &vmcommon.VMOutput{
				ReturnCode: vmcommon.OutOfGas,
			}, nil
		},
	}
	target, _ := NewSCQueryService(
		&mock.VMContainerMock{
			GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
				return mockVM, nil
			},
		},
		uint64(math.MaxUint64),
	)

	query := process.SCQuery{
		ScAddress: []byte(DummyScAddress),
		FuncName:  "function",
		Arguments: [][]byte{},
	}

	returnedData, err := target.ExecuteQuery(&query)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error running vm func")
	assert.Nil(t, returnedData)
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

	target, _ := NewSCQueryService(
		&mock.VMContainerMock{
			GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
				return mockVM, nil
			},
		},
		uint64(math.MaxUint64),
	)

	noOfGoRoutines := 1000
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
