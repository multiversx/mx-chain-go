package block_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	commonMocks "github.com/multiversx/mx-chain-go/testscommon/common"
	"github.com/multiversx/mx-chain-go/testscommon/components"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	busyIdentifier = "busy"
	idleIdentifier = "idle"
)

func haveTime() time.Duration {
	return 2000 * time.Millisecond
}

func createArgBaseProcessor(
	coreComponents *mock.CoreComponentsMock,
	dataComponents *mock.DataComponentsMock,
	bootstrapComponents *mock.BootstrapComponentsMock,
	statusComponents *mock.StatusComponentsMock,
) blproc.ArgBaseProcessor {
	nodesCoordinatorInstance := shardingMocks.NewNodesCoordinatorMock()
	argsHeaderValidator := blproc.ArgsHeaderValidator{
		Hasher:      &hashingMocks.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
	}
	headerValidator, _ := blproc.NewHeaderValidator(argsHeaderValidator)

	startHeaders := createGenesisBlocks(mock.NewOneShardCoordinatorMock())

	accountsDb := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accountsDb[state.UserAccountsState] = &stateMock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return nil, nil
		},
	}

	statusCoreComponents := &mock.StatusCoreComponentsStub{
		AppStatusHandlerField: &statusHandlerMock.AppStatusHandlerStub{},
	}

	return blproc.ArgBaseProcessor{
		CoreComponents:       coreComponents,
		DataComponents:       dataComponents,
		BootstrapComponents:  bootstrapComponents,
		StatusComponents:     statusComponents,
		StatusCoreComponents: statusCoreComponents,
		Config:               config.Config{},
		AccountsDB:           accountsDb,
		ForkDetector:         &mock.ForkDetectorMock{},
		NodesCoordinator:     nodesCoordinatorInstance,
		FeeHandler:           &mock.FeeAccumulatorStub{},
		RequestHandler:       &testscommon.RequestHandlerStub{},
		BlockChainHook:       &testscommon.BlockChainHookStub{},
		TxCoordinator:        &testscommon.TransactionCoordinatorMock{},
		EpochStartTrigger:    &mock.EpochStartTriggerStub{},
		HeaderValidator:      headerValidator,
		BootStorer: &mock.BoostrapStorerMock{
			PutCalled: func(round int64, bootData bootstrapStorage.BootstrapData) error {
				return nil
			},
		},
		BlockTracker:                   mock.NewBlockTrackerMock(bootstrapComponents.ShardCoordinator(), startHeaders),
		BlockSizeThrottler:             &mock.BlockSizeThrottlerStub{},
		Version:                        "softwareVersion",
		HistoryRepository:              &dblookupext.HistoryRepositoryStub{},
		GasHandler:                     &mock.GasHandlerMock{},
		ScheduledTxsExecutionHandler:   &testscommon.ScheduledTxsExecutionStub{},
		OutportDataProvider:            &outport.OutportDataProviderStub{},
		ScheduledMiniBlocksEnableEpoch: 2,
		ProcessedMiniBlocksTracker:     &testscommon.ProcessedMiniBlocksTrackerStub{},
		ReceiptsRepository:             &testscommon.ReceiptsRepositoryStub{},
		BlockProcessingCutoffHandler:   &testscommon.BlockProcessingCutoffStub{},
		ManagedPeersHolder:             &testscommon.ManagedPeersHolderStub{},
		SentSignaturesTracker:          &testscommon.SentSignatureTrackerStub{},
		RunTypeComponents:              components.GetRunTypeComponents(),
	}
}

func createTestBlockchain() *testscommon.ChainHandlerStub {
	return &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}
}

func generateTestCache() storage.Cacher {
	cache, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 1000, Shards: 1, SizeInBytes: 0})
	return cache
}

func generateTestUnit() storage.Storer {
	storer, _ := storageunit.NewStorageUnit(
		generateTestCache(),
		database.NewMemDB(),
	)

	return storer
}

func createShardedDataChacherNotifier(
	handler data.TransactionHandler,
	testHash []byte,
) func() dataRetriever.ShardedDataCacherNotifier {
	return func() dataRetriever.ShardedDataCacherNotifier {
		return &testscommon.ShardedDataStub{
			ShardDataStoreCalled: func(id string) (c storage.Cacher) {
				return &testscommon.CacherStub{
					PeekCalled: func(key []byte) (value interface{}, ok bool) {
						if reflect.DeepEqual(key, testHash) {
							return handler, true
						}
						return nil, false
					},
					KeysCalled: func() [][]byte {
						return [][]byte{[]byte("key1"), []byte("key2")}
					},
					LenCalled: func() int {
						return 0
					},
					MaxSizeCalled: func() int {
						return 1000
					},
				}
			},
			RemoveSetOfDataFromPoolCalled: func(keys [][]byte, id string) {},
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				if reflect.DeepEqual(key, []byte("tx1_hash")) {
					return handler, true
				}
				return nil, false
			},
			AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheId string) {
			},
		}
	}
}

func initDataPool(testHash []byte) *dataRetrieverMock.PoolsHolderStub {
	rwdTx := &rewardTx.RewardTx{
		Round:   1,
		Epoch:   0,
		Value:   big.NewInt(10),
		RcvAddr: []byte("receiver"),
	}
	txCalled := createShardedDataChacherNotifier(&transaction.Transaction{Nonce: 10}, testHash)
	unsignedTxCalled := createShardedDataChacherNotifier(&transaction.Transaction{Nonce: 10}, testHash)
	rewardTransactionsCalled := createShardedDataChacherNotifier(rwdTx, testHash)

	sdp := &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled:         txCalled,
		UnsignedTransactionsCalled: unsignedTxCalled,
		RewardTransactionsCalled:   rewardTransactionsCalled,
		MetaBlocksCalled: func() storage.Cacher {
			return &testscommon.CacherStub{
				GetCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
				KeysCalled: func() [][]byte {
					return nil
				},
				LenCalled: func() int {
					return 0
				},
				MaxSizeCalled: func() int {
					return 1000
				},
				PeekCalled: func(key []byte) (value interface{}, ok bool) {
					if reflect.DeepEqual(key, []byte("tx1_hash")) {
						return &transaction.Transaction{Nonce: 10}, true
					}
					return nil, false
				},
				RegisterHandlerCalled: func(i func(key []byte, value interface{})) {},
				RemoveCalled:          func(key []byte) {},
			}
		},
		MiniBlocksCalled: func() storage.Cacher {
			cs := testscommon.NewCacherStub()
			cs.RegisterHandlerCalled = func(i func(key []byte, value interface{})) {
			}
			cs.GetCalled = func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal([]byte("bbb"), key) {
					return make(block.MiniBlockSlice, 0), true
				}

				return nil, false
			}
			cs.PeekCalled = func(key []byte) (value interface{}, ok bool) {
				if bytes.Equal([]byte("bbb"), key) {
					return make(block.MiniBlockSlice, 0), true
				}

				return nil, false
			}
			cs.RegisterHandlerCalled = func(i func(key []byte, value interface{})) {}
			cs.RemoveCalled = func(key []byte) {}
			cs.LenCalled = func() int {
				return 0
			}
			cs.MaxSizeCalled = func() int {
				return 300
			}
			cs.KeysCalled = func() [][]byte {
				return nil
			}
			return cs
		},
		HeadersCalled: func() dataRetriever.HeadersPool {
			cs := &mock.HeadersCacherStub{}
			cs.RegisterHandlerCalled = func(i func(header data.HeaderHandler, key []byte)) {
			}
			cs.GetHeaderByHashCalled = func(hash []byte) (data.HeaderHandler, error) {
				return nil, process.ErrMissingHeader
			}
			cs.RemoveHeaderByHashCalled = func(key []byte) {
			}
			cs.LenCalled = func() int {
				return 0
			}
			cs.MaxSizeCalled = func() int {
				return 1000
			}
			cs.NoncesCalled = func(shardId uint32) []uint64 {
				return nil
			}
			return cs
		},
	}

	return sdp
}

func initStore() *dataRetriever.ChainStorer {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.RewardTransactionUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, generateTestUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ReceiptsUnit, generateTestUnit())
	store.AddStorer(dataRetriever.TrieEpochRootHashUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ScheduledSCRsUnit, generateTestUnit())
	return store
}

func createDummyMetaBlock(destShardId uint32, senderShardId uint32, miniBlockHashes ...[]byte) *block.MetaBlock {
	metaBlock := &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{
				ShardID:               senderShardId,
				ShardMiniBlockHeaders: make([]block.MiniBlockHeader, len(miniBlockHashes)),
			},
		},
	}

	for idx, mbHash := range miniBlockHashes {
		metaBlock.ShardInfo[0].ShardMiniBlockHeaders[idx].ReceiverShardID = destShardId
		metaBlock.ShardInfo[0].ShardMiniBlockHeaders[idx].SenderShardID = senderShardId
		metaBlock.ShardInfo[0].ShardMiniBlockHeaders[idx].Hash = mbHash
	}

	return metaBlock
}

func createDummyMiniBlock(
	txHash string,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	destShardId uint32,
	senderShardId uint32) (*block.MiniBlock, []byte) {

	miniblock := &block.MiniBlock{
		TxHashes:        [][]byte{[]byte(txHash)},
		ReceiverShardID: destShardId,
		SenderShardID:   senderShardId,
	}

	buff, _ := marshalizer.Marshal(miniblock)
	hash := hasher.Compute(string(buff))

	return miniblock, hash
}

func isInTxHashes(searched []byte, list [][]byte) bool {
	for _, txHash := range list {
		if bytes.Equal(txHash, searched) {
			return true
		}
	}
	return false
}

type wrongBody struct {
}

// Clone -
func (wr *wrongBody) Clone() data.BodyHandler {
	wrCopy := *wr

	return &wrCopy
}

// SetMiniBlocks -
func (wr *wrongBody) SetMiniBlocks(_ []data.MiniBlockHandler) error {
	return nil
}

// IntegrityAndValidity -
func (wr *wrongBody) IntegrityAndValidity() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (wr *wrongBody) IsInterfaceNil() bool {
	return wr == nil
}

func createComponentHolderMocks() (
	*mock.CoreComponentsMock,
	*mock.DataComponentsMock,
	*mock.BootstrapComponentsMock,
	*mock.StatusComponentsMock,
) {
	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})

	coreComponents := &mock.CoreComponentsMock{
		IntMarsh:                  &mock.MarshalizerMock{},
		Hash:                      &mock.HasherStub{},
		UInt64ByteSliceConv:       &mock.Uint64ByteSliceConverterMock{},
		StatusField:               &statusHandlerMock.AppStatusHandlerStub{},
		RoundField:                &mock.RoundHandlerMock{},
		ProcessStatusHandlerField: &testscommon.ProcessStatusHandlerStub{},
		EpochNotifierField:        &epochNotifier.EpochNotifierStub{},
		EnableEpochsHandlerField:  enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
		RoundNotifierField:        &epochNotifier.RoundNotifierStub{},
		EnableRoundsHandlerField:  &testscommon.EnableRoundsHandlerStub{},
	}

	dataComponents := &mock.DataComponentsMock{
		Storage:    initStore(),
		DataPool:   initDataPool([]byte("")),
		BlockChain: blkc,
	}

	boostrapComponents := &mock.BootstrapComponentsMock{
		Coordinator:          mock.NewOneShardCoordinatorMock(),
		HdrIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		VersionedHdrFactory: &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32) data.HeaderHandler {
				return &block.Header{}
			},
		},
	}

	statusComponents := &mock.StatusComponentsMock{
		Outport: &outport.OutportStub{},
	}

	return coreComponents, dataComponents, boostrapComponents, statusComponents
}

func CreateMockArguments(
	coreComponents *mock.CoreComponentsMock,
	dataComponents *mock.DataComponentsMock,
	bootstrapComponents *mock.BootstrapComponentsMock,
	statusComponents *mock.StatusComponentsMock,
) blproc.ArgShardProcessor {
	return blproc.ArgShardProcessor{
		ArgBaseProcessor: createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents),
	}
}

