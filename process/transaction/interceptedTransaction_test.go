package transaction_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/versioning"
	"github.com/multiversx/mx-chain-core-go/data"
	dataTransaction "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/interceptors"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/transaction"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errSingleSignKeyGenMock = errors.New("errSingleSignKeyGenMock")
var errSignerMockVerifySigFails = errors.New("errSignerMockVerifySigFails")

var senderShard = uint32(2)
var recvShard = uint32(3)
var senderAddress = []byte("12345678901234567890123456789012")
var recvAddress = []byte("23456789012345678901234567890123")
var sigBad = []byte("bad-signature")
var sigOk = []byte("signature")

func createMockPubKeyConverter() *testscommon.PubkeyConverterMock {
	return testscommon.NewPubkeyConverterMock(32)
}

func createDummySigner() crypto.SingleSigner {
	return &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			if !bytes.Equal(sig, sigOk) {
				return errSignerMockVerifySigFails
			}
			return nil
		},
	}
}

func createKeyGenMock() crypto.KeyGenerator {
	return &mock.SingleSignKeyGenMock{
		PublicKeyFromByteArrayCalled: func(b []byte) (key crypto.PublicKey, e error) {
			if string(b) == "" {
				return nil, errSingleSignKeyGenMock
			}

			return &mock.SingleSignPublicKey{}, nil
		},
	}
}

func createFreeTxFeeHandler() *economicsmocks.EconomicsHandlerStub {
	return &economicsmocks.EconomicsHandlerStub{
		CheckValidityTxValuesCalled: func(tx data.TransactionWithFeeHandler) error {
			return nil
		},
	}
}

func createInterceptedTxWithTxFeeHandlerAndVersionChecker(tx *dataTransaction.Transaction, txFeeHandler process.FeeHandler, txVerChecker *testscommon.TxVersionCheckerStub) (*transaction.InterceptedTransaction, error) {
	marshaller := &marshallerMock.MarshalizerMock{}
	txBuff, err := marshaller.Marshal(tx)
	if err != nil {
		return nil, err
	}

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = 6
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, senderAddress) {
			return senderShard
		}
		if bytes.Equal(address, recvAddress) {
			return recvShard
		}

		return shardCoordinator.CurrentShard
	}

	return transaction.NewInterceptedTransaction(
		txBuff,
		marshaller,
		marshaller,
		&hashingMocks.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&testscommon.PubkeyConverterStub{
			LenCalled: func() int {
				return 32
			},
		},
		shardCoordinator,
		txFeeHandler,
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("T"),
		false,
		&hashingMocks.HasherMock{},
		txVerChecker,
	)
}

func createInterceptedTxFromPlainTx(tx *dataTransaction.Transaction, txFeeHandler process.FeeHandler, chainID []byte, minTxVersion uint32) (*transaction.InterceptedTransaction, error) {
	marshalizer := &mock.MarshalizerMock{}
	txBuff, err := marshalizer.Marshal(tx)
	if err != nil {
		return nil, err
	}

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = 6
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, senderAddress) {
			return senderShard
		}
		if bytes.Equal(address, recvAddress) {
			return recvShard
		}

		return shardCoordinator.CurrentShard
	}

	return transaction.NewInterceptedTransaction(
		txBuff,
		marshalizer,
		marshalizer,
		&hashingMocks.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&testscommon.PubkeyConverterStub{
			LenCalled: func() int {
				return 32
			},
		},
		shardCoordinator,
		txFeeHandler,
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		chainID,
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(minTxVersion),
	)
}

func createInterceptedTxFromPlainTxWithArgParser(tx *dataTransaction.Transaction) (*transaction.InterceptedTransaction, error) {
	marshalizer := &mock.MarshalizerMock{}
	txBuff, err := marshalizer.Marshal(tx)
	if err != nil {
		return nil, err
	}

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = 0
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, senderAddress) {
			return senderShard
		}
		if bytes.Equal(address, recvAddress) {
			return recvShard
		}

		return shardCoordinator.CurrentShard
	}

	return transaction.NewInterceptedTransaction(
		txBuff,
		marshalizer,
		marshalizer,
		&hashingMocks.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&testscommon.PubkeyConverterStub{
			LenCalled: func() int {
				return 32
			},
		},
		shardCoordinator,
		createFreeTxFeeHandler(),
		&testscommon.WhiteListHandlerStub{},
		smartContract.NewArgumentParser(),
		tx.ChainID,
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(tx.Version),
	)
}

// ------- NewInterceptedTransaction

