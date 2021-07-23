package node_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	coreMock "github.com/ElrondNetwork/elrond-go-core/core/mock"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dblookupext"
	"github.com/ElrondNetwork/elrond-go/node"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	dblookupextMock "github.com/ElrondNetwork/elrond-go/testscommon/dblookupext"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNode_GetTransactionInvalidHashShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode()
	_, err := n.GetTransaction("zzz", false)
	assert.Error(t, err)
}

func TestNode_GetTransactionFromPool(t *testing.T) {
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

	actualA, err := n.GetTransaction(hex.EncodeToString([]byte("a")), false)
	require.Nil(t, err)
	actualB, err := n.GetTransaction(hex.EncodeToString([]byte("b")), false)
	require.Nil(t, err)
	actualC, err := n.GetTransaction(hex.EncodeToString([]byte("c")), false)
	require.Nil(t, err)

	require.Equal(t, txA.Nonce, actualA.Nonce)
	require.Equal(t, uint32(1), actualA.SourceShard)
	require.Equal(t, uint32(2), actualA.DestinationShard)

	require.Equal(t, txB.Nonce, actualB.Nonce)
	require.Equal(t, uint32(2), actualB.SourceShard)
	require.Equal(t, uint32(1), actualB.DestinationShard)

	require.Equal(t, txC.Nonce, actualC.Nonce)
	require.Equal(t, uint32(1), actualC.SourceShard)
	require.Equal(t, uint32(1), actualC.DestinationShard)

	require.Equal(t, transaction.TxStatusPending, actualA.Status)
	require.Equal(t, transaction.TxStatusPending, actualB.Status)
	require.Equal(t, transaction.TxStatusPending, actualC.Status)

	// Reward transactions

	txD := &rewardTx.RewardTx{Round: 42, RcvAddr: []byte("alice")}
	dataPool.RewardTransactions().AddData([]byte("d"), txD, 42, "foo")

	actualD, err := n.GetTransaction(hex.EncodeToString([]byte("d")), false)
	require.Nil(t, err)
	require.Equal(t, txD.Round, actualD.Round)
	require.Equal(t, transaction.TxStatusPending, actualD.Status)

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

	actualE, err := n.GetTransaction(hex.EncodeToString([]byte("e")), false)
	require.Nil(t, err)
	actualF, err := n.GetTransaction(hex.EncodeToString([]byte("f")), false)
	require.Nil(t, err)
	actualG, err := n.GetTransaction(hex.EncodeToString([]byte("g")), false)
	require.Nil(t, err)

	require.Equal(t, txE.GasLimit, actualE.GasLimit)
	require.Equal(t, txF.GasLimit, actualF.GasLimit)
	require.Equal(t, txG.GasLimit, actualG.GasLimit)
	require.Equal(t, transaction.TxStatusPending, actualE.Status)
	require.Equal(t, transaction.TxStatusPending, actualF.Status)
	require.Equal(t, transaction.TxStatusPending, actualG.Status)
}

