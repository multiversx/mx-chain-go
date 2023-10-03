package transactionAPI

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/process"
	processMocks "github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	dblookupextMock "github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgAPITransactionProcessor() *ArgAPITransactionProcessor {
	return &ArgAPITransactionProcessor{
		RoundDuration:            0,
		GenesisTime:              time.Time{},
		Marshalizer:              &mock.MarshalizerFake{},
		AddressPubKeyConverter:   &testscommon.PubkeyConverterMock{},
		ShardCoordinator:         createShardCoordinator(),
		HistoryRepository:        &dblookupextMock.HistoryRepositoryStub{},
		StorageService:           &storageStubs.ChainStorerStub{},
		DataPool:                 &dataRetrieverMock.PoolsHolderMock{},
		Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
		FeeComputer:              &testscommon.FeeComputerStub{},
		TxTypeHandler:            &testscommon.TxTypeHandlerMock{},
		LogsFacade:               &testscommon.LogsFacadeStub{},
		DataFieldParser: &testscommon.DataFieldParserStub{
			ParseCalled: func(dataField []byte, sender, receiver []byte, _ uint32) *datafield.ResponseParseData {
				return &datafield.ResponseParseData{}
			},
		},
	}
}

func TestNewAPITransactionProcessor(t *testing.T) {
	t.Parallel()

	t.Run("NilArg", func(t *testing.T) {
		t.Parallel()

		_, err := NewAPITransactionProcessor(nil)
		require.Equal(t, ErrNilAPITransactionProcessorArg, err)
	})

	t.Run("NilMarshalizer", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPITransactionProcessor()
		arguments.Marshalizer = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, process.ErrNilMarshalizer, err)
	})

	t.Run("NilDataPool", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPITransactionProcessor()
		arguments.DataPool = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, process.ErrNilDataPoolHolder, err)
	})

	t.Run("NilHistoryRepository", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPITransactionProcessor()
		arguments.HistoryRepository = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, process.ErrNilHistoryRepository, err)
	})

	t.Run("NilShardCoordinator", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPITransactionProcessor()
		arguments.ShardCoordinator = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, process.ErrNilShardCoordinator, err)
	})

	t.Run("NilPubKeyConverter", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPITransactionProcessor()
		arguments.AddressPubKeyConverter = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, process.ErrNilPubkeyConverter, err)
	})

	t.Run("NilStorageService", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPITransactionProcessor()
		arguments.StorageService = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, process.ErrNilStorage, err)
	})

	t.Run("NilUint64Converter", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPITransactionProcessor()
		arguments.Uint64ByteSliceConverter = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, process.ErrNilUint64Converter, err)
	})

	t.Run("NilTxFeeComputer", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPITransactionProcessor()
		arguments.FeeComputer = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, ErrNilFeeComputer, err)
	})

	t.Run("NilTypeHandler", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPITransactionProcessor()
		arguments.TxTypeHandler = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, process.ErrNilTxTypeHandler, err)
	})

	t.Run("NilLogsFacade", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPITransactionProcessor()
		arguments.LogsFacade = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, ErrNilLogsFacade, err)
	})

	t.Run("NilDataFieldParser", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgAPITransactionProcessor()
		arguments.DataFieldParser = nil

		_, err := NewAPITransactionProcessor(arguments)
		require.Equal(t, ErrNilDataFieldParser, err)
	})
}

func TestNode_GetTransactionInvalidHashShouldErr(t *testing.T) {
	t.Parallel()

	n, _, _, _ := createAPITransactionProc(t, 0, false)
	_, err := n.GetTransaction("zzz", false)
	assert.Error(t, err)
}