func TestNewInterceptedTransaction_NilBufferShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		nil,
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubKeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(1),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilBuffer, err)
}

func TestNewInterceptedTransaction_NilArgsParser(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubKeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.WhiteListHandlerStub{},
		nil,
		[]byte("chainID"),
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(1),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilArgumentParser, err)
}

func TestNewInterceptedTransaction_NilVersionChecker(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubKeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		&hashingMocks.HasherMock{},
		nil,
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilTransactionVersionChecker, err)
}

func TestNewInterceptedTransaction_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		nil,
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubKeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(1),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedTransaction_NilSignMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		nil,
		&hashingMocks.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubKeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(1),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedTransaction_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		nil,
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubKeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(1),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedTransaction_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		nil,
		&mock.SignerMock{},
		createMockPubKeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(1),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilKeyGen, err)
}

func TestNewInterceptedTransaction_NilSignerShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		nil,
		createMockPubKeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(1),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilSingleSigner, err)
}

func TestNewInterceptedTransaction_NilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		nil,
		mock.NewOneShardCoordinatorMock(),
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(1),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewInterceptedTransaction_NilCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubKeyConverter(),
		nil,
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(1),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptedTransaction_NilFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubKeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		nil,
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(1),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func TestNewInterceptedTransaction_NilWhiteListerVerifiedTxsShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubKeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&economicsmocks.EconomicsHandlerStub{},
		nil,
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(1),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilWhiteListHandler, err)
}

func TestNewInterceptedTransaction_InvalidChainIDShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubKeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		nil,
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(1),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrInvalidChainID, err)
}

func TestNewInterceptedTransaction_NilTxSignHasherShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubKeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		nil,
		versioning.NewTxVersionChecker(1),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedTransaction_UnmarshalingTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errExpected
			},
		},
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubKeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&economicsmocks.EconomicsHandlerStub{},
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(1),
	)

	assert.Nil(t, txi)
	assert.Equal(t, errExpected, err)
}

func TestNewInterceptedTransaction_ShouldWork(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}

	txi, err := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)

	assert.False(t, check.IfNil(txi))
	assert.Nil(t, err)
	assert.Equal(t, tx, txi.Transaction())
}

// ------- CheckValidity

func TestInterceptedTransaction_CheckValidityNilSignatureShouldErr(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: nil,
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)

	err := txi.CheckValidity()

	assert.Equal(t, process.ErrNilSignature, err)
}

func TestInterceptedTransaction_CheckValidityNilRecvAddressShouldErr(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   nil,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)

	err := txi.CheckValidity()

	assert.Equal(t, process.ErrInvalidRcvAddr, err)
}

func TestInterceptedTransaction_CheckValidityInvalidRecvAddressShouldErr(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   append(recvAddress, 0),
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)

	err := txi.CheckValidity()

	assert.Equal(t, process.ErrInvalidRcvAddr, err)
}

func TestInterceptedTransaction_CheckValidityNilSenderAddressShouldErr(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   nil,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)

	err := txi.CheckValidity()

	assert.Equal(t, process.ErrInvalidSndAddr, err)
}

func TestInterceptedTransaction_CheckValidityInvalidSenderAddressShouldErr(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   append(senderAddress, 0),
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)

	err := txi.CheckValidity()

	assert.Equal(t, process.ErrInvalidSndAddr, err)
}

func TestInterceptedTransaction_CheckValidityNilValueShouldErr(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     nil,
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	txi, err := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)
	assert.Nil(t, err)

	err = txi.CheckValidity()

	assert.Equal(t, process.ErrNilValue, err)
}

func TestInterceptedTransaction_CheckValidityInvalidUserNameLength(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:       1,
		Value:       big.NewInt(2),
		Data:        []byte("data"),
		GasLimit:    3,
		GasPrice:    4,
		RcvAddr:     recvAddress,
		SndAddr:     senderAddress,
		Signature:   sigOk,
		RcvUserName: []byte("username1111111111111111111111111111111111111111111111"),
		ChainID:     chainID,
		Version:     minTxVersion,
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)

	err := txi.CheckValidity()
	assert.Equal(t, process.ErrInvalidUserNameLength, err)

	tx.RcvUserName = nil
	tx.SndUserName = []byte("username11111111111111111111111111111111111111111111111111111111")
	txi, _ = createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)
	err = txi.CheckValidity()
	assert.Equal(t, process.ErrInvalidUserNameLength, err)

	tx.RcvUserName = []byte("12345678901234567890123456789012")
	tx.SndUserName = []byte("12345678901234567890123456789012")
	txi, _ = createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)
	err = txi.CheckValidity()
	assert.Nil(t, err)
}