func TestNode_GetTransactionFromStorage(t *testing.T) {
	t.Parallel()

	n, chainStorer, _, _ := createNode(t, 0, false)

	// Normal transactions

	// Cross-shard, we are source
	internalMarshalizer := n.GetCoreComponents().InternalMarshalizer()
	txA := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	_ = chainStorer.Transactions.PutWithMarshalizer([]byte("a"), txA, internalMarshalizer)
	// Cross-shard, we are destination
	txB := &transaction.Transaction{Nonce: 7, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	_ = chainStorer.Transactions.PutWithMarshalizer([]byte("b"), txB, internalMarshalizer)
	// Intra-shard
	txC := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	_ = chainStorer.Transactions.PutWithMarshalizer([]byte("c"), txC, internalMarshalizer)

	actualA, err := n.GetTransaction(hex.EncodeToString([]byte("a")), false)
	require.Nil(t, err)
	actualB, err := n.GetTransaction(hex.EncodeToString([]byte("b")), false)
	require.Nil(t, err)
	actualC, err := n.GetTransaction(hex.EncodeToString([]byte("c")), false)
	require.Nil(t, err)

	require.Equal(t, txA.Nonce, actualA.Nonce)
	require.Equal(t, txB.Nonce, actualB.Nonce)
	require.Equal(t, txC.Nonce, actualC.Nonce)
	require.Equal(t, transaction.TxStatusPending, actualA.Status)
	require.Equal(t, transaction.TxStatusSuccess, actualB.Status)
	require.Equal(t, transaction.TxStatusSuccess, actualC.Status)

	// Reward transactions

	txD := &rewardTx.RewardTx{Round: 42, RcvAddr: []byte("alice")}
	_ = chainStorer.Rewards.PutWithMarshalizer([]byte("d"), txD, internalMarshalizer)

	actualD, err := n.GetTransaction(hex.EncodeToString([]byte("d")), false)
	require.Nil(t, err)
	require.Equal(t, txD.Round, actualD.Round)
	require.Equal(t, transaction.TxStatusSuccess, actualD.Status)

	// Unsigned transactions

	// Cross-shard, we are source
	txE := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	_ = chainStorer.Unsigned.PutWithMarshalizer([]byte("e"), txE, internalMarshalizer)
	// Cross-shard, we are destination
	txF := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	_ = chainStorer.Unsigned.PutWithMarshalizer([]byte("f"), txF, internalMarshalizer)
	// Intra-shard
	txG := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	_ = chainStorer.Unsigned.PutWithMarshalizer([]byte("g"), txG, internalMarshalizer)

	actualE, err := n.GetTransaction(hex.EncodeToString([]byte("e")), false)
	require.Nil(t, err)
	actualF, err := n.GetTransaction(hex.EncodeToString([]byte("f")), false)
	require.Nil(t, err)
	actualG, err := n.GetTransaction(hex.EncodeToString([]byte("g")), false)
	require.Nil(t, err)

	require.Equal(t, txE.GasLimit, actualE.GasLimit)
	require.Equal(t, txF.GasLimit, actualF.GasLimit)
	require.Equal(t, txG.GasLimit, actualG.GasLimit)
	require.Equal(t, transaction.TxStatusPending, actualE.Status)
	require.Equal(t, transaction.TxStatusSuccess, actualF.Status)
	require.Equal(t, transaction.TxStatusSuccess, actualG.Status)

	// Missing transaction
	tx, err := n.GetTransaction(hex.EncodeToString([]byte("missing")), false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "transaction not found")
	require.Nil(t, tx)

	// Badly serialized transaction
	_ = chainStorer.Transactions.Put([]byte("badly-serialized"), []byte("this isn't good"))
	tx, err = n.GetTransaction(hex.EncodeToString([]byte("badly-serialized")), false)
	require.NotNil(t, err)
	require.Nil(t, tx)
}

func TestNode_GetTransactionWithResultsFromStorage(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerFake{}
	txHash := hex.EncodeToString([]byte("txHash"))
	tx := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	scResultHash := []byte("scHash")
	scResult := &smartContractResult.SmartContractResult{
		OriginalTxHash: []byte("txHash"),
	}

	resultHashesByTxHash := &dblookupext.ResultsHashesByTxHash{
		ScResultsHashesAndEpoch: []*dblookupext.ScResultsHashesAndEpoch{
			{
				Epoch:           0,
				ScResultsHashes: [][]byte{scResultHash},
			},
		},
	}

	chainStorer := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			switch unitType {
			case dataRetriever.TransactionUnit:
				return &testscommon.StorerStub{
					GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
						return marshalizer.Marshal(tx)
					},
				}
			case dataRetriever.UnsignedTransactionUnit:
				return &testscommon.StorerStub{
					GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
						return marshalizer.Marshal(scResult)
					},
				}
			case dataRetriever.TxLogsUnit:
				return &testscommon.StorerStub{
					GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
						return nil, errors.New("dummy")
					},
				}
			default:
				return nil
			}
		},
	}
	historyRepo := &dblookupextMock.HistoryRepositoryStub{
		GetMiniblockMetadataByTxHashCalled: func(hash []byte) (*dblookupext.MiniblockMetadata, error) {
			return &dblookupext.MiniblockMetadata{}, nil
		},
		GetEventsHashesByTxHashCalled: func(hash []byte, epoch uint32) (*dblookupext.ResultsHashesByTxHash, error) {
			return resultHashesByTxHash, nil
		},
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = marshalizer
	coreComponents.AddrPubKeyConv = &mock.PubkeyConverterMock{}
	dataComponents := getDefaultDataComponents()
	dataComponents.DataPool = testscommon.NewPoolsHolderMock()
	dataComponents.Store = chainStorer
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = &mock.ShardCoordinatorMock{}
	processComponents.HistoryRepositoryInternal = historyRepo

	n, err := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithProcessComponents(processComponents),
	)
	require.NoError(t, err)

	expectedTx := &transaction.ApiTransactionResult{
		Tx:            &transaction.Transaction{Nonce: tx.Nonce, RcvAddr: tx.RcvAddr, SndAddr: tx.SndAddr, Value: tx.Value},
		Nonce:         tx.Nonce,
		Receiver:      hex.EncodeToString(tx.RcvAddr),
		Sender:        hex.EncodeToString(tx.SndAddr),
		Status:        transaction.TxStatusSuccess,
		MiniBlockType: block.TxBlock.String(),
		Type:          "normal",
		Value:         "<nil>",
		SmartContractResults: []*transaction.ApiSmartContractResult{
			{
				Hash:           hex.EncodeToString(scResultHash),
				OriginalTxHash: txHash,
			},
		},
	}

	apiTx, err := n.GetTransaction(txHash, true)
	require.Nil(t, err)
	require.Equal(t, expectedTx, apiTx)
}