func TestNode_GetTransactionFromPool(t *testing.T) {
	t.Parallel()

	n, _, dataPool, _ := createAPITransactionProc(t, 42, false)

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

	n, chainStorer, _, _ := createAPITransactionProc(t, 0, false)

	// Cross-shard, we are source
	internalMarshalizer := &mock.MarshalizerFake{}
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

func TestNode_GetTransactionWithResultsFromStorageMissingStorer(t *testing.T) {
	t.Parallel()

	t.Run("missing TransactionUnit", testWithMissingStorer(dataRetriever.TransactionUnit))
	t.Run("missing UnsignedTransactionUnit", testWithMissingStorer(dataRetriever.UnsignedTransactionUnit))
	t.Run("missing RewardTransactionUnit", testWithMissingStorer(dataRetriever.RewardTransactionUnit))
}
func testWithMissingStorer(missingUnit dataRetriever.UnitType) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		args := createMockArgAPITransactionProcessor()
		args.StorageService = &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				if unitType == missingUnit {
					return nil, fmt.Errorf("%w for %s", storage.ErrKeyNotFound, missingUnit.String())
				}
				return &storageStubs.StorerStub{
					SearchFirstCalled: func(key []byte) ([]byte, error) {
						return nil, errors.New("dummy")
					},
				}, nil
			},
		}

		apiTransactionProc, _ := NewAPITransactionProcessor(args)
		_, err := apiTransactionProc.getTransactionFromStorage([]byte("txHash"))
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), ErrTransactionNotFound.Error()))
	}
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

	chainStorer := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			switch unitType {
			case dataRetriever.TransactionUnit:
				return &storageStubs.StorerStub{
					GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
						return marshalizer.Marshal(tx)
					},
				}, nil
			case dataRetriever.UnsignedTransactionUnit:
				return &storageStubs.StorerStub{
					GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
						return marshalizer.Marshal(scResult)
					},
				}, nil
			case dataRetriever.RewardTransactionUnit:
				return &storageStubs.StorerStub{
					GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
						return nil, errors.New("dummy")
					},
				}, nil
			default:
				return nil, storage.ErrKeyNotFound
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

	feeComputer := &testscommon.FeeComputerStub{
		ComputeTransactionFeeCalled: func(tx *transaction.ApiTransactionResult) *big.Int {
			return big.NewInt(1000)
		},
	}

	args := &ArgAPITransactionProcessor{
		RoundDuration:            0,
		GenesisTime:              time.Time{},
		Marshalizer:              &mock.MarshalizerFake{},
		AddressPubKeyConverter:   &testscommon.PubkeyConverterMock{},
		ShardCoordinator:         &mock.ShardCoordinatorMock{},
		HistoryRepository:        historyRepo,
		StorageService:           chainStorer,
		DataPool:                 dataRetrieverMock.NewPoolsHolderMock(),
		Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
		FeeComputer:              feeComputer,
		TxTypeHandler:            &testscommon.TxTypeHandlerMock{},
		LogsFacade:               &testscommon.LogsFacadeStub{},
		DataFieldParser: &testscommon.DataFieldParserStub{
			ParseCalled: func(dataField []byte, sender, receiver []byte, _ uint32) *datafield.ResponseParseData {
				return &datafield.ResponseParseData{}
			},
		},
	}
	apiTransactionProc, _ := NewAPITransactionProcessor(args)

	expectedTx := &transaction.ApiTransactionResult{
		Tx:                          &transaction.Transaction{Nonce: tx.Nonce, RcvAddr: tx.RcvAddr, SndAddr: tx.SndAddr, Value: tx.Value},
		Hash:                        "747848617368",
		ProcessingTypeOnSource:      process.MoveBalance.String(),
		ProcessingTypeOnDestination: process.MoveBalance.String(),
		Nonce:                       tx.Nonce,
		Receiver:                    hex.EncodeToString(tx.RcvAddr),
		Sender:                      hex.EncodeToString(tx.SndAddr),
		Status:                      transaction.TxStatusSuccess,
		MiniBlockType:               block.TxBlock.String(),
		Type:                        string(transaction.TxTypeNormal),
		Value:                       "<nil>",
		SmartContractResults: []*transaction.ApiSmartContractResult{
			{
				Hash:           hex.EncodeToString(scResultHash),
				OriginalTxHash: txHash,
				Receivers:      []string{},
			},
		},
		InitiallyPaidFee: "1000",
		Receivers:        []string{},
		Fee:              "0",
	}

	apiTx, err := apiTransactionProc.GetTransaction(txHash, true)
	require.Nil(t, err)
	require.Equal(t, expectedTx, apiTx)
}