func createMockTransactionCoordinatorArguments(
	accountAdapter state.AccountsAdapter,
	poolsHolder dataRetriever.PoolsHolder,
	preProcessorsContainer process.PreProcessorsContainer,
) coordinator.ArgTransactionCoordinator {
	argsTransactionCoordinator := coordinator.ArgTransactionCoordinator{
		Hasher:           &hashingMocks.HasherMock{},
		Marshalizer:      &mock.MarshalizerMock{},
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(3),
		Accounts:         accountAdapter,
		MiniBlockPool:    poolsHolder.MiniBlocks(),
		RequestHandler:   &testscommon.RequestHandlerStub{},
		PreProcessors:    preProcessorsContainer,
		InterProcessors: &mock.InterimProcessorContainerMock{
			KeysCalled: func() []block.Type {
				return []block.Type{block.SmartContractResultBlock}
			},
		},
		GasHandler:                   &testscommon.GasHandlerStub{},
		FeeHandler:                   &mock.FeeAccumulatorStub{},
		BlockSizeComputation:         &testscommon.BlockSizeComputationStub{},
		BalanceComputation:           &testscommon.BalanceComputationStub{},
		EconomicsFee:                 &economicsmocks.EconomicsHandlerStub{},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.PanicDoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
		TxExecutionOrderHandler:      &commonMocks.TxExecutionOrderHandlerStub{},
	}

	return argsTransactionCoordinator
}

func TestCheckProcessorNilParameters(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

	tests := []struct {
		args        func() blproc.ArgBaseProcessor
		expectedErr error
	}{
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.AccountsDB[state.UserAccountsState] = nil
				return args
			},
			expectedErr: process.ErrNilAccountsAdapter,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				return createArgBaseProcessor(coreComponents, nil, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilDataComponentsHolder,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				return createArgBaseProcessor(nil, dataComponents, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilCoreComponentsHolder,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.BootstrapComponents = nil
				return args
			},
			expectedErr: process.ErrNilBootstrapComponentsHolder,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				return createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, nil)
			},
			expectedErr: process.ErrNilStatusComponentsHolder,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.ForkDetector = nil
				return args
			},
			expectedErr: process.ErrNilForkDetector,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.Hash = nil
				args := createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
				return args
			},
			expectedErr: process.ErrNilHasher,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.IntMarsh = nil
				args := createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
				return args
			},
			expectedErr: process.ErrNilMarshalizer,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				dataCompCopy := *dataComponents
				dataCompCopy.Storage = nil
				args := createArgBaseProcessor(coreComponents, &dataCompCopy, bootstrapComponents, statusComponents)
				return args
			},
			expectedErr: process.ErrNilStorage,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.BootstrapComponents = &mainFactoryMocks.BootstrapComponentsStub{ShCoordinator: nil}
				return args
			},
			expectedErr: process.ErrNilShardCoordinator,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.NodesCoordinator = nil
				return args
			},
			expectedErr: process.ErrNilNodesCoordinator,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.UInt64ByteSliceConv = nil
				return createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilUint64Converter,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.RequestHandler = nil
				return args
			},
			expectedErr: process.ErrNilRequestHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.EpochStartTrigger = nil
				return args
			},
			expectedErr: process.ErrNilEpochStartTrigger,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.RoundNotifierField = nil
				return createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilRoundNotifier,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.RoundField = nil
				return createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilRoundHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.BootStorer = nil
				return args
			},
			expectedErr: process.ErrNilStorage,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.BlockChainHook = nil
				return args
			},
			expectedErr: process.ErrNilBlockChainHook,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.TxCoordinator = nil
				return args
			},
			expectedErr: process.ErrNilTransactionCoordinator,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.HeaderValidator = nil
				return args
			},
			expectedErr: process.ErrNilHeaderValidator,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.BlockTracker = nil
				return args
			},
			expectedErr: process.ErrNilBlockTracker,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.FeeHandler = nil
				return args
			},
			expectedErr: process.ErrNilEconomicsFeeHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				dataComp := &mock.DataComponentsMock{
					Storage:    dataComponents.Storage,
					DataPool:   dataComponents.DataPool,
					BlockChain: nil,
				}
				return createArgBaseProcessor(coreComponents, dataComp, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilBlockChain,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.BlockSizeThrottler = nil
				return args
			},
			expectedErr: process.ErrNilBlockSizeThrottler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				statusCompCopy := *statusComponents
				statusCompCopy.Outport = nil
				return createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, &statusCompCopy)
			},
			expectedErr: process.ErrNilOutportHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.HistoryRepository = nil
				return args
			},
			expectedErr: process.ErrNilHistoryRepository,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				bootStrapCopy := *bootstrapComponents
				bootStrapCopy.HdrIntegrityVerifier = nil
				return createArgBaseProcessor(coreComponents, dataComponents, &bootStrapCopy, statusComponents)
			},
			expectedErr: process.ErrNilHeaderIntegrityVerifier,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.EnableRoundsHandlerField = nil
				args := createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
				return args
			},
			expectedErr: process.ErrNilEnableRoundsHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.StatusCoreComponents = nil
				return args
			},
			expectedErr: process.ErrNilStatusCoreComponentsHolder,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.StatusCoreComponents = &mock.StatusCoreComponentsStub{
					AppStatusHandlerField: nil,
				}
				return args
			},
			expectedErr: process.ErrNilAppStatusHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.EnableEpochsHandlerField = nil
				return createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilEnableEpochsHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.EpochNotifierField = nil
				return createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilEpochNotifier,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.GasHandler = nil
				return args
			},
			expectedErr: process.ErrNilGasHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.ScheduledTxsExecutionHandler = nil
				return args
			},
			expectedErr: process.ErrNilScheduledTxsExecutionHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.ProcessedMiniBlocksTracker = nil
				return args
			},
			expectedErr: process.ErrNilProcessedMiniBlocksTracker,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.ReceiptsRepository = nil
				return args
			},
			expectedErr: process.ErrNilReceiptsRepository,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				bootstrapCopy := *bootstrapComponents
				bootstrapCopy.VersionedHdrFactory = nil
				return createArgBaseProcessor(coreComponents, dataComponents, &bootstrapCopy, statusComponents)
			},
			expectedErr: process.ErrNilVersionedHeaderFactory,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.ManagedPeersHolder = nil
				return args
			},
			expectedErr: process.ErrNilManagedPeersHolder,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				return createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
			},
			expectedErr: nil,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.RunTypeComponents = nil
				return args
			},
			expectedErr: errorsMx.ErrNilRunTypeComponents,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				rtMock := getRunTypeComponentsMock()
				rtMock.AccountCreator = nil
				args.RunTypeComponents = rtMock
				return args
			},
			expectedErr: state.ErrNilAccountFactory,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				rtMock := getRunTypeComponentsMock()
				rtMock.OutGoingOperationsPool = nil
				args.RunTypeComponents = rtMock
				return args
			},
			expectedErr: errorsMx.ErrNilOutGoingOperationsPool,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				rtMock := getRunTypeComponentsMock()
				rtMock.DataCodec = nil
				args.RunTypeComponents = rtMock
				return args
			},
			expectedErr: errorsMx.ErrNilDataCodec,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				rtMock := getRunTypeComponentsMock()
				rtMock.TopicsChecker = nil
				args.RunTypeComponents = rtMock
				return args
			},
			expectedErr: errorsMx.ErrNilTopicsChecker,
		},
	}

	for _, test := range tests {
		err := blproc.CheckProcessorNilParameters(test.args())
		require.Equal(t, test.expectedErr, err)
	}

	coreCompCopy := *coreComponents
	coreCompCopy.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStubWithNoFlagsDefined()
	args := createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
	err := blproc.CheckProcessorNilParameters(args)
	require.True(t, errors.Is(err, core.ErrInvalidEnableEpochsHandler))
}

func getRunTypeComponentsMock() *mock.RunTypeComponentsStub {
	rt := mock.NewRunTypeComponentsStub()
	return &mock.RunTypeComponentsStub{
		AccountCreator:         rt.AccountsCreator(),
		OutGoingOperationsPool: rt.OutGoingOperationsPoolHandler(),
		DataCodec:              rt.DataDecoderHandler(),
		TopicsChecker:          rt.TopicsCheckerHandler(),
	}
}

func TestBlockProcessor_CheckBlockValidity(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.Hash = &hashingMocks.HasherMock{}
	blkc := createTestBlockchain()
	dataComponents.BlockChain = blkc
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	bp, _ := blproc.NewShardProcessor(arguments)

	body := &block.Body{}
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.Round = 1
	hdr.TimeStamp = 0
	hdr.PrevHash = []byte("X")
	err := bp.CheckBlockValidity(hdr, body)
	assert.Equal(t, process.ErrBlockHashDoesNotMatch, err)

	hdr.PrevHash = []byte("")
	err = bp.CheckBlockValidity(hdr, body)
	assert.Nil(t, err)

	hdr.Nonce = 2
	err = bp.CheckBlockValidity(hdr, body)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)

	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &block.Header{Round: 1, Nonce: 1}
	}
	prevHash := []byte("X")
	blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
		return prevHash
	}
	hdr = &block.Header{}

	err = bp.CheckBlockValidity(hdr, body)
	assert.Equal(t, process.ErrLowerRoundInBlock, err)

	hdr.Round = 2
	hdr.Nonce = 1
	err = bp.CheckBlockValidity(hdr, body)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)

	hdr.Nonce = 2
	hdr.PrevHash = []byte("XX")
	err = bp.CheckBlockValidity(hdr, body)
	assert.Equal(t, process.ErrBlockHashDoesNotMatch, err)

	hdr.PrevHash = blkc.GetCurrentBlockHeaderHash()
	hdr.PrevRandSeed = []byte("X")
	err = bp.CheckBlockValidity(hdr, body)
	assert.Equal(t, process.ErrRandSeedDoesNotMatch, err)

	hdr.PrevRandSeed = []byte("")
	err = bp.CheckBlockValidity(hdr, body)
	assert.Nil(t, err)
}

func TestVerifyStateRoot_ShouldWork(t *testing.T) {
	t.Parallel()
	rootHash := []byte("root hash to be tested")
	accounts := &stateMock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
	}

	arguments := CreateMockArguments(createComponentHolderMocks())
	arguments.AccountsDB[state.UserAccountsState] = accounts
	bp, _ := blproc.NewShardProcessor(arguments)

	assert.True(t, bp.VerifyStateRoot(rootHash))
}

func TestBaseProcessor_SetIndexOfFirstTxProcessed(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments(createComponentHolderMocks())
	processedMiniBlocksTracker := processedMb.NewProcessedMiniBlocksTracker()
	arguments.ProcessedMiniBlocksTracker = processedMiniBlocksTracker
	bp, _ := blproc.NewShardProcessor(arguments)

	metaHash := []byte("meta_hash")
	mbHash := []byte("mb_hash")
	miniBlockHeader := &block.MiniBlockHeader{
		Hash: mbHash,
	}

	processedMbInfo := &processedMb.ProcessedMiniBlockInfo{
		FullyProcessed:         false,
		IndexOfLastTxProcessed: 8,
	}
	processedMiniBlocksTracker.SetProcessedMiniBlockInfo(metaHash, mbHash, processedMbInfo)
	err := bp.SetIndexOfFirstTxProcessed(miniBlockHeader)
	assert.Nil(t, err)
	assert.Equal(t, int32(9), miniBlockHeader.GetIndexOfFirstTxProcessed())
}

