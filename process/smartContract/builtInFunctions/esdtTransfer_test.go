package builtInFunctions

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func TestESDTTransfer_ProcessBuiltInFunctionErrors(t *testing.T) {
	t.Parallel()

	esdt, _ := NewESDTTransferFunc(10, &mock.MarshalizerMock{}, &mock.PauseHandlerStub{})
	_ = esdt.setPayableHandler(&mock.PayableHandlerStub{})
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
	assert.Nil(t, err)

	input.GasProvided = esdt.funcGasCost - 1
	accSnd := state.NewEmptyUserAccount()
	_, err = esdt.ProcessBuiltinFunction(accSnd, nil, input)
	assert.Equal(t, err, process.ErrNotEnoughGas)
}

func TestESDTTransfer_ProcessBuiltInFunctionSingleShard(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	esdt, _ := NewESDTTransferFunc(10, marshalizer, &mock.PauseHandlerStub{})
	_ = esdt.setPayableHandler(&mock.PayableHandlerStub{})

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
	}
	key := []byte("key")
	value := big.NewInt(10).Bytes()
	input.Arguments = [][]byte{key, value}
	accSnd, _ := state.NewUserAccount([]byte("snd"))
	accDst, _ := state.NewUserAccount([]byte("dst"))

	_, err := esdt.ProcessBuiltinFunction(accSnd, accDst, input)
	assert.Equal(t, err, process.ErrInsufficientFunds)

	esdtKey := append(esdt.keyPrefix, key...)
	esdtToken := &ESDigitalToken{Value: big.NewInt(100)}
	marshaledData, _ := marshalizer.Marshal(esdtToken)
	_ = accSnd.DataTrieTracker().SaveKeyValue(esdtKey, marshaledData)

	_, err = esdt.ProcessBuiltinFunction(accSnd, accDst, input)
	assert.Nil(t, err)
	marshaledData, _ = accSnd.DataTrieTracker().RetrieveValue(esdtKey)
	_ = marshalizer.Unmarshal(esdtToken, marshaledData)
	assert.True(t, esdtToken.Value.Cmp(big.NewInt(90)) == 0)

	marshaledData, _ = accDst.DataTrieTracker().RetrieveValue(esdtKey)
	_ = marshalizer.Unmarshal(esdtToken, marshaledData)
	assert.True(t, esdtToken.Value.Cmp(big.NewInt(10)) == 0)
}

func TestESDTTransfer_ProcessBuiltInFunctionSenderInShard(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	esdt, _ := NewESDTTransferFunc(10, marshalizer, &mock.PauseHandlerStub{})
	_ = esdt.setPayableHandler(&mock.PayableHandlerStub{})

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
	}
	key := []byte("key")
	value := big.NewInt(10).Bytes()
	input.Arguments = [][]byte{key, value}
	accSnd, _ := state.NewUserAccount([]byte("snd"))

	esdtKey := append(esdt.keyPrefix, key...)
	esdtToken := &ESDigitalToken{Value: big.NewInt(100)}
	marshaledData, _ := marshalizer.Marshal(esdtToken)
	_ = accSnd.DataTrieTracker().SaveKeyValue(esdtKey, marshaledData)

	_, err := esdt.ProcessBuiltinFunction(accSnd, nil, input)
	assert.Nil(t, err)
	marshaledData, _ = accSnd.DataTrieTracker().RetrieveValue(esdtKey)
	_ = marshalizer.Unmarshal(esdtToken, marshaledData)
	assert.True(t, esdtToken.Value.Cmp(big.NewInt(90)) == 0)
}

func TestESDTTransfer_ProcessBuiltInFunctionDestInShard(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	esdt, _ := NewESDTTransferFunc(10, marshalizer, &mock.PauseHandlerStub{})
	_ = esdt.setPayableHandler(&mock.PayableHandlerStub{})

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
	}
	key := []byte("key")
	value := big.NewInt(10).Bytes()
	input.Arguments = [][]byte{key, value}
	accDst, _ := state.NewUserAccount([]byte("dst"))

	_, err := esdt.ProcessBuiltinFunction(nil, accDst, input)
	assert.Nil(t, err)
	esdtKey := append(esdt.keyPrefix, key...)
	esdtToken := &ESDigitalToken{}
	marshaledData, _ := accDst.DataTrieTracker().RetrieveValue(esdtKey)
	_ = marshalizer.Unmarshal(esdtToken, marshaledData)
	assert.True(t, esdtToken.Value.Cmp(big.NewInt(10)) == 0)
}