func TestNode_lookupHistoricalTransaction(t *testing.T) {
	t.Parallel()

	n, chainStorer, _, historyRepo := createNode(t, 42, true)

	// Normal transactions

	// Cross-shard, we are source
	internalMarshalizer := n.GetCoreComponents().InternalMarshalizer()
	txA := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	_ = chainStorer.Transactions.PutWithMarshalizer([]byte("a"), txA, internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.TxBlock, 1, 2, 42, nil, 0)

	actualA, err := n.GetTransaction(hex.EncodeToString([]byte("a")), false)
	require.Nil(t, err)
	require.Equal(t, txA.Nonce, actualA.Nonce)
	require.Equal(t, 42, int(actualA.Epoch))
	require.Equal(t, transaction.TxStatusPending, actualA.Status)

	// Cross-shard, we are destination
	txB := &transaction.Transaction{Nonce: 7, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	_ = chainStorer.Transactions.PutWithMarshalizer([]byte("b"), txB, internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.TxBlock, 2, 1, 42, nil, 0)

	actualB, err := n.GetTransaction(hex.EncodeToString([]byte("b")), false)
	require.Nil(t, err)
	require.Equal(t, txB.Nonce, actualB.Nonce)
	require.Equal(t, 42, int(actualB.Epoch))
	require.Equal(t, transaction.TxStatusSuccess, actualB.Status)

	// Intra-shard
	txC := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	_ = chainStorer.Transactions.PutWithMarshalizer([]byte("c"), txC, internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.TxBlock, 1, 1, 42, nil, 0)

	actualC, err := n.GetTransaction(hex.EncodeToString([]byte("c")), false)
	require.Nil(t, err)
	require.Equal(t, txC.Nonce, actualC.Nonce)
	require.Equal(t, 42, int(actualC.Epoch))
	require.Equal(t, transaction.TxStatusSuccess, actualC.Status)

	// Invalid transaction
	txInvalid := &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	_ = chainStorer.Transactions.PutWithMarshalizer([]byte("invalid"), txInvalid, n.GetCoreComponents().InternalMarshalizer())
	setupGetMiniblockMetadataByTxHash(historyRepo, block.InvalidBlock, 1, 1, 42, nil, 0)

	actualInvalid, err := n.GetTransaction(hex.EncodeToString([]byte("invalid")), false)
	require.Nil(t, err)
	require.Equal(t, txInvalid.Nonce, actualInvalid.Nonce)
	require.Equal(t, 42, int(actualInvalid.Epoch))
	require.Equal(t, string(transaction.TxTypeInvalid), actualInvalid.Type)
	require.Equal(t, transaction.TxStatusInvalid, actualInvalid.Status)

	// Reward transactions
	headerHash := []byte("hash")
	headerNonce := uint64(1)
	nonceBytes := n.GetCoreComponents().Uint64ByteSliceConverter().ToByteSlice(headerNonce)
	_ = chainStorer.HdrNonce.Put(nonceBytes, headerHash)
	txD := &rewardTx.RewardTx{Round: 42, RcvAddr: []byte("alice")}
	_ = chainStorer.Rewards.PutWithMarshalizer([]byte("d"), txD, internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.RewardsBlock, core.MetachainShardId, 1, 42, headerHash, headerNonce)

	actualD, err := n.GetTransaction(hex.EncodeToString([]byte("d")), false)
	require.Nil(t, err)
	require.Equal(t, 42, int(actualD.Epoch))
	require.Equal(t, string(transaction.TxTypeReward), actualD.Type)
	require.Equal(t, transaction.TxStatusSuccess, actualD.Status)

	// Unsigned transactions

	// Cross-shard, we are source
	txE := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}
	_ = chainStorer.Unsigned.PutWithMarshalizer([]byte("e"), txE, internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.SmartContractResultBlock, 1, 2, 42, nil, 0)

	actualE, err := n.GetTransaction(hex.EncodeToString([]byte("e")), false)
	require.Nil(t, err)
	require.Equal(t, 42, int(actualE.Epoch))
	require.Equal(t, txE.GasLimit, actualE.GasLimit)
	require.Equal(t, string(transaction.TxTypeUnsigned), actualE.Type)
	require.Equal(t, transaction.TxStatusPending, actualE.Status)

	// Cross-shard, we are destination
	txF := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("bob"), RcvAddr: []byte("alice")}
	_ = chainStorer.Unsigned.PutWithMarshalizer([]byte("f"), txF, internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.SmartContractResultBlock, 2, 1, 42, nil, 0)

	actualF, err := n.GetTransaction(hex.EncodeToString([]byte("f")), false)
	require.Nil(t, err)
	require.Equal(t, 42, int(actualF.Epoch))
	require.Equal(t, txF.GasLimit, actualF.GasLimit)
	require.Equal(t, string(transaction.TxTypeUnsigned), actualF.Type)
	require.Equal(t, transaction.TxStatusSuccess, actualF.Status)

	// Intra-shard
	txG := &smartContractResult.SmartContractResult{GasLimit: 15, SndAddr: []byte("alice"), RcvAddr: []byte("alice")}
	_ = chainStorer.Unsigned.PutWithMarshalizer([]byte("g"), txG, internalMarshalizer)
	setupGetMiniblockMetadataByTxHash(historyRepo, block.SmartContractResultBlock, 1, 1, 42, nil, 0)

	actualG, err := n.GetTransaction(hex.EncodeToString([]byte("g")), false)
	require.Nil(t, err)
	require.Equal(t, 42, int(actualG.Epoch))
	require.Equal(t, txG.GasLimit, actualG.GasLimit)
	require.Equal(t, string(transaction.TxTypeUnsigned), actualG.Type)
	require.Equal(t, transaction.TxStatusSuccess, actualG.Status)

	// Missing transaction
	historyRepo.GetMiniblockMetadataByTxHashCalled = func(hash []byte) (*dblookupext.MiniblockMetadata, error) {
		return nil, fmt.Errorf("fooError")
	}
	tx, err := n.GetTransaction(hex.EncodeToString([]byte("g")), false)
	require.Nil(t, tx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "transaction not found")
	require.Contains(t, err.Error(), "fooError")

	// Badly serialized transaction
	_ = chainStorer.Transactions.Put([]byte("badly-serialized"), []byte("this isn't good"))
	historyRepo.GetMiniblockMetadataByTxHashCalled = func(hash []byte) (*dblookupext.MiniblockMetadata, error) {
		return &dblookupext.MiniblockMetadata{}, nil
	}
	tx, err = n.GetTransaction(hex.EncodeToString([]byte("badly-serialized")), false)
	require.NotNil(t, err)
	require.Nil(t, tx)

	// Reward reverted transaction
	wrongHeaderHash := []byte("wrong-hash")
	headerHash = []byte("hash")
	headerNonce = uint64(1)
	nonceBytes = n.GetCoreComponents().Uint64ByteSliceConverter().ToByteSlice(headerNonce)
	_ = chainStorer.HdrNonce.Put(nonceBytes, headerHash)
	txH := &rewardTx.RewardTx{Round: 50, RcvAddr: []byte("alice")}
	_ = chainStorer.Rewards.PutWithMarshalizer([]byte("h"), txH, n.GetCoreComponents().InternalMarshalizer())
	setupGetMiniblockMetadataByTxHash(historyRepo, block.RewardsBlock, core.MetachainShardId, 1, 42, wrongHeaderHash, headerNonce)

	actualH, err := n.GetTransaction(hex.EncodeToString([]byte("h")), false)
	require.Nil(t, err)
	require.Equal(t, 42, int(actualD.Epoch))
	require.Equal(t, string(transaction.TxTypeReward), actualH.Type)
	require.Equal(t, transaction.TxStatusRewardReverted, actualH.Status)

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

	node.PutMiniblockFieldsInTransaction(tx, metadata)

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

