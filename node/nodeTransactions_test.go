package node

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_GetTransaction_InvalidHashShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := NewNode()
	_, err := n.GetTransaction("zzz")
	assert.Error(t, err)
}

func TestNode_GetTransaction_FromPool(t *testing.T) {
	t.Parallel()

	n, _, dataPool, _ := createNode(t, 42, false)

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

	n, chainStorer, _, _ := createNode(t, 0, false)

	// Normal transactions

	// Cross-shard, we are source
	txA := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	chainStorer.Transactions.PutWithMarshalizer([]byte("a"), txA, n.internalMarshalizer)
	// Cross-shard, we are destination
	txB := &transaction.Transaction{Nonce: 7, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	chainStorer.Transactions.PutWithMarshalizer([]byte("b"), txB, n.internalMarshalizer)
	// Intra-shard
	txC := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	chainStorer.Transactions.PutWithMarshalizer([]byte("c"), txC, n.internalMarshalizer)

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
	chainStorer.Rewards.PutWithMarshalizer([]byte("d"), txD, n.internalMarshalizer)

	actualD, err := n.GetTransaction(hex.EncodeToString([]byte("d")))
	require.Nil(t, err)
	require.Equal(t, txD.Round, actualD.Round)
	require.Equal(t, transaction.TxStatusExecuted, actualD.Status)

	// Unsigned transactions

	// Cross-shard, we are source
	txE := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	chainStorer.Unsigned.PutWithMarshalizer([]byte("e"), txE, n.internalMarshalizer)
	// Cross-shard, we are destination
	txF := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	chainStorer.Unsigned.PutWithMarshalizer([]byte("f"), txF, n.internalMarshalizer)
	// Intra-shard
	txG := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	chainStorer.Unsigned.PutWithMarshalizer([]byte("g"), txG, n.internalMarshalizer)

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
	require.Contains(t, err.Error(), "transaction not found")
	require.Nil(t, tx)

	// Badly serialized transaction
	chainStorer.Transactions.Put([]byte("badly-serialized"), []byte("this isn't good"))
	tx, err = n.GetTransaction(hex.EncodeToString([]byte("badly-serialized")))
	require.NotNil(t, err)
	require.Nil(t, tx)
}

func TestNode_lookupHistoricalTransaction(t *testing.T) {
	t.Parallel()

	n, chainStorer, _, historyRepo := createNode(t, 42, true)

	// Normal transactions

	// Cross-shard, we are source
	txA := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	chainStorer.Transactions.PutWithMarshalizer([]byte("a"), txA, n.internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.TxBlock, 1, 2, 42)

	actualA, err := n.GetTransaction(hex.EncodeToString([]byte("a")))
	require.Nil(t, err)
	require.Equal(t, txA.Nonce, actualA.Nonce)
	require.Equal(t, 42, int(actualA.Epoch))
	require.Equal(t, transaction.TxStatusPartiallyExecuted, actualA.Status)

	// Cross-shard, we are destination
	txB := &transaction.Transaction{Nonce: 7, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	chainStorer.Transactions.PutWithMarshalizer([]byte("b"), txB, n.internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.TxBlock, 2, 1, 42)

	actualB, err := n.GetTransaction(hex.EncodeToString([]byte("b")))
	require.Nil(t, err)
	require.Equal(t, txB.Nonce, actualB.Nonce)
	require.Equal(t, 42, int(actualB.Epoch))
	require.Equal(t, transaction.TxStatusExecuted, actualB.Status)

	// Intra-shard
	txC := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	chainStorer.Transactions.PutWithMarshalizer([]byte("c"), txC, n.internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.TxBlock, 1, 1, 42)

	actualC, err := n.GetTransaction(hex.EncodeToString([]byte("c")))
	require.Nil(t, err)
	require.Equal(t, txC.Nonce, actualC.Nonce)
	require.Equal(t, 42, int(actualC.Epoch))
	require.Equal(t, transaction.TxStatusExecuted, actualC.Status)

	// Reward transactions

	txD := &rewardTx.RewardTx{Round: 42, RcvAddr: []byte("alice")}
	chainStorer.Rewards.PutWithMarshalizer([]byte("d"), txD, n.internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.RewardsBlock, core.MetachainShardId, 1, 42)

	actualD, err := n.GetTransaction(hex.EncodeToString([]byte("d")))
	require.Nil(t, err)
	require.Equal(t, 42, int(actualD.Epoch))
	require.Equal(t, transaction.TxStatusExecuted, actualD.Status)

	// Unsigned transactions

	// Cross-shard, we are source
	txE := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	chainStorer.Unsigned.PutWithMarshalizer([]byte("e"), txE, n.internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.SmartContractResultBlock, 1, 2, 42)

	actualE, err := n.GetTransaction(hex.EncodeToString([]byte("e")))
	require.Nil(t, err)
	require.Equal(t, 42, int(actualE.Epoch))
	require.Equal(t, txE.GasLimit, actualE.GasLimit)
	require.Equal(t, transaction.TxStatusPartiallyExecuted, actualE.Status)

	// Cross-shard, we are destination
	txF := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	chainStorer.Unsigned.PutWithMarshalizer([]byte("f"), txF, n.internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.SmartContractResultBlock, 2, 1, 42)

	actualF, err := n.GetTransaction(hex.EncodeToString([]byte("f")))
	require.Nil(t, err)
	require.Equal(t, 42, int(actualF.Epoch))
	require.Equal(t, txF.GasLimit, actualF.GasLimit)
	require.Equal(t, transaction.TxStatusExecuted, actualF.Status)

	// Intra-shard
	txG := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	chainStorer.Unsigned.PutWithMarshalizer([]byte("g"), txG, n.internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.SmartContractResultBlock, 1, 1, 42)

	actualG, err := n.GetTransaction(hex.EncodeToString([]byte("g")))
	require.Nil(t, err)
	require.Equal(t, 42, int(actualG.Epoch))
	require.Equal(t, txG.GasLimit, actualG.GasLimit)
	require.Equal(t, transaction.TxStatusExecuted, actualG.Status)

	// Missing transaction
	historyRepo.GetMiniblockMetadataByTxHashCalled = func(hash []byte) (*dblookupext.MiniblockMetadata, error) {
		return nil, fmt.Errorf("fooError")
	}
	tx, err := n.GetTransaction(hex.EncodeToString([]byte("g")))
	require.Contains(t, err.Error(), "transaction not found")
	require.Contains(t, err.Error(), "fooError")
	require.Nil(t, tx)

	// Badly serialized transaction
	chainStorer.Transactions.Put([]byte("badly-serialized"), []byte("this isn't good"))
	historyRepo.GetMiniblockMetadataByTxHashCalled = func(hash []byte) (*dblookupext.MiniblockMetadata, error) {
		return &dblookupext.MiniblockMetadata{}, nil
	}
	tx, err = n.GetTransaction(hex.EncodeToString([]byte("badly-serialized")))
	require.NotNil(t, err)
	require.Nil(t, tx)
}

func TestNode_PutHistoryFieldsInTransaction(t *testing.T) {
	tx := &transaction.ApiTransactionResult{}
	metadata := &dblookupext.MiniblockMetadata{
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

	PutMiniblockFieldsInTransaction(tx, metadata)

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

func createNode(t *testing.T, epoch uint32, withDbLookupExt bool) (*Node, *genericmocks.ChainStorerMock, *testscommon.PoolsHolderMock, *testscommon.HistoryRepositoryStub) {
	chainStorer := genericmocks.NewChainStorerMock(epoch)
	dataPool := testscommon.NewPoolsHolderMock()
	marshalizer := &mock.MarshalizerFake{}

	historyRepo := &testscommon.HistoryRepositoryStub{
		IsEnabledCalled: func() bool {
			return withDbLookupExt
		},
	}

	n, err := NewNode(
		WithDataPool(dataPool),
		WithDataStore(chainStorer),
		WithInternalMarshalizer(marshalizer, 0),
		WithAddressPubkeyConverter(&mock.PubkeyConverterMock{}),
		WithShardCoordinator(createShardCoordinator()),
		WithHistoryRepository(historyRepo),
	)

	require.Nil(t, err)
	return n, chainStorer, dataPool, historyRepo
}

func createShardCoordinator() *mock.ShardCoordinatorMock {
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

	return shardCoordinator
}

func setupGetMiniblockMetadataByTxHash(historyRepo *testscommon.HistoryRepositoryStub, blockType block.Type, sourceShard uint32, destinationShard uint32, epoch uint32) {
	historyRepo.GetMiniblockMetadataByTxHashCalled = func(hash []byte) (*dblookupext.MiniblockMetadata, error) {
		return &dblookupext.MiniblockMetadata{
			Type:               int32(blockType),
			SourceShardID:      sourceShard,
			DestinationShardID: destinationShard,
			Epoch:              epoch,
		}, nil
	}
}