func TestBaseProcessor_SetIndexOfLastTxProcessed(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments(createComponentHolderMocks())
	bp, _ := blproc.NewShardProcessor(arguments)

	mbHash := []byte("mb_hash")
	processedMiniBlocksDestMeInfo := make(map[string]*processedMb.ProcessedMiniBlockInfo)
	miniBlockHeader := &block.MiniBlockHeader{
		Hash:    mbHash,
		TxCount: 100,
	}

	err := bp.SetIndexOfLastTxProcessed(miniBlockHeader, processedMiniBlocksDestMeInfo)
	assert.Nil(t, err)
	assert.Equal(t, int32(99), miniBlockHeader.GetIndexOfLastTxProcessed())

	processedMbInfo := &processedMb.ProcessedMiniBlockInfo{
		FullyProcessed:         false,
		IndexOfLastTxProcessed: 8,
	}
	processedMiniBlocksDestMeInfo[string(mbHash)] = processedMbInfo

	err = bp.SetIndexOfLastTxProcessed(miniBlockHeader, processedMiniBlocksDestMeInfo)
	assert.Nil(t, err)
	assert.Equal(t, int32(8), miniBlockHeader.GetIndexOfLastTxProcessed())
}

func TestBaseProcessor_SetProcessingTypeAndConstructionStateForScheduledMb(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments(createComponentHolderMocks())
	bp, _ := blproc.NewShardProcessor(arguments)

	mbHash := []byte("mb_hash")
	processedMiniBlocksDestMeInfo := make(map[string]*processedMb.ProcessedMiniBlockInfo)
	miniBlockHeader := &block.MiniBlockHeader{
		Hash: mbHash,
	}

	processedMbInfo := &processedMb.ProcessedMiniBlockInfo{
		FullyProcessed: false,
	}

	miniBlockHeader.SenderShardID = 0
	err := bp.SetProcessingTypeAndConstructionStateForScheduledMb(miniBlockHeader, processedMiniBlocksDestMeInfo)
	assert.Nil(t, err)
	assert.Equal(t, int32(block.Proposed), miniBlockHeader.GetConstructionState())
	assert.Equal(t, int32(block.Scheduled), miniBlockHeader.GetProcessingType())

	miniBlockHeader.SenderShardID = 1

	err = bp.SetProcessingTypeAndConstructionStateForScheduledMb(miniBlockHeader, processedMiniBlocksDestMeInfo)
	assert.Nil(t, err)
	assert.Equal(t, int32(block.Final), miniBlockHeader.GetConstructionState())
	assert.Equal(t, int32(block.Scheduled), miniBlockHeader.GetProcessingType())

	processedMiniBlocksDestMeInfo[string(mbHash)] = processedMbInfo

	err = bp.SetProcessingTypeAndConstructionStateForScheduledMb(miniBlockHeader, processedMiniBlocksDestMeInfo)
	assert.Nil(t, err)
	assert.Equal(t, int32(block.PartialExecuted), miniBlockHeader.GetConstructionState())
	assert.Equal(t, int32(block.Scheduled), miniBlockHeader.GetProcessingType())
}

func TestBaseProcessor_SetProcessingTypeAndConstructionStateForNormalMb(t *testing.T) {
	t.Parallel()

	t.Run("set processing/construction for normal mini blocks not processed, should work", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		mbHash := []byte("mb_hash")
		processedMiniBlocksDestMeInfo := make(map[string]*processedMb.ProcessedMiniBlockInfo)
		miniBlockHeader := &block.MiniBlockHeader{
			Hash: mbHash,
		}

		processedMbInfo := &processedMb.ProcessedMiniBlockInfo{
			FullyProcessed: false,
		}

		err := bp.SetProcessingTypeAndConstructionStateForNormalMb(miniBlockHeader, processedMiniBlocksDestMeInfo)
		assert.Nil(t, err)
		assert.Equal(t, int32(block.Final), miniBlockHeader.GetConstructionState())
		assert.Equal(t, int32(block.Normal), miniBlockHeader.GetProcessingType())

		processedMiniBlocksDestMeInfo[string(mbHash)] = processedMbInfo

		err = bp.SetProcessingTypeAndConstructionStateForNormalMb(miniBlockHeader, processedMiniBlocksDestMeInfo)
		assert.Nil(t, err)
		assert.Equal(t, int32(block.PartialExecuted), miniBlockHeader.GetConstructionState())
		assert.Equal(t, int32(block.Normal), miniBlockHeader.GetProcessingType())
	})

	t.Run("set processing/construction for normal mini blocks already processed, should work", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			IsMiniBlockExecutedCalled: func(i []byte) bool {
				return true
			},
		}
		bp, _ := blproc.NewShardProcessor(arguments)

		mbHash := []byte("mb_hash")
		processedMiniBlocksDestMeInfo := make(map[string]*processedMb.ProcessedMiniBlockInfo)
		miniBlockHeader := &block.MiniBlockHeader{
			Hash: mbHash,
		}

		processedMbInfo := &processedMb.ProcessedMiniBlockInfo{
			FullyProcessed: false,
		}

		err := bp.SetProcessingTypeAndConstructionStateForNormalMb(miniBlockHeader, processedMiniBlocksDestMeInfo)
		assert.Nil(t, err)
		assert.Equal(t, int32(block.Final), miniBlockHeader.GetConstructionState())
		assert.Equal(t, int32(block.Processed), miniBlockHeader.GetProcessingType())

		processedMiniBlocksDestMeInfo[string(mbHash)] = processedMbInfo

		err = bp.SetProcessingTypeAndConstructionStateForNormalMb(miniBlockHeader, processedMiniBlocksDestMeInfo)
		assert.Nil(t, err)
		assert.Equal(t, int32(block.PartialExecuted), miniBlockHeader.GetConstructionState())
		assert.Equal(t, int32(block.Processed), miniBlockHeader.GetProcessingType())
	})
}

// ------- RevertState
func TestBaseProcessor_RevertStateRecreateTrieFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("err")
	arguments := CreateMockArguments(createComponentHolderMocks())
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash []byte) error {
			return expectedErr
		},
	}

	bp, _ := blproc.NewShardProcessor(arguments)

	hdr := block.Header{Nonce: 37}
	err := bp.RevertStateToBlock(&hdr, hdr.RootHash)
	assert.Equal(t, expectedErr, err)
}

// removeHeadersBehindNonceFromPools
func TestBaseProcessor_RemoveHeadersBehindNonceFromPools(t *testing.T) {
	t.Parallel()

	removeFromDataPoolWasCalled := false
	dataPool := initDataPool([]byte(""))
	dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
		cs := &mock.HeadersCacherStub{}
		cs.RegisterHandlerCalled = func(i func(header data.HeaderHandler, key []byte)) {
		}
		cs.GetHeaderByHashCalled = func(hash []byte) (handler data.HeaderHandler, err error) {
			return nil, err
		}
		cs.RemoveHeaderByHashCalled = func(key []byte) {
		}
		cs.LenCalled = func() int {
			return 0
		}
		cs.MaxSizeCalled = func() int {
			return 1000
		}
		cs.NoncesCalled = func(shardId uint32) []uint64 {
			return []uint64{1, 2, 3}
		}
		cs.GetHeaderByNonceAndShardIdCalled = func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
			hdrs := make([]data.HeaderHandler, 0)
			hdrs = append(hdrs, &block.Header{Nonce: 2})
			return hdrs, nil, nil
		}

		return cs
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = dataPool
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
		RemoveBlockDataFromPoolCalled: func(body *block.Body) error {
			removeFromDataPoolWasCalled = true
			return nil
		},
	}
	bp, _ := blproc.NewShardProcessor(arguments)

	bp.RemoveHeadersBehindNonceFromPools(true, 0, 4)

	assert.True(t, removeFromDataPoolWasCalled)
}

// ------- ComputeNewNoncePrevHash

func TestBlockProcessor_computeHeaderHashMarshalizerFail1ShouldErr(t *testing.T) {
	t.Parallel()
	marshalizer := &mock.MarshalizerStub{}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	bp, _ := blproc.NewShardProcessor(arguments)
	hdr, txBlock := createTestHdrTxBlockBody()
	expectedError := errors.New("marshalizer fail")
	marshalizer.MarshalCalled = func(obj interface{}) (bytes []byte, e error) {
		if hdr == obj {
			return nil, expectedError
		}

		if reflect.DeepEqual(txBlock, obj) {
			return []byte("txBlockBodyMarshalized"), nil
		}
		return nil, nil
	}
	_, err := bp.ComputeHeaderHash(hdr)
	assert.Equal(t, expectedError, err)
}

func TestBlockPorcessor_ComputeNewNoncePrevHashShouldWork(t *testing.T) {
	t.Parallel()
	marshalizer := &mock.MarshalizerStub{}
	hasher := &mock.HasherStub{}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.IntMarsh = marshalizer
	coreComponents.Hash = hasher
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	bp, _ := blproc.NewShardProcessor(arguments)
	hdr, txBlock := createTestHdrTxBlockBody()
	marshalizer.MarshalCalled = func(obj interface{}) (bytes []byte, e error) {
		if hdr == obj {
			return []byte("hdrHeaderMarshalized"), nil
		}
		if reflect.DeepEqual(txBlock, obj) {
			return []byte("txBlockBodyMarshalized"), nil
		}
		return nil, nil
	}
	hasher.ComputeCalled = func(s string) []byte {
		if s == "hdrHeaderMarshalized" {
			return []byte("hdr hash")
		}
		if s == "txBlockBodyMarshalized" {
			return []byte("tx block body hash")
		}
		return nil
	}
	_, err := bp.ComputeHeaderHash(hdr)
	assert.Nil(t, err)
}

func createShardProcessHeadersToSaveLastNotarized(
	highestNonce uint64,
	genesisHdr data.HeaderHandler,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) []data.HeaderHandler {
	rootHash := []byte("roothash")
	processedHdrs := make([]data.HeaderHandler, 0)

	headerMarsh, _ := marshalizer.Marshal(genesisHdr)
	headerHash := hasher.Compute(string(headerMarsh))

	for i := uint64(1); i <= highestNonce; i++ {
		hdr := &block.Header{
			Nonce:         i,
			Round:         i,
			Signature:     rootHash,
			RandSeed:      rootHash,
			PrevRandSeed:  rootHash,
			PubKeysBitmap: rootHash,
			RootHash:      rootHash,
			PrevHash:      headerHash}
		processedHdrs = append(processedHdrs, hdr)

		headerMarsh, _ = marshalizer.Marshal(hdr)
		headerHash = hasher.Compute(string(headerMarsh))
	}

	return processedHdrs
}

func createMetaProcessHeadersToSaveLastNoterized(
	highestNonce uint64,
	genesisHdr data.HeaderHandler,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) []data.HeaderHandler {
	rootHash := []byte("roothash")
	processedHdrs := make([]data.HeaderHandler, 0)

	headerMarsh, _ := marshalizer.Marshal(genesisHdr)
	headerHash := hasher.Compute(string(headerMarsh))

	for i := uint64(1); i <= highestNonce; i++ {
		hdr := &block.MetaBlock{
			Nonce:         i,
			Round:         i,
			Signature:     rootHash,
			RandSeed:      rootHash,
			PrevRandSeed:  rootHash,
			PubKeysBitmap: rootHash,
			RootHash:      rootHash,
			PrevHash:      headerHash}
		processedHdrs = append(processedHdrs, hdr)

		headerMarsh, _ = marshalizer.Marshal(hdr)
		headerHash = hasher.Compute(string(headerMarsh))
	}

	return processedHdrs
}

func TestBaseProcessor_SaveLastNotarizedInOneShardHdrsSliceForShardIsNil(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.Hash = &hashingMocks.HasherMock{}
	coreComponents.IntMarsh = &mock.MarshalizerMock{}
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)
	prHdrs := createShardProcessHeadersToSaveLastNotarized(10, &block.Header{}, &hashingMocks.HasherMock{}, &mock.MarshalizerMock{})

	err := sp.SaveLastNotarizedHeader(2, prHdrs)

	assert.Equal(t, process.ErrNotarizedHeadersSliceForShardIsNil, err)
}

