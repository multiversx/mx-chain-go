package unsigned_test

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"testing"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/unsigned"
	"github.com/stretchr/testify/assert"
)

var senderShard = uint32(2)
var recvShard = uint32(3)
var senderAddress = []byte("sender")
var recvAddress = []byte("receiver")

func createInterceptedScrFromPlainScr(scr *smartContractResult.SmartContractResult) (*unsigned.InterceptedUnsignedTransaction, error) {
	marshalizer := &mock.MarshalizerMock{}
	txBuff, _ := marshalizer.Marshal(scr)

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

	return unsigned.NewInterceptedUnsignedTransaction(
		txBuff,
		marshalizer,
		mock.HasherMock{},
		&mock.PubkeyConverterStub{},
		shardCoordinator,
	)
}

func createMockPubkeyConverter() *mock.PubkeyConverterMock {
	return mock.NewPubkeyConverterMock(32)
}

//------- NewInterceptedUnsignedTransaction

func TestNewInterceptedUnsignedTransaction_NilBufferShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := unsigned.NewInterceptedUnsignedTransaction(
		nil,
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilBuffer, err)
}

func TestNewInterceptedUnsignedTransaction_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := unsigned.NewInterceptedUnsignedTransaction(
		make([]byte, 0),
		nil,
		mock.HasherMock{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedUnsignedTransaction_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := unsigned.NewInterceptedUnsignedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		nil,
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedUnsignedTransaction_NilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := unsigned.NewInterceptedUnsignedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		nil,
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilPubkeyConverter, err)
}

func TestNewInterceptedUnsignedTransaction_NilCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	txi, err := unsigned.NewInterceptedUnsignedTransaction(
		make([]byte, 0),
		&mock.MarshalizerMock{},
		mock.HasherMock{},
		createMockPubkeyConverter(),
		nil,
	)

	assert.Nil(t, txi)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptedUnsignedTransaction_UnmarshalingTxFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")

	txi, err := unsigned.NewInterceptedUnsignedTransaction(
		make([]byte, 0),
		&mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return errExpected
			},
		},
		mock.HasherMock{},
		createMockPubkeyConverter(),
		mock.NewOneShardCoordinatorMock(),
	)

	assert.Nil(t, txi)
	assert.Equal(t, errExpected, err)
}

func TestNewInterceptedUnsignedTransaction_ShouldWork(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:      1,
		Value:      big.NewInt(2),
		Data:       []byte("data"),
		RcvAddr:    recvAddress,
		SndAddr:    senderAddress,
		PrevTxHash: []byte("TX"),
	}
	txi, err := createInterceptedScrFromPlainScr(tx)

	assert.False(t, check.IfNil(txi))
	assert.Nil(t, err)
}

//------- CheckValidity

func TestInterceptedUnsignedTransaction_CheckValidityNilTxHashShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:      1,
		Value:      big.NewInt(2),
		Data:       []byte("data"),
		RcvAddr:    recvAddress,
		SndAddr:    senderAddress,
		PrevTxHash: nil,
	}
	txi, _ := createInterceptedScrFromPlainScr(tx)

	err := txi.CheckValidity()

	assert.Equal(t, process.ErrNilTxHash, err)
}

func TestInterceptedUnsignedTransaction_CheckValidityNilSenderAddressShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:      1,
		Value:      big.NewInt(2),
		Data:       []byte("data"),
		RcvAddr:    recvAddress,
		SndAddr:    nil,
		PrevTxHash: []byte("TX"),
	}
	txi, _ := createInterceptedScrFromPlainScr(tx)

	err := txi.CheckValidity()

	assert.Equal(t, process.ErrNilSndAddr, err)
}

func TestInterceptedUnsignedTransaction_CheckValidityNilRecvAddressShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:      1,
		Value:      big.NewInt(2),
		Data:       []byte("data"),
		RcvAddr:    nil,
		SndAddr:    senderAddress,
		PrevTxHash: []byte("TX"),
	}
	txi, _ := createInterceptedScrFromPlainScr(tx)

	err := txi.CheckValidity()

	assert.Equal(t, process.ErrNilRcvAddr, err)
}

func TestInterceptedUnsignedTransaction_CheckValidityNilValueShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:      1,
		Value:      nil,
		Data:       []byte("data"),
		RcvAddr:    recvAddress,
		SndAddr:    senderAddress,
		PrevTxHash: []byte("TX"),
	}
	txi, _ := createInterceptedScrFromPlainScr(tx)

	err := txi.CheckValidity()

	assert.Equal(t, process.ErrNilValue, err)
}