func TestInterceptedTransaction_CheckValidityNilNegativeValueShouldErr(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(-2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)

	err := txi.CheckValidity()

	assert.Equal(t, process.ErrNegativeValue, err)
}

func TestNewInterceptedTransaction_InsufficientFeeShouldErr(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	gasLimit := uint64(3)
	gasPrice := uint64(4)
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  gasLimit,
		GasPrice:  gasPrice,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	errExpected := errors.New("insufficient fee")
	feeHandler := &economicsmocks.EconomicsHandlerStub{
		CheckValidityTxValuesCalled: func(tx data.TransactionWithFeeHandler) error {
			return errExpected
		},
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, feeHandler, chainID, minTxVersion)

	err := txi.CheckValidity()

	assert.Equal(t, errExpected, err)
}

func TestInterceptedTransaction_CheckValidityInvalidSenderShouldErr(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   []byte(""),
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)

	err := txi.CheckValidity()

	assert.NotNil(t, err)
}

func TestInterceptedTransaction_CheckValidityVerifyFailsShouldErr(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: []byte("wrong sig"),
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)

	err := txi.CheckValidity()

	assert.Equal(t, errSignerMockVerifySigFails, err)
}

func TestInterceptedTransaction_CheckValidityWrongChainIDShouldErr(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	wrongChainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   wrongChainID,
		Version:   minTxVersion,
	}

	correctChainID := []byte("correct")
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), correctChainID, minTxVersion)

	err := txi.CheckValidity()
	assert.Equal(t, process.ErrInvalidChainID, err)
}

func TestInterceptedTransaction_CheckValidityInvalidVersionShouldErr(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(2)
	wrongChainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   wrongChainID,
		Version:   1,
	}

	correctChainID := []byte("correct")
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), correctChainID, minTxVersion)

	err := txi.CheckValidity()
	assert.Equal(t, process.ErrInvalidTransactionVersion, err)
}

func TestInterceptedTransaction_TransactionWithNilChainIDShouldErr(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   nil,
		Version:   minTxVersion,
	}

	chainID := []byte("chain")
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)

	err := txi.CheckValidity()
	assert.Equal(t, process.ErrInvalidChainID, err)
}

func TestInterceptedTransaction_CheckValidityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)

	err := txi.CheckValidity()

	assert.Nil(t, err)
}

func TestInterceptedTransaction_CheckValiditySignedWithHashButNotEnabled(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion + 1,
		Options:   dataTransaction.MaskSignedWithHash,
	}

	marshalizer := &mock.MarshalizerMock{}
	txBuff, _ := marshalizer.Marshal(tx)
	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = 6
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, senderAddress) {
			return senderShard
		}
		if bytes.Equal(address, recvAddress) {
			return recvShard
		}

		return shardCoordinator.CurrentShard
	}

	txi, _ := transaction.NewInterceptedTransaction(
		txBuff,
		marshalizer,
		marshalizer,
		&hashingMocks.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&testscommon.PubkeyConverterStub{
			LenCalled: func() int {
				return 32
			},
		},
		shardCoordinator,
		createFreeTxFeeHandler(),
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		chainID,
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(minTxVersion),
	)

	err := txi.CheckValidity()
	assert.Equal(t, process.ErrTransactionSignedWithHashIsNotEnabled, err)
}

func TestInterceptedTransaction_CheckValiditySignedWithHashShouldWork(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion + 1,
		Options:   dataTransaction.MaskSignedWithHash,
	}

	marshalizer := &mock.MarshalizerMock{}
	txBuff, _ := marshalizer.Marshal(tx)
	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = 6
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, senderAddress) {
			return senderShard
		}
		if bytes.Equal(address, recvAddress) {
			return recvShard
		}

		return shardCoordinator.CurrentShard
	}

	txi, _ := transaction.NewInterceptedTransaction(
		txBuff,
		marshalizer,
		marshalizer,
		&hashingMocks.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&testscommon.PubkeyConverterStub{
			LenCalled: func() int {
				return 32
			},
		},
		shardCoordinator,
		createFreeTxFeeHandler(),
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		chainID,
		true,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(minTxVersion),
	)

	err := txi.CheckValidity()
	assert.Nil(t, err)
}

func TestInterceptedTransaction_OkValsGettersShouldWork(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}

	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)

	assert.Equal(t, senderShard, txi.SenderShardId())
	assert.Equal(t, recvShard, txi.ReceiverShardId())
	assert.False(t, txi.IsForCurrentShard())
	assert.Equal(t, tx, txi.Transaction())
}

