package systemSmartContracts

import (
	"math/big"
	"testing"

	vmData "github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
)

func TestCheckIfNil_NilArgs(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(nil)

	assert.Equal(t, vm.ErrInputArgsIsNil, err)
}

func TestCheckIfNil_NilCallerAddr(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(&vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  nil,
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    0,
			GasProvided: 0,
			CallType:    vmData.DirectCall,
		},
		RecipientAddr: []byte("dummyAddress"),
		Function:      "something",
	})

	assert.Equal(t, vm.ErrInputCallerAddrIsNil, err)
}

func TestCheckIfNil_NilCallValue(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(&vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("dummyAddress"),
			Arguments:   nil,
			CallValue:   nil,
			GasPrice:    0,
			GasProvided: 0,
			CallType:    vmData.DirectCall,
		},
		RecipientAddr: []byte("dummyAddress"),
		Function:      "something",
	})

	assert.Equal(t, vm.ErrInputCallValueIsNil, err)
}

func TestCheckIfNil_NilRecipientAddr(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(&vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("dummyAddress"),
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    0,
			GasProvided: 0,
			CallType:    vmData.DirectCall,
		},
		RecipientAddr: nil,
		Function:      "something",
	})

	assert.Equal(t, vm.ErrInputRecipientAddrIsNil, err)
}

func TestCheckIfNil_NilFunction(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(&vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("dummyAddress"),
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    0,
			GasProvided: 0,
			CallType:    vmData.DirectCall,
		},
		RecipientAddr: []byte("dummyAddress"),
		Function:      "",
	})

	assert.Equal(t, vm.ErrInputFunctionIsNil, err)
}

func TestCheckIfNil(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(&vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("dummyAddress"),
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    0,
			GasProvided: 0,
			CallType:    vmData.DirectCall,
		},
		RecipientAddr: []byte("dummyAddress"),
		Function:      "something",
	})

	assert.Nil(t, err)
}