func TestBaseProcessor_SaveLastNotarizedInMultiShardHdrsSliceForShardIsNil(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.Hash = &hashingMocks.HasherMock{}
	coreComponents.IntMarsh = &mock.MarshalizerMock{}
	bootstrapComponents.Coordinator = shardCoordinator
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	prHdrs := createShardProcessHeadersToSaveLastNotarized(10, &block.Header{}, &hashingMocks.HasherMock{}, &mock.MarshalizerMock{})

	err := sp.SaveLastNotarizedHeader(6, prHdrs)

	assert.Equal(t, process.ErrNotarizedHeadersSliceForShardIsNil, err)
}

func TestBaseProcessor_SaveLastNotarizedHdrShardGood(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.Hash = &hashingMocks.HasherMock{}
	coreComponents.IntMarsh = &mock.MarshalizerMock{}
	bootstrapComponents.Coordinator = shardCoordinator
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	sp, _ := blproc.NewShardProcessor(arguments)
	argsHeaderValidator := blproc.ArgsHeaderValidator{
		Hasher:      coreComponents.Hasher(),
		Marshalizer: coreComponents.InternalMarshalizer(),
	}
	headerValidator, _ := blproc.NewHeaderValidator(argsHeaderValidator)
	sp.SetHeaderValidator(headerValidator)

	genesisBlcks := createGenesisBlocks(shardCoordinator)

	highestNonce := uint64(10)
	shardId := uint32(0)
	prHdrs := createShardProcessHeadersToSaveLastNotarized(
		highestNonce,
		genesisBlcks[shardId],
		coreComponents.Hasher(),
		coreComponents.InternalMarshalizer())

	err := sp.SaveLastNotarizedHeader(shardId, prHdrs)
	assert.Nil(t, err)

	assert.Equal(t, highestNonce, sp.LastNotarizedHdrForShard(shardId).GetNonce())
}

func TestBaseProcessor_SaveLastNotarizedHdrMetaGood(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.Hash = &hashingMocks.HasherMock{}
	coreComponents.IntMarsh = &mock.MarshalizerMock{}
	bootstrapComponents.Coordinator = shardCoordinator
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	argsHeaderValidator := blproc.ArgsHeaderValidator{
		Hasher:      coreComponents.Hasher(),
		Marshalizer: coreComponents.InternalMarshalizer(),
	}
	headerValidator, _ := blproc.NewHeaderValidator(argsHeaderValidator)
	sp.SetHeaderValidator(headerValidator)

	genesisBlcks := createGenesisBlocks(shardCoordinator)

	highestNonce := uint64(10)
	prHdrs := createMetaProcessHeadersToSaveLastNoterized(
		highestNonce,
		genesisBlcks[core.MetachainShardId],
		coreComponents.Hasher(),
		coreComponents.InternalMarshalizer())

	err := sp.SaveLastNotarizedHeader(core.MetachainShardId, prHdrs)
	assert.Nil(t, err)

	assert.Equal(t, highestNonce, sp.LastNotarizedHdrForShard(core.MetachainShardId).GetNonce())
}

func TestShardProcessor_ProcessBlockEpochDoesNotMatchShouldErr(t *testing.T) {
	t.Parallel()
	blockChain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Epoch: 2,
			}
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}
	coreComponents, dataComponents, bootstrapComponents, statusComponents := CreateCoreComponentsMultiShard()
	dataComponents.BlockChain = blockChain
	arguments := CreateMockArgumentsMultiShard(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)
	header := &block.Header{Round: 10, Nonce: 1}

	blk := &block.Body{}
	_, _, err := sp.ProcessBlock(header, blk, func() time.Duration { return time.Second })

	assert.True(t, errors.Is(err, process.ErrEpochDoesNotMatch))
}

func TestShardProcessor_ProcessBlockEpochDoesNotMatchShouldErr2(t *testing.T) {
	t.Parallel()

	randSeed := []byte("randseed")
	blockChain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Epoch:           1,
				RandSeed:        randSeed,
				AccumulatedFees: big.NewInt(0),
				DeveloperFees:   big.NewInt(0),
			}
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := CreateCoreComponentsMultiShard()
	dataComponents.BlockChain = blockChain
	arguments := CreateMockArgumentsMultiShard(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		EpochCalled: func() uint32 {
			return 1
		},
	}

	sp, _ := blproc.NewShardProcessor(arguments)
	header := &block.Header{Round: 10, Nonce: 1, Epoch: 5, RandSeed: randSeed, PrevRandSeed: randSeed}

	blk := &block.Body{}
	_, _, err := sp.ProcessBlock(header, blk, func() time.Duration { return time.Second })

	assert.True(t, errors.Is(err, process.ErrEpochDoesNotMatch))
}

func TestShardProcessor_ProcessBlockEpochDoesNotMatchShouldErr3(t *testing.T) {
	t.Parallel()

	randSeed := []byte("randseed")
	blockChain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Epoch:    3,
				RandSeed: randSeed,
			}
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := CreateCoreComponentsMultiShard()
	dataComponents.BlockChain = blockChain
	arguments := CreateMockArgumentsMultiShard(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		EpochCalled: func() uint32 {
			return 2
		},
		IsEpochStartCalled: func() bool {
			return true
		},
	}

	sp, _ := blproc.NewShardProcessor(arguments)
	header := &block.Header{Round: 10, Nonce: 1, Epoch: 5, RandSeed: randSeed, PrevRandSeed: randSeed}

	blk := &block.Body{}
	_, _, err := sp.ProcessBlock(header, blk, func() time.Duration { return time.Second })

	assert.True(t, errors.Is(err, process.ErrEpochDoesNotMatch))
}

func TestShardProcessor_ProcessBlockEpochDoesNotMatchShouldErrMetaHashDoesNotMatch(t *testing.T) {
	t.Parallel()

	randSeed := []byte("randseed")
	chain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Epoch:    2,
				RandSeed: randSeed,
			}
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}
	hasher := &mock.HasherStub{ComputeCalled: func(s string) []byte {
		return nil
	}}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := CreateCoreComponentsMultiShard()
	dataComponents.BlockChain = chain
	coreComponents.Hash = hasher
	arguments := CreateMockArgumentsMultiShard(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	epochStartTrigger := &mock.EpochStartTriggerStub{
		EpochCalled: func() uint32 {
			return 2
		},
		MetaEpochCalled: func() uint32 {
			return 3
		},
		IsEpochStartCalled: func() bool {
			return true
		},
		EpochFinalityAttestingRoundCalled: func() uint64 {
			return 100
		},
	}
	arguments.EpochStartTrigger = epochStartTrigger

	sp, _ := blproc.NewShardProcessor(arguments)
	rootHash, _ := arguments.AccountsDB[state.UserAccountsState].RootHash()
	epochStartHash := []byte("epochStartHash")
	header := &block.Header{
		Round:              10,
		Nonce:              1,
		Epoch:              3,
		RandSeed:           randSeed,
		PrevRandSeed:       randSeed,
		EpochStartMetaHash: epochStartHash,
		RootHash:           rootHash,
		AccumulatedFees:    big.NewInt(0),
		DeveloperFees:      big.NewInt(0),
	}

	blk := &block.Body{}
	_, _, err := sp.ProcessBlock(header, blk, func() time.Duration { return time.Second })
	assert.True(t, errors.Is(err, process.ErrEpochDoesNotMatch))

	epochStartTrigger.EpochStartMetaHdrHashCalled = func() []byte {
		return header.EpochStartMetaHash
	}
	_, _, err = sp.ProcessBlock(header, blk, func() time.Duration { return time.Second })
	assert.Nil(t, err)
}

func TestShardProcessor_ProcessBlockEpochDoesNotMatchShouldErrMetaHashDoesNotMatchForOldEpoch(t *testing.T) {
	t.Parallel()

	randSeed := []byte("randseed")
	chain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Epoch:    2,
				RandSeed: randSeed,
			}
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}
	hasher := &mock.HasherStub{ComputeCalled: func(s string) []byte {
		return nil
	}}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := CreateCoreComponentsMultiShard()
	coreComponents.Hash = hasher
	dataComponents.BlockChain = chain
	arguments := CreateMockArgumentsMultiShard(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		EpochCalled: func() uint32 {
			return 5
		},
		MetaEpochCalled: func() uint32 {
			return 6
		},
		IsEpochStartCalled: func() bool {
			return true
		},
		EpochFinalityAttestingRoundCalled: func() uint64 {
			return 100
		},
	}

	sp, _ := blproc.NewShardProcessor(arguments)
	rootHash, _ := arguments.AccountsDB[state.UserAccountsState].RootHash()
	epochStartHash := []byte("epochStartHash")
	header := &block.Header{
		Round:              10,
		Nonce:              1,
		Epoch:              3,
		RandSeed:           randSeed,
		PrevRandSeed:       randSeed,
		EpochStartMetaHash: epochStartHash,
		RootHash:           rootHash,
		AccumulatedFees:    big.NewInt(0),
		DeveloperFees:      big.NewInt(0),
	}

	blk := &block.Body{}
	_, _, err := sp.ProcessBlock(header, blk, func() time.Duration { return time.Second })
	assert.True(t, errors.Is(err, process.ErrMissingHeader))

	metaHdr := &block.MetaBlock{}
	metaHdrData, _ := coreComponents.InternalMarshalizer().Marshal(metaHdr)
	_ = dataComponents.StorageService().Put(dataRetriever.MetaBlockUnit, header.EpochStartMetaHash, metaHdrData)

	_, _, err = sp.ProcessBlock(header, blk, func() time.Duration { return time.Second })
	assert.True(t, errors.Is(err, process.ErrEpochDoesNotMatch))

	metaHdr = &block.MetaBlock{Epoch: 3, EpochStart: block.EpochStart{
		LastFinalizedHeaders: []block.EpochStartShardData{{}},
		Economics:            block.Economics{},
	}}
	metaHdrData, _ = coreComponents.InternalMarshalizer().Marshal(metaHdr)
	_ = dataComponents.StorageService().Put(dataRetriever.MetaBlockUnit, header.EpochStartMetaHash, metaHdrData)

	_, _, err = sp.ProcessBlock(header, blk, func() time.Duration { return time.Second })
	assert.Nil(t, err)
}

func TestBlockProcessor_PruneStateOnRollbackPrunesPeerTrieIfAccPruneIsDisabled(t *testing.T) {
	t.Parallel()

	pruningCalled := 0
	peerAccDb := &stateMock.AccountsStub{
		PruneTrieCalled: func(rootHash []byte, identifier state.TriePruningIdentifier, _ state.PruningHandler) {
			pruningCalled++
		},
		CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
			pruningCalled++
		},
		IsPruningEnabledCalled: func() bool {
			return true
		},
	}

	arguments := CreateMockArguments(createComponentHolderMocks())
	arguments.AccountsDB[state.PeerAccountsState] = peerAccDb
	bp, _ := blproc.NewShardProcessor(arguments)

	prevHeader := &block.MetaBlock{
		RootHash:               []byte("prevRootHash"),
		ValidatorStatsRootHash: []byte("prevValidatorRootHash"),
	}
	currHeader := &block.MetaBlock{
		RootHash:               []byte("prevRootHash"),
		ValidatorStatsRootHash: []byte("currValidatorRootHash"),
	}

	bp.PruneStateOnRollback(currHeader, []byte("currHeaderHash"), prevHeader, []byte("prevHeaderHash"))
	assert.Equal(t, 2, pruningCalled)
}

