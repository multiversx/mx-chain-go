package node_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/fullHistory"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_GetTransaction_InvalidHashShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()
	_, err := n.GetTransaction("zzz")
	assert.Error(t, err)
}

func TestNode_GetTransaction_FromPool(t *testing.T) {
	t.Parallel()

	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfShardId: 1,
		ComputeIdCalled: func(address []byte) uint32 {
			if address == nil {
				return core.MetachainShardId
			}
			if bytes.Equal(address, []byte("alice")) {
				return 1
			}
			if bytes.Equal(address, []byte("bob")) {
				return 2
			}
			panic("bad test")
		},
	}

	dataPool := testscommon.NewPoolsHolderMock()
	n, err := node.NewNode(
		node.WithDataPool(dataPool),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(shardCoordinator),
	)
	require.Nil(t, err)

	// Normal transactions

	// Cross-shard, we are source
	txA := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	dataPool.Transactions().AddData([]byte("a"), txA, 42, "1")
	// Cross-shard, we are destination
	txB := &transaction.Transaction{Nonce: 7, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	dataPool.Transactions().AddData([]byte("b"), txB, 42, "1")
	// Intra-shard
	txC := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	dataPool.Transactions().AddData([]byte("c"), txC, 42, "1")

	actualA, err := n.GetTransaction(hex.EncodeToString([]byte("a")))
	require.Nil(t, err)
	actualB, err := n.GetTransaction(hex.EncodeToString([]byte("b")))
	require.Nil(t, err)
	actualC, err := n.GetTransaction(hex.EncodeToString([]byte("c")))
	require.Nil(t, err)

	require.Equal(t, txA.Nonce, actualA.Nonce)
	require.Equal(t, txB.Nonce, actualB.Nonce)
	require.Equal(t, txC.Nonce, actualC.Nonce)
	require.Equal(t, transaction.TxStatusReceived, actualA.Status)
	require.Equal(t, transaction.TxStatusPartiallyExecuted, actualB.Status)
	require.Equal(t, transaction.TxStatusReceived, actualC.Status)

	// Reward transactions

	txD := &rewardTx.RewardTx{Round: 42, RcvAddr: []byte("alice")}
	dataPool.RewardTransactions().AddData([]byte("d"), txD, 42, "foo")

	actualD, err := n.GetTransaction(hex.EncodeToString([]byte("d")))
	require.Nil(t, err)
	require.Equal(t, txD.Round, actualD.Round)
	require.Equal(t, transaction.TxStatusPartiallyExecuted, actualD.Status)

	// Unsigned transactions

	// Cross-shard, we are source
	txE := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	dataPool.UnsignedTransactions().AddData([]byte("e"), txE, 42, "foo")
	// Cross-shard, we are destination
	txF := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	dataPool.UnsignedTransactions().AddData([]byte("f"), txF, 42, "foo")
	// Intra-shard
	txG := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	dataPool.UnsignedTransactions().AddData([]byte("g"), txG, 42, "foo")

	actualE, err := n.GetTransaction(hex.EncodeToString([]byte("e")))
	require.Nil(t, err)
	actualF, err := n.GetTransaction(hex.EncodeToString([]byte("f")))
	require.Nil(t, err)
	actualG, err := n.GetTransaction(hex.EncodeToString([]byte("g")))
	require.Nil(t, err)

	require.Equal(t, txE.GasLimit, actualE.GasLimit)
	require.Equal(t, txF.GasLimit, actualF.GasLimit)
	require.Equal(t, txG.GasLimit, actualG.GasLimit)
	require.Equal(t, transaction.TxStatusReceived, actualE.Status)
	require.Equal(t, transaction.TxStatusPartiallyExecuted, actualF.Status)
	require.Equal(t, transaction.TxStatusReceived, actualG.Status)
}

