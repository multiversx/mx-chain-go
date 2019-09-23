package rewardTransaction_test

import (
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/rewardTransaction"
	"github.com/stretchr/testify/assert"
)

func TestNewRewardTxInterceptor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	rti, err := rewardTransaction.NewRewardTxInterceptor(
		nil,
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.AddressConverterMock{},
		&mock.HasherMock{},
		mock.NewMultiShardsCoordinatorMock(3))

	assert.Nil(t, rti)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewRewardTxInterceptor_NilRewardTxPoolShouldErr(t *testing.T) {
	t.Parallel()

	rti, err := rewardTransaction.NewRewardTxInterceptor(
		&mock.MarshalizerMock{},
		nil,
		&mock.StorerStub{},
		&mock.AddressConverterMock{},
		&mock.HasherMock{},
		mock.NewMultiShardsCoordinatorMock(3))

	assert.Nil(t, rti)
	assert.Equal(t, process.ErrNilRewardTxDataPool, err)
}

func TestNewRewardTxInterceptor_NilRewardTxStorerShouldErr(t *testing.T) {
	t.Parallel()

	rti, err := rewardTransaction.NewRewardTxInterceptor(
		&mock.MarshalizerMock{},
		&mock.ShardedDataStub{},
		nil,
		&mock.AddressConverterMock{},
		&mock.HasherMock{},
		mock.NewMultiShardsCoordinatorMock(3))

	assert.Nil(t, rti)
	assert.Equal(t, process.ErrNilRewardsTxStorage, err)
}

func TestNewRewardTxInterceptor_NilAddressConverterShouldErr(t *testing.T) {
	t.Parallel()

	rti, err := rewardTransaction.NewRewardTxInterceptor(
		&mock.MarshalizerMock{},
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		nil,
		&mock.HasherMock{},
		mock.NewMultiShardsCoordinatorMock(3))

	assert.Nil(t, rti)
	assert.Equal(t, process.ErrNilAddressConverter, err)
}

func TestNewRewardTxInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	rti, err := rewardTransaction.NewRewardTxInterceptor(
		&mock.MarshalizerMock{},
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.AddressConverterMock{},
		nil,
		mock.NewMultiShardsCoordinatorMock(3))

	assert.Nil(t, rti)
	assert.Equal(t, process.ErrNilHasher, err)
}

func TestNewRewardTxInterceptor_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	rti, err := rewardTransaction.NewRewardTxInterceptor(
		&mock.MarshalizerMock{},
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.AddressConverterMock{},
		&mock.HasherMock{},
		nil)

	assert.Nil(t, rti)
	assert.Equal(t, process.ErrNilShardCoordinator, err)
}

func TestNewRewardTxInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	rti, err := rewardTransaction.NewRewardTxInterceptor(
		&mock.MarshalizerMock{},
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.AddressConverterMock{},
		&mock.HasherMock{},
		mock.NewMultiShardsCoordinatorMock(3))

	assert.NotNil(t, rti)
	assert.Nil(t, err)
	assert.False(t, rti.IsInterfaceNil())
}

func TestRewardTxInterceptor_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	rti, _ := rewardTransaction.NewRewardTxInterceptor(
		&mock.MarshalizerMock{},
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.AddressConverterMock{},
		&mock.HasherMock{},
		mock.NewMultiShardsCoordinatorMock(3))

	err := rti.ProcessReceivedMessage(nil)
	assert.Equal(t, process.ErrNilMessage, err)
}

func TestRewardTxInterceptor_ProcessReceivedMessageNilDataShouldErr(t *testing.T) {
	t.Parallel()

	rti, _ := rewardTransaction.NewRewardTxInterceptor(
		&mock.MarshalizerMock{},
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.AddressConverterMock{},
		&mock.HasherMock{},
		mock.NewMultiShardsCoordinatorMock(3))

	message := &mock.P2PMessageMock{
		DataField: nil,
	}

	err := rti.ProcessReceivedMessage(message)
	assert.Equal(t, process.ErrNilDataToProcess, err)
}