func TestBlockProcessor_PruneStateOnRollbackPrunesPeerTrieIfSameRootHashButDifferentValidatorRootHash(t *testing.T) {
	t.Parallel()

	pruningCalled := 0
	peerAccDb := &stateMock.AccountsStub{
		PruneTrieCalled: func(rootHash []byte, identifier state.TriePruningIdentifier, _ state.PruningHandler) {
			pruningCalled++
		},
		CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
			pruningCalled++
		},
		IsPruningEnabledCalled: func() bool {
			return true
		},
	}

	accDb := &stateMock.AccountsStub{
		PruneTrieCalled: func(rootHash []byte, identifier state.TriePruningIdentifier, _ state.PruningHandler) {
			pruningCalled++
		},
		CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
			pruningCalled++
		},
		IsPruningEnabledCalled: func() bool {
			return true
		},
	}

	arguments := CreateMockArguments(createComponentHolderMocks())
	arguments.AccountsDB[state.PeerAccountsState] = peerAccDb
	arguments.AccountsDB[state.UserAccountsState] = accDb
	bp, _ := blproc.NewShardProcessor(arguments)

	prevHeader := &block.MetaBlock{
		RootHash:               []byte("prevRootHash"),
		ValidatorStatsRootHash: []byte("prevValidatorRootHash"),
	}
	currHeader := &block.MetaBlock{
		RootHash:               []byte("prevRootHash"),
		ValidatorStatsRootHash: []byte("currValidatorRootHash"),
	}

	bp.PruneStateOnRollback(currHeader, []byte("currHeaderHash"), prevHeader, []byte("prevHeaderHash"))
	assert.Equal(t, 2, pruningCalled)
}

func TestBlockProcessor_RequestHeadersIfMissingShouldWorkWhenSortedHeadersListIsEmpty(t *testing.T) {
	t.Parallel()

	var requestedNonces []uint64
	var mutRequestedNonces sync.Mutex

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	roundHandler := &mock.RoundHandlerMock{}
	coreComponents.RoundField = roundHandler
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	requestHandlerStub := &testscommon.RequestHandlerStub{
		RequestMetaHeaderByNonceCalled: func(nonce uint64) {
			mutRequestedNonces.Lock()
			requestedNonces = append(requestedNonces, nonce)
			mutRequestedNonces.Unlock()
		},
	}
	arguments.RequestHandler = requestHandlerStub
	sp, _ := blproc.NewShardProcessor(arguments)

	sortedHeaders := make([]data.HeaderHandler, 0)

	requestedNonces = make([]uint64, 0)
	roundHandler.RoundIndex = process.MaxHeaderRequestsAllowed + 5
	_ = sp.RequestHeadersIfMissing(sortedHeaders, core.MetachainShardId)
	time.Sleep(100 * time.Millisecond)
	mutRequestedNonces.Lock()
	sort.Slice(requestedNonces, func(i, j int) bool {
		return requestedNonces[i] < requestedNonces[j]
	})
	mutRequestedNonces.Unlock()
	expectedNonces := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	assert.Equal(t, expectedNonces, requestedNonces)

	requestedNonces = make([]uint64, 0)
	roundHandler.RoundIndex = 5
	_ = sp.RequestHeadersIfMissing(sortedHeaders, core.MetachainShardId)
	time.Sleep(100 * time.Millisecond)
	mutRequestedNonces.Lock()
	sort.Slice(requestedNonces, func(i, j int) bool {
		return requestedNonces[i] < requestedNonces[j]
	})
	mutRequestedNonces.Unlock()
	expectedNonces = []uint64{1, 2, 3}
	assert.Equal(t, expectedNonces, requestedNonces)
}

func TestBlockProcessor_RequestHeadersIfMissingShouldWork(t *testing.T) {
	t.Parallel()

	var requestedNonces []uint64
	var mutRequestedNonces sync.Mutex

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	roundHandler := &mock.RoundHandlerMock{}
	coreComponents.RoundField = roundHandler
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	requestHandlerStub := &testscommon.RequestHandlerStub{
		RequestMetaHeaderByNonceCalled: func(nonce uint64) {
			mutRequestedNonces.Lock()
			requestedNonces = append(requestedNonces, nonce)
			mutRequestedNonces.Unlock()
		},
	}
	arguments.RequestHandler = requestHandlerStub
	sp, _ := blproc.NewShardProcessor(arguments)

	sortedHeaders := make([]data.HeaderHandler, 0)

	crossNotarizedHeader := &block.MetaBlock{
		Nonce: 5,
		Round: 5,
	}
	arguments.BlockTracker.AddCrossNotarizedHeader(core.MetachainShardId, crossNotarizedHeader, []byte("hash"))

	hdr1 := &block.MetaBlock{
		Nonce: 1,
		Round: 1,
	}
	sortedHeaders = append(sortedHeaders, hdr1)

	hdr2 := &block.MetaBlock{
		Nonce: 8,
		Round: 8,
	}
	sortedHeaders = append(sortedHeaders, hdr2)

	hdr3 := &block.MetaBlock{
		Nonce: 10,
		Round: 10,
	}
	sortedHeaders = append(sortedHeaders, hdr3)

	requestedNonces = make([]uint64, 0)
	roundHandler.RoundIndex = 15
	_ = sp.RequestHeadersIfMissing(sortedHeaders, core.MetachainShardId)
	time.Sleep(100 * time.Millisecond)
	mutRequestedNonces.Lock()
	sort.Slice(requestedNonces, func(i, j int) bool {
		return requestedNonces[i] < requestedNonces[j]
	})
	mutRequestedNonces.Unlock()
	expectedNonces := []uint64{6, 7, 9, 11, 12, 13}
	assert.Equal(t, expectedNonces, requestedNonces)

	requestedNonces = make([]uint64, 0)
	roundHandler.RoundIndex = process.MaxHeaderRequestsAllowed + 10
	_ = sp.RequestHeadersIfMissing(sortedHeaders, core.MetachainShardId)
	time.Sleep(100 * time.Millisecond)
	mutRequestedNonces.Lock()
	sort.Slice(requestedNonces, func(i, j int) bool {
		return requestedNonces[i] < requestedNonces[j]
	})
	mutRequestedNonces.Unlock()
	expectedNonces = []uint64{6, 7, 9, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}
	assert.Equal(t, expectedNonces, requestedNonces)
}

func TestBlockProcessor_RequestHeadersIfMissingShouldAddHeaderIntoTrackerPool(t *testing.T) {
	t.Parallel()

	var addedNonces []uint64
	poolsHolderStub := initDataPool([]byte(""))
	poolsHolderStub.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{
			GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
				addedNonces = append(addedNonces, hdrNonce)
				return []data.HeaderHandler{&block.MetaBlock{Nonce: 1}}, [][]byte{[]byte("hash")}, nil
			},
		}
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = poolsHolderStub
	roundHandler := &mock.RoundHandlerMock{}
	coreComponents.RoundField = roundHandler
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	sp, _ := blproc.NewShardProcessor(arguments)

	sortedHeaders := make([]data.HeaderHandler, 0)

	crossNotarizedHeader := &block.MetaBlock{
		Nonce: 5,
		Round: 5,
	}
	arguments.BlockTracker.AddCrossNotarizedHeader(core.MetachainShardId, crossNotarizedHeader, []byte("hash"))

	hdr1 := &block.MetaBlock{
		Nonce: 1,
		Round: 1,
	}
	sortedHeaders = append(sortedHeaders, hdr1)

	hdr2 := &block.MetaBlock{
		Nonce: 8,
		Round: 8,
	}
	sortedHeaders = append(sortedHeaders, hdr2)

	hdr3 := &block.MetaBlock{
		Nonce: 10,
		Round: 10,
	}
	sortedHeaders = append(sortedHeaders, hdr3)

	addedNonces = make([]uint64, 0)

	roundHandler.RoundIndex = 12
	_ = sp.RequestHeadersIfMissing(sortedHeaders, core.MetachainShardId)

	expectedAddedNonces := []uint64{6, 7, 9}
	assert.Equal(t, expectedAddedNonces, addedNonces)
}

func TestAddHeaderIntoTrackerPool_ShouldWork(t *testing.T) {
	t.Parallel()

	var wasCalled bool
	shardID := core.MetachainShardId
	nonce := uint64(1)
	poolsHolderStub := initDataPool([]byte(""))
	poolsHolderStub.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{
			GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
				if hdrNonce == nonce && shardId == shardID {
					wasCalled = true
					return []data.HeaderHandler{&block.MetaBlock{Nonce: 1}}, [][]byte{[]byte("hash")}, nil
				}

				return nil, nil, errors.New("error")
			},
		}
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = poolsHolderStub
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	wasCalled = false
	sp.AddHeaderIntoTrackerPool(nonce+1, shardID)
	assert.False(t, wasCalled)

	wasCalled = false
	sp.AddHeaderIntoTrackerPool(nonce, shardID)
	assert.True(t, wasCalled)
}

func TestBaseProcessor_commitTrieEpochRootHashIfNeededNilStorerShouldErr(t *testing.T) {
	t.Parallel()

	epoch := uint32(37)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	store := dataRetriever.NewChainStorer()
	dataComponents.Storage = store
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	mb := &block.MetaBlock{Epoch: epoch}
	err := sp.CommitTrieEpochRootHashIfNeeded(mb, []byte("root"))
	require.NotNil(t, err)
	require.True(t, strings.Contains(err.Error(), dataRetriever.ErrStorerNotFound.Error()))
	require.True(t, strings.Contains(err.Error(), dataRetriever.TrieEpochRootHashUnit.String()))
}

func TestBaseProcessor_commitTrieEpochRootHashIfNeededDisabledStorerShouldNotErr(t *testing.T) {
	t.Parallel()

	epoch := uint32(37)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.Storage.AddStorer(dataRetriever.TrieEpochRootHashUnit, &storageunit.NilStorer{})
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	sp, _ := blproc.NewShardProcessor(arguments)

	mb := &block.MetaBlock{Epoch: epoch}
	err := sp.CommitTrieEpochRootHashIfNeeded(mb, []byte("root"))
	require.NoError(t, err)
}

func TestBaseProcessor_commitTrieEpochRootHashIfNeededCannotFindUserAccountStateShouldErr(t *testing.T) {
	t.Parallel()

	epoch := uint32(37)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB = map[state.AccountsDbIdentifier]state.AccountsAdapter{}

	sp, _ := blproc.NewShardProcessor(arguments)

	mb := &block.MetaBlock{Epoch: epoch}
	err := sp.CommitTrieEpochRootHashIfNeeded(mb, []byte("root"))
	require.True(t, errors.Is(err, process.ErrNilAccountsAdapter))
}

func TestBaseProcessor_commitTrieEpochRootHashIfNeededShouldWork(t *testing.T) {
	t.Parallel()

	epoch := uint32(37)
	rootHash := []byte("root-hash")

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.UInt64ByteSliceConv = uint64ByteSlice.NewBigEndianConverter()
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TrieEpochRootHashUnit,
		&storageStubs.StorerStub{
			PutCalled: func(key, data []byte) error {
				restoredEpoch, err := coreComponents.UInt64ByteSliceConv.ToUint64(key)
				require.NoError(t, err)
				require.Equal(t, epoch, uint32(restoredEpoch))
				return nil
			},
		},
	)
	dataComponents.Storage = store

	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB = map[state.AccountsDbIdentifier]state.AccountsAdapter{
		state.UserAccountsState: &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return rootHash, nil
			},
			GetAllLeavesCalled: func(channels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.TrieLeafParser) error {
				close(channels.LeavesChan)
				channels.ErrChan.Close()
				return nil
			},
		},
	}

	sp, _ := blproc.NewShardProcessor(arguments)

	mb := &block.MetaBlock{Epoch: epoch}
	err := sp.CommitTrieEpochRootHashIfNeeded(mb, []byte("root"))
	require.NoError(t, err)
}