func TestNode_lookupHistoricalTransaction(t *testing.T) {
	t.Parallel()

	n, chainStorer, _, historyRepo := createAPITransactionProc(t, 42, true)

	// Normal transactions

	// Cross-shard, we are source
	internalMarshalizer := n.marshalizer
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
	_ = chainStorer.Transactions.PutWithMarshalizer([]byte("invalid"), txInvalid, n.marshalizer)
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
	nonceBytes := n.uint64ByteSliceConverter.ToByteSlice(headerNonce)
	_ = chainStorer.MetaHdrNonce.Put(nonceBytes, headerHash)
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
	nonceBytes = n.uint64ByteSliceConverter.ToByteSlice(headerNonce)
	_ = chainStorer.MetaHdrNonce.Put(nonceBytes, headerHash)
	txH := &rewardTx.RewardTx{Round: 50, RcvAddr: []byte("alice")}
	_ = chainStorer.Rewards.PutWithMarshalizer([]byte("h"), txH, n.marshalizer)
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

	putMiniblockFieldsInTransaction(tx, metadata)

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

func TestApiTransactionProcessor_GetTransactionsPool(t *testing.T) {
	t.Parallel()

	txHash0, txHash1, txHash2, txHash3 := []byte("txHash0"), []byte("txHash1"), []byte("txHash2"), []byte("txHash3")
	expectedTxs := [][]byte{txHash0, txHash1}
	expectedScrs := [][]byte{txHash2}
	expectedRwds := [][]byte{txHash3}
	args := createMockArgAPITransactionProcessor()
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				KeysCalled: func() [][]byte {
					return expectedTxs
				},
				SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
					return createTx(key, "alice", 1).Tx, true
				},
			}
		},
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				KeysCalled: func() [][]byte {
					return expectedScrs
				},
				SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
					return &smartContractResult.SmartContractResult{}, true
				},
			}
		},
		RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				KeysCalled: func() [][]byte {
					return expectedRwds
				},
				SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
					return &rewardTx.RewardTx{}, true
				},
			}
		},
	}
	atp, err := NewAPITransactionProcessor(args)
	require.NoError(t, err)
	require.NotNil(t, atp)

	res, err := atp.GetTransactionsPool("")
	require.NoError(t, err)

	regularTxs := []common.Transaction{
		{
			TxFields: map[string]interface{}{
				"hash": hex.EncodeToString(txHash0),
			},
		},
		{
			TxFields: map[string]interface{}{
				"hash": hex.EncodeToString(txHash1),
			},
		},
	}
	require.Equal(t, regularTxs, res.RegularTransactions)

	scrTxs := []common.Transaction{
		{
			TxFields: map[string]interface{}{
				"hash": hex.EncodeToString(txHash2),
			},
		},
	}
	require.Equal(t, scrTxs, res.SmartContractResults)

	rewardTxs := []common.Transaction{
		{
			TxFields: map[string]interface{}{
				"hash": hex.EncodeToString(txHash3),
			},
		},
	}
	require.Equal(t, rewardTxs, res.Rewards)
}

func createTx(hash []byte, sender string, nonce uint64) *txcache.WrappedTransaction {
	tx := &transaction.Transaction{
		SndAddr: []byte(sender),
		Nonce:   nonce,
		Value:   big.NewInt(100000 + int64(nonce)),
	}

	return &txcache.WrappedTransaction{
		Tx:     tx,
		TxHash: hash,
		Size:   128,
	}
}