func TestInterceptedTransaction_ScTxDeployRecvShardIdShouldBeSendersShardId(t *testing.T) {
	t.Parallel()

	senderAddressInShard1 := make([]byte, 32)
	senderAddressInShard1[31] = 1

	minTxVersion := uint32(1)
	recvAddressDeploy := make([]byte, 32)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddressDeploy,
		SndAddr:   senderAddressInShard1,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	marshalizer := &mock.MarshalizerMock{}
	txBuff, _ := marshalizer.Marshal(tx)

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = 1
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(address, senderAddressInShard1) {
			return 1
		}
		if bytes.Equal(address, recvAddressDeploy) {
			return 0
		}

		return shardCoordinator.CurrentShard
	}

	txIntercepted, err := transaction.NewInterceptedTransaction(
		txBuff,
		marshalizer,
		marshalizer,
		&hashingMocks.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&testscommon.PubkeyConverterStub{},
		shardCoordinator,
		createFreeTxFeeHandler(),
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		chainID,
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(minTxVersion),
	)

	assert.Nil(t, err)
	assert.Equal(t, uint32(1), txIntercepted.ReceiverShardId())
	assert.Equal(t, uint32(1), txIntercepted.SenderShardId())
}

func TestInterceptedTransaction_GetNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     nonce,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}

	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)

	result := txi.Nonce()
	assert.Equal(t, nonce, result)
}

func TestInterceptedTransaction_SenderShardId(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}

	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)

	result := txi.SenderShardId()
	assert.Equal(t, senderShard, result)
}

func TestInterceptedTransaction_GetSenderAddress(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}

	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler(), chainID, minTxVersion)
	result := txi.SenderAddress()
	assert.NotNil(t, result)
}

func TestInterceptedTransaction_CheckValiditySecondTimeDoesNotVerifySig(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}

	var sigVerified bool
	signer := &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			sigVerified = true
			return nil
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	txBuff, err := marshalizer.Marshal(tx)
	require.Nil(t, err)

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = 6
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		return shardCoordinator.CurrentShard
	}

	cache := testscommon.NewCacherMock()
	whiteListerVerifiedTxs, err := interceptors.NewWhiteListDataVerifier(cache)
	require.Nil(t, err)

	txi, err := transaction.NewInterceptedTransaction(
		txBuff,
		marshalizer,
		marshalizer,
		&hashingMocks.HasherMock{},
		createKeyGenMock(),
		signer,
		&testscommon.PubkeyConverterStub{
			LenCalled: func() int {
				return 32
			},
		},
		shardCoordinator,
		createFreeTxFeeHandler(),
		whiteListerVerifiedTxs,
		&mock.ArgumentParserMock{},
		chainID,
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(minTxVersion),
	)
	require.Nil(t, err)

	// first check should verify sig
	sigVerified = false
	err = txi.CheckValidity()
	require.Nil(t, err)
	require.True(t, sigVerified)

	// second check should find txi in whitelist and should not verify sig
	sigVerified = false
	err = txi.CheckValidity()
	require.Nil(t, err)
	require.False(t, sigVerified)
}

func TestInterceptedTransaction_CheckValidityOfRelayedTx(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("relayedTx"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	txi, _ := createInterceptedTxFromPlainTxWithArgParser(tx)
	err := txi.CheckValidity()
	assert.Equal(t, err, process.ErrInvalidArguments)

	tx.Data = []byte("relayedTx@00@11")
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.Equal(t, err, process.ErrInvalidArguments)

	tx.Data = []byte("relayedTx@0011")
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.NotNil(t, err)

	userTx := &dataTransaction.Transaction{
		SndAddr:   recvAddress,
		RcvAddr:   senderAddress,
		Data:      []byte("hello"),
		GasLimit:  3,
		GasPrice:  4,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	marshalizer := &mock.MarshalizerMock{}
	userTxData, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxData))
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.ErrorIs(t, err, data.ErrNilValue)

	userTx.Value = big.NewInt(0)
	userTxData, _ = marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxData))
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.Nil(t, err)

	userTx.Signature = []byte("notOk")
	userTxData, _ = marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxData))
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.ErrorIs(t, err, errSignerMockVerifySigFails)
	assert.Contains(t, err.Error(), "inner transaction")

	userTx.Signature = sigOk
	userTx.SndAddr = []byte("otherAddress")
	userTxData, _ = marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxData))
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.Equal(t, err, process.ErrRelayedTxBeneficiaryDoesNotMatchReceiver)

	userTx.SndAddr = recvAddress
	userTx.Data = []byte(core.RelayedTransaction)
	userTxData, _ = marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxData))
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.Equal(t, process.ErrRecursiveRelayedTxIsNotAllowed, err)
}

