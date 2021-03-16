package builtInFunctions

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestEsdtNFTCreateRoleTransfer_Constructor(t *testing.T) {
	t.Parallel()

	e, err := NewESDTNFTCreateRoleTransfer(nil, &mock.AccountsStub{}, mock.NewMultiShardsCoordinatorMock(2))
	assert.Nil(t, e)
	assert.Equal(t, err, process.ErrNilMarshalizer)

	e, err = NewESDTNFTCreateRoleTransfer(&mock.MarshalizerMock{}, nil, mock.NewMultiShardsCoordinatorMock(2))
	assert.Nil(t, e)
	assert.Equal(t, err, process.ErrNilAccountsAdapter)

	e, err = NewESDTNFTCreateRoleTransfer(&mock.MarshalizerMock{}, &mock.AccountsStub{}, nil)
	assert.Nil(t, e)
	assert.Equal(t, err, process.ErrNilShardCoordinator)

	e, err = NewESDTNFTCreateRoleTransfer(&mock.MarshalizerMock{}, &mock.AccountsStub{}, mock.NewMultiShardsCoordinatorMock(2))
	assert.Nil(t, err)
	assert.NotNil(t, e)
	assert.False(t, e.IsInterfaceNil())

	e.SetNewGasConfig(&process.GasCost{})
}

func TestESDTNFTCreateRoleTransfer_ProcessWithErrors(t *testing.T) {
	t.Parallel()

	e, err := NewESDTNFTCreateRoleTransfer(&mock.MarshalizerMock{}, &mock.AccountsStub{}, mock.NewMultiShardsCoordinatorMock(2))
	assert.Nil(t, err)
	assert.NotNil(t, e)

	vmOutput, err := e.ProcessBuiltinFunction(nil, nil, nil)
	assert.Equal(t, err, process.ErrNilVmInput)
	assert.Nil(t, vmOutput)

	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, &vmcommon.ContractCallInput{})
	assert.Equal(t, err, process.ErrNilValue)
	assert.Nil(t, vmOutput)

	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, &vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(10)}})
	assert.Equal(t, err, process.ErrBuiltInFunctionCalledWithValue)
	assert.Nil(t, vmOutput)

	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, &vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(0)}})
	assert.Equal(t, err, process.ErrInvalidArguments)
	assert.Nil(t, vmOutput)

	vmInput := &vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(0)}}
	vmInput.Arguments = [][]byte{{1}, {2}}
	vmOutput, err = e.ProcessBuiltinFunction(&mock.UserAccountStub{}, nil, vmInput)
	assert.Equal(t, err, process.ErrInvalidArguments)
	assert.Nil(t, vmOutput)

	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Equal(t, err, process.ErrNilUserAccount)
	assert.Nil(t, vmOutput)
}