func TestApiTransactionProcessor_GetTransactionsPoolForSender(t *testing.T) {
	t.Parallel()

	txHash0, txHash1, txHash2 := []byte("txHash0"), []byte("txHash1"), []byte("txHash2")
	sender := "alice"
	txCacheIntraShard, _ := txcache.NewTxCache(txcache.ConfigSourceMe{
		Name:                       "test",
		NumChunks:                  4,
		NumBytesPerSenderThreshold: 1_048_576, // 1 MB
		CountPerSenderThreshold:    math.MaxUint32,
	}, &txcachemocks.TxGasHandlerMock{
		MinimumGasMove:       1,
		MinimumGasPrice:      1,
		GasProcessingDivisor: 1,
	})
	txCacheIntraShard.AddTx(createTx(txHash2, sender, 3))
	txCacheIntraShard.AddTx(createTx(txHash0, sender, 1))
	txCacheIntraShard.AddTx(createTx(txHash1, sender, 2))

	txHash3, txHash4 := []byte("txHash3"), []byte("txHash4")
	txCacheWithMeta, _ := txcache.NewTxCache(txcache.ConfigSourceMe{
		Name:                       "test-meta",
		NumChunks:                  4,
		NumBytesPerSenderThreshold: 1_048_576, // 1 MB
		CountPerSenderThreshold:    math.MaxUint32,
	}, &txcachemocks.TxGasHandlerMock{
		MinimumGasMove:       1,
		MinimumGasPrice:      1,
		GasProcessingDivisor: 1,
	})
	txCacheWithMeta.AddTx(createTx(txHash3, sender, 4))
	txCacheWithMeta.AddTx(createTx(txHash4, sender, 5))

	args := createMockArgAPITransactionProcessor()
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(cacheID string) storage.Cacher {
					if len(cacheID) == 1 { // self shard
						return txCacheIntraShard
					}

					return txCacheWithMeta
				},
			}
		},
	}
	args.AddressPubKeyConverter = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(humanReadable string) ([]byte, error) {
			return []byte(humanReadable), nil
		},
		EncodeCalled: func(pkBytes []byte) (string, error) {
			return string(pkBytes), nil
		},
		SilentEncodeCalled: func(pkBytes []byte, log core.Logger) string {
			return string(pkBytes)
		},
	}
	args.ShardCoordinator = &processMocks.ShardCoordinatorStub{
		NumberOfShardsCalled: func() uint32 {
			return 1
		},
	}
	atp, err := NewAPITransactionProcessor(args)
	require.NoError(t, err)
	require.NotNil(t, atp)

	res, err := atp.GetTransactionsPoolForSender(sender, "sender,value")
	require.NoError(t, err)
	expectedHashes := []string{hex.EncodeToString(txHash0), hex.EncodeToString(txHash1), hex.EncodeToString(txHash2), hex.EncodeToString(txHash3), hex.EncodeToString(txHash4)}
	expectedValues := []string{"100001", "100002", "100003", "100004", "100005"}
	for i, tx := range res.Transactions {
		require.Equal(t, expectedHashes[i], tx.TxFields[hashField])
		require.Equal(t, expectedValues[i], tx.TxFields[valueField])
		require.Equal(t, sender, tx.TxFields["sender"])
	}

	// if no tx is found in pool for a sender, it isn't an error, but return empty slice
	newSender := "new-sender"
	res, err = atp.GetTransactionsPoolForSender(newSender, "")
	require.NoError(t, err)
	require.Equal(t, &common.TransactionsPoolForSenderApiResponse{
		Transactions: []common.Transaction{},
	}, res)
}

func TestApiTransactionProcessor_GetLastPoolNonceForSender(t *testing.T) {
	t.Parallel()

	txHash0, txHash1, txHash2, txHash3, txHash4 := []byte("txHash0"), []byte("txHash1"), []byte("txHash2"), []byte("txHash3"), []byte("txHash4")
	sender := "alice"
	lastNonce := uint64(10)
	txCacheIntraShard, _ := txcache.NewTxCache(txcache.ConfigSourceMe{
		Name:                       "test",
		NumChunks:                  4,
		NumBytesPerSenderThreshold: 1_048_576, // 1 MB
		CountPerSenderThreshold:    math.MaxUint32,
	}, &txcachemocks.TxGasHandlerMock{
		MinimumGasMove:       1,
		MinimumGasPrice:      1,
		GasProcessingDivisor: 1,
	})
	txCacheIntraShard.AddTx(createTx(txHash2, sender, 3))
	txCacheIntraShard.AddTx(createTx(txHash0, sender, 1))
	txCacheIntraShard.AddTx(createTx(txHash1, sender, 2))
	txCacheIntraShard.AddTx(createTx(txHash3, sender, lastNonce))
	txCacheIntraShard.AddTx(createTx(txHash4, sender, 5))

	args := createMockArgAPITransactionProcessor()
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(cacheID string) storage.Cacher {
					return txCacheIntraShard
				},
			}
		},
	}
	args.AddressPubKeyConverter = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(humanReadable string) ([]byte, error) {
			return []byte(humanReadable), nil
		},
		EncodeCalled: func(pkBytes []byte) (string, error) {
			return string(pkBytes), nil
		},
	}
	args.ShardCoordinator = &processMocks.ShardCoordinatorStub{
		NumberOfShardsCalled: func() uint32 {
			return 1
		},
	}
	atp, err := NewAPITransactionProcessor(args)
	require.NoError(t, err)
	require.NotNil(t, atp)

	res, err := atp.GetLastPoolNonceForSender(sender)
	require.NoError(t, err)
	require.Equal(t, lastNonce, res)
}

