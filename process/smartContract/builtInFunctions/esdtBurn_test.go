package builtInFunctions

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func TestESDTBurn_ProcessBuiltInFunctionErrors(t *testing.T) {
	t.Parallel()

	pauseHandler := &mock.PauseHandlerStub{}
	esdt, _ := NewESDTBurnFunc(10, &mock.MarshalizerMock{}, pauseHandler)
	_, err := esdt.ProcessBuiltinFunction(nil, nil, nil)
	assert.Equal(t, err, process.ErrNilVmInput)

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue: big.NewInt(0),
		},
	}
	_, err = esdt.ProcessBuiltinFunction(nil, nil, input)
	assert.Equal(t, err, process.ErrInvalidArguments)

	input = &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
	}
	key := []byte("key")
	value := []byte("value")
	input.Arguments = [][]byte{key, value}
	_, err = esdt.ProcessBuiltinFunction(nil, nil, input)
	assert.Equal(t, err, process.ErrAddressIsNotESDTSystemSC)

	input.RecipientAddr = vm.ESDTSCAddress
	input.GasProvided = esdt.funcGasCost - 1
	accSnd, _ := state.NewUserAccount([]byte("dst"))
	_, err = esdt.ProcessBuiltinFunction(accSnd, nil, input)
	assert.Equal(t, err, process.ErrNotEnoughGas)

	_, err = esdt.ProcessBuiltinFunction(nil, nil, input)
	assert.Equal(t, err, process.ErrNilUserAccount)

	pauseHandler.IsPausedCalled = func(token []byte) bool {
		return true
	}
	input.GasProvided = esdt.funcGasCost
	_, err = esdt.ProcessBuiltinFunction(accSnd, nil, input)
	assert.Equal(t, err, process.ErrESDTTokenIsPaused)
}

func TestESDTBurn_ProcessBuiltInFunctionSenderBurns(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	pauseHandler := &mock.PauseHandlerStub{}
	esdt, _ := NewESDTBurnFunc(10, marshalizer, pauseHandler)

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
		RecipientAddr: vm.ESDTSCAddress,
	}
	key := []byte("key")
	value := big.NewInt(10).Bytes()
	input.Arguments = [][]byte{key, value}
	accSnd, _ := state.NewUserAccount([]byte("snd"))

	esdtFrozen := ESDTUserMetadata{Frozen: true}
	esdtNotFrozen := ESDTUserMetadata{Frozen: false}

	esdtKey := append(esdt.keyPrefix, key...)
	esdtToken := &ESDigitalToken{Value: big.NewInt(100), Properties: esdtFrozen.ToBytes()}
	marshaledData, _ := marshalizer.Marshal(esdtToken)
	accSnd.DataTrieTracker().SaveKeyValue(esdtKey, marshaledData)

	_, err := esdt.ProcessBuiltinFunction(accSnd, nil, input)
	assert.Equal(t, err, process.ErrESDTIsFrozenForAccount)

	pauseHandler.IsPausedCalled = func(token []byte) bool {
		return true
	}
	esdtToken = &ESDigitalToken{Value: big.NewInt(100), Properties: esdtNotFrozen.ToBytes()}
	marshaledData, _ = marshalizer.Marshal(esdtToken)
	accSnd.DataTrieTracker().SaveKeyValue(esdtKey, marshaledData)

	_, err = esdt.ProcessBuiltinFunction(accSnd, nil, input)
	assert.Equal(t, err, process.ErrESDTTokenIsPaused)

	pauseHandler.IsPausedCalled = func(token []byte) bool {
		return false
	}
	_, err = esdt.ProcessBuiltinFunction(accSnd, nil, input)
	assert.Nil(t, err)

	marshaledData, _ = accSnd.DataTrieTracker().RetrieveValue(esdtKey)
	_ = marshalizer.Unmarshal(esdtToken, marshaledData)
	assert.True(t, esdtToken.Value.Cmp(big.NewInt(90)) == 0)

	value = big.NewInt(100).Bytes()
	input.Arguments = [][]byte{key, value}
	_, err = esdt.ProcessBuiltinFunction(accSnd, nil, input)
	assert.Equal(t, err, process.ErrInsufficientFunds)

	value = big.NewInt(90).Bytes()
	input.Arguments = [][]byte{key, value}
	_, err = esdt.ProcessBuiltinFunction(accSnd, nil, input)
	assert.Nil(t, err)

	marshaledData, _ = accSnd.DataTrieTracker().RetrieveValue(esdtKey)
	_ = marshalizer.Unmarshal(esdtToken, marshaledData)
	assert.True(t, esdtToken.Value.Cmp(big.NewInt(0)) == 0)
}
