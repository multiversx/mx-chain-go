package rewardTransaction_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/rewardTransaction"
	"github.com/stretchr/testify/assert"
)

func TestNewInterceptedRewardTransaction_NilTxBuffShouldErr(t *testing.T) {
	t.Parallel()

	irt, err := rewardTransaction.NewInterceptedRewardTransaction(
		nil,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3))

	assert.Nil(t, irt)
	assert.Equal(t, process.ErrNilBuffer, err)
}

func TestNewInterceptedRewardTransaction_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	txBuff := []byte("tx")
	irt, err := rewardTransaction.NewInterceptedRewardTransaction(
		txBuff,
		nil,
		&mock.HasherMock{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3))

	assert.Nil(t, irt)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewInterceptedRewardTransaction_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	txBuff := []byte("tx")
	irt, err := rewardTransaction.NewInterceptedRewardTransaction(
		txBuff,
		&mock.MarshalizerMock{},
		nil,
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3))

	assert.Nil(t, irt)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewInterceptedRewardTransaction_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	txBuff := []byte("tx")
	irt, err := rewardTransaction.NewInterceptedRewardTransaction(
		txBuff,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		nil,
		mock.NewMultiShardsCoordinatorMock(3))

	assert.Nil(t, irt)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewInterceptedRewardTransaction_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	txBuff := []byte("tx")
	irt, err := rewardTransaction.NewInterceptedRewardTransaction(
		txBuff,
		&mock.MarshalizerMock{},
		&mock.HasherMock{},
		&mock.AddressConverterMock{},
		nil)

	assert.Nil(t, irt)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewInterceptedRewardTransaction_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	rewTx := rewardTx.RewardTx{
		Round:   0,
		Epoch:   0,
		Value:   new(big.Int).SetInt64(100),
		RcvAddr: []byte("receiver"),
		ShardId: 0,
	}

	marshalizer := &mock.MarshalizerMock{}
	txBuff, _ := marshalizer.Marshal(rewTx)
	irt, err := rewardTransaction.NewInterceptedRewardTransaction(
		txBuff,
		marshalizer,
		&mock.HasherMock{},
		&mock.AddressConverterMock{},
		mock.NewMultiShardsCoordinatorMock(3))

	assert.NotNil(t, irt)
	assert.Nil(t, err)
}

func TestNewInterceptedRewardTransaction_TestGetters(t *testing.T) {
	t.Parallel()

	shardId := uint32(0)
	rewTx := rewardTx.RewardTx{
		Round:   0,
		Epoch:   0,
		Value:   new(big.Int).SetInt64(100),
		RcvAddr: []byte("receiver"),
		ShardId: shardId,
	}

	marshalizer := &mock.MarshalizerMock{}
	shardCoord := mock.NewMultiShardsCoordinatorMock(3)
	shardCoord.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		return shardId
	}

	txBuff, _ := marshalizer.Marshal(rewTx)
	irt, err := rewardTransaction.NewInterceptedRewardTransaction(
		txBuff,
		marshalizer,
		&mock.HasherMock{},
		&mock.AddressConverterMock{},
		shardCoord)

	assert.NotNil(t, irt)
	assert.Nil(t, err)

	assert.Equal(t, shardId, irt.RcvShard())
	assert.Equal(t, shardId, irt.SndShard())
	assert.Equal(t, &rewTx, irt.RewardTransaction())
	assert.False(t, irt.IsAddressedToOtherShards())

	txHash := irt.Hasher().Compute(string(txBuff))
	assert.Equal(t, txHash, irt.Hash())
}