func TestBaseProcessor_commitTrieEpochRootHashIfNeeded_GetAllLeaves(t *testing.T) {
	t.Parallel()

	epoch := uint32(37)
	rootHash := []byte("root-hash")

	t.Run("error on getting the leaves", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.UInt64ByteSliceConv = uint64ByteSlice.NewBigEndianConverter()
		store := dataRetriever.NewChainStorer()
		store.AddStorer(dataRetriever.TrieEpochRootHashUnit,
			&storageStubs.StorerStub{
				PutCalled: func(key, data []byte) error {
					return nil
				},
			},
		)
		dataComponents.Storage = store

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		expectedErr := errors.New("expected error")
		arguments.AccountsDB = map[state.AccountsDbIdentifier]state.AccountsAdapter{
			state.UserAccountsState: &stateMock.AccountsStub{
				RootHashCalled: func() ([]byte, error) {
					return rootHash, nil
				},
				GetAllLeavesCalled: func(channels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.TrieLeafParser) error {
					close(channels.LeavesChan)
					channels.ErrChan.Close()
					return expectedErr
				},
			},
		}

		sp, _ := blproc.NewShardProcessor(arguments)

		mb := &block.MetaBlock{Epoch: epoch}
		err := sp.CommitTrieEpochRootHashIfNeeded(mb, []byte("root"))
		require.Equal(t, expectedErr, err)
	})

	t.Run("error on trie iterator chan", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.UInt64ByteSliceConv = uint64ByteSlice.NewBigEndianConverter()
		store := dataRetriever.NewChainStorer()
		store.AddStorer(dataRetriever.TrieEpochRootHashUnit,
			&storageStubs.StorerStub{
				PutCalled: func(key, data []byte) error {
					return nil
				},
			},
		)
		dataComponents.Storage = store

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		expectedErr := errors.New("expected error")
		arguments.AccountsDB = map[state.AccountsDbIdentifier]state.AccountsAdapter{
			state.UserAccountsState: &stateMock.AccountsStub{
				RootHashCalled: func() ([]byte, error) {
					return rootHash, nil
				},
				GetAllLeavesCalled: func(channels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeafParser common.TrieLeafParser) error {
					channels.ErrChan.WriteInChanNonBlocking(expectedErr)
					close(channels.LeavesChan)
					return nil
				},
			},
		}

		sp, _ := blproc.NewShardProcessor(arguments)

		mb := &block.MetaBlock{Epoch: epoch}
		err := sp.CommitTrieEpochRootHashIfNeeded(mb, []byte("root"))
		require.Equal(t, expectedErr, err)
	})
}

func TestBaseProcessor_commitTrieEpochRootHashIfNeededShouldUseDataTrieIfNeededWork(t *testing.T) {
	t.Parallel()

	var processDataTrieTests = []struct {
		processDataTrie        bool
		calledWithUserRootHash bool
	}{
		{false, false},
		{true, true},
	}

	for _, tt := range processDataTrieTests {
		epoch := uint32(37)
		rootHash := []byte("userAcc-root-hash")

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.UInt64ByteSliceConv = uint64ByteSlice.NewBigEndianConverter()
		coreComponents.IntMarsh = &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				userAccount := obj.(state.UserAccountHandler)
				userAccount.SetRootHash(rootHash)
				return nil
			},
		}

		store := dataRetriever.NewChainStorer()
		store.AddStorer(dataRetriever.TrieEpochRootHashUnit,
			&storageStubs.StorerStub{
				PutCalled: func(key, data []byte) error {
					restoredEpoch, err := coreComponents.UInt64ByteSliceConv.ToUint64(key)
					require.NoError(t, err)
					require.Equal(t, epoch, uint32(restoredEpoch))
					return nil
				},
			},
		)
		dataComponents.Storage = store
		calledWithUserAccountRootHash := false
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.AccountsDB = map[state.AccountsDbIdentifier]state.AccountsAdapter{
			state.UserAccountsState: &stateMock.AccountsStub{
				GetAllLeavesCalled: func(channels *common.TrieIteratorChannels, ctx context.Context, rh []byte, _ common.TrieLeafParser) error {
					if bytes.Equal(rootHash, rh) {
						calledWithUserAccountRootHash = true
						close(channels.LeavesChan)
						channels.ErrChan.Close()
						return nil
					}

					go func() {
						channels.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("address"), []byte("bytes"))
						close(channels.LeavesChan)
						channels.ErrChan.Close()
					}()

					return nil
				},
			},
		}

		arguments.Config.Debug.EpochStart.ProcessDataTrieOnCommitEpoch = tt.processDataTrie
		sp, _ := blproc.NewShardProcessor(arguments)

		mb := &block.MetaBlock{Epoch: epoch}
		err := sp.CommitTrieEpochRootHashIfNeeded(mb, []byte("root"))
		require.NoError(t, err)

		require.Equal(t, tt.calledWithUserRootHash, calledWithUserAccountRootHash)
	}
}

func TestBaseProcessor_updateState(t *testing.T) {
	t.Parallel()

	var pruneRootHash []byte
	var cancelPruneRootHash []byte

	poolMock := dataRetrieverMock.NewPoolsHolderMock()

	numHeaders := 5
	headers := make([]block.Header, numHeaders)
	for i := 0; i < numHeaders; i++ {
		headers[i] = block.Header{
			Nonce:    uint64(i),
			RootHash: []byte(strconv.Itoa(i)),
		}
	}

	hdrStore := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			if len(headers) != 0 {
				header := headers[0]
				headers = headers[1:]
				return json.Marshal(header)
			}

			return nil, nil
		},
	}

	storer := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return hdrStore, nil
		},
	}

	shardC := mock.NewMultiShardsCoordinatorMock(3)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = poolMock
	dataComponents.Storage = storer
	bootstrapComponents.Coordinator = shardC
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	arguments.BlockTracker = &mock.BlockTrackerMock{}
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		IsPruningEnabledCalled: func() bool {
			return true
		},
		PruneTrieCalled: func(rootHashParam []byte, identifier state.TriePruningIdentifier, _ state.PruningHandler) {
			pruneRootHash = rootHashParam
		},
		CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
			cancelPruneRootHash = rootHash
		},
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	prevRootHash := []byte("rootHash")
	for i := range headers {
		sp.UpdateState(
			&headers[i],
			headers[i].RootHash,
			prevRootHash,
			arguments.AccountsDB[state.UserAccountsState],
		)

		assert.Equal(t, prevRootHash, pruneRootHash)
		assert.Equal(t, prevRootHash, cancelPruneRootHash)

		prevRootHash = headers[i].RootHash
	}

	assert.Equal(t, []byte(strconv.Itoa(len(headers)-2)), pruneRootHash)
	assert.Equal(t, []byte(strconv.Itoa(len(headers)-2)), cancelPruneRootHash)
}

func TestBaseProcessor_ProcessScheduledBlockShouldFail(t *testing.T) {
	t.Parallel()

	t.Run("execute all scheduled txs fail", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		processHandler := arguments.CoreComponents.ProcessStatusHandler()
		mockProcessHandler := processHandler.(*testscommon.ProcessStatusHandlerStub)
		busyIdleCalled := make([]string, 0)
		mockProcessHandler.SetIdleCalled = func() {
			busyIdleCalled = append(busyIdleCalled, idleIdentifier)
		}
		mockProcessHandler.SetBusyCalled = func(reason string) {
			busyIdleCalled = append(busyIdleCalled, busyIdentifier)
		}

		localErr := errors.New("execute all err")
		scheduledTxsExec := &testscommon.ScheduledTxsExecutionStub{
			ExecuteAllCalled: func(func() time.Duration) error {
				return localErr
			},
		}

		arguments.ScheduledTxsExecutionHandler = scheduledTxsExec
		bp, _ := blproc.NewShardProcessor(arguments)

		err := bp.ProcessScheduledBlock(
			&block.MetaBlock{}, &block.Body{}, haveTime,
		)

		assert.Equal(t, localErr, err)
		assert.Equal(t, []string{busyIdentifier, idleIdentifier}, busyIdleCalled)
	})
	t.Run("get root hash fail", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		processHandler := arguments.CoreComponents.ProcessStatusHandler()
		mockProcessHandler := processHandler.(*testscommon.ProcessStatusHandlerStub)
		busyIdleCalled := make([]string, 0)
		mockProcessHandler.SetIdleCalled = func() {
			busyIdleCalled = append(busyIdleCalled, idleIdentifier)
		}
		mockProcessHandler.SetBusyCalled = func(reason string) {
			busyIdleCalled = append(busyIdleCalled, busyIdentifier)
		}

		localErr := errors.New("root hash err")
		accounts := &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return nil, localErr
			},
		}
		arguments.AccountsDB[state.UserAccountsState] = accounts

		bp, _ := blproc.NewShardProcessor(arguments)

		err := bp.ProcessScheduledBlock(
			&block.MetaBlock{}, &block.Body{}, haveTime,
		)

		assert.Equal(t, localErr, err)
		assert.Equal(t, []string{busyIdentifier, idleIdentifier}, busyIdleCalled)
	})
}

func TestBaseProcessor_ProcessScheduledBlockShouldWork(t *testing.T) {
	t.Parallel()
	rootHash := []byte("root hash to be tested")
	accounts := &stateMock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
	}

	initialGasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(11),
		DeveloperFees:   big.NewInt(12),
		GasProvided:     13,
		GasPenalized:    14,
		GasRefunded:     15,
	}

	finalGasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(101),
		DeveloperFees:   big.NewInt(103),
		GasProvided:     105,
		GasPenalized:    107,
		GasRefunded:     109,
	}

	feeHandler := createFeeHandlerMockForProcessScheduledBlock(initialGasAndFees, finalGasAndFees)
	gasHandler := createGasHandlerMockForProcessScheduledBlock(initialGasAndFees, finalGasAndFees)

	expectedGasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(90),
		DeveloperFees:   big.NewInt(91),
		GasProvided:     92,
		GasPenalized:    93,
		GasRefunded:     94,
	}

	wasCalledSetScheduledRootHash := false
	wasCalledSetScheduledGasAndFees := false
	scheduledTxsExec := &testscommon.ScheduledTxsExecutionStub{
		ExecuteAllCalled: func(func() time.Duration) error {
			return nil
		},
		SetScheduledRootHashCalled: func(hash []byte) {
			wasCalledSetScheduledRootHash = true
			require.Equal(t, rootHash, hash)
		},
		SetScheduledGasAndFeesCalled: func(gasAndFees scheduled.GasAndFees) {
			wasCalledSetScheduledGasAndFees = true
			require.Equal(t, expectedGasAndFees, gasAndFees)
		},
	}

	arguments := CreateMockArguments(createComponentHolderMocks())
	processHandler := arguments.CoreComponents.ProcessStatusHandler()
	mockProcessHandler := processHandler.(*testscommon.ProcessStatusHandlerStub)
	busyIdleCalled := make([]string, 0)
	mockProcessHandler.SetIdleCalled = func() {
		busyIdleCalled = append(busyIdleCalled, idleIdentifier)
	}
	mockProcessHandler.SetBusyCalled = func(reason string) {
		busyIdleCalled = append(busyIdleCalled, busyIdentifier)
	}

	arguments.AccountsDB[state.UserAccountsState] = accounts
	arguments.ScheduledTxsExecutionHandler = scheduledTxsExec
	arguments.FeeHandler = feeHandler
	arguments.GasHandler = gasHandler
	bp, _ := blproc.NewShardProcessor(arguments)

	err := bp.ProcessScheduledBlock(
		&block.MetaBlock{}, &block.Body{}, haveTime,
	)
	require.Nil(t, err)

	assert.True(t, wasCalledSetScheduledGasAndFees)
	assert.True(t, wasCalledSetScheduledRootHash)
	assert.Equal(t, []string{busyIdentifier, idleIdentifier}, busyIdleCalled) // the order is important
}

// get initial fees on first getGasAndFees call and final fees on second call
func createFeeHandlerMockForProcessScheduledBlock(initial, final scheduled.GasAndFees) process.TransactionFeeHandler {
	runCount := 0
	return &mock.FeeAccumulatorStub{
		GetAccumulatedFeesCalled: func() *big.Int {
			if runCount%4 >= 2 {
				return final.AccumulatedFees
			}
			runCount++
			return initial.AccumulatedFees
		},
		GetDeveloperFeesCalled: func() *big.Int {
			if runCount%4 >= 2 {
				return final.DeveloperFees
			}
			runCount++
			return initial.DeveloperFees
		},
	}
}