func TestNode_GetTransaction_FromStorage(t *testing.T) {
	t.Parallel()

	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfShardId: 1,
		ComputeIdCalled: func(address []byte) uint32 {
			if address == nil {
				return core.MetachainShardId
			}
			if bytes.Equal(address, []byte("alice")) {
				return 1
			}
			if bytes.Equal(address, []byte("bob")) {
				return 2
			}
			panic("bad test")
		},
	}

	transactionsStorer := genericmocks.NewStorerMock()
	rewardsStorer := genericmocks.NewStorerMock()
	unsignedStorer := genericmocks.NewStorerMock()
	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			if unitType == dataRetriever.TransactionUnit {
				return transactionsStorer
			}
			if unitType == dataRetriever.RewardTransactionUnit {
				return rewardsStorer
			}
			if unitType == dataRetriever.UnsignedTransactionUnit {
				return unsignedStorer
			}
			panic("bad test")
		},
	}

	marshalizer := &mock.MarshalizerFake{}
	dataPool := testscommon.NewPoolsHolderMock()

	n, err := node.NewNode(
		node.WithDataPool(dataPool),
		node.WithDataStore(storer),
		node.WithInternalMarshalizer(marshalizer, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(shardCoordinator),
		node.WithHistoryRepository(&testscommon.HistoryRepositoryStub{
			IsEnabledCalled: func() bool {
				return false
			},
		}),
	)
	require.Nil(t, err)

	// Normal transactions

	// Cross-shard, we are source
	txA := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	transactionsStorer.PutWithMarshalizer([]byte("a"), txA, marshalizer)
	// Cross-shard, we are destination
	txB := &transaction.Transaction{Nonce: 7, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	transactionsStorer.PutWithMarshalizer([]byte("b"), txB, marshalizer)
	// Intra-shard
	txC := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	transactionsStorer.PutWithMarshalizer([]byte("c"), txC, marshalizer)

	actualA, err := n.GetTransaction(hex.EncodeToString([]byte("a")))
	require.Nil(t, err)
	actualB, err := n.GetTransaction(hex.EncodeToString([]byte("b")))
	require.Nil(t, err)
	actualC, err := n.GetTransaction(hex.EncodeToString([]byte("c")))
	require.Nil(t, err)

	require.Equal(t, txA.Nonce, actualA.Nonce)
	require.Equal(t, txB.Nonce, actualB.Nonce)
	require.Equal(t, txC.Nonce, actualC.Nonce)
	require.Equal(t, transaction.TxStatusPartiallyExecuted, actualA.Status)
	require.Equal(t, transaction.TxStatusExecuted, actualB.Status)
	require.Equal(t, transaction.TxStatusExecuted, actualC.Status)

	// Reward transactions

	txD := &rewardTx.RewardTx{Round: 42, RcvAddr: []byte("alice")}
	rewardsStorer.PutWithMarshalizer([]byte("d"), txD, marshalizer)

	actualD, err := n.GetTransaction(hex.EncodeToString([]byte("d")))
	require.Nil(t, err)
	require.Equal(t, txD.Round, actualD.Round)
	require.Equal(t, transaction.TxStatusExecuted, actualD.Status)

	// Unsigned transactions

	// Cross-shard, we are source
	txE := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	unsignedStorer.PutWithMarshalizer([]byte("e"), txE, marshalizer)
	// Cross-shard, we are destination
	txF := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	unsignedStorer.PutWithMarshalizer([]byte("f"), txF, marshalizer)
	// Intra-shard
	txG := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	unsignedStorer.PutWithMarshalizer([]byte("g"), txG, marshalizer)

	actualE, err := n.GetTransaction(hex.EncodeToString([]byte("e")))
	require.Nil(t, err)
	actualF, err := n.GetTransaction(hex.EncodeToString([]byte("f")))
	require.Nil(t, err)
	actualG, err := n.GetTransaction(hex.EncodeToString([]byte("g")))
	require.Nil(t, err)

	require.Equal(t, txE.GasLimit, actualE.GasLimit)
	require.Equal(t, txF.GasLimit, actualF.GasLimit)
	require.Equal(t, txG.GasLimit, actualG.GasLimit)
	require.Equal(t, transaction.TxStatusPartiallyExecuted, actualE.Status)
	require.Equal(t, transaction.TxStatusExecuted, actualF.Status)
	require.Equal(t, transaction.TxStatusExecuted, actualG.Status)

	// Missing transaction
	tx, err := n.GetTransaction(hex.EncodeToString([]byte("missing")))
	require.Equal(t, node.ErrTransactionNotFound, err)
	require.Nil(t, tx)

	// Badly serialized transaction
	transactionsStorer.Put([]byte("badly-serialized"), []byte("this isn't good"))
	tx, err = n.GetTransaction(hex.EncodeToString([]byte("badly-serialized")))
	require.NotNil(t, err)
	require.Nil(t, tx)
}

func TestNode_GetFullHistoryTransaction(t *testing.T) {
	t.Parallel()

	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfShardId: 1,
		ComputeIdCalled: func(address []byte) uint32 {
			if address == nil {
				return core.MetachainShardId
			}
			if bytes.Equal(address, []byte("alice")) {
				return 1
			}
			if bytes.Equal(address, []byte("bob")) {
				return 2
			}
			panic("bad test")
		},
	}

	transactionsStorer := genericmocks.NewStorerMock()
	rewardsStorer := genericmocks.NewStorerMock()
	unsignedStorer := genericmocks.NewStorerMock()
	epochByHashStorer := genericmocks.NewStorerMock()
	miniblockHashByTxHashStorer := genericmocks.NewStorerMock()
	miniblocksMetadataStorer := genericmocks.NewStorerMock()

	storer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			if unitType == dataRetriever.TransactionUnit {
				return transactionsStorer
			}
			if unitType == dataRetriever.RewardTransactionUnit {
				return rewardsStorer
			}
			if unitType == dataRetriever.UnsignedTransactionUnit {
				return unsignedStorer
			}
			panic("bad test")
		},
	}

	marshalizer := &mock.MarshalizerFake{}
	dataPool := testscommon.NewPoolsHolderMock()

	repoHistory, err := fullHistory.NewHistoryRepository(fullHistory.HistoryRepositoryArguments{
		SelfShardID:                 1,
		Marshalizer:                 marshalizer,
		Hasher:                      &mock.HasherFake{},
		EpochByHashStorer:           epochByHashStorer,
		MiniblockHashByTxHashStorer: miniblockHashByTxHashStorer,
		MiniblocksMetadataStorer:    miniblocksMetadataStorer,
	})
	require.Nil(t, err)

	n, err := node.NewNode(
		node.WithDataPool(dataPool),
		node.WithDataStore(storer),
		node.WithInternalMarshalizer(marshalizer, 0),
		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		node.WithShardCoordinator(shardCoordinator),
		node.WithHistoryRepository(repoHistory),
	)
	require.Nil(t, err)

	// dummyTx, _ := getDummyNormalTx()
	// expectedTx := &transaction.ApiTransactionResult{
	// 	Type:             "normal",
	// 	Nonce:            dummyTx.Nonce,
	// 	Round:            round,
	// 	Epoch:            epoch,
	// 	Value:            dummyTx.Value.String(),
	// 	Receiver:         hex.EncodeToString(dummyTx.RcvAddr),
	// 	Sender:           hex.EncodeToString(dummyTx.SndAddr),
	// 	GasPrice:         dummyTx.GasPrice,
	// 	GasLimit:         dummyTx.GasLimit,
	// 	Data:             dummyTx.Data,
	// 	Code:             "",
	// 	Signature:        hex.EncodeToString(dummyTx.Signature),
	// 	SourceShard:      sndShard,
	// 	DestinationShard: rcvShard,
	// 	BlockNonce:       blockNonce,
	// 	MiniBlockHash:    hex.EncodeToString(mbHash),
	// 	BlockHash:        hex.EncodeToString(blockHash),
	// 	Status:           transaction.TxStatusExecuted,
	// }

	// Normal transactions

	// Cross-shard, we are source
	txA := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	transactionsStorer.PutWithMarshalizer([]byte("a"), txA, marshalizer)
	// Cross-shard, we are destination
	txB := &transaction.Transaction{Nonce: 7, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	transactionsStorer.PutWithMarshalizer([]byte("b"), txB, marshalizer)
	// Intra-shard
	txC := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	transactionsStorer.PutWithMarshalizer([]byte("c"), txC, marshalizer)

	actualA, err := n.GetTransaction(hex.EncodeToString([]byte("a")))
	require.Nil(t, err)
	actualB, err := n.GetTransaction(hex.EncodeToString([]byte("b")))
	require.Nil(t, err)
	actualC, err := n.GetTransaction(hex.EncodeToString([]byte("c")))
	require.Nil(t, err)

	require.Equal(t, txA.Nonce, actualA.Nonce)
	require.Equal(t, txB.Nonce, actualB.Nonce)
	require.Equal(t, txC.Nonce, actualC.Nonce)
	require.Equal(t, transaction.TxStatusPartiallyExecuted, actualA.Status)
	require.Equal(t, transaction.TxStatusExecuted, actualB.Status)
	require.Equal(t, transaction.TxStatusExecuted, actualC.Status)
}

