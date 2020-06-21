package transaction_test

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errSingleSignKeyGenMock = errors.New("errSingleSignKeyGenMock")
var errSignerMockVerifySigFails = errors.New("errSignerMockVerifySigFails")

var senderShard = uint32(2)
var recvShard = uint32(3)
var senderAddress = []byte("sender")
var recvAddress = []byte("receiver")
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

func createInterceptedTxFromPlainTx(tx *dataTransaction.Transaction, txFeeHandler process.FeeHandler) (*transaction.InterceptedTransaction, error) {
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
		&mock.PubkeyConverterStub{},
		shardCoordinator,
		txFeeHandler,
		&mock.WhiteListHandlerStub{},
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
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilWhiteListHandler, err)
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
	)

	assert.Nil(t, txi)
	assert.Equal(t, errExpected, err)
}

func TestNewInterceptedTransaction_ShouldWork(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}

	txi, err := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler())

	assert.False(t, check.IfNil(txi))
	assert.Nil(t, err)
	assert.Equal(t, tx, txi.Transaction())
}

//------- CheckValidity

func TestInterceptedTransaction_CheckValidityNilSignatureShouldErr(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: nil,
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler())

	err := txi.CheckValidity()

	assert.Equal(t, process.ErrNilSignature, err)
}

func TestInterceptedTransaction_CheckValidityNilRecvAddressShouldErr(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   nil,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler())

	err := txi.CheckValidity()

	assert.Equal(t, process.ErrNilRcvAddr, err)
}

func TestInterceptedTransaction_CheckValidityNilSenderAddressShouldErr(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   nil,
		Signature: sigOk,
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler())

	err := txi.CheckValidity()

	assert.Equal(t, process.ErrNilSndAddr, err)
}

func TestInterceptedTransaction_CheckValidityNilValueShouldErr(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     nil,
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}
	txi, err := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler())
	assert.Nil(t, err)

	err = txi.CheckValidity()

	assert.Equal(t, process.ErrNilValue, err)
}

func TestInterceptedTransaction_CheckValidityInvalidUserNameLength(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:       1,
		Value:       big.NewInt(2),
		Data:        []byte("data"),
		GasLimit:    3,
		GasPrice:    4,
		RcvAddr:     recvAddress,
		SndAddr:     senderAddress,
		Signature:   sigOk,
		RcvUserName: []byte("username"),
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler())

	err := txi.CheckValidity()
	assert.Equal(t, process.ErrInvalidUserNameLength, err)

	tx.RcvUserName = nil
	tx.SndUserName = []byte("username")
	txi, _ = createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler())
	err = txi.CheckValidity()
	assert.Equal(t, process.ErrInvalidUserNameLength, err)

	tx.RcvUserName = []byte("12345678901234567890123456789012")
	tx.SndUserName = []byte("12345678901234567890123456789012")
	txi, _ = createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler())
	err = txi.CheckValidity()
	assert.Nil(t, err)
}

func TestInterceptedTransaction_CheckValidityNilNegativeValueShouldErr(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(-2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler())

	err := txi.CheckValidity()

	assert.Equal(t, process.ErrNegativeValue, err)
}

func TestNewInterceptedTransaction_InsufficientFeeShouldErr(t *testing.T) {
	t.Parallel()

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
	}
	errExpected := errors.New("insufficient fee")
	feeHandler := &mock.FeeHandlerStub{
		CheckValidityTxValuesCalled: func(tx process.TransactionWithFeeHandler) error {
			return errExpected
		},
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, feeHandler)

	err := txi.CheckValidity()

	assert.Equal(t, errExpected, err)
}

func TestInterceptedTransaction_CheckValidityInvalidSenderShouldErr(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   []byte(""),
		Signature: sigOk,
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler())

	err := txi.CheckValidity()

	assert.NotNil(t, err)
}

func TestInterceptedTransaction_CheckValidityVerifyFailsShouldErr(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: []byte("wrong sig"),
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler())

	err := txi.CheckValidity()

	assert.Equal(t, errSignerMockVerifySigFails, err)
}

func TestInterceptedTransaction_CheckValidityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}
	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler())

	err := txi.CheckValidity()

	assert.Nil(t, err)
}

func TestInterceptedTransaction_OkValsGettersShouldWork(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}

	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler())

	assert.Equal(t, senderShard, txi.SenderShardId())
	assert.Equal(t, recvShard, txi.ReceiverShardId())
	assert.False(t, txi.IsForCurrentShard())
	assert.Equal(t, tx, txi.Transaction())
}

func TestInterceptedTransaction_ScTxDeployRecvShardIdShouldBeSendersShardId(t *testing.T) {
	t.Parallel()

	senderAddressInShard1 := make([]byte, 32)
	senderAddressInShard1[31] = 1

	recvAddressDeploy := make([]byte, 32)

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddressDeploy,
		SndAddr:   senderAddressInShard1,
		Signature: sigOk,
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
	)

	assert.Nil(t, err)
	assert.Equal(t, uint32(1), txIntercepted.ReceiverShardId())
	assert.Equal(t, uint32(1), txIntercepted.SenderShardId())
}

func TestInterceptedTransaction_GetNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)

	tx := &dataTransaction.Transaction{
		Nonce:     nonce,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}

	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler())

	result := txi.Nonce()
	assert.Equal(t, nonce, result)
}

func TestInterceptedTransaction_SenderShardId(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}

	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler())

	result := txi.SenderShardId()
	assert.Equal(t, senderShard, result)
}

func TestInterceptedTransaction_FeeCallsTxFeeHandler(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}

	computeFeeCalled := false
	txFeeHandler := createFreeTxFeeHandler()
	txi, _ := createInterceptedTxFromPlainTx(tx, txFeeHandler)
	txFeeHandler.ComputeFeeCalled = func(tx process.TransactionWithFeeHandler) *big.Int {
		computeFeeCalled = true

		return big.NewInt(0)
	}

	_ = txi.Fee()

	assert.True(t, computeFeeCalled)
}

func TestInterceptedTransaction_GetSenderAddress(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}

	txi, _ := createInterceptedTxFromPlainTx(tx, createFreeTxFeeHandler())
	result := txi.SenderAddress()
	assert.NotNil(t, result)
}

func TestInterceptedTransaction_CheckValiditySecondTimeDoesNotVerifySig(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     0,
		Value:     big.NewInt(2),
		Data:      []byte("data"),
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
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
		&mock.PubkeyConverterStub{},
		shardCoordinator,
		createFreeTxFeeHandler(),
		whiteListerVerifiedTxs,
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

//------- IsInterfaceNil

func TestInterceptedTransaction_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var txi *transaction.InterceptedTransaction

	assert.True(t, check.IfNil(txi))
}
