package node_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNode_GetTransaction_ThrottlerCannotProcessShouldErr(t *testing.T) {
	t.Parallel()

	throttler := &mock.ThrottlerStub{
		CanProcessCalled: func() bool {
			return false
		},
	}
	n, _ := node.NewNode(
		node.WithApiTransactionByHashThrottler(throttler),
	)
	_, err := n.GetTransaction("aaa")
	assert.Equal(t, node.ErrSystemBusyTxHash, err)
}

func TestNode_GetTransaction_InvalidHashShouldErr(t *testing.T) {
	t.Parallel()

	throttler := &mock.ThrottlerStub{
		CanProcessCalled: func() bool {
			return true
		},
	}
	n, _ := node.NewNode(
		node.WithApiTransactionByHashThrottler(throttler),
	)
	_, err := n.GetTransaction("zzz")
	assert.Error(t, err)
}

func TestNode_GetTransaction_ShouldFindInTxCacheAndReturn(t *testing.T) {
	t.Parallel()

	throttler := &mock.ThrottlerStub{
		CanProcessCalled: func() bool {
			return true
		},
	}
	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled: getCacherHandler(true, ""),
	}
	n, _ := node.NewNode(
		node.WithApiTransactionByHashThrottler(throttler),
		node.WithDataPool(dataPool),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
	)
	expectedTx, _ := getDummyNormalTx()
	tx, err := n.GetTransaction("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, expectedTx.Nonce, tx.Nonce)
}

func TestNode_GetTransaction_ShouldFindInRwdTxCacheAndReturn(t *testing.T) {
	t.Parallel()

	throttler := &mock.ThrottlerStub{
		CanProcessCalled: func() bool {
			return true
		},
	}
	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:       getCacherHandler(false, ""),
		RewardTransactionsCalled: getCacherHandler(true, "reward"),
	}
	n, _ := node.NewNode(
		node.WithApiTransactionByHashThrottler(throttler),
		node.WithDataPool(dataPool),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
	)
	expectedTx, _ := getDummyRewardTx()
	tx, err := n.GetTransaction("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, hex.EncodeToString(expectedTx.RcvAddr), tx.Receiver)
}

func TestNode_GetTransaction_ShouldFindInUnsignedTxCacheAndReturn(t *testing.T) {
	t.Parallel()

	throttler := &mock.ThrottlerStub{
		CanProcessCalled: func() bool {
			return true
		},
	}
	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:         getCacherHandler(false, ""),
		RewardTransactionsCalled:   getCacherHandler(false, ""),
		UnsignedTransactionsCalled: getCacherHandler(true, "unsigned"),
	}
	n, _ := node.NewNode(
		node.WithApiTransactionByHashThrottler(throttler),
		node.WithDataPool(dataPool),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
	)
	expectedTx, _ := getUnsignedTx()
	tx, err := n.GetTransaction("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, expectedTx.Nonce, tx.Nonce)
}

func TestNode_GetTransaction_ShouldFindInTxStorageAndReturn(t *testing.T) {
	t.Parallel()

	throttler := &mock.ThrottlerStub{
		CanProcessCalled: func() bool {
			return true
		},
	}
	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:         getCacherHandler(false, ""),
		RewardTransactionsCalled:   getCacherHandler(false, ""),
		UnsignedTransactionsCalled: getCacherHandler(false, ""),
	}
	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return getStorerStub(true)
		},
	}
	n, _ := node.NewNode(
		node.WithApiTransactionByHashThrottler(throttler),
		node.WithDataPool(dataPool),
		node.WithDataStore(storer),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
	)
	expectedTx, _ := getDummyNormalTx()
	tx, err := n.GetTransaction("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, expectedTx.Nonce, tx.Nonce)
}

func TestNode_GetTransaction_ShouldFindInRwdTxStorageAndReturn(t *testing.T) {
	t.Parallel()

	throttler := &mock.ThrottlerStub{
		CanProcessCalled: func() bool {
			return true
		},
	}
	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:         getCacherHandler(false, ""),
		RewardTransactionsCalled:   getCacherHandler(false, ""),
		UnsignedTransactionsCalled: getCacherHandler(false, ""),
	}
	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			if unitType == dataRetriever.TransactionUnit {
				return getStorerStub(false)
			}

			return getStorerStub(true)
		},
	}
	n, _ := node.NewNode(
		node.WithApiTransactionByHashThrottler(throttler),
		node.WithDataPool(dataPool),
		node.WithDataStore(storer),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
	)
	expectedTx, _ := getDummyNormalTx()
	tx, err := n.GetTransaction("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, hex.EncodeToString(expectedTx.RcvAddr), tx.Receiver)
}

