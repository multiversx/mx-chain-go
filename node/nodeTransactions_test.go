package node_test

import (
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

func TestNode_GetTransactionStatus_ThrottlerCannotProcessShouldErr(t *testing.T) {
	t.Parallel()

	throttler := &mock.ThrottlerStub{
		CanProcessCalled: func() bool {
			return false
		},
	}
	n, _ := node.NewNode(
		node.WithApiTransactionByHashThrottler(throttler),
	)
	_, err := n.GetTransactionStatus("aaa")
	assert.Equal(t, node.ErrSystemBusyTxHash, err)
}

func TestNode_GetTransactionStatus_InvalidHashShouldErr(t *testing.T) {
	t.Parallel()

	throttler := &mock.ThrottlerStub{
		CanProcessCalled: func() bool {
			return true
		},
	}
	n, _ := node.NewNode(
		node.WithApiTransactionByHashThrottler(throttler),
	)
	_, err := n.GetTransactionStatus("zzz")
	assert.Error(t, err)
}

func TestNode_GetTransactionStatus_ShouldFindInTxCacheAndReturnReceived(t *testing.T) {
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
	)
	res, err := n.GetTransactionStatus("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, string(core.TxStatusReceived), res)
}

func TestNode_GetTransactionStatus_ShouldFindInRwdTxCacheAndReturnReceived(t *testing.T) {
	t.Parallel()

	throttler := &mock.ThrottlerStub{
		CanProcessCalled: func() bool {
			return true
		},
	}
	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:       getCacherHandler(false, ""),
		RewardTransactionsCalled: getCacherHandler(true, ""),
	}
	n, _ := node.NewNode(
		node.WithApiTransactionByHashThrottler(throttler),
		node.WithDataPool(dataPool),
	)
	res, err := n.GetTransactionStatus("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, string(core.TxStatusReceived), res)
}

func TestNode_GetTransactionStatus_ShouldFindInUnsignedTxCacheAndReturnReceived(t *testing.T) {
	t.Parallel()

	throttler := &mock.ThrottlerStub{
		CanProcessCalled: func() bool {
			return true
		},
	}
	dataPool := &testscommon.PoolsHolderStub{
		TransactionsCalled:         getCacherHandler(false, ""),
		RewardTransactionsCalled:   getCacherHandler(false, ""),
		UnsignedTransactionsCalled: getCacherHandler(true, ""),
	}
	n, _ := node.NewNode(
		node.WithApiTransactionByHashThrottler(throttler),
		node.WithDataPool(dataPool),
	)
	res, err := n.GetTransactionStatus("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, string(core.TxStatusReceived), res)
}

func TestNode_GetTransactionStatus_ShouldFindInTxStorageAndReturnExecuted(t *testing.T) {
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
	)
	res, err := n.GetTransactionStatus("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, string(core.TxStatusExecuted), res)
}

func TestNode_GetTransactionStatus_ShouldFindInRwdTxStorageAndReturnExecuted(t *testing.T) {
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
	)
	res, err := n.GetTransactionStatus("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, string(core.TxStatusExecuted), res)
}

func TestNode_GetTransactionStatus_ShouldFindInUnsignedTxStorageAndReturnExecuted(t *testing.T) {
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
	)
	res, err := n.GetTransactionStatus("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, string(core.TxStatusExecuted), res)
}

func TestNode_GetTransactionStatus_ShouldNotFindAndReturnUnknown(t *testing.T) {
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
	res, err := n.GetTransactionStatus("aaaa")
	assert.NoError(t, err)
	assert.Equal(t, string(core.TxStatusUnknown), res)
}

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