// get initial gas consumed on first getGasAndFees call and final gas consumed on second call
func createGasHandlerMockForProcessScheduledBlock(initial, final scheduled.GasAndFees) process.GasHandler {
	runCount := 0
	return &mock.GasHandlerMock{
		TotalGasProvidedCalled: func() uint64 {
			return initial.GasProvided
		},
		TotalGasPenalizedCalled: func() uint64 {
			if runCount%4 >= 2 {
				return final.GasPenalized
			}
			runCount++
			return initial.GasPenalized
		},
		TotalGasRefundedCalled: func() uint64 {
			if runCount%4 >= 2 {
				return final.GasRefunded
			}
			runCount++
			return initial.GasRefunded
		},
		TotalGasProvidedWithScheduledCalled: func() uint64 {
			return final.GasProvided
		},
	}
}

func TestBaseProcessor_gasAndFeesDelta(t *testing.T) {
	zeroGasAndFees := process.GetZeroGasAndFees()

	t.Run("final accumulatedFees lower then initial accumulatedFees", func(t *testing.T) {
		t.Parallel()

		initialGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(100),
		}

		finalGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(10),
		}

		gasAndFees := blproc.GasAndFeesDelta(initialGasAndFees, finalGasAndFees)
		assert.Equal(t, zeroGasAndFees, gasAndFees)
	})
	t.Run("final devFees lower then initial devFees", func(t *testing.T) {
		t.Parallel()

		initialGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(10),
			DeveloperFees:   big.NewInt(100),
		}

		finalGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(100),
			DeveloperFees:   big.NewInt(10),
		}

		gasAndFees := blproc.GasAndFeesDelta(initialGasAndFees, finalGasAndFees)
		assert.Equal(t, zeroGasAndFees, gasAndFees)
	})
	t.Run("final gasProvided lower then initial gasProvided", func(t *testing.T) {
		t.Parallel()

		initialGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(11),
			DeveloperFees:   big.NewInt(12),
			GasProvided:     100,
		}

		finalGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(101),
			DeveloperFees:   big.NewInt(102),
			GasProvided:     10,
		}

		gasAndFees := blproc.GasAndFeesDelta(initialGasAndFees, finalGasAndFees)
		assert.Equal(t, zeroGasAndFees, gasAndFees)
	})
	t.Run("final gasPenalized lower then initial gasPenalized", func(t *testing.T) {
		t.Parallel()

		initialGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(11),
			DeveloperFees:   big.NewInt(12),
			GasProvided:     13,
			GasPenalized:    100,
		}

		finalGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(101),
			DeveloperFees:   big.NewInt(102),
			GasProvided:     103,
			GasPenalized:    10,
		}

		gasAndFees := blproc.GasAndFeesDelta(initialGasAndFees, finalGasAndFees)
		assert.Equal(t, zeroGasAndFees, gasAndFees)
	})
	t.Run("final gasRefunded lower then initial gasRefunded", func(t *testing.T) {
		t.Parallel()

		initialGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(11),
			DeveloperFees:   big.NewInt(12),
			GasProvided:     13,
			GasPenalized:    14,
			GasRefunded:     100,
		}

		finalGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(101),
			DeveloperFees:   big.NewInt(102),
			GasProvided:     103,
			GasPenalized:    104,
			GasRefunded:     10,
		}

		gasAndFees := blproc.GasAndFeesDelta(initialGasAndFees, finalGasAndFees)
		assert.Equal(t, zeroGasAndFees, gasAndFees)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		initialGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(11),
			DeveloperFees:   big.NewInt(12),
			GasProvided:     13,
			GasPenalized:    14,
			GasRefunded:     15,
		}

		finalGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(101),
			DeveloperFees:   big.NewInt(103),
			GasProvided:     105,
			GasPenalized:    107,
			GasRefunded:     109,
		}

		expectedGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(0).Sub(finalGasAndFees.AccumulatedFees, initialGasAndFees.AccumulatedFees),
			DeveloperFees:   big.NewInt(0).Sub(finalGasAndFees.DeveloperFees, initialGasAndFees.DeveloperFees),
			GasProvided:     finalGasAndFees.GasProvided - initialGasAndFees.GasProvided,
			GasPenalized:    finalGasAndFees.GasPenalized - initialGasAndFees.GasPenalized,
			GasRefunded:     finalGasAndFees.GasRefunded - initialGasAndFees.GasRefunded,
		}

		gasAndFees := blproc.GasAndFeesDelta(initialGasAndFees, finalGasAndFees)

		assert.Equal(t, expectedGasAndFees, gasAndFees)
	})

}

func TestBaseProcessor_getIndexOfFirstMiniBlockToBeExecuted(t *testing.T) {
	t.Parallel()

	t.Run("scheduledMiniBlocks flag not set", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		index := bp.GetIndexOfFirstMiniBlockToBeExecuted(&block.MetaBlock{})
		assert.Equal(t, 0, index)
	})

	t.Run("scheduledMiniBlocks flag is set, empty block", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		index := bp.GetIndexOfFirstMiniBlockToBeExecuted(&block.MetaBlock{})
		assert.Equal(t, 0, index)
	})

	t.Run("get first index for the miniBlockHeader which is not processed executionType", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		mbh1 := block.MiniBlockHeader{}
		mbhReserved1 := block.MiniBlockHeaderReserved{ExecutionType: block.Processed}
		mbh1.Reserved, _ = mbhReserved1.Marshal()

		mbh2 := block.MiniBlockHeader{}
		mbhReserved2 := block.MiniBlockHeaderReserved{ExecutionType: block.Normal}
		mbh2.Reserved, _ = mbhReserved2.Marshal()

		metaBlock := &block.MetaBlock{
			MiniBlockHeaders: []block.MiniBlockHeader{
				mbh1,
				mbh2,
			},
		}

		index := bp.GetIndexOfFirstMiniBlockToBeExecuted(metaBlock)
		assert.Equal(t, 1, index)
	})
}

func TestBaseProcessor_getFinalMiniBlocks(t *testing.T) {
	t.Parallel()

	t.Run("scheduledMiniBlocks flag not set", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		body, err := bp.GetFinalMiniBlocks(&block.MetaBlock{}, &block.Body{})
		assert.Nil(t, err)
		assert.Equal(t, &block.Body{}, body)
	})

	t.Run("scheduledMiniBlocks flag is set, empty body", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		body, err := bp.GetFinalMiniBlocks(&block.MetaBlock{}, &block.Body{})
		assert.Nil(t, err)
		assert.Equal(t, &block.Body{}, body)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		mb1 := &block.MiniBlock{
			TxHashes: [][]byte{[]byte("txHash1")},
		}
		mb2 := &block.MiniBlock{
			TxHashes: [][]byte{[]byte("txHash2")},
		}
		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				mb1,
				mb2,
			},
		}

		mbh1 := block.MiniBlockHeader{}
		mbhReserved1 := block.MiniBlockHeaderReserved{State: block.Proposed}
		mbh1.Reserved, _ = mbhReserved1.Marshal()

		mbh2 := block.MiniBlockHeader{}
		mbhReserved2 := block.MiniBlockHeaderReserved{State: block.Final}
		mbh2.Reserved, _ = mbhReserved2.Marshal()

		metaBlock := &block.MetaBlock{
			MiniBlockHeaders: []block.MiniBlockHeader{
				mbh1,
				mbh2,
			},
		}

		expectedBody := &block.Body{MiniBlocks: block.MiniBlockSlice{mb2}}

		retBody, err := bp.GetFinalMiniBlocks(metaBlock, body)
		assert.Nil(t, err)
		assert.Equal(t, expectedBody, retBody)
	})
}

func TestBaseProcessor_getScheduledMiniBlocksFromMe(t *testing.T) {
	t.Parallel()

	t.Run("wrong body type", func(t *testing.T) {
		t.Parallel()

		retBody, err := blproc.GetScheduledMiniBlocksFromMe(&block.Header{}, &wrongBody{})
		assert.Equal(t, process.ErrWrongTypeAssertion, err)
		assert.Nil(t, retBody)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		mb1 := &block.MiniBlock{
			TxHashes: [][]byte{[]byte("txHash1")},
		}
		mb2 := &block.MiniBlock{
			TxHashes: [][]byte{[]byte("txHash2")},
		}
		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				mb1,
				mb2,
			},
		}

		mbh1 := block.MiniBlockHeader{
			SenderShardID: 1,
		}
		mbhReserved1 := block.MiniBlockHeaderReserved{ExecutionType: block.Normal}
		mbh1.Reserved, _ = mbhReserved1.Marshal()

		mbh2 := block.MiniBlockHeader{
			SenderShardID: 1,
		}
		mbhReserved2 := block.MiniBlockHeaderReserved{ExecutionType: block.Scheduled}
		mbh2.Reserved, _ = mbhReserved2.Marshal()

		header := &block.Header{
			ShardID: 1,
			MiniBlockHeaders: []block.MiniBlockHeader{
				mbh1,
				mbh2,
			},
		}

		retBody, err := blproc.GetScheduledMiniBlocksFromMe(header, body)
		assert.Nil(t, err)
		assert.Equal(t, block.MiniBlockSlice{mb2}, retBody)
	})
}

func TestBaseProcessor_checkScheduledMiniBlockValidity(t *testing.T) {
	t.Parallel()

	hash1 := []byte("Hash1")

	t.Run("scheduledMiniBlocks flag not set", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		err := bp.CheckScheduledMiniBlocksValidity(&block.MetaBlock{})
		assert.Nil(t, err)
	})

	t.Run("fail to calculate hash", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		expectedErr := errors.New("expected error")
		coreComponents.IntMarsh = &marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			GetScheduledMiniBlocksCalled: func() block.MiniBlockSlice {
				return block.MiniBlockSlice{&block.MiniBlock{
					TxHashes: [][]byte{hash1},
				}}
			},
		}

		bp, _ := blproc.NewShardProcessor(arguments)

		header := &block.Header{
			MiniBlockHeaders: []block.MiniBlockHeader{
				{Hash: []byte("differentHash")},
			},
		}

		err := bp.CheckScheduledMiniBlocksValidity(header)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("scheduled miniblocks mismatch", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		coreComponents.Hash = &mock.HasherStub{
			ComputeCalled: func(s string) []byte {
				return hash1
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			GetScheduledMiniBlocksCalled: func() block.MiniBlockSlice {
				return block.MiniBlockSlice{&block.MiniBlock{
					TxHashes: [][]byte{hash1},
				}}
			},
		}

		bp, _ := blproc.NewShardProcessor(arguments)

		header := &block.Header{
			MiniBlockHeaders: []block.MiniBlockHeader{
				{Hash: []byte("differentHash")},
			},
		}

		err := bp.CheckScheduledMiniBlocksValidity(header)
		assert.Equal(t, process.ErrScheduledMiniBlocksMismatch, err)
	})

	t.Run("num header miniblocks lower than scheduled miniblocks, should fail", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			GetScheduledMiniBlocksCalled: func() block.MiniBlockSlice {
				return block.MiniBlockSlice{
					&block.MiniBlock{
						TxHashes: [][]byte{hash1},
					},
					&block.MiniBlock{
						TxHashes: [][]byte{[]byte("hash2")},
					},
				}
			},
		}

		bp, _ := blproc.NewShardProcessor(arguments)

		header := &block.Header{
			MiniBlockHeaders: []block.MiniBlockHeader{
				{Hash: hash1},
			},
		}

		err := bp.CheckScheduledMiniBlocksValidity(header)
		assert.Equal(t, process.ErrScheduledMiniBlocksMismatch, err)
	})

	t.Run("same hash, should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.Hash = &mock.HasherStub{
			ComputeCalled: func(s string) []byte {
				return hash1
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			GetScheduledMiniBlocksCalled: func() block.MiniBlockSlice {
				return block.MiniBlockSlice{
					&block.MiniBlock{
						TxHashes: [][]byte{hash1},
					},
				}
			},
		}

		bp, _ := blproc.NewShardProcessor(arguments)

		header := &block.Header{
			MiniBlockHeaders: []block.MiniBlockHeader{
				{Hash: hash1},
			},
		}

		err := bp.CheckScheduledMiniBlocksValidity(header)
		assert.Nil(t, err)
	})
}