func TestInterceptedTransaction_CheckValidityOfRelayedTxV2(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("relayedTxV2"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	txi, _ := createInterceptedTxFromPlainTxWithArgParser(tx)
	err := txi.CheckValidity()
	assert.Equal(t, err, process.ErrInvalidArguments)

	tx.Data = []byte("relayedTxV2@00@11")
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.Equal(t, err, process.ErrInvalidArguments)

	tx.Data = []byte("relayedTxV2@00@11@22@33")
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.NotNil(t, err)

	tx.Data = []byte("relayedTxV2@notHex")
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.Nil(t, err)

	userTx := &dataTransaction.Transaction{
		SndAddr:   recvAddress,
		RcvAddr:   senderAddress,
		Data:      []byte("hello"),
		GasLimit:  3,
		GasPrice:  4,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	userTx.Signature = []byte("notOk")
	tx.Data = []byte(core.RelayedTransactionV2 + "@" + hex.EncodeToString(userTx.RcvAddr) + "@" + hex.EncodeToString(big.NewInt(0).SetUint64(userTx.Nonce).Bytes()) + "@" + hex.EncodeToString(userTx.Data) + "@" + hex.EncodeToString(userTx.Signature))
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.ErrorIs(t, err, errSignerMockVerifySigFails)
	assert.Contains(t, err.Error(), "inner transaction")

	userTx.Signature = sigOk
	userTx.SndAddr = []byte("otherAddress")
	tx.Data = []byte(core.RelayedTransactionV2 + "@" + hex.EncodeToString(userTx.RcvAddr) + "@" + hex.EncodeToString(big.NewInt(0).SetUint64(userTx.Nonce).Bytes()) + "@" + hex.EncodeToString([]byte(core.RelayedTransaction)) + "@" + hex.EncodeToString(userTx.Signature))
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.Equal(t, process.ErrRecursiveRelayedTxIsNotAllowed, err)

	userTx.Signature = sigOk
	userTx.SndAddr = []byte("otherAddress")
	tx.Data = []byte(core.RelayedTransactionV2 + "@" + hex.EncodeToString(userTx.RcvAddr) + "@" + hex.EncodeToString(big.NewInt(0).SetUint64(userTx.Nonce).Bytes()) + "@" + hex.EncodeToString(userTx.Data) + "@" + hex.EncodeToString(userTx.Signature))
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.Nil(t, err)
}

func TestInterceptedTransaction_CheckValidityOfRelayedTxV3(t *testing.T) {
	t.Parallel()

	minTxVersion := uint32(1)
	chainID := []byte("chain")
	innerTx := &dataTransaction.Transaction{
		Nonce:       1,
		Value:       big.NewInt(2),
		Data:        []byte("data inner tx 1"),
		GasLimit:    3,
		GasPrice:    4,
		RcvAddr:     []byte("34567890123456789012345678901234"),
		SndAddr:     recvAddress,
		Signature:   sigOk,
		ChainID:     chainID,
		Version:     minTxVersion,
		RelayedAddr: senderAddress,
	}

	tx := &dataTransaction.Transaction{
		Nonce:            1,
		Value:            big.NewInt(0),
		GasLimit:         10,
		GasPrice:         4,
		RcvAddr:          recvAddress,
		SndAddr:          senderAddress,
		Signature:        sigOk,
		ChainID:          chainID,
		Version:          minTxVersion,
		InnerTransaction: innerTx,
	}
	txi, _ := createInterceptedTxFromPlainTxWithArgParser(tx)
	err := txi.CheckValidity()
	assert.Nil(t, err)

	innerTx.RelayedAddr = nil
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.Equal(t, process.ErrRelayedTxV3EmptyRelayer, err)
	innerTx.RelayedAddr = senderAddress

	innerTx.SndAddr = []byte("34567890123456789012345678901234")
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.Equal(t, process.ErrRelayedTxV3BeneficiaryDoesNotMatchReceiver, err)
	innerTx.SndAddr = recvAddress

	innerTx.Signature = nil
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.NotNil(t, err)

	innerTx.Signature = sigBad
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.NotNil(t, err)

	innerTx2 := &dataTransaction.Transaction{
		Nonce:     2,
		Value:     big.NewInt(3),
		Data:      []byte("data inner tx 2"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
		ChainID:   chainID,
		Version:   minTxVersion,
	}
	innerTx.InnerTransaction = innerTx2
	tx.InnerTransaction = innerTx
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	err = txi.CheckValidity()
	assert.NotNil(t, err)
}

// ------- IsInterfaceNil
func TestInterceptedTransaction_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var txi *transaction.InterceptedTransaction

	assert.True(t, check.IfNil(txi))
}

func TestRelayTransaction_NotAddedToWhitelistUntilIntegrityChecked(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	whiteListHandler, _ := interceptors.NewWhiteListDataVerifier(testscommon.NewCacherMock())

	userTx := &dataTransaction.Transaction{
		SndAddr:   recvAddress,
		RcvAddr:   senderAddress,
		Data:      []byte("hello"),
		Value:     big.NewInt(10),
		GasLimit:  3,
		GasPrice:  4,
		Signature: sigOk,
		ChainID:   []byte("chain"),
		Version:   1,
	}

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("relayedTx@abba"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigBad,
		ChainID:   []byte("chain"),
		Version:   1,
	}

	// Bad signature -> not whitelisted
	txi, _ := createInterceptedTxFromPlainTxWithArgParser(tx)
	txi.SetWhitelistHandler(whiteListHandler)

	err := txi.CheckValidity()
	require.Equal(t, errSignerMockVerifySigFails, err)
	require.False(t, whiteListHandler.IsWhiteListed(txi))

	// Good wrapper signature, but user tx is not valid -> not whitelisted
	tx.Signature = sigOk
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	txi.SetWhitelistHandler(whiteListHandler)

	err = txi.CheckValidity()
	require.NotNil(t, err)
	require.False(t, whiteListHandler.IsWhiteListed(txi))

	// Good wrapper signature, bad user tx signature -> not whitelisted
	userTx.Signature = sigBad
	userTxData, _ := marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxData))
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	txi.SetWhitelistHandler(whiteListHandler)

	err = txi.CheckValidity()
	require.NotNil(t, err)
	require.False(t, whiteListHandler.IsWhiteListed(txi))

	// Good transaction -> whitelisted
	userTx.Signature = sigOk
	userTxData, _ = marshalizer.Marshal(userTx)
	tx.Data = []byte(core.RelayedTransaction + "@" + hex.EncodeToString(userTxData))
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	txi.SetWhitelistHandler(whiteListHandler)

	err = txi.CheckValidity()
	require.Nil(t, err)
	require.True(t, whiteListHandler.IsWhiteListed(txi))

	// Good signature (regular transaction) -> whitelisted
	tx.Data = []byte("test")
	txi, _ = createInterceptedTxFromPlainTxWithArgParser(tx)
	txi.SetWhitelistHandler(whiteListHandler)

	err = txi.CheckValidity()
	require.Nil(t, err)
	require.True(t, whiteListHandler.IsWhiteListed(txi))
}

