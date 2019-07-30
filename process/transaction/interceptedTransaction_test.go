package transaction_test

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/stretchr/testify/assert"
)

var errSingleSignKeyGenMock = errors.New("errSingleSignKeyGenMock")
var errSignerMockVerifySigFails = errors.New("errSignerMockVerifySigFails")

var senderShard = uint32(2)
var recvShard = uint32(3)
var senderAddress = []byte("sender")
var recvAddress = []byte("receiver")
var sigOk = []byte("signature")

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

func createInterceptedTxFromPlainTx(tx *dataTransaction.Transaction) (*transaction.InterceptedTransaction, error) {
	marshalizer := &mock.MarshalizerMock{}
	txBuff, _ := marshalizer.Marshal(tx)

	shardCoordinator := mock.NewMultipleShardsCoordinatorMock()
	shardCoordinator.CurrentShard = 6
	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		if bytes.Equal(address.Bytes(), senderAddress) {
			return senderShard
		}
		if bytes.Equal(address.Bytes(), recvAddress) {
			return recvShard
		}

		return shardCoordinator.CurrentShard
	}

	return transaction.NewInterceptedTransaction(
		txBuff,
		marshalizer,
		mock.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return mock.NewAddressMock(pubKey), nil
			},
		},
		shardCoordinator,
	)
}

func TestNewInterceptedTransaction_NilBufferShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		nil,
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilBuffer, err)
}

func TestNewInterceptedTransaction_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		nil,
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedTransaction_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		nil,
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedTransaction_NilKeyGenShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		nil,
		&mock.SignerMock{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilKeyGen, err)
}

func TestNewInterceptedTransaction_NilSignerShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		nil,
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilSingleSigner, err)
}

func TestNewInterceptedTransaction_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		nil,
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewInterceptedTransaction_NilCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.AddressConverterMock{},
		nil,
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
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
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, errExpected, err)
}

func TestNewInterceptedTransaction_MarshalingCopiedTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")

	txi, err := transaction.NewInterceptedTransaction(
		make([]byte, 0),
		&mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) (bytes []byte, e error) {
				return nil, errExpected
			},
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return nil
			},
		},
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, errExpected, err)
}

func TestNewInterceptedTransaction_AddrConvFailsShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := transaction.NewInterceptedTransaction(
		[]byte("{}"),
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		&mock.SingleSignKeyGenMock{},
		&mock.SignerMock{},
		&mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return nil, errors.New("expected error")
			},
		},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrInvalidSndAddr, err)
}

func TestNewInterceptedTransaction_NilSignatureShouldErr(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      "data",
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: nil,
	}

	txi, err := createInterceptedTxFromPlainTx(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilSignature, err)
}

func TestNewInterceptedTransaction_NilSenderAddressShouldErr(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      "data",
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   nil,
		Signature: sigOk,
	}

	txi, err := createInterceptedTxFromPlainTx(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilSndAddr, err)
}

func TestNewInterceptedTransaction_NilRecvAddressShouldErr(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      "data",
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   nil,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}

	txi, err := createInterceptedTxFromPlainTx(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilRcvAddr, err)
}

func TestNewInterceptedTransaction_NilValueShouldErr(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     nil,
		Data:      "data",
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}

	txi, err := createInterceptedTxFromPlainTx(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilValue, err)
}

func TestNewInterceptedTransaction_NilNegativeValueShouldErr(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(-2),
		Data:      "data",
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}

	txi, err := createInterceptedTxFromPlainTx(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNegativeValue, err)
}

func TestNewInterceptedTransaction_InvalidSenderShouldErr(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      "data",
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   []byte(""),
		Signature: sigOk,
	}

	txi, err := createInterceptedTxFromPlainTx(tx)

	assert.Nil(t, txi)
	assert.Equal(t, errSingleSignKeyGenMock, err)
}

func TestNewInterceptedTransaction_VerifyFailsShouldErr(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      "data",
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: []byte("wrong sig"),
	}

	txi, err := createInterceptedTxFromPlainTx(tx)

	assert.Nil(t, txi)
	assert.Equal(t, errSignerMockVerifySigFails, err)
}

func TestNewInterceptedTransaction_ShouldWork(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      "data",
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}

	txi, err := createInterceptedTxFromPlainTx(tx)

	assert.NotNil(t, txi)
	assert.Nil(t, err)
	assert.Equal(t, tx, txi.Transaction())
}

func TestNewInterceptedTransaction_OkValsGettersShouldWork(t *testing.T) {
	t.Parallel()

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      "data",
		GasLimit:  3,
		GasPrice:  4,
		RcvAddr:   recvAddress,
		SndAddr:   senderAddress,
		Signature: sigOk,
	}

	txi, _ := createInterceptedTxFromPlainTx(tx)

	assert.Equal(t, senderShard, txi.SndShard())
	assert.Equal(t, recvShard, txi.RcvShard())
	assert.True(t, txi.IsAddressedToOtherShards())
	assert.Equal(t, tx, txi.Transaction())
}

func TestNewInterceptedTransaction_ScTxDeployRecvShardIdShouldBeSendersShardId(t *testing.T) {
	t.Parallel()

	senderAddressInShard1 := make([]byte, 32)
	senderAddressInShard1[31] = 1

	recvAddressDeploy := make([]byte, 32)

	tx := &dataTransaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(2),
		Data:      "data",
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
	shardCoordinator.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		if bytes.Equal(address.Bytes(), senderAddressInShard1) {
			return 1
		}
		if bytes.Equal(address.Bytes(), recvAddressDeploy) {
			return 0
		}

		return shardCoordinator.CurrentShard
	}

	txIntercepted, err := transaction.NewInterceptedTransaction(
		txBuff,
		marshalizer,
		mock.HasherMock{},
		createKeyGenMock(),
		createDummySigner(),
		&mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return mock.NewAddressMock(pubKey), nil
			},
		},
		shardCoordinator,
	)

	assert.Nil(t, err)
	assert.Equal(t, uint32(1), txIntercepted.RcvShard())
	assert.Equal(t, uint32(1), txIntercepted.SndShard())
}
