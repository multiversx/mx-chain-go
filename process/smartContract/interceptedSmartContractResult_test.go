package smartContract_test

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/smartContract"
	"github.com/stretchr/testify/assert"
)

var senderShard = uint32(2)
var recvShard = uint32(3)
var senderAddress = []byte("sender")
var recvAddress = []byte("receiver")

func createInterceptedScrFromPlainScr(scr *smartContractResult.SmartContractResult) (*smartContract.InterceptedSmartContractResult, error) {
	marshalizer := &mock.MarshalizerMock{}
	txBuff, _ := marshalizer.Marshal(scr)

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

	return smartContract.NewInterceptedSmartContractResult(
		txBuff,
		marshalizer,
		mock.HasherMock{},
		&mock.AddressConverterStub{
			CreateAddressFromPublicKeyBytesCalled: func(pubKey []byte) (container state.AddressContainer, e error) {
				return mock.NewAddressMock(pubKey), nil
			},
		},
		shardCoordinator,
	)
}

func TestNewInterceptedSmartContractResult_NilBufferShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := smartContract.NewInterceptedSmartContractResult(
		nil,
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilBuffer, err)
}

func TestNewInterceptedSmartContractResult_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := smartContract.NewInterceptedSmartContractResult(
		make([]byte, 0),
		nil,
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedSmartContractResult_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := smartContract.NewInterceptedSmartContractResult(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		nil,
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedSmartContractResult_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := smartContract.NewInterceptedSmartContractResult(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		nil,
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewInterceptedSmartContractResult_NilCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := smartContract.NewInterceptedSmartContractResult(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		nil,
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptedSmartContractResult_UnmarshalingTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")

	txi, err := smartContract.NewInterceptedSmartContractResult(
		make([]byte, 0),
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errExpected
			},
		},
		mock.HasherMock{},
		&mock.AddressConverterMock{},
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, errExpected, err)
}

func TestNewInterceptedSmartContractResult_AddrConvFailsShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := smartContract.NewInterceptedSmartContractResult(
		[]byte("{}"),
		&mock.MarshalizerMock{},
		mock.HasherMock{},
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

func TestNewInterceptedSmartContractResult_NilTxHashShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(2),
		Data:    []byte("data"),
		RcvAddr: recvAddress,
		SndAddr: senderAddress,
		TxHash:  nil,
	}

	txi, err := createInterceptedScrFromPlainScr(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilTxHash, err)
}

func TestNewInterceptedSmartContractResult_NilSenderAddressShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(2),
		Data:    []byte("data"),
		RcvAddr: recvAddress,
		SndAddr: nil,
		TxHash:  []byte("TX"),
	}

	txi, err := createInterceptedScrFromPlainScr(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilSndAddr, err)
}

func TestNewInterceptedSmartContractResult_NilRecvAddressShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(2),
		Data:    []byte("data"),
		RcvAddr: nil,
		SndAddr: senderAddress,
		TxHash:  []byte("TX"),
	}

	txi, err := createInterceptedScrFromPlainScr(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilRcvAddr, err)
}

func TestNewInterceptedSmartContractResult_NilValueShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   nil,
		Data:    []byte("data"),
		RcvAddr: recvAddress,
		SndAddr: senderAddress,
		TxHash:  []byte("TX"),
	}

	txi, err := createInterceptedScrFromPlainScr(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilValue, err)
}

func TestNewInterceptedSmartContractResult_NilNegativeValueShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(-2),
		Data:    []byte("data"),
		RcvAddr: recvAddress,
		SndAddr: senderAddress,
		TxHash:  []byte("TX"),
	}

	txi, err := createInterceptedScrFromPlainScr(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNegativeValue, err)
}

func TestNewInterceptedSmartContractResult_InvalidSenderShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(2),
		Data:    []byte("data"),
		RcvAddr: recvAddress,
		SndAddr: []byte(""),
		TxHash:  []byte("TX"),
	}

	txi, err := createInterceptedScrFromPlainScr(tx)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilSndAddr, err)
}

func TestNewInterceptedSmartContractResult_ShouldWork(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(2),
		Data:    []byte("data"),
		RcvAddr: recvAddress,
		SndAddr: senderAddress,
		TxHash:  []byte("TX"),
	}

	txi, err := createInterceptedScrFromPlainScr(tx)

	assert.NotNil(t, txi)
	assert.Nil(t, err)
	assert.Equal(t, tx, txi.SmartContractResult())
}

func TestNewInterceptedSmartContractResult_OkValsGettersShouldWork(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:   1,
		Value:   big.NewInt(2),
		Data:    []byte("data"),
		RcvAddr: recvAddress,
		SndAddr: senderAddress,
		TxHash:  []byte("TX"),
	}

	txi, _ := createInterceptedScrFromPlainScr(tx)

	assert.Equal(t, senderShard, txi.SndShard())
	assert.Equal(t, recvShard, txi.RcvShard())
	assert.True(t, txi.IsAddressedToOtherShards())
	assert.Equal(t, tx, txi.SmartContractResult())
}
