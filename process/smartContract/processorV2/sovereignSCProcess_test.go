package processorV2

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/evictionWaitingList"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageMock "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/trie"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	mock2 "github.com/multiversx/mx-chain-vm-common-go/mock"
	"github.com/stretchr/testify/require"
)

func createAccountsDB(
	hasher hashing.Hasher,
	marshaller marshal.Marshalizer,
	enableEpochsHandler common.EnableEpochsHandler,
) *state.AccountsDB {
	argsAccCreator := state.ArgsAccountCreation{
		Hasher:              hasher,
		Marshaller:          marshaller,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	accCreator, _ := factory.NewAccountCreator(argsAccCreator)

	storageManagerArgs := storageMock.GetStorageManagerArgs()
	storageManagerArgs.Marshalizer = marshaller
	storageManagerArgs.Hasher = hasher
	storageManagerArgs.MainStorer = createMemUnit()
	storageManagerArgs.CheckpointsStorer = createMemUnit()

	trieFactoryManager, _ := trie.CreateTrieStorageManager(storageManagerArgs, storageMock.GetStorageManagerOptions())

	tr, _ := trie.NewTrie(trieFactoryManager, marshaller, hasher, enableEpochsHandler, 5)
	ewlArgs := evictionWaitingList.MemoryEvictionWaitingListArgs{
		RootHashesSize: 100,
		HashesSize:     10000,
	}
	ewl, _ := evictionWaitingList.NewMemoryEvictionWaitingList(ewlArgs)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 10)

	args := state.ArgsAccountsDB{
		Trie:                  tr,
		Hasher:                hasher,
		Marshaller:            marshaller,
		AccountFactory:        accCreator,
		StoragePruningManager: spm,
		ProcessingMode:        common.Normal,
		ProcessStatusHandler:  &testscommon.ProcessStatusHandlerStub{},
		AppStatusHandler:      &statusHandlerMock.AppStatusHandlerStub{},
		AddressConverter:      &testscommon.PubkeyConverterMock{},
	}
	adb, _ := state.NewAccountsDB(args)
	return adb
}

func createMemUnit() storage.Storer {
	capacity := uint32(10)
	shards := uint32(1)
	sizeInBytes := uint64(0)
	cache, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: capacity, Shards: shards, SizeInBytes: sizeInBytes})
	persist, _ := database.NewlruDB(100000)
	unit, _ := storageunit.NewStorageUnit(cache, persist)

	return unit
}

func createESDTDataStorage(marshaller marshal.Marshalizer, accountsDB state.AccountsAdapter) vmcommon.ESDTNFTStorageHandler {
	esdtDataStorage, _ := builtInFunctions.NewESDTDataStorage(builtInFunctions.ArgsNewESDTDataStorage{
		Accounts:              accountsDB,
		GlobalSettingsHandler: &testscommon.ESDTGlobalSettingsHandlerStub{},
		Marshalizer:           marshaller,
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsSaveToSystemAccountFlagEnabledField:  true,
			IsOptimizeNFTStoreFlagEnabledField:     true,
			IsSendAlwaysFlagEnabledField:           true,
			IsESDTNFTImprovementV1FlagEnabledField: true,
		},
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(1),
	})

	return esdtDataStorage
}

