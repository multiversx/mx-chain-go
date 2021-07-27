package coordinator

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	vmData "github.com/ElrondNetwork/elrond-go-core/data/vm"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-vm-common/builtInFunctions"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
	"github.com/stretchr/testify/assert"
)

func createMockArguments() ArgNewTxTypeHandler {
	esdtTransferParser, _ := parsers.NewESDTTransferParser(&mock.MarshalizerMock{})
	return ArgNewTxTypeHandler{
		PubkeyConverter:    createMockPubkeyConverter(),
		ShardCoordinator:   mock.NewMultiShardsCoordinatorMock(3),
		BuiltInFunctions:   builtInFunctions.NewBuiltInFunctionContainer(),
		ArgumentParser:     parsers.NewCallArgsParser(),
		EpochNotifier:      &mock.EpochNotifierStub{},
		ESDTTransferParser: esdtTransferParser,
	}
}

func createMockPubkeyConverter() *mock.PubkeyConverterMock {
	return mock.NewPubkeyConverterMock(32)
}

func TestNewTxTypeHandler_NilAddrConv(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.PubkeyConverter = nil
	tth, err := NewTxTypeHandler(arg)

	assert.Nil(t, tth)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
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
	arg.BuiltInFunctions = nil
	tth, err := NewTxTypeHandler(arg)

	assert.Nil(t, tth)
	assert.Equal(t, process.ErrNilBuiltInFunction, err)
}

func TestNewTxTypeHandler_NilEpochNotifier(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.EpochNotifier = nil
	tth, err := NewTxTypeHandler(arg)

	assert.Nil(t, tth)
	assert.Equal(t, process.ErrNilEpochNotifier, err)
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

	txTypeIn, txTypeCross := tth.ComputeTransactionType(nil)
	assert.Equal(t, process.InvalidTransaction, txTypeIn)
	assert.Equal(t, process.InvalidTransaction, txTypeCross)
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
	txTypeIn, txTypeCross := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.InvalidTransaction, txTypeIn)
	assert.Equal(t, process.InvalidTransaction, txTypeCross)
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

	txTypeIn, txTypeCross := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.InvalidTransaction, txTypeIn)
	assert.Equal(t, process.InvalidTransaction, txTypeCross)
}

func TestTxTypeHandler_ComputeTransactionTypeScDeployment(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, createMockPubkeyConverter().Len())
	tx.Data = []byte("data")
	tx.Value = big.NewInt(45)

	txTypeIn, txTypeCross := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.SCDeployment, txTypeIn)
	assert.Equal(t, process.SCDeployment, txTypeCross)
}

func TestTxTypeHandler_ComputeTransactionTypeBuiltInFunctionCallNftTransfer(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.BuiltInFunctions = builtInFunctions.NewBuiltInFunctionContainer()
	_ = arg.BuiltInFunctions.Add(core.BuiltInFunctionESDTNFTTransfer, &mock.BuiltInFunctionStub{})

	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	scAddress := bytes.Repeat([]byte{0}, core.NumInitCharactersForScAddress-core.VMTypeLen)
	scAddressSuffix := bytes.Repeat([]byte{1}, 32-len(scAddress))
	scAddress = append(scAddress, scAddressSuffix...)

	addr := bytes.Repeat([]byte{1}, arg.PubkeyConverter.Len())
	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = addr
	tx.RcvAddr = scAddress
	tx.Data = []byte(core.BuiltInFunctionESDTNFTTransfer +
		"@" + hex.EncodeToString([]byte("token")) +
		"@" + hex.EncodeToString([]byte("rcv")) +
		"@" + hex.EncodeToString([]byte("attr")) +
		"@" + hex.EncodeToString([]byte("attr")) +
		"@" + hex.EncodeToString(big.NewInt(10).Bytes()))

	tx.Value = big.NewInt(45)

	txTypeIn, txTypeCross := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.BuiltInFunctionCall, txTypeIn)
	assert.Equal(t, process.SCInvoking, txTypeCross)
}

func TestTxTypeHandler_ComputeTransactionTypeBuiltInFunctionCallEsdtTransfer(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()

	arg.BuiltInFunctions = builtInFunctions.NewBuiltInFunctionContainer()
	_ = arg.BuiltInFunctions.Add(core.BuiltInFunctionESDTTransfer, &mock.BuiltInFunctionStub{})

	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	addr := bytes.Repeat([]byte{1}, arg.PubkeyConverter.Len())
	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = addr
	tx.RcvAddr = addr
	tx.Data = []byte(core.BuiltInFunctionESDTTransfer +
		"@" + hex.EncodeToString([]byte("token")) +
		"@" + hex.EncodeToString([]byte("rcv")) +
		"@" + hex.EncodeToString(big.NewInt(10).Bytes()))
	tx.Value = big.NewInt(45)

	txTypeIn, txTypeCross := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.BuiltInFunctionCall, txTypeIn)
	assert.Equal(t, process.BuiltInFunctionCall, txTypeCross)
}