func createNode(t *testing.T, epoch uint32, withDbLookupExt bool) (*node.Node, *genericMocks.ChainStorerMock, *testscommon.PoolsHolderMock, *dblookupextMock.HistoryRepositoryStub) {
	chainStorer := genericMocks.NewChainStorerMock(epoch)
	dataPool := testscommon.NewPoolsHolderMock()
	marshalizer := &mock.MarshalizerFake{}

	historyRepo := &dblookupextMock.HistoryRepositoryStub{
		IsEnabledCalled: func() bool {
			return withDbLookupExt
		},
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.IntMarsh = marshalizer
	coreComponents.AddrPubKeyConv = &mock.PubkeyConverterMock{}
	coreComponents.UInt64ByteSliceConv = mock.NewNonceHashConverterMock()
	dataComponents := getDefaultDataComponents()
	dataComponents.DataPool = dataPool
	dataComponents.Store = chainStorer
	processComponents := getDefaultProcessComponents()
	processComponents.ShardCoord = createShardCoordinator()
	processComponents.HistoryRepositoryInternal = historyRepo

	n, err := node.NewNode(
		node.WithCoreComponents(coreComponents),
		node.WithDataComponents(dataComponents),
		node.WithProcessComponents(processComponents),
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

func setupGetMiniblockMetadataByTxHash(
	historyRepo *dblookupextMock.HistoryRepositoryStub,
	blockType block.Type,
	sourceShard uint32,
	destinationShard uint32,
	epoch uint32,
	headerHash []byte,
	headerNonce uint64,
) {
	historyRepo.GetMiniblockMetadataByTxHashCalled = func(hash []byte) (*dblookupext.MiniblockMetadata, error) {
		return &dblookupext.MiniblockMetadata{
			Type:               int32(blockType),
			SourceShardID:      sourceShard,
			DestinationShardID: destinationShard,
			Epoch:              epoch,
			HeaderNonce:        headerNonce,
			HeaderHash:         headerHash,
		}, nil
	}
}

func TestPrepareUnsignedTx(t *testing.T) {
	t.Parallel()
	addrSize := 32
	scr1 := &smartContractResult.SmartContractResult{
		Nonce:          1,
		Value:          big.NewInt(2),
		SndAddr:        bytes.Repeat([]byte{0}, addrSize),
		RcvAddr:        bytes.Repeat([]byte{1}, addrSize),
		OriginalSender: []byte("invalid original sender"),
	}

	coreComponents := getDefaultCoreComponents()
	coreComponents.AddrPubKeyConv, _ = pubkeyConverter.NewBech32PubkeyConverter(addrSize, &coreMock.LoggerMock{})

	n, err := node.NewNode(
		node.WithCoreComponents(coreComponents),
	)
	assert.Nil(t, err)

	scrResult1, err := n.PrepareUnsignedTx(scr1)
	assert.Nil(t, err)
	expectedScr1 := &transaction.ApiTransactionResult{
		Tx:             scr1,
		Nonce:          1,
		Type:           string(transaction.TxTypeUnsigned),
		Value:          "2",
		Receiver:       "erd1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqsl6e0p7",
		Sender:         "erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu",
		OriginalSender: "",
	}
	assert.Equal(t, scrResult1, expectedScr1)

	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          3,
		Value:          big.NewInt(4),
		SndAddr:        bytes.Repeat([]byte{5}, addrSize),
		RcvAddr:        bytes.Repeat([]byte{6}, addrSize),
		OriginalSender: bytes.Repeat([]byte{7}, addrSize),
	}

	scrResult2, err := n.PrepareUnsignedTx(scr2)
	assert.Nil(t, err)
	expectedScr2 := &transaction.ApiTransactionResult{
		Tx:             scr2,
		Nonce:          3,
		Type:           string(transaction.TxTypeUnsigned),
		Value:          "4",
		Receiver:       "erd1qcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqwkh39e",
		Sender:         "erd1q5zs2pg9q5zs2pg9q5zs2pg9q5zs2pg9q5zs2pg9q5zs2pg9q5zsrqsks3",
		OriginalSender: "erd1qurswpc8qurswpc8qurswpc8qurswpc8qurswpc8qurswpc8qurstywtnm",
	}
	assert.Equal(t, scrResult2, expectedScr2)
}

func TestNode_ComputeTimestampForRound(t *testing.T) {
	genesis := getTime(t, "1596117600")
	n, _ := node.NewNode(
		node.WithGenesisTime(genesis),
		node.WithRoundDuration(6000),
	)

	res := n.ComputeTimestampForRound(0)
	require.Equal(t, int64(0), res)

	res = n.ComputeTimestampForRound(4837403)
	require.Equal(t, int64(1625142018), res)
}

func getTime(t *testing.T, timestamp string) time.Time {
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		require.NoError(t, err)
	}
	tm := time.Unix(i, 0)

	return tm
}