func TestNode_GetTransaction_ShouldFindInUnsignedTxStorageAndReturn(t *testing.T) {
	t.Parallel()

	throttler := &mock.ThrottlerStub{
		CanProcessCalled: func() bool {
			return true
		},
	}
	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:         getCacherHandler(false, ""),
		RewardTransactionsCalled:   getCacherHandler(false, ""),
		UnsignedTransactionsCalled: getCacherHandler(false, ""),
	}
	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			switch unitType {
			case dataRetriever.UnsignedTransactionUnit:
				return getStorerStub(true)
			default:
				return getStorerStub(false)
			}
		},
	}
	n, _ := node.NewNode(
		node.WithApiTransactionByHashThrottler(throttler),
		node.WithDataPool(dataPool),
		node.WithDataStore(storer),
		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
	)
	expectedTx, _ := getDummyNormalTx()
	tx, err := n.GetTransaction("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, expectedTx.Nonce, tx.Nonce)
}

func TestNode_GetTransaction_ShouldFindInStorageButErrorUnmarshaling(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("error unmarshalling")

	throttler := &mock.ThrottlerStub{
		CanProcessCalled: func() bool {
			return true
		},
	}
	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:         getCacherHandler(false, ""),
		RewardTransactionsCalled:   getCacherHandler(false, ""),
		UnsignedTransactionsCalled: getCacherHandler(false, ""),
	}
	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			switch unitType {
			case dataRetriever.UnsignedTransactionUnit:
				return getStorerStub(true)
			default:
				return getStorerStub(false)
			}
		},
	}
	n, _ := node.NewNode(
		node.WithApiTransactionByHashThrottler(throttler),
		node.WithDataPool(dataPool),
		node.WithDataStore(storer),
		node.WithInternalMarshalizer(&mock.MarshalizerMock{
			UnmarshalHandler: func(_ interface{}, _ []byte) error {
				return expectedErr
			},
		}, 0),
		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
	)
	tx, err := n.GetTransaction("aaaa")
	assert.Nil(t, tx)
	assert.Equal(t, expectedErr, err)
}

func TestNode_GetTransaction_ShouldNotFindAndReturnUnknown(t *testing.T) {
	t.Parallel()

	throttler := &mock.ThrottlerStub{
		CanProcessCalled: func() bool {
			return true
		},
	}
	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:         getCacherHandler(false, ""),
		RewardTransactionsCalled:   getCacherHandler(false, ""),
		UnsignedTransactionsCalled: getCacherHandler(false, ""),
	}
	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return getStorerStub(false)
		},
	}
	n, _ := node.NewNode(
		node.WithApiTransactionByHashThrottler(throttler),
		node.WithDataPool(dataPool),
		node.WithDataStore(storer),
	)
	tx, err := n.GetTransaction("aaaa")
	assert.Nil(t, tx)
	assert.Error(t, err)
}

