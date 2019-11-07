package smartContract

import (
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

func TestNewSCDataGetter_NilVmShouldErr(t *testing.T) {
	t.Parallel()

	target, err := NewSCDataGetter(nil)

	assert.Nil(t, target)
	assert.Equal(t, process.ErrNoVM, err)
}

func TestNewSCDataGetter_ShouldWork(t *testing.T) {
	t.Parallel()

	target, err := NewSCDataGetter(&mock.VMContainerMock{})

	assert.NotNil(t, target)
	assert.Nil(t, err)
}

func TestRunAndGetVMOutput_GetNilAddressShouldErr(t *testing.T) {
	t.Parallel()

	target, _ := NewSCDataGetter(&mock.VMContainerMock{})

	command := CommandRunFunction{
		ScAddress: nil,
		FuncName:  "function",
		Arguments: []*big.Int{},
	}

	output, err := target.RunAndGetVMOutput(&command)

	assert.Nil(t, output)
	assert.Equal(t, process.ErrNilScAddress, err)
}

func TestRunAndGetVMOutput_EmptyFunctionShouldErr(t *testing.T) {
	t.Parallel()

	target, _ := NewSCDataGetter(&mock.VMContainerMock{})

	command := CommandRunFunction{
		ScAddress: []byte{0},
		FuncName:  "",
		Arguments: []*big.Int{},
	}

	output, err := target.RunAndGetVMOutput(&command)

	assert.Nil(t, output)
	assert.Equal(t, process.ErrEmptyFunctionName, err)
}

func TestRunAndGetVMOutput_ShouldReceiveCommandCorrectly(t *testing.T) {
	t.Parallel()

	funcName := "function"
	scAddress := []byte(DummyScAddress)
	args := []*big.Int{big.NewInt(42), big.NewInt(43)}
	runWasCalled := false

	mockVM := &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
			runWasCalled = true
			assert.Equal(t, int64(42), input.Arguments[0].Int64())
			assert.Equal(t, int64(43), input.Arguments[1].Int64())
			assert.Equal(t, scAddress, input.CallerAddr)
			assert.Equal(t, funcName, input.Function)

			return &vmcommon.VMOutput{
				ReturnCode: vmcommon.Ok,
			}, nil
		},
	}

	target, _ := NewSCDataGetter(
		&mock.VMContainerMock{
			GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
				return mockVM, nil
			},
		},
	)

	command := CommandRunFunction{
		ScAddress: scAddress,
		FuncName:  funcName,
		Arguments: args,
	}

	_, _ = target.RunAndGetVMOutput(&command)
	assert.True(t, runWasCalled)
}

func TestRunAndGetVMOutput_ReturnsCorrectly(t *testing.T) {
	t.Parallel()

	data := []*big.Int{big.NewInt(90), big.NewInt(91)}

	mockVM := &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
			return &vmcommon.VMOutput{
				ReturnCode: vmcommon.Ok,
				ReturnData: data,
			}, nil
		},
	}

	target, _ := NewSCDataGetter(
		&mock.VMContainerMock{
			GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
				return mockVM, nil
			},
		},
	)

	command := CommandRunFunction{
		ScAddress: []byte(DummyScAddress),
		FuncName:  "function",
		Arguments: []*big.Int{},
	}

	vmOutput, err := target.RunAndGetVMOutput(&command)

	assert.Nil(t, err)
	assert.Equal(t, data[0], vmOutput.ReturnData[0])
	assert.Equal(t, data[1], vmOutput.ReturnData[1])
}

func TestRunAndGetVMOutput_WhenNotOkCodeShouldErr(t *testing.T) {
	t.Parallel()

	mockVM := &mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
			return &vmcommon.VMOutput{
				ReturnCode: vmcommon.OutOfGas,
			}, nil
		},
	}
	target, _ := NewSCDataGetter(
		&mock.VMContainerMock{
			GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
				return mockVM, nil
			},
		},
	)

	command := CommandRunFunction{
		ScAddress: []byte(DummyScAddress),
		FuncName:  "function",
		Arguments: []*big.Int{},
	}

	returnedData, err := target.RunAndGetVMOutput(&command)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error running vm func")
	assert.Nil(t, returnedData)
}

func TestRunAndGetVMOutput_ShouldCallRunScSequentially(t *testing.T) {
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

	target, _ := NewSCDataGetter(
		&mock.VMContainerMock{
			GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
				return mockVM, nil
			},
		},
	)

	noOfGoRoutines := 1000
	wg := sync.WaitGroup{}
	wg.Add(noOfGoRoutines)
	for i := 0; i < noOfGoRoutines; i++ {
		go func() {
			command := CommandRunFunction{
				ScAddress: []byte(DummyScAddress),
				FuncName:  "function",
				Arguments: []*big.Int{},
			}

			_, _ = target.RunAndGetVMOutput(&command)
			wg.Done()
		}()
	}

	wg.Wait()
}