func TestRewardTxInterceptor_ProcessReceivedMessageIntraShardShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false
	rti, _ := rewardTransaction.NewRewardTxInterceptor(
		&mock.MarshalizerMock{},
		&mock.ShardedDataStub{
			AddDataCalled: func(key []byte, data interface{}, cacheId string) {
				wasCalled = true
			},
		},
		&mock.StorerStub{},
		&mock.AddressConverterMock{},
		&mock.HasherMock{},
		mock.NewMultiShardsCoordinatorMock(3))

	rewardTx1 := rewardTx.RewardTx{
		Round:   1,
		Epoch:   0,
		Value:   new(big.Int).SetInt64(157),
		RcvAddr: []byte("rcvr1"),
		ShardId: 0,
	}
	rewardTxBytes1, _ := rti.Marshalizer().Marshal(rewardTx1)

	rewardTx2 := rewardTx.RewardTx{
		Round:   0,
		Epoch:   1,
		Value:   new(big.Int).SetInt64(157),
		RcvAddr: []byte("rcvr2"),
		ShardId: 0,
	}
	rewardTxBytes2, _ := rti.Marshalizer().Marshal(rewardTx2)

	var rewardTxsSlice [][]byte
	rewardTxsSlice = append(rewardTxsSlice, rewardTxBytes1, rewardTxBytes2)
	rewardTxsBuff, _ := json.Marshal(rewardTxsSlice)

	message := &mock.P2PMessageMock{
		DataField: rewardTxsBuff,
	}

	err := rti.ProcessReceivedMessage(message)
	time.Sleep(20 * time.Millisecond)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestRewardTxInterceptor_ProcessReceivedMessageCrossShardShouldNotAdd(t *testing.T) {
	t.Parallel()

	wasCalled := false
	shardCoord := mock.NewMultiShardsCoordinatorMock(3)
	shardCoord.ComputeIdCalled = func(address state.AddressContainer) uint32 {
		return uint32(1)
	}
	rti, _ := rewardTransaction.NewRewardTxInterceptor(
		&mock.MarshalizerMock{},
		&mock.ShardedDataStub{
			AddDataCalled: func(key []byte, data interface{}, cacheId string) {
				wasCalled = true
			},
		},
		&mock.StorerStub{},
		&mock.AddressConverterMock{},
		&mock.HasherMock{},
		shardCoord)

	rewardTx1 := rewardTx.RewardTx{
		Round:   1,
		Epoch:   0,
		Value:   new(big.Int).SetInt64(157),
		RcvAddr: []byte("rcvr1"),
		ShardId: 1,
	}
	rewardTxBytes1, _ := rti.Marshalizer().Marshal(rewardTx1)

	rewardTx2 := rewardTx.RewardTx{
		Round:   0,
		Epoch:   1,
		Value:   new(big.Int).SetInt64(157),
		RcvAddr: []byte("rcvr2"),
		ShardId: 1,
	}
	rewardTxBytes2, _ := rti.Marshalizer().Marshal(rewardTx2)

	var rewardTxsSlice [][]byte
	rewardTxsSlice = append(rewardTxsSlice, rewardTxBytes1, rewardTxBytes2)
	rewardTxsBuff, _ := json.Marshal(rewardTxsSlice)

	message := &mock.P2PMessageMock{
		DataField: rewardTxsBuff,
	}

	err := rti.ProcessReceivedMessage(message)
	time.Sleep(20 * time.Millisecond)
	assert.Nil(t, err)
	// check that AddData was not called, as tx is cross shard
	assert.False(t, wasCalled)
}

func TestRewardTxInterceptor_SetBroadcastCallback(t *testing.T) {
	t.Parallel()

	rti, _ := rewardTransaction.NewRewardTxInterceptor(
		&mock.MarshalizerMock{},
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.AddressConverterMock{},
		&mock.HasherMock{},
		mock.NewMultiShardsCoordinatorMock(3))

	bytesToSend := []byte("test")
	var bytesToReceive []byte
	rti.SetBroadcastCallback(func(buffToSend []byte) {
		bytesToReceive = buffToSend
		return
	})

	rti.BroadcastCallbackHandler(bytesToSend)
	assert.Equal(t, bytesToSend, bytesToReceive)
}
