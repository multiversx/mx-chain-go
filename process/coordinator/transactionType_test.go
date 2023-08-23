package coordinator

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmData "github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	"github.com/stretchr/testify/assert"
)

func createMockArguments() ArgNewTxTypeHandler {
	esdtTransferParser, _ := parsers.NewESDTTransferParser(&mock.MarshalizerMock{})
	return ArgNewTxTypeHandler{
		PubkeyConverter:    createMockPubkeyConverter(),
		ShardCoordinator:   mock.NewMultiShardsCoordinatorMock(3),
		BuiltInFunctions:   builtInFunctions.NewBuiltInFunctionContainer(),
		ArgumentParser:     parsers.NewCallArgsParser(),
		ESDTTransferParser: esdtTransferParser,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.ESDTMetadataContinuousCleanupFlag
			},
		},
	}
}

func createMockPubkeyConverter() *testscommon.PubkeyConverterMock {
	return testscommon.NewPubkeyConverterMock(32)
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
	arg.PubkeyConverter = &testscommon.PubkeyConverterStub{
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
	arg.PubkeyConverter = &testscommon.PubkeyConverterStub{
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
	arg.PubkeyConverter = &testscommon.PubkeyConverterStub{
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
	arg.PubkeyConverter = &testscommon.PubkeyConverterStub{
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
	arg.PubkeyConverter = &testscommon.PubkeyConverterStub{
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
	arg.PubkeyConverter = &testscommon.PubkeyConverterStub{
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
	arg.PubkeyConverter = &testscommon.PubkeyConverterStub{
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