func TestTxTypeHandler_ComputeTransactionTypeRecv0AddressWrongTransaction(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("SRC")
	tx.RcvAddr = make([]byte, createMockPubkeyConverter().Len())
	tx.Data = nil
	tx.Value = big.NewInt(45)

	txTypeIn, txTypeCross := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.InvalidTransaction, txTypeIn)
	assert.Equal(t, process.InvalidTransaction, txTypeCross)
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

	txTypeIn, txTypeCross := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.SCInvoking, txTypeIn)
	assert.Equal(t, process.SCInvoking, txTypeCross)
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
	arg.PubkeyConverter = &mock.PubkeyConverterStub{
		LenCalled: func() int {
			return len(tx.RcvAddr)
		},
	}
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txTypeIn, txTypeCross := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.MoveBalance, txTypeIn)
	assert.Equal(t, process.MoveBalance, txTypeCross)
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
	arg.PubkeyConverter = &mock.PubkeyConverterStub{
		LenCalled: func() int {
			return len(tx.RcvAddr)
		},
	}
	builtIn := "builtIn"
	arg.BuiltInFunctions = builtInFunctions.NewBuiltInFunctionContainer()
	_ = arg.BuiltInFunctions.Add(builtIn, &mock.BuiltInFunctionStub{})
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txTypeIn, txTypeCross := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.BuiltInFunctionCall, txTypeIn)
	assert.Equal(t, process.BuiltInFunctionCall, txTypeCross)
}

func TestTxTypeHandler_ComputeTransactionTypeBuiltInFuncNotActiveMoveBalance(t *testing.T) {
	t.Parallel()

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("000")
	tx.RcvAddr = []byte("001")
	tx.Data = []byte("builtIn")
	tx.Value = big.NewInt(45)

	arg := createMockArguments()
	arg.PubkeyConverter = &mock.PubkeyConverterStub{
		LenCalled: func() int {
			return len(tx.RcvAddr)
		},
	}
	builtIn := "builtIn"
	arg.BuiltInFunctions = builtInFunctions.NewBuiltInFunctionContainer()
	_ = arg.BuiltInFunctions.Add(builtIn, &mock.BuiltInFunctionStub{IsActiveCalled: func() bool {
		return false
	}})
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txTypeIn, txTypeCross := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.MoveBalance, txTypeIn)
	assert.Equal(t, process.MoveBalance, txTypeCross)
}

func TestTxTypeHandler_ComputeTransactionTypeBuiltInFuncNotActiveSCCall(t *testing.T) {
	t.Parallel()

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("000")
	tx.RcvAddr = vm.ESDTSCAddress
	tx.Data = []byte("builtIn")
	tx.Value = big.NewInt(45)

	arg := createMockArguments()
	arg.PubkeyConverter = &mock.PubkeyConverterStub{
		LenCalled: func() int {
			return len(tx.RcvAddr)
		},
	}
	builtIn := "builtIn"
	arg.BuiltInFunctions = builtInFunctions.NewBuiltInFunctionContainer()
	_ = arg.BuiltInFunctions.Add(builtIn, &mock.BuiltInFunctionStub{IsActiveCalled: func() bool {
		return false
	}})
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txTypeIn, txTypeCross := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.SCInvoking, txTypeIn)
	assert.Equal(t, process.SCInvoking, txTypeCross)
}

func TestTxTypeHandler_ComputeTransactionTypeRelayedFunc(t *testing.T) {
	t.Parallel()

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("000")
	tx.RcvAddr = []byte("001")
	tx.Data = []byte(core.RelayedTransaction)
	tx.Value = big.NewInt(45)

	arg := createMockArguments()
	arg.PubkeyConverter = &mock.PubkeyConverterStub{
		LenCalled: func() int {
			return len(tx.RcvAddr)
		},
	}
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txTypeIn, txTypeCross := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.RelayedTx, txTypeIn)
	assert.Equal(t, process.RelayedTx, txTypeCross)
}

func TestTxTypeHandler_ComputeTransactionTypeRelayedV2Func(t *testing.T) {
	t.Parallel()

	tx := &transaction.Transaction{}
	tx.Nonce = 0
	tx.SndAddr = []byte("000")
	tx.RcvAddr = []byte("001")
	tx.Data = []byte(core.RelayedTransactionV2)
	tx.Value = big.NewInt(45)

	arg := createMockArguments()
	arg.PubkeyConverter = &mock.PubkeyConverterStub{
		LenCalled: func() int {
			return len(tx.RcvAddr)
		},
	}
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txTypeIn, txTypeCross := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.RelayedTxV2, txTypeIn)
	assert.Equal(t, process.RelayedTxV2, txTypeCross)

	tth.flagRelayedTxV2.Unset()
	txTypeIn, txTypeCross = tth.ComputeTransactionType(tx)
	assert.Equal(t, process.MoveBalance, txTypeIn)
	assert.Equal(t, process.MoveBalance, txTypeCross)
}

func TestTxTypeHandler_ComputeTransactionTypeForSCRCallBack(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{}
	tx.Nonce = 0
	tx.SndAddr = []byte("000")
	tx.RcvAddr = []byte("001")
	tx.Data = []byte("00")
	tx.CallType = vmData.AsynchronousCallBack
	tx.Value = big.NewInt(45)

	arg := createMockArguments()
	arg.PubkeyConverter = &mock.PubkeyConverterStub{
		LenCalled: func() int {
			return len(tx.RcvAddr)
		},
	}
	tth, err := NewTxTypeHandler(arg)

	assert.NotNil(t, tth)
	assert.Nil(t, err)

	txTypeIn, txTypeCross := tth.ComputeTransactionType(tx)
	assert.Equal(t, process.SCInvoking, txTypeIn)
	assert.Equal(t, process.SCInvoking, txTypeCross)
}
