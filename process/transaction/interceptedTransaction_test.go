package transaction_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"testing"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/versioning"
	"github.com/ElrondNetwork/elrond-go/crypto"
	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/testscommon"
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

func createMockPubkeyConverter() *mock.PubkeyConverterMock {
	return mock.NewPubkeyConverterMock(32)
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

func createFreeTxFeeHandler() *mock.FeeHandlerStub {
	return &mock.FeeHandlerStub{
		CheckValidityTxValuesCalled: func(tx process.TransactionWithFeeHandler) error {
			return nil
		},
	}
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
		mock.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&mock.PubkeyConverterStub{
			LenCalled: func() int {
				return 32
			},
		},
		shardCoordinator,
		txFeeHandler,
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		chainID,
		false,
		mock.HasherMock{},
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
		mock.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&mock.PubkeyConverterStub{
			LenCalled: func() int {
				return 32
			},
		},
		shardCoordinator,
		createFreeTxFeeHandler(),
		&mock.WhiteListHandlerStub{},
		smartContract.NewArgumentParser(),
		tx.ChainID,
		false,
		mock.HasherMock{},
		versioning.NewTxVersionChecker(tx.Version),
	)
}

//------- NewInterceptedTransaction

func TestNewInterceptedTransaction_NilBufferShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		nil,
		&mock.MarshalizerMock{},
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.FeeHandlerStub{},
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		mock.HasherMock{},
		versioning.NewTxVersionChecker(1),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilBuffer, err)
}

func TestNewInterceptedTransaction_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		nil,
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.FeeHandlerStub{},
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		mock.HasherMock{},
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
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.FeeHandlerStub{},
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		mock.HasherMock{},
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
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.FeeHandlerStub{},
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		mock.HasherMock{},
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
		mock.HasherMock{},
		nil,
		&mock.SignerMock{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.FeeHandlerStub{},
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		mock.HasherMock{},
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
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		nil,
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.FeeHandlerStub{},
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		mock.HasherMock{},
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
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		nil,
		mock.NewOneShardCoordinatorMock(),
		&mock.FeeHandlerStub{},
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		mock.HasherMock{},
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
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubkeyConverter(),
		nil,
		&mock.FeeHandlerStub{},
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		mock.HasherMock{},
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
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		nil,
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		mock.HasherMock{},
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
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.FeeHandlerStub{},
		nil,
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		mock.HasherMock{},
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
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.FeeHandlerStub{},
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		nil,
		false,
		mock.HasherMock{},
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
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.FeeHandlerStub{},
		&mock.WhiteListHandlerStub{},
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
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
		&mock.FeeHandlerStub{},
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("chainID"),
		false,
		mock.HasherMock{},
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

//------- CheckValidity

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
	feeHandler := &mock.FeeHandlerStub{
		CheckValidityTxValuesCalled: func(tx process.TransactionWithFeeHandler) error {
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
		Options:   versioning.MaskSignedWithHash,
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
		mock.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&mock.PubkeyConverterStub{
			LenCalled: func() int {
				return 32
			},
		},
		shardCoordinator,
		createFreeTxFeeHandler(),
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		chainID,
		false,
		mock.HasherMock{},
		versioning.NewTxVersionChecker(minTxVersion),
	)

	err := txi.CheckValidity()
	assert.Equal(t, process.ErrTransactionSignedWithHashIsNotEnabled, err)
}

func TestInterceptedTransaction_CheckValiditySignedWithHashShoudWork(t *testing.T) {
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
		Options:   versioning.MaskSignedWithHash,
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
		mock.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&mock.PubkeyConverterStub{
			LenCalled: func() int {
				return 32
			},
		},
		shardCoordinator,
		createFreeTxFeeHandler(),
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		chainID,
		true,
		mock.HasherMock{},
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
		mock.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&mock.PubkeyConverterStub{},
		shardCoordinator,
		createFreeTxFeeHandler(),
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		chainID,
		false,
		mock.HasherMock{},
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
		mock.HasherMock{},
		createKeyGenMock(),
		signer,
		&mock.PubkeyConverterStub{
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
		mock.HasherMock{},
		versioning.NewTxVersionChecker(minTxVersion),
	)
	require.Nil(t, err)

	// first check should verify sig
	sigVerified = false
	err = txi.CheckValidity()
	require.Nil(t, err)
	require.True(t, sigVerified)

	//second check should find txi in whitelist and should not verify sig
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
	assert.Equal(t, process.ErrNilValue, err)

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
	assert.Equal(t, errSignerMockVerifySigFails, err)

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

//------- IsInterfaceNil
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
		mock.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&mock.PubkeyConverterStub{},
		shardCoordinator,
		createFreeTxFeeHandler(),
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("T"),
		false,
		mock.HasherMock{},
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

	tx := &dataTransaction.Transaction{
		Nonce:   nonce,
		RcvAddr: rcvAdrr,
		SndAddr: sndAddr,
		Value:   value,
	}

	marshalizer := &mock.MarshalizerMock{}
	txBuff, _ := marshalizer.Marshal(tx)

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()

	txin, _ := transaction.NewInterceptedTransaction(
		txBuff,
		marshalizer,
		marshalizer,
		mock.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&mock.PubkeyConverterStub{},
		shardCoordinator,
		createFreeTxFeeHandler(),
		&mock.WhiteListHandlerStub{},
		&mock.ArgumentParserMock{},
		[]byte("T"),
		false,
		mock.HasherMock{},
		versioning.NewTxVersionChecker(0),
	)

	expectedFormat := fmt.Sprintf(
		"sender=%s, nonce=%d, value=%s, recv=%s",
		logger.DisplayByteSlice(sndAddr), nonce, value.String(), logger.DisplayByteSlice(rcvAdrr),
	)

	assert.Equal(t, expectedFormat, txin.String())
}