func TestBaseProcessor_setMiniBlockHeaderReservedField(t *testing.T) {
	t.Parallel()

	miniBlockHash := []byte("miniBlockHash")

	t.Run("scheduledMiniBlocks flag not set", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		err := bp.SetMiniBlockHeaderReservedField(&block.MiniBlock{}, &block.MiniBlockHeader{Hash: []byte{}}, make(map[string]*processedMb.ProcessedMiniBlockInfo))
		assert.Nil(t, err)
	})

	t.Run("no scheduled miniBlock, miniBlock Not executed", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			IsScheduledTxCalled: func(hash []byte) bool {
				return false
			},
			IsMiniBlockExecutedCalled: func(hash []byte) bool {
				assert.Equal(t, miniBlockHash, hash)
				return false
			},
		}
		bp, _ := blproc.NewShardProcessor(arguments)

		mbHandler := &block.MiniBlockHeader{
			Hash: miniBlockHash,
		}

		err := bp.SetMiniBlockHeaderReservedField(&block.MiniBlock{}, mbHandler, make(map[string]*processedMb.ProcessedMiniBlockInfo))
		assert.Nil(t, err)
		assert.Equal(t, int32(block.Normal), mbHandler.GetProcessingType())
		assert.Equal(t, int32(block.Final), mbHandler.GetConstructionState())
	})

	t.Run("no scheduled miniBlock, miniBlock executed", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			IsScheduledTxCalled: func(hash []byte) bool {
				return false
			},
			IsMiniBlockExecutedCalled: func(hash []byte) bool {
				assert.Equal(t, miniBlockHash, hash)
				return true
			},
		}
		bp, _ := blproc.NewShardProcessor(arguments)

		mbHandler := &block.MiniBlockHeader{
			Hash: miniBlockHash,
		}

		err := bp.SetMiniBlockHeaderReservedField(&block.MiniBlock{}, mbHandler, make(map[string]*processedMb.ProcessedMiniBlockInfo))
		assert.Nil(t, err)
		assert.Equal(t, int32(block.Processed), mbHandler.GetProcessingType())
		assert.Equal(t, int32(block.Final), mbHandler.GetConstructionState())
	})

	t.Run("is scheduled miniBlock, different shardId", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		bootstrapComponents.Coordinator = &testscommon.ShardsCoordinatorMock{
			SelfIDCalled: func() uint32 {
				return 1
			},
		}

		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			IsScheduledTxCalled: func(hash []byte) bool {
				return true
			},
		}
		bp, _ := blproc.NewShardProcessor(arguments)

		mb := &block.MiniBlock{
			TxHashes: [][]byte{[]byte("hash")},
		}

		mbHandler := &block.MiniBlockHeader{
			Hash:          miniBlockHash,
			SenderShardID: 2,
		}

		err := bp.SetMiniBlockHeaderReservedField(mb, mbHandler, make(map[string]*processedMb.ProcessedMiniBlockInfo))
		assert.Nil(t, err)
		assert.Equal(t, int32(block.Scheduled), mbHandler.GetProcessingType())
		assert.Equal(t, int32(block.Final), mbHandler.GetConstructionState())
	})

	t.Run("is scheduled miniBlock, same shardId", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		shardId := uint32(1)
		bootstrapComponents.Coordinator = &testscommon.ShardsCoordinatorMock{
			SelfIDCalled: func() uint32 {
				return shardId
			},
		}

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			IsScheduledTxCalled: func(hash []byte) bool {
				return true
			},
		}
		bp, _ := blproc.NewShardProcessor(arguments)

		mb := &block.MiniBlock{
			TxHashes: [][]byte{[]byte("hash")},
		}

		mbHandler := &block.MiniBlockHeader{
			Hash:          miniBlockHash,
			SenderShardID: shardId,
		}

		err := bp.SetMiniBlockHeaderReservedField(mb, mbHandler, make(map[string]*processedMb.ProcessedMiniBlockInfo))
		assert.Nil(t, err)
		assert.Equal(t, int32(block.Scheduled), mbHandler.GetProcessingType())
		assert.Equal(t, int32(block.Proposed), mbHandler.GetConstructionState())
	})
}

func TestMetaProcessor_RestoreBlockBodyIntoPoolsShouldErrNilBlockBody(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
	dataComponents.Storage = initStore()
	arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	mp, _ := blproc.NewMetaProcessor(arguments)

	err := mp.RestoreBlockBodyIntoPools(nil)
	assert.Equal(t, err, process.ErrNilBlockBody)
}

func TestMetaProcessor_RestoreBlockBodyIntoPoolsShouldErrWhenRestoreBlockDataFromStorageFails(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("error")

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
	dataComponents.Storage = initStore()
	arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
		RestoreBlockDataFromStorageCalled: func(body *block.Body) (int, error) {
			return 0, expectedError
		},
	}
	mp, _ := blproc.NewMetaProcessor(arguments)

	err := mp.RestoreBlockBodyIntoPools(&block.Body{})
	assert.Equal(t, err, expectedError)
}

func TestMetaProcessor_RestoreBlockBodyIntoPoolsShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
	dataComponents.Storage = initStore()
	arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
		RestoreBlockDataFromStorageCalled: func(body *block.Body) (int, error) {
			return 1, nil
		},
	}
	mp, _ := blproc.NewMetaProcessor(arguments)

	err := mp.RestoreBlockBodyIntoPools(&block.Body{})
	assert.Nil(t, err)
}

func TestBaseProcessor_getPruningHandler(t *testing.T) {
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.Config = config.Config{}
	arguments.StatusCoreComponents = &mock.StatusCoreComponentsStub{
		AppStatusHandlerField: &statusHandlerMock.AppStatusHandlerStub{},
	}
	bp, _ := blproc.NewShardProcessor(arguments)

	bp.SetLastRestartNonce(1)
	ph := bp.GetPruningHandler(10)
	assert.False(t, ph.IsPruningEnabled())

	bp.SetLastRestartNonce(1)
	ph = bp.GetPruningHandler(11)
	assert.False(t, ph.IsPruningEnabled())

	bp.SetLastRestartNonce(1)
	ph = bp.GetPruningHandler(14)
	assert.True(t, ph.IsPruningEnabled())
}

func TestBaseProcessor_getPruningHandlerSetsDefaulPruningDelay(t *testing.T) {
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.Config = config.Config{}
	bp, _ := blproc.NewShardProcessor(arguments)

	bp.SetLastRestartNonce(0)
	ph := bp.GetPruningHandler(9)
	assert.False(t, ph.IsPruningEnabled())
}

func TestBaseProcessor_checkConstructionStateAndIndexesCorrectness(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments(createComponentHolderMocks())
	bp, _ := blproc.NewShardProcessor(arguments)

	mbh := &block.MiniBlockHeader{
		TxCount: 5,
	}

	_ = mbh.SetConstructionState(int32(block.PartialExecuted))

	_ = mbh.SetIndexOfLastTxProcessed(int32(mbh.TxCount))
	err := bp.CheckConstructionStateAndIndexesCorrectness(mbh)
	assert.Nil(t, err)

	_ = mbh.SetIndexOfLastTxProcessed(int32(mbh.TxCount) - 2)
	err = bp.CheckConstructionStateAndIndexesCorrectness(mbh)
	assert.Nil(t, err)

	_ = mbh.SetIndexOfLastTxProcessed(int32(mbh.TxCount) - 1)
	err = bp.CheckConstructionStateAndIndexesCorrectness(mbh)
	assert.Equal(t, process.ErrIndexDoesNotMatchWithPartialExecutedMiniBlock, err)

	_ = mbh.SetConstructionState(int32(block.Final))

	_ = mbh.SetIndexOfLastTxProcessed(int32(mbh.TxCount))
	err = bp.CheckConstructionStateAndIndexesCorrectness(mbh)
	assert.Equal(t, process.ErrIndexDoesNotMatchWithFullyExecutedMiniBlock, err)

	_ = mbh.SetIndexOfLastTxProcessed(int32(mbh.TxCount) - 2)
	err = bp.CheckConstructionStateAndIndexesCorrectness(mbh)
	assert.Equal(t, process.ErrIndexDoesNotMatchWithFullyExecutedMiniBlock, err)

	_ = mbh.SetIndexOfLastTxProcessed(int32(mbh.TxCount) - 1)
	err = bp.CheckConstructionStateAndIndexesCorrectness(mbh)
	assert.Nil(t, err)
}

func TestBaseProcessor_ConcurrentCallsNonceOfFirstCommittedBlock(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments(createComponentHolderMocks())
	bp, _ := blproc.NewShardProcessor(arguments)

	numCalls := 1000
	wg := &sync.WaitGroup{}
	wg.Add(numCalls)

	mutValuesRead := sync.Mutex{}
	values := make(map[uint64]int)
	noValues := 0
	lastValRead := uint64(0)

	for i := 0; i < numCalls; i++ {
		go func(idx int) {
			time.Sleep(time.Millisecond * 10)

			switch idx % 2 {
			case 0:
				val := bp.NonceOfFirstCommittedBlock()

				mutValuesRead.Lock()
				if val.HasValue {
					values[val.Value]++
					lastValRead = val.Value
				} else {
					noValues++
				}
				mutValuesRead.Unlock()
			case 1:
				bp.SetNonceOfFirstCommittedBlock(uint64(idx))
			}

			wg.Done()
		}(i)
	}

	wg.Wait()

	mutValuesRead.Lock()
	defer mutValuesRead.Unlock()

	assert.True(t, len(values) <= 1) // we can have the situation when all reads are done before the first set
	assert.Equal(t, numCalls/2, values[lastValRead]+noValues)
}

func TestBaseProcessor_CheckSentSignaturesAtCommitTime(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("nodes coordinator errors, should return error", func(t *testing.T) {
		nodesCoordinatorInstance := shardingMocks.NewNodesCoordinatorMock()
		nodesCoordinatorInstance.ComputeValidatorsGroupCalled = func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			return nil, expectedErr
		}

		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.SentSignaturesTracker = &testscommon.SentSignatureTrackerStub{
			ResetCountersForManagedBlockSignerCalled: func(signerPk []byte) {
				assert.Fail(t, "should have not called ResetCountersManagedBlockSigners")
			},
		}
		arguments.NodesCoordinator = nodesCoordinatorInstance
		bp, _ := blproc.NewShardProcessor(arguments)

		err := bp.CheckSentSignaturesAtCommitTime(&block.Header{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work with bitmap", func(t *testing.T) {
		validator0, _ := nodesCoordinator.NewValidator([]byte("pk0"), 0, 0)
		validator1, _ := nodesCoordinator.NewValidator([]byte("pk1"), 1, 1)
		validator2, _ := nodesCoordinator.NewValidator([]byte("pk2"), 2, 2)

		nodesCoordinatorInstance := shardingMocks.NewNodesCoordinatorMock()
		nodesCoordinatorInstance.ComputeValidatorsGroupCalled = func(randomness []byte, round uint64, shardId uint32, epoch uint32) (validatorsGroup []nodesCoordinator.Validator, err error) {
			return []nodesCoordinator.Validator{validator0, validator1, validator2}, nil
		}

		resetCountersCalled := make([][]byte, 0)
		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.SentSignaturesTracker = &testscommon.SentSignatureTrackerStub{
			ResetCountersForManagedBlockSignerCalled: func(signerPk []byte) {
				resetCountersCalled = append(resetCountersCalled, signerPk)
			},
		}
		arguments.NodesCoordinator = nodesCoordinatorInstance
		bp, _ := blproc.NewShardProcessor(arguments)

		err := bp.CheckSentSignaturesAtCommitTime(&block.Header{
			PubKeysBitmap: []byte{0b00000101},
		})
		assert.Nil(t, err)

		assert.Equal(t, [][]byte{validator0.PubKey(), validator2.PubKey()}, resetCountersCalled)
	})
}