func createBuiltInFuncContainer(marshaller marshal.Marshalizer, accountsDB state.AccountsAdapter, esdtDataStorage vmcommon.ESDTNFTStorageHandler) vmcommon.BuiltInFunctionContainer {
	builtInFuncContainer := builtInFunctions.NewBuiltInFunctionContainer()
	esdtMultiTransfer, _ := builtInFunctions.NewESDTNFTMultiTransferFunc(
		10,
		marshaller,
		&mock2.GlobalSettingsHandlerStub{},
		accountsDB,
		mock.NewMultiShardsCoordinatorMock(1),
		vmcommon.BaseOperationCost{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		&mock2.ESDTRoleHandlerStub{},
		esdtDataStorage,
	)

	_ = esdtMultiTransfer.SetPayableChecker(&mock2.PayableHandlerStub{})
	_ = builtInFuncContainer.Add(core.BuiltInFunctionMultiESDTNFTTransfer, esdtMultiTransfer)

	return builtInFuncContainer
}

func createSovBlockchainHook(
	marshaller marshal.Marshalizer,
	accountsDB state.AccountsAdapter,
) process.BlockChainHookHandler {
	esdtStorage := createESDTDataStorage(marshaller, accountsDB)
	builtInFuncContainer := createBuiltInFuncContainer(marshaller, accountsDB, esdtStorage)

	args := hooks.ArgBlockChainHook{
		Accounts:              accountsDB,
		PubkeyConv:            createMockPubkeyConverter(),
		StorageService:        genericMocks.NewChainStorerMock(0),
		DataPool:              dataRetriever.NewPoolsHolderMock(),
		BlockChain:            &testscommon.ChainHandlerMock{},
		ShardCoordinator:      mock.NewMultiShardsCoordinatorMock(1),
		Marshalizer:           marshaller,
		Uint64Converter:       uint64ByteSlice.NewBigEndianConverter(),
		BuiltInFunctions:      builtInFuncContainer,
		NFTStorageHandler:     esdtStorage,
		GlobalSettingsHandler: &testscommon.ESDTGlobalSettingsHandlerStub{},
		CompiledSCPool:        &testscommon.TxCacherStub{},
		ConfigSCStorage:       config.StorageConfig{},
		EnableEpochs:          config.EnableEpochs{},
		EpochNotifier:         forking.NewGenericEpochNotifier(),
		EnableEpochsHandler:   &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		WorkingDir:            "",
		NilCompiledSCStore:    true,
		GasSchedule: &testscommon.GasScheduleNotifierMock{
			LatestGasScheduleCalled: func() map[string]map[string]uint64 {
				return map[string]map[string]uint64{}
			},
		},
		Counter:                  &testscommon.BlockChainHookCounterStub{},
		MissingTrieNodesNotifier: syncer.NewMissingTrieNodesNotifier(),
	}

	blockChainHook, _ := hooks.NewBlockChainHookImpl(args)
	sovBlockchainHook, _ := hooks.NewSovereignBlockChainHook(blockChainHook)
	return sovBlockchainHook
}

func createSovereignSmartContractProcessorArguments() scrCommon.ArgsNewSmartContractProcessor {
	gasSchedule := make(map[string]map[string]uint64)
	gasSchedule[common.BaseOpsAPICost] = make(map[string]uint64)
	gasSchedule[common.BaseOpsAPICost][common.AsyncCallStepField] = 1000
	gasSchedule[common.BaseOpsAPICost][common.AsyncCallbackGasLockField] = 3000
	gasSchedule[common.BuiltInCost] = make(map[string]uint64)
	gasSchedule[common.BuiltInCost][core.BuiltInFunctionESDTTransfer] = 2000

	hasher := sha256.NewSha256()
	marshaller := &marshal.GogoProtoMarshalizer{}
	accAdapter := createAccountsDB(hasher, marshaller, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
	sovBlock := createSovBlockchainHook(marshaller, accAdapter)

	return scrCommon.ArgsNewSmartContractProcessor{
		VmContainer:      &mock.VMContainerMock{},
		ArgsParser:       smartContract.NewArgumentParser(),
		Hasher:           &hashingMocks.HasherMock{},
		Marshalizer:      &marshal.GogoProtoMarshalizer{},
		AccountsDB:       accAdapter,
		BlockChainHook:   sovBlock,
		BuiltInFunctions: builtInFunctions.NewBuiltInFunctionContainer(),
		PubkeyConv:       createMockPubkeyConverter(),
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(1),
		ScrForwarder:     &mock.IntermediateTransactionHandlerMock{},
		BadTxForwarder:   &mock.IntermediateTransactionHandlerMock{},
		TxFeeHandler:     &mock.FeeAccumulatorStub{},
		TxLogsProcessor:  &mock.TxLogsProcessorStub{},
		EconomicsFee: &economicsmocks.EconomicsHandlerStub{
			DeveloperPercentageCalled: func() float64 {
				return 0.0
			},
			ComputeTxFeeCalled: func(tx data.TransactionWithFeeHandler) *big.Int {
				return core.SafeMul(tx.GetGasLimit(), tx.GetGasPrice())
			},
			ComputeFeeForProcessingCalled: func(tx data.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
				return core.SafeMul(tx.GetGasPrice(), gasToUse)
			},
		},
		TxTypeHandler: &testscommon.TxTypeHandlerMock{},
		GasHandler:    &testscommon.GasHandlerStub{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsSCDeployFlagEnabledField: true,
		},
		GasSchedule:        testscommon.NewGasScheduleNotifierMock(gasSchedule),
		WasmVMChangeLocker: &sync.RWMutex{},
		VMOutputCacher:     txcache.NewDisabledCache(),
	}
}

const baseESDTKeyPrefix = core.ProtectedKeyPrefix + core.ESDTKeyIdentifier

var keyPrefix = []byte(baseESDTKeyPrefix)

func computeESDTNFTTokenKey(esdtTokenKey []byte, nonce uint64) []byte {
	return append(esdtTokenKey, big.NewInt(0).SetUint64(nonce).Bytes()...)
}

func testNFTTokenShouldExist(
	tb testing.TB,
	marshaller vmcommon.Marshalizer,
	account vmcommon.AccountHandler,
	tokenName []byte,
	nonce uint64,
	expectedValue *big.Int,
) {
	tokenId := append(keyPrefix, tokenName...)
	esdtNFTTokenKey := computeESDTNFTTokenKey(tokenId, nonce)
	esdtData := &esdt.ESDigitalToken{Value: big.NewInt(0), Type: uint32(core.Fungible)}
	marshaledData, _, _ := account.(vmcommon.UserAccountHandler).AccountDataHandler().RetrieveValue(esdtNFTTokenKey)
	_ = marshaller.Unmarshal(esdtData, marshaledData)
	require.Equal(tb, expectedValue, esdtData.Value)
}

func createNFTMetaData(marshaller marshal.Marshalizer, value *big.Int, nonce uint64, creator []byte) ([]byte, error) {
	esdtData := &esdt.ESDigitalToken{
		Type:  uint32(core.NonFungible),
		Value: value,
		TokenMetaData: &esdt.MetaData{
			URIs:       [][]byte{[]byte("uri1"), []byte("uri2"), []byte("uri3")},
			Nonce:      nonce,
			Hash:       []byte("NFT hash"),
			Name:       []byte("name nft"),
			Attributes: []byte("attributes"),
			Creator:    creator,
		},
	}

	return marshaller.Marshal(esdtData)
}

func TestSovereignSCProcessor_ProcessSmartContractResultExecuteSCIfMetaAndBuiltIn(t *testing.T) {
	t.Parallel()

	scAddress := []byte("000000000001234567890123456789012")
	dstScAddress := createAccount(scAddress)
	dstScAddress.SetCode([]byte("code"))
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if bytes.Equal(scAddress, address) {
			return shardCoordinator.SelfId()
		}
		return 0
	}
	shardCoordinator.CurrentShard = core.MetachainShardId

	arguments := createSovereignSmartContractProcessorArguments()
	arguments.ShardCoordinator = shardCoordinator
	arguments.VmContainer = &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return &mock.VMExecutionHandlerStub{
				RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (output *vmcommon.VMOutput, e error) {
					return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, nil
				},
			}, nil
		},
	}
	arguments.TxTypeHandler = &testscommon.TxTypeHandlerMock{
		ComputeTransactionTypeCalled: func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
			return process.BuiltInFunctionCall, process.BuiltInFunctionCall
		},
	}
	enableEpochsHandlerStub := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	arguments.EnableEpochsHandler = enableEpochsHandlerStub

	sc, _ := NewSmartContractProcessorV2(arguments)
	sovProc, _ := NewSovereignSCRProcessor(sc)

	metaData, _ := createNFTMetaData(arguments.Marshalizer, big.NewInt(100), 4, scAddress)
	dataScr := []byte("MultiESDTNFTTransfer@02@746f6b656e31@04@")
	dataScr = append(dataScr, []byte(hex.EncodeToString(metaData))...)
	dataScr = append(dataScr, []byte("@746f6b656e32@@32")...)

	scr := smartContractResult.SmartContractResult{
		SndAddr: core.ESDTSCAddress,
		RcvAddr: scAddress,
		Data:    dataScr,
		Value:   big.NewInt(0),
	}
	_, err := sovProc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)

	acc, _ := sovProc.accounts.LoadAccount(scAddress)

	_, _ = arguments.AccountsDB.Commit()

	testNFTTokenShouldExist(t, arguments.Marshalizer, acc, []byte("token1"), 4, big.NewInt(100))
	testNFTTokenShouldExist(t, arguments.Marshalizer, acc, []byte("token2"), 0, big.NewInt(50))
}