// func TestNode_GetFullHistoryTransaction_TxNotInHistoryStorerShouldErr(t *testing.T) {
// 	dataPool := &testscommon.PoolsHolderStub{
// 		TransactionsCalled:         getCacherHandler(false, ""),
// 		RewardTransactionsCalled:   getCacherHandler(false, ""),
// 		UnsignedTransactionsCalled: getCacherHandler(false, ""),
// 	}
// 	n, _ := node.NewNode(
// 		node.WithDataPool(dataPool),
// 		node.WithInternalMarshalizer(&mock.MarshalizerFake{}, 0),
// 		node.WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
// 		node.WithShardCoordinator(&mock.ShardCoordinatorMock{}),
// 		node.WithHistoryRepository(&testscommon.HistoryRepositoryStub{
// 			IsEnabledCalled: func() bool {
// 				return true
// 			},
// 			GetMiniblockMetadataByTxHashCalled: func(hash []byte) (*fullHistory.MiniblockMetadata, error) {
// 				return nil, fmt.Errorf("cannot")
// 			},
// 		}),
// 	)

// 	tx, err := n.GetTransaction("aaaa")
// 	assert.Equal(t, node.ErrTransactionNotFound, err)
// 	assert.Nil(t, tx)
// }

func TestNode_PutHistoryFieldsInTransaction(t *testing.T) {
	tx := &transaction.ApiTransactionResult{}
	metadata := &fullHistory.MiniblockMetadata{
		Epoch:                             42,
		Round:                             4321,
		MiniblockHash:                     []byte{15},
		DestinationShardID:                12,
		SourceShardID:                     11,
		HeaderNonce:                       4300,
		HeaderHash:                        []byte{14},
		NotarizedAtSourceInMetaNonce:      4250,
		NotarizedAtSourceInMetaHash:       []byte{13},
		NotarizedAtDestinationInMetaNonce: 4253,
		NotarizedAtDestinationInMetaHash:  []byte{12},
	}

	node.PutHistoryFieldsInTransaction(tx, metadata)

	require.Equal(t, 42, int(tx.Epoch))
	require.Equal(t, 4321, int(tx.Round))
	require.Equal(t, "0f", tx.MiniBlockHash)
	require.Equal(t, 12, int(tx.DestinationShard))
	require.Equal(t, 11, int(tx.SourceShard))
	require.Equal(t, 4300, int(tx.BlockNonce))
	require.Equal(t, "0e", tx.BlockHash)
	require.Equal(t, 4250, int(tx.NotarizedAtSourceInMetaNonce))
	require.Equal(t, "0d", tx.NotarizedAtSourceInMetaHash)
	require.Equal(t, 4253, int(tx.NotarizedAtDestinationInMetaNonce))
	require.Equal(t, "0c", tx.NotarizedAtDestinationInMetaHash)
}