func TestInterceptedTransaction_Type(t *testing.T) {
	t.Parallel()

	expectedType := "intercepted tx"

	intx := &transaction.InterceptedTransaction{}

	assert.Equal(t, expectedType, intx.Type())
}

func TestInterceptedTransaction_Fee(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:    1,
		GasLimit: 3,
		GasPrice: 4,
		RcvAddr:  nil,
	}
	marshalizer := &mock.MarshalizerMock{}
	txBuff, _ := marshalizer.Marshal(tx)

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()

	txin, _ := transaction.NewInterceptedTransaction(
		txBuff,
		marshalizer,
		marshalizer,
		&hashingMocks.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&testscommon.PubkeyConverterStub{},
		shardCoordinator,
		createFreeTxFeeHandler(),
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("T"),
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(0),
	)

	assert.Equal(t, big.NewInt(0), txin.Fee())
}

func TestInterceptedTransaction_String(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	value := big.NewInt(150)
	sndAddr := []byte("snd")
	rcvAdrr := []byte("rcv")
	dataField := []byte("data")

	tx := &dataTransaction.Transaction{
		Nonce:   nonce,
		RcvAddr: rcvAdrr,
		SndAddr: sndAddr,
		Value:   value,
		Data:    dataField,
	}

	marshalizer := &mock.MarshalizerMock{}
	txBuff, _ := marshalizer.Marshal(tx)

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()

	txin, _ := transaction.NewInterceptedTransaction(
		txBuff,
		marshalizer,
		marshalizer,
		&hashingMocks.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&testscommon.PubkeyConverterStub{},
		shardCoordinator,
		createFreeTxFeeHandler(),
		&testscommon.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("T"),
		false,
		&hashingMocks.HasherMock{},
		versioning.NewTxVersionChecker(0),
	)

	expectedFormat := fmt.Sprintf(
		"sender=%s, nonce=%d, value=%s, recv=%s, data=%s",
		logger.DisplayByteSlice(sndAddr), nonce, value.String(), logger.DisplayByteSlice(rcvAdrr), hex.EncodeToString(dataField),
	)

	assert.Equal(t, expectedFormat, txin.String())
}