func TestApiTransactionProcessor_GetTransactionsPoolNonceGapsForSender(t *testing.T) {
	t.Parallel()

	txHash1, txHash2, txHash3, txHash4 := []byte("txHash1"), []byte("txHash2"), []byte("txHash3"), []byte("txHash4")
	sender := "alice"
	txCacheIntraShard, _ := txcache.NewTxCache(txcache.ConfigSourceMe{
		Name:                       "test",
		NumChunks:                  4,
		NumBytesPerSenderThreshold: 1_048_576, // 1 MB
		CountPerSenderThreshold:    math.MaxUint32,
	}, &txcachemocks.TxGasHandlerMock{
		MinimumGasMove:       1,
		MinimumGasPrice:      1,
		GasProcessingDivisor: 1,
	})

	txCacheWithMeta, _ := txcache.NewTxCache(txcache.ConfigSourceMe{
		Name:                       "test-meta",
		NumChunks:                  4,
		NumBytesPerSenderThreshold: 1_048_576, // 1 MB
		CountPerSenderThreshold:    math.MaxUint32,
	}, &txcachemocks.TxGasHandlerMock{
		MinimumGasMove:       1,
		MinimumGasPrice:      1,
		GasProcessingDivisor: 1,
	})

	accountNonce := uint64(20)
	// expected nonce gaps: 21-31, 33-33, 36-38
	firstNonceInPool := uint64(32)
	firstNonceAfterGap1 := uint64(34)
	lastNonceBeforeGap2 := uint64(35)
	firstNonceAfterGap2 := uint64(39)
	txCacheIntraShard.AddTx(createTx(txHash1, sender, firstNonceInPool))
	txCacheIntraShard.AddTx(createTx(txHash2, sender, firstNonceAfterGap1))
	txCacheIntraShard.AddTx(createTx(txHash3, sender, lastNonceBeforeGap2))
	txCacheIntraShard.AddTx(createTx(txHash4, sender, firstNonceAfterGap2))

	args := createMockArgAPITransactionProcessor()
	args.DataPool = &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataStub{
				ShardDataStoreCalled: func(cacheID string) storage.Cacher {
					if len(cacheID) == 1 { // self shard
						return txCacheIntraShard
					}

					return txCacheWithMeta
				},
			}
		},
	}
	args.AddressPubKeyConverter = &testscommon.PubkeyConverterStub{
		DecodeCalled: func(humanReadable string) ([]byte, error) {
			return []byte(humanReadable), nil
		},
		EncodeCalled: func(pkBytes []byte) (string, error) {
			return string(pkBytes), nil
		},
	}
	args.ShardCoordinator = &processMocks.ShardCoordinatorStub{
		NumberOfShardsCalled: func() uint32 {
			return 1
		},
	}
	atp, err := NewAPITransactionProcessor(args)
	require.NoError(t, err)
	require.NotNil(t, atp)

	expectedResponse := &common.TransactionsPoolNonceGapsForSenderApiResponse{
		Sender: sender,
		Gaps: []common.NonceGapApiResponse{
			{
				From: accountNonce,
				To:   firstNonceInPool - 1,
			},
			{
				From: firstNonceInPool + 1,
				To:   firstNonceAfterGap1 - 1,
			},
			{
				From: lastNonceBeforeGap2 + 1,
				To:   firstNonceAfterGap2 - 1,
			},
		},
	}
	res, err := atp.GetTransactionsPoolNonceGapsForSender(sender, accountNonce)
	require.NoError(t, err)
	require.Equal(t, expectedResponse, res)

	// if no tx is found in pool for a sender, it isn't an error, but return empty slice
	newSender := "new-sender"
	res, err = atp.GetTransactionsPoolNonceGapsForSender(newSender, 0)
	require.NoError(t, err)
	require.Equal(t, &common.TransactionsPoolNonceGapsForSenderApiResponse{
		Sender: newSender,
		Gaps:   []common.NonceGapApiResponse{},
	}, res)
}

