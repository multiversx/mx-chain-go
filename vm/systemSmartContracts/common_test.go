package systemSmartContracts

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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
			GasPrice:    big.NewInt(0),
			GasProvided: big.NewInt(0),
			Header:      nil,
		},
		RecipientAddr: []byte("tralala"),
		Function:      "something",
	})

	assert.Equal(t, vm.ErrInputCallerAddrIsNil, err)
}

func TestCheckIfNil_NilCallValue(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(&vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("tralala"),
			Arguments:   nil,
			CallValue:   nil,
			GasPrice:    big.NewInt(0),
			GasProvided: big.NewInt(0),
			Header:      nil,
		},
		RecipientAddr: []byte("tralala"),
		Function:      "something",
	})

	assert.Equal(t, vm.ErrInputCallValueIsNil, err)
}

func TestCheckIfNil_NilGasPrice(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(&vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("tralala"),
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    nil,
			GasProvided: big.NewInt(0),
			Header:      nil,
		},
		RecipientAddr: []byte("tralala"),
		Function:      "something",
	})

	assert.Equal(t, vm.ErrInputGasPriceIsNil, err)
}

func TestCheckIfNil_NilGasProvided(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(&vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("tralala"),
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    big.NewInt(0),
			GasProvided: nil,
			Header:      nil,
		},
		RecipientAddr: []byte("tralala"),
		Function:      "something",
	})

	assert.Equal(t, vm.ErrInputGasProvidedIsNil, err)
}

func TestCheckIfNil_NilRecipientAddr(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(&vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("tralala"),
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    big.NewInt(0),
			GasProvided: big.NewInt(0),
			Header:      nil,
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
			CallerAddr:  []byte("tralala"),
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    big.NewInt(0),
			GasProvided: big.NewInt(0),
			Header:      nil,
		},
		RecipientAddr: []byte("tralala"),
		Function:      "",
	})

	assert.Equal(t, vm.ErrInputFunctionIsNil, err)
}

func TestCheckIfNil(t *testing.T) {
	t.Parallel()

	err := CheckIfNil(&vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallerAddr:  []byte("tralala"),
			Arguments:   nil,
			CallValue:   big.NewInt(0),
			GasPrice:    big.NewInt(0),
			GasProvided: big.NewInt(0),
			Header:      nil,
		},
		RecipientAddr: []byte("tralala"),
		Function:      "something",
	})

	assert.Nil(t, err)
}