func TestInterceptedTransaction_checkMaxGasPrice(t *testing.T) {
	t.Parallel()

	maxAllowedGasPriceSetGuardian := uint64(2000000)
	feeHandler := &economicsmocks.EconomicsHandlerStub{
		MaxGasPriceSetGuardianCalled: func() uint64 {
			return maxAllowedGasPriceSetGuardian
		},
	}
	setGuardianBuiltinCallData := []byte("SetGuardian@xxxx")
	testTx1 := &dataTransaction.Transaction{
		GasPrice: maxAllowedGasPriceSetGuardian / 2,
		Data:     setGuardianBuiltinCallData,
	}
	testTx2 := &dataTransaction.Transaction{
		GasPrice: maxAllowedGasPriceSetGuardian * 2,
		Data:     setGuardianBuiltinCallData,
	}

	t.Run("guardedTx returns always OK no matter the gas price", func(t *testing.T) {
		txVersionChecker := &testscommon.TxVersionCheckerStub{
			IsGuardedTransactionCalled: func(tx *dataTransaction.Transaction) bool {
				return true
			},
		}
		inTx1, err := createInterceptedTxWithTxFeeHandlerAndVersionChecker(testTx1, feeHandler, txVersionChecker)
		require.Nil(t, err)
		inTx2, err := createInterceptedTxWithTxFeeHandlerAndVersionChecker(testTx2, feeHandler, txVersionChecker)
		require.Nil(t, err)

		errMaxGasPrice := inTx1.CheckMaxGasPrice()
		require.Nil(t, errMaxGasPrice)

		errMaxGasPrice = inTx2.CheckMaxGasPrice()
		require.Nil(t, errMaxGasPrice)
	})
	t.Run("not guarded Tx, not setGuardian always OK", func(t *testing.T) {
		tx1 := *testTx1
		tx1.Data = []byte("dummy")
		tx2 := *testTx2
		tx2.Data = []byte("dummy")

		txVersionChecker := &testscommon.TxVersionCheckerStub{
			IsGuardedTransactionCalled: func(tx *dataTransaction.Transaction) bool {
				return false
			},
		}

		inTx1, err := createInterceptedTxWithTxFeeHandlerAndVersionChecker(&tx1, feeHandler, txVersionChecker)
		require.Nil(t, err)
		inTx2, err := createInterceptedTxWithTxFeeHandlerAndVersionChecker(&tx2, feeHandler, txVersionChecker)
		require.Nil(t, err)

		errMaxGasPrice := inTx1.CheckMaxGasPrice()
		require.Nil(t, errMaxGasPrice)

		errMaxGasPrice = inTx2.CheckMaxGasPrice()
		require.Nil(t, errMaxGasPrice)
	})
	t.Run("not guarded Tx with setGuardian call and price lower than max or equal OK", func(t *testing.T) {
		tx1 := *testTx1
		tx1.GasPrice = maxAllowedGasPriceSetGuardian
		tx2 := *testTx2
		tx2.GasPrice = maxAllowedGasPriceSetGuardian / 2

		txVersionChecker := &testscommon.TxVersionCheckerStub{
			IsGuardedTransactionCalled: func(tx *dataTransaction.Transaction) bool {
				return false
			},
		}

		inTx1, err := createInterceptedTxWithTxFeeHandlerAndVersionChecker(&tx1, feeHandler, txVersionChecker)
		require.Nil(t, err)
		inTx2, err := createInterceptedTxWithTxFeeHandlerAndVersionChecker(&tx2, feeHandler, txVersionChecker)
		require.Nil(t, err)

		errMaxGasPrice := inTx1.CheckMaxGasPrice()
		require.Nil(t, errMaxGasPrice)

		errMaxGasPrice = inTx2.CheckMaxGasPrice()
		require.Nil(t, errMaxGasPrice)
	})
	t.Run("not guarded Tx with setGuardian call and price higher than max err", func(t *testing.T) {
		tx1 := *testTx1
		tx1.GasPrice = maxAllowedGasPriceSetGuardian * 2
		txVersionChecker := &testscommon.TxVersionCheckerStub{
			IsGuardedTransactionCalled: func(tx *dataTransaction.Transaction) bool {
				return false
			},
		}

		inTx1, err := createInterceptedTxWithTxFeeHandlerAndVersionChecker(&tx1, feeHandler, txVersionChecker)
		require.Nil(t, err)

		errMaxGasPrice := inTx1.CheckMaxGasPrice()
		require.Equal(t, process.ErrGasPriceTooHigh, errMaxGasPrice)
	})
}