func TestNode_ComputeTransactionStatus(t *testing.T) {
	t.Parallel()

	throttler := &mock.ThrottlerStub{
		CanProcessCalled: func() bool {
			return true
		},
	}
	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return getStorerStub(false)
		},
	}

	shardZeroAddr := []byte("addrShard0")
	shardOneAddr := []byte("addrShard1")
	shardCoordinator := &mock.ShardCoordinatorMock{
		ComputeIdCalled: func(addr []byte) uint32 {
			if bytes.Equal(shardZeroAddr, addr) {
				return 0
			}
			return 1
		},
	}

	n, _ := node.NewNode(
		node.WithApiTransactionByHashThrottler(throttler),
		node.WithDataStore(storer),
		node.WithShardCoordinator(shardCoordinator),
	)

	rwdTxCrossShard := &rewardTx.RewardTx{RcvAddr: shardZeroAddr}
	normalTxIntraShard := &transaction.Transaction{RcvAddr: shardZeroAddr, SndAddr: shardZeroAddr}
	normalTxCrossShard := &transaction.Transaction{RcvAddr: shardOneAddr, SndAddr: shardZeroAddr}
	unsignedTxIntraShard := &smartContractResult.SmartContractResult{RcvAddr: shardZeroAddr, SndAddr: shardZeroAddr}
	unsignedTxCrossShard := &smartContractResult.SmartContractResult{RcvAddr: shardOneAddr, SndAddr: shardZeroAddr}

	// cross shard reward tx in storage source shard
	shardCoordinator.SelfShardId = core.MetachainShardId
	txStatus := n.ComputeTransactionStatus(rwdTxCrossShard, false)
	assert.Equal(t, core.TxStatusPartiallyExecuted, txStatus)

	// cross shard reward tx in pool source shard
	shardCoordinator.SelfShardId = core.MetachainShardId
	txStatus = n.ComputeTransactionStatus(rwdTxCrossShard, true)
	assert.Equal(t, core.TxStatusReceived, txStatus)

	// intra shard transaction in storage
	shardCoordinator.SelfShardId = 0
	txStatus = n.ComputeTransactionStatus(normalTxIntraShard, false)
	assert.Equal(t, core.TxStatusExecuted, txStatus)

	// intra shard transaction in pool
	shardCoordinator.SelfShardId = 0
	txStatus = n.ComputeTransactionStatus(normalTxIntraShard, true)
	assert.Equal(t, core.TxStatusReceived, txStatus)

	// cross shard transaction in storage source shard
	shardCoordinator.SelfShardId = 0
	txStatus = n.ComputeTransactionStatus(normalTxCrossShard, false)
	assert.Equal(t, core.TxStatusPartiallyExecuted, txStatus)

	// cross shard transaction in pool source shard
	shardCoordinator.SelfShardId = 0
	txStatus = n.ComputeTransactionStatus(normalTxCrossShard, true)
	assert.Equal(t, core.TxStatusReceived, txStatus)

	// cross shard transaction in storage destination shard
	shardCoordinator.SelfShardId = 1
	txStatus = n.ComputeTransactionStatus(normalTxCrossShard, false)
	assert.Equal(t, core.TxStatusExecuted, txStatus)

	// cross shard transaction in pool destination shard
	shardCoordinator.SelfShardId = 1
	txStatus = n.ComputeTransactionStatus(normalTxCrossShard, true)
	assert.Equal(t, core.TxStatusPartiallyExecuted, txStatus)

	// intra shard scr in storage source shard
	shardCoordinator.SelfShardId = 0
	txStatus = n.ComputeTransactionStatus(unsignedTxIntraShard, false)
	assert.Equal(t, core.TxStatusExecuted, txStatus)

	// intra shard scr in pool source shard
	shardCoordinator.SelfShardId = 0
	txStatus = n.ComputeTransactionStatus(unsignedTxIntraShard, true)
	assert.Equal(t, core.TxStatusReceived, txStatus)

	// cross shard scr in storage source shard
	shardCoordinator.SelfShardId = 0
	txStatus = n.ComputeTransactionStatus(unsignedTxCrossShard, false)
	assert.Equal(t, core.TxStatusPartiallyExecuted, txStatus)

	// cross shard scr in pool source shard
	shardCoordinator.SelfShardId = 0
	txStatus = n.ComputeTransactionStatus(unsignedTxCrossShard, true)
	assert.Equal(t, core.TxStatusReceived, txStatus)
}

func getCacherHandler(find bool, cacherType string) func() dataRetriever.ShardedDataCacherNotifier {
	return func() dataRetriever.ShardedDataCacherNotifier {
		switch cacherType {
		case "reward":
			return &testscommon.ShardedDataStub{
				SearchFirstDataCalled: func(_ []byte) (interface{}, bool) {
					if find {
						tx, _ := getDummyRewardTx()
						return tx, true
					}

					return nil, false
				},
			}
		case "unsigned":
			return &testscommon.ShardedDataStub{
				SearchFirstDataCalled: func(_ []byte) (interface{}, bool) {
					if find {
						tx, _ := getUnsignedTx()
						return tx, true
					}

					return nil, false
				},
			}
		default:
			return &testscommon.ShardedDataStub{
				SearchFirstDataCalled: func(_ []byte) (interface{}, bool) {
					if find {
						tx, _ := getDummyNormalTx()
						return tx, true
					}

					return nil, false
				},
			}
		}
	}
}

func getStorerStub(find bool) storage.Storer {
	return &mock.StorerStub{
		HasCalled: func(_ []byte) error {
			if !find {
				return errors.New("key not found")
			}
			return nil
		},
		SearchFirstCalled: func(_ []byte) ([]byte, error) {
			if !find {
				return nil, errors.New("key not found")
			}
			_, txBytes := getDummyNormalTx()
			return txBytes, nil
		},
	}
}

func getDummyNormalTx() (*transaction.Transaction, []byte) {
	tx := transaction.Transaction{Nonce: 37, RcvAddr: []byte("rcvr")}
	marshalizer := &mock.MarshalizerFake{}
	txBytes, _ := marshalizer.Marshal(&tx)
	return &tx, txBytes
}

func getDummyRewardTx() (*rewardTx.RewardTx, []byte) {
	tx := rewardTx.RewardTx{RcvAddr: []byte("rcvr")}
	marshalizer := &mock.MarshalizerFake{}
	txBytes, _ := marshalizer.Marshal(&tx)
	return &tx, txBytes
}

func getUnsignedTx() (*smartContractResult.SmartContractResult, []byte) {
	tx := smartContractResult.SmartContractResult{RcvAddr: []byte("rcvr")}
	marshalizer := &mock.MarshalizerFake{}
	txBytes, _ := marshalizer.Marshal(&tx)
	return &tx, txBytes
}