func createAPITransactionProc(t *testing.T, epoch uint32, withDbLookupExt bool) (*apiTransactionProcessor, *genericMocks.ChainStorerMock, *dataRetrieverMock.PoolsHolderMock, *dblookupextMock.HistoryRepositoryStub) {
	chainStorer := genericMocks.NewChainStorerMock(epoch)
	dataPool := dataRetrieverMock.NewPoolsHolderMock()

	historyRepo := &dblookupextMock.HistoryRepositoryStub{
		IsEnabledCalled: func() bool {
			return withDbLookupExt
		},
	}
	dataFieldParser := &testscommon.DataFieldParserStub{
		ParseCalled: func(dataField []byte, sender, receiver []byte, _ uint32) *datafield.ResponseParseData {
			return &datafield.ResponseParseData{}
		},
	}

	args := &ArgAPITransactionProcessor{
		RoundDuration:            0,
		GenesisTime:              time.Time{},
		Marshalizer:              &mock.MarshalizerFake{},
		AddressPubKeyConverter:   &testscommon.PubkeyConverterMock{},
		ShardCoordinator:         createShardCoordinator(),
		HistoryRepository:        historyRepo,
		StorageService:           chainStorer,
		DataPool:                 dataPool,
		Uint64ByteSliceConverter: mock.NewNonceHashConverterMock(),
		FeeComputer:              &testscommon.FeeComputerStub{},
		TxTypeHandler:            &testscommon.TxTypeHandlerMock{},
		LogsFacade:               &testscommon.LogsFacadeStub{},
		DataFieldParser:          dataFieldParser,
	}
	apiTransactionProc, err := NewAPITransactionProcessor(args)
	require.Nil(t, err)

	return apiTransactionProc, chainStorer, dataPool, historyRepo
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

	n, _, _, _ := createAPITransactionProc(t, 0, true)
	n.txUnmarshaller.addressPubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(addrSize, "erd")
	n.addressPubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(addrSize, "erd")

	scrResult1 := n.txUnmarshaller.prepareUnsignedTx(scr1)
	expectedScr1 := &transaction.ApiTransactionResult{
		Tx:             scr1,
		Nonce:          1,
		Type:           string(transaction.TxTypeUnsigned),
		Value:          "2",
		Receiver:       "erd1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqsl6e0p7",
		Sender:         "erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu",
		OriginalSender: "",
		CallType:       vm.DirectCallStr,
	}
	assert.Equal(t, scrResult1, expectedScr1)

	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          3,
		Value:          big.NewInt(4),
		SndAddr:        bytes.Repeat([]byte{5}, addrSize),
		RcvAddr:        bytes.Repeat([]byte{6}, addrSize),
		OriginalSender: bytes.Repeat([]byte{7}, addrSize),
	}

	scrResult2 := n.txUnmarshaller.prepareUnsignedTx(scr2)
	expectedScr2 := &transaction.ApiTransactionResult{
		Tx:             scr2,
		Nonce:          3,
		Type:           string(transaction.TxTypeUnsigned),
		Value:          "4",
		Receiver:       "erd1qcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqwkh39e",
		Sender:         "erd1q5zs2pg9q5zs2pg9q5zs2pg9q5zs2pg9q5zs2pg9q5zs2pg9q5zsrqsks3",
		OriginalSender: "erd1qurswpc8qurswpc8qurswpc8qurswpc8qurswpc8qurswpc8qurstywtnm",
		CallType:       vm.DirectCallStr,
	}
	assert.Equal(t, scrResult2, expectedScr2)
}