func TestInterceptedTransaction_VerifyGuardianSig(t *testing.T) {
	t.Parallel()

	testTxVersionChecker := testscommon.TxVersionCheckerStub{
		IsGuardedTransactionCalled: func(tx *dataTransaction.Transaction) bool {
			return true
		},
	}
	feeHandler := &economicsmocks.EconomicsHandlerStub{
		MaxGasPriceSetGuardianCalled: func() uint64 {
			return 1000
		},
	}
	testTx := dataTransaction.Transaction{
		Data:              []byte("some data"),
		GuardianAddr:      []byte("guardian addr"),
		GuardianSignature: []byte("guardian signature"),
	}

	t.Run("get data for signing with error", func(t *testing.T) {
		tx := testTx
		txVersionChecker := testTxVersionChecker
		txVersionChecker.IsSignedWithHashCalled = func(tx *dataTransaction.Transaction) bool {
			return true
		}
		inTx, err := createInterceptedTxWithTxFeeHandlerAndVersionChecker(&tx, feeHandler, &txVersionChecker)
		require.Nil(t, err)
		err = inTx.VerifyGuardianSig(&tx)
		require.Equal(t, process.ErrTransactionSignedWithHashIsNotEnabled, err)
	})
	t.Run("nil guardian sig", func(t *testing.T) {
		tx := testTx
		tx.GuardianSignature = nil
		inTx, err := createInterceptedTxWithTxFeeHandlerAndVersionChecker(&tx, feeHandler, &testTxVersionChecker)
		require.Nil(t, err)

		err = inTx.VerifyGuardianSig(&tx)
		require.ErrorIs(t, err, errSignerMockVerifySigFails)
		require.Contains(t, err.Error(), "guardian's signature")
	})
	t.Run("normal TX with not empty guardian address", func(t *testing.T) {
		tx := testTx
		tx.GuardianAddr = []byte("guardian addr")
		txVersionChecker := testTxVersionChecker
		txVersionChecker.IsGuardedTransactionCalled = func(tx *dataTransaction.Transaction) bool {
			return false
		}
		inTx, err := createInterceptedTxWithTxFeeHandlerAndVersionChecker(&tx, feeHandler, &txVersionChecker)
		require.Nil(t, err)

		err = inTx.VerifyGuardianSig(&tx)
		require.True(t, errors.Is(err, process.ErrGuardianAddressNotExpected))
	})
	t.Run("normal TX with guardian sig", func(t *testing.T) {
		tx := testTx
		tx.GuardianAddr = nil
		tx.GuardianSignature = []byte("guardian signature")
		txVersionChecker := testTxVersionChecker
		txVersionChecker.IsGuardedTransactionCalled = func(tx *dataTransaction.Transaction) bool {
			return false
		}
		inTx, err := createInterceptedTxWithTxFeeHandlerAndVersionChecker(&tx, feeHandler, &txVersionChecker)
		require.Nil(t, err)

		err = inTx.VerifyGuardianSig(&tx)
		require.True(t, errors.Is(err, process.ErrGuardianSignatureNotExpected))
	})
	t.Run("wrong guardian sig", func(t *testing.T) {
		tx := testTx
		tx.GuardianSignature = sigBad
		inTx, err := createInterceptedTxWithTxFeeHandlerAndVersionChecker(&tx, feeHandler, &testTxVersionChecker)
		require.Nil(t, err)

		err = inTx.VerifyGuardianSig(&tx)
		require.ErrorIs(t, err, errSignerMockVerifySigFails)
		require.Contains(t, err.Error(), "guardian's signature")
	})
	t.Run("correct guardian sig", func(t *testing.T) {
		tx := testTx
		tx.GuardianSignature = sigOk
		inTx, err := createInterceptedTxWithTxFeeHandlerAndVersionChecker(&tx, feeHandler, &testTxVersionChecker)
		require.Nil(t, err)

		err = inTx.VerifyGuardianSig(&tx)
		require.Nil(t, err)
	})
}
