package coordinator

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func createMockArguments() ArgNewTxTypeHandler {
	return ArgNewTxTypeHandler{
		AddressConverter: &mock.AddressConverterMock{},
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(3),
		BuiltInFuncNames: make(map[string]struct{}),
		ArgumentParser:   vmcommon.NewAtArgumentParser(),
	}
}

func TestNewTxTypeHandler_NilAddrConv(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.AddressConverter = nil
	tth, err := NewTxTypeHandler(arg)

	assert.Nil(t, tth)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewTxTypeHandler_NilShardCoord(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.ShardCoordinator = nil
	tth, err := NewTxTypeHandler(arg)

	assert.Nil(t, tth)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewTxTypeHandler_NilArgParser(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.ArgumentParser = nil
	tth, err := NewTxTypeHandler(arg)

	assert.Nil(t, tth)
	assert.Equal(t, process.ErrNilArgumentParser, err)
}

func TestNewTxTypeHandler_NilBuiltInFuncs(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.BuiltInFuncNames = nil
	tth, err := NewTxTypeHandler(arg)

	assert.Nil(t, tth)
	assert.Equal(t, process.ErrNilBuiltInFunction, err)
}

func TestNewTxTypeHandler_ValsOk(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)
	assert.False(t, tth.IsInterfaceNil())
}

func TestTxTypeHandler_ComputeTransactionTypeNil(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	_, err = tth.ComputeTransactionType(nil)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestTxTypeHandler_ComputeTransactionTypeNilTx(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte("DST")
	tx.Value = big.NewInt(45)

	tx = nil
	_, err = tth.ComputeTransactionType(tx)
	assert.Equal(t, process.ErrNilTransaction, err)
}

func TestTxTypeHandler_ComputeTransactionTypeErrWrongTransaction(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = nil
	tx.Value = big.NewInt(45)

	_, err = tth.ComputeTransactionType(tx)
	assert.Equal(t, process.ErrWrongTransaction, err)
}

func TestTxTypeHandler_ComputeTransactionTypeScDeployment(t *testing.T) {
	t.Parallel()

	addressConverter := &mock.AddressConverterMock{}
	arg := createMockArguments()
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, addressConverter.AddressLen())
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	txType, err := tth.ComputeTransactionType(tx)
	assert.Nil(t, err)
	assert.Equal(t, process.SCDeployment, txType)
}

func TestTxTypeHandler_ComputeTransactionTypeScInvoking(t *testing.T) {
	t.Parallel()

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255}
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	arg := createMockArguments()
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txType, err := tth.ComputeTransactionType(tx)
	assert.Nil(t, err)
	assert.Equal(t, process.SCInvoking, txType)
}

func TestTxTypeHandler_ComputeTransactionTypeMoveBalance(t *testing.T) {
	t.Parallel()

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("000")
	tx.RcvAddr = []byte("001")
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	arg := createMockArguments()
	arg.AddressConverter = &mock.AddressConverterStub{
		AddressLenCalled: func() int {
			return len(tx.RcvAddr)
		},
		CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, err error) {
			return mock.NewAddressMock(pubKey), nil
		}}
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txType, err := tth.ComputeTransactionType(tx)
	assert.Nil(t, err)
	assert.Equal(t, process.MoveBalance, txType)
}

func TestTxTypeHandler_ComputeTransactionTypeBuiltInFunc(t *testing.T) {
	t.Parallel()

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("000")
	tx.RcvAddr = []byte("001")
	tx.Data = []byte("builtIn")
	tx.Value = big.NewInt(45)

	arg := createMockArguments()
	arg.AddressConverter = &mock.AddressConverterStub{
		AddressLenCalled: func() int {
			return len(tx.RcvAddr)
		},
		CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, err error) {
			return mock.NewAddressMock(pubKey), nil
		}}
	builtIn := "builtIn"
	arg.BuiltInFuncNames[builtIn] = struct{}{}
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txType, err := tth.ComputeTransactionType(tx)
	assert.Nil(t, err)
	assert.Equal(t, process.SCInvoking, txType)
}