func TestInterceptedUnsignedTransaction_CheckValidityNilNegativeValueShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:      1,
		Value:      big.NewInt(-2),
		Data:       []byte("data"),
		RcvAddr:    recvAddress,
		SndAddr:    senderAddress,
		PrevTxHash: []byte("TX"),
	}
	txi, _ := createInterceptedScrFromPlainScr(tx)

	err := txi.CheckValidity()

	assert.Equal(t, process.ErrNegativeValue, err)
}

func TestInterceptedUnsignedTransaction_CheckValidityInvalidSenderShouldErr(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:      1,
		Value:      big.NewInt(2),
		Data:       []byte("data"),
		RcvAddr:    recvAddress,
		SndAddr:    []byte(""),
		PrevTxHash: []byte("TX"),
	}
	txi, _ := createInterceptedScrFromPlainScr(tx)

	err := txi.CheckValidity()

	assert.Equal(t, process.ErrNilSndAddr, err)
}

func TestInterceptedUnsignedTransaction_CheckValidityShouldWork(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:      1,
		Value:      big.NewInt(2),
		Data:       []byte("data"),
		RcvAddr:    recvAddress,
		SndAddr:    senderAddress,
		PrevTxHash: []byte("TX"),
	}
	txi, _ := createInterceptedScrFromPlainScr(tx)

	err := txi.CheckValidity()
	assert.Nil(t, err)
	assert.Equal(t, tx, txi.Transaction())
}

//------- getters

func TestInterceptedUnsignedTransaction_OkValsGettersShouldWork(t *testing.T) {
	t.Parallel()

	nonce := uint64(45)
	tx := &smartContractResult.SmartContractResult{
		Nonce:      nonce,
		Value:      big.NewInt(2),
		Data:       []byte("data"),
		RcvAddr:    recvAddress,
		SndAddr:    senderAddress,
		PrevTxHash: []byte("TX"),
	}
	txi, _ := createInterceptedScrFromPlainScr(tx)

	marshalizer := &mock.MarshalizerMock{}
	hasher := mock.HasherMock{}
	expectedHash, _ := core.CalculateHash(marshalizer, hasher, tx)

	assert.Equal(t, senderShard, txi.SenderShardId())
	assert.Equal(t, recvShard, txi.ReceiverShardId())
	assert.False(t, txi.IsForCurrentShard())
	assert.Equal(t, tx, txi.Transaction())
	assert.Equal(t, expectedHash, txi.Hash())
	assert.Equal(t, nonce, txi.Nonce())
	assert.Equal(t, big.NewInt(0), txi.Fee())
	assert.Equal(t, senderAddress, txi.SenderAddress())
}

//------- IsInterfaceNil

func TestInterceptedTransaction_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var utxi *unsigned.InterceptedUnsignedTransaction

	assert.True(t, check.IfNil(utxi))
}

func TestInterceptedUnsignedTransaction_Type(t *testing.T) {
	t.Parallel()

	tx := &smartContractResult.SmartContractResult{
		Nonce:      1,
		Value:      big.NewInt(2),
		Data:       []byte("data"),
		RcvAddr:    recvAddress,
		SndAddr:    senderAddress,
		PrevTxHash: nil,
	}
	txi, _ := createInterceptedScrFromPlainScr(tx)

	expectedType := "intercepted unsigned tx"

	assert.Equal(t, expectedType, txi.Type())
}

func TestInterceptedUnsignedTransaction_String(t *testing.T) {
	t.Parallel()

	nonce := uint64(1)
	value := big.NewInt(200)
	tx := &smartContractResult.SmartContractResult{
		Nonce:      nonce,
		Value:      value,
		Data:       []byte("data"),
		RcvAddr:    recvAddress,
		SndAddr:    senderAddress,
		PrevTxHash: nil,
	}
	txi, _ := createInterceptedScrFromPlainScr(tx)

	expectedFormat := fmt.Sprintf(
		"sender=%s, nonce=%d, value=%s, recv=%s",
		logger.DisplayByteSlice(senderAddress), nonce, value, logger.DisplayByteSlice(recvAddress),
	)

	assert.Equal(t, expectedFormat, txi.String())
}