func TestNode_ComputeTimestampForRound(t *testing.T) {
	genesis := getTime(t, "1596117600")
	n, _, _, _ := createAPITransactionProc(t, 0, false)
	n.genesisTime = genesis
	n.roundDuration = 6000

	res := n.computeTimestampForRound(0)
	require.Equal(t, int64(0), res)

	res = n.computeTimestampForRound(4837403)
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

func TestApiTransactionProcessor_GetTransactionPopulatesComputedFields(t *testing.T) {
	dataPool := dataRetrieverMock.NewPoolsHolderMock()
	feeComputer := &testscommon.FeeComputerStub{}
	txTypeHandler := &testscommon.TxTypeHandlerMock{}

	arguments := createMockArgAPITransactionProcessor()
	arguments.DataPool = dataPool
	arguments.FeeComputer = feeComputer
	arguments.TxTypeHandler = txTypeHandler

	processor, err := NewAPITransactionProcessor(arguments)
	require.Nil(t, err)
	require.NotNil(t, processor)

	t.Run("InitiallyPaidFee", func(t *testing.T) {
		feeComputer.ComputeTransactionFeeCalled = func(tx *transaction.ApiTransactionResult) *big.Int {
			return big.NewInt(1000)
		}

		dataPool.Transactions().AddData([]byte{0, 0}, &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}, 42, "1")
		tx, err := processor.GetTransaction("0000", true)

		require.Nil(t, err)
		require.Equal(t, "1000", tx.InitiallyPaidFee)
	})

	t.Run("InitiallyPaidFee (missing on unsigned transaction)", func(t *testing.T) {
		feeComputer.ComputeTransactionFeeCalled = func(tx *transaction.ApiTransactionResult) *big.Int {
			return big.NewInt(1000)
		}

		scr := &smartContractResult.SmartContractResult{GasLimit: 0, Data: []byte("@ok"), Value: big.NewInt(0)}
		dataPool.UnsignedTransactions().AddData([]byte{0, 1}, scr, 42, "foo")
		tx, err := processor.GetTransaction("0001", true)

		require.Nil(t, err)
		require.Equal(t, "", tx.InitiallyPaidFee)
	})

	t.Run("ProcessingType", func(t *testing.T) {
		txTypeHandler.ComputeTransactionTypeCalled = func(data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.MoveBalance, process.SCDeployment
		}

		dataPool.Transactions().AddData([]byte{0, 2}, &transaction.Transaction{Nonce: 7, SndAddr: []byte("alice"), RcvAddr: []byte("bob")}, 42, "1")
		tx, err := processor.GetTransaction("0002", true)

		require.Nil(t, err)
		require.Equal(t, process.MoveBalance.String(), tx.ProcessingTypeOnSource)
		require.Equal(t, process.SCDeployment.String(), tx.ProcessingTypeOnDestination)
	})

	t.Run("IsRefund (false)", func(t *testing.T) {
		scr := &smartContractResult.SmartContractResult{GasLimit: 0, Data: []byte("@ok"), Value: big.NewInt(0)}
		dataPool.UnsignedTransactions().AddData([]byte{0, 3}, scr, 42, "foo")
		tx, err := processor.GetTransaction("0003", true)

		require.Nil(t, err)
		require.Equal(t, false, tx.IsRefund)
	})

	t.Run("IsRefund (true)", func(t *testing.T) {
		scr := &smartContractResult.SmartContractResult{GasLimit: 0, Data: []byte("@6f6b"), Value: big.NewInt(500)}
		dataPool.UnsignedTransactions().AddData([]byte{0, 4}, scr, 42, "foo")
		tx, err := processor.GetTransaction("0004", true)

		require.Nil(t, err)
		require.Equal(t, true, tx.IsRefund)
	})
}

func TestApiTransactionProcessor_PopulateComputedFields(t *testing.T) {
	feeComputer := &testscommon.FeeComputerStub{}
	txTypeHandler := &testscommon.TxTypeHandlerMock{}

	arguments := createMockArgAPITransactionProcessor()
	arguments.Marshalizer = &marshal.GogoProtoMarshalizer{}
	arguments.FeeComputer = feeComputer
	arguments.TxTypeHandler = txTypeHandler

	processor, err := NewAPITransactionProcessor(arguments)
	require.Nil(t, err)
	require.NotNil(t, processor)

	txTypeHandler.ComputeTransactionTypeCalled = func(data.TransactionHandler) (process.TransactionType, process.TransactionType) {
		return process.MoveBalance, process.SCDeployment
	}

	feeComputer.ComputeTransactionFeeCalled = func(tx *transaction.ApiTransactionResult) *big.Int {
		return big.NewInt(1000)
	}

	apiTx := &transaction.ApiTransactionResult{Type: string(transaction.TxTypeNormal)}
	processor.PopulateComputedFields(apiTx)

	require.Equal(t, "MoveBalance", apiTx.ProcessingTypeOnSource)
	require.Equal(t, "SCDeployment", apiTx.ProcessingTypeOnDestination)
	require.Equal(t, "1000", apiTx.InitiallyPaidFee)
}
