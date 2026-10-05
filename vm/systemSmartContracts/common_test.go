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

func TestIsNumArgsCorrectToStake(t *testing.T) {
	t.Parallel()

	// Empty args
	assert.False(t, isNumArgsCorrectToStake(nil))
	assert.False(t, isNumArgsCorrectToStake([][]byte{}))

	// Huge args[0] exceeding uint64 (e.g. 32 bytes 0xFF)
	hugeBytes := make([]byte, 32)
	for i := range hugeBytes {
		hugeBytes[i] = 0xff
	}
	assert.False(t, isNumArgsCorrectToStake([][]byte{hugeBytes}))

	// maxNodesToRun overflow boundary (math.MaxUint64)
	maxUint64Bytes := big.NewInt(0).SetUint64(^uint64(0)).Bytes()
	assert.False(t, isNumArgsCorrectToStake([][]byte{maxUint64Bytes}))

	// 1 node: min args = 2*1+1 = 3, max args = 2*1+1+2 = 5
	oneNodeBytes := big.NewInt(1).Bytes()
	assert.False(t, isNumArgsCorrectToStake([][]byte{oneNodeBytes}))                     // len 1 -> false
	assert.False(t, isNumArgsCorrectToStake([][]byte{oneNodeBytes, {1}}))                // len 2 -> false
	assert.True(t, isNumArgsCorrectToStake([][]byte{oneNodeBytes, {1}, {2}}))            // len 3 -> true
	assert.True(t, isNumArgsCorrectToStake([][]byte{oneNodeBytes, {1}, {2}, {3}}))       // len 4 -> true
	assert.True(t, isNumArgsCorrectToStake([][]byte{oneNodeBytes, {1}, {2}, {3}, {4}}))  // len 5 -> true
	assert.False(t, isNumArgsCorrectToStake([][]byte{oneNodeBytes, {1}, {2}, {3}, {4}, {5}})) // len 6 -> false
}

