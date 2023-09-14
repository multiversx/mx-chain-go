package processorV2

import (
	"encoding/hex"
	"math/big"
	"math/rand"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager"
	"github.com/multiversx/mx-chain-go/state/storagePruningManager/evictionWaitingList"
	"github.com/multiversx/mx-chain-go/state/syncer"
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
	mockVm "github.com/multiversx/mx-chain-vm-common-go/mock"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	"github.com/stretchr/testify/require"
)

var (
	esdtKeyPrefix = []byte(core.ProtectedKeyPrefix + core.ESDTKeyIdentifier)
	sovHasher     = sha256.NewSha256()
	sovMarshaller = &marshal.GogoProtoMarshalizer{}
	sovPubKeyConv = createMockPubkeyConverter()
	sovShardCoord = mock.NewMultiShardsCoordinatorMock(1)

	sovEnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsSaveToSystemAccountFlagEnabledField:  true,
		IsOptimizeNFTStoreFlagEnabledField:     true,
		IsSendAlwaysFlagEnabledField:           true,
		IsESDTNFTImprovementV1FlagEnabledField: true,
	}
)

func createSovereignSmartContractProcessorArguments() scrCommon.ArgsNewSmartContractProcessor {
	accAdapter := createAccountsDB()
	esdtStorage := createESDTDataStorage(accAdapter)
	builtInFuncContainer := createBuiltInFuncContainer(accAdapter, esdtStorage)
	sovBlock := createSovBlockchainHook(accAdapter, builtInFuncContainer, esdtStorage)
	txTypeHandler := createTxTypeHandler(builtInFuncContainer)

	return scrCommon.ArgsNewSmartContractProcessor{
		VmContainer:         &mock.VMContainerMock{},
		ArgsParser:          smartContract.NewArgumentParser(),
		Hasher:              &hashingMocks.HasherMock{},
		Marshalizer:         &marshal.GogoProtoMarshalizer{},
		AccountsDB:          accAdapter,
		BlockChainHook:      sovBlock,
		BuiltInFunctions:    builtInFuncContainer,
		PubkeyConv:          sovPubKeyConv,
		ShardCoordinator:    sovShardCoord,
		ScrForwarder:        &mock.IntermediateTransactionHandlerMock{},
		BadTxForwarder:      &mock.IntermediateTransactionHandlerMock{},
		TxFeeHandler:        &mock.FeeAccumulatorStub{},
		TxLogsProcessor:     &mock.TxLogsProcessorStub{},
		EconomicsFee:        &economicsmocks.EconomicsHandlerStub{},
		TxTypeHandler:       txTypeHandler,
		GasHandler:          &testscommon.GasHandlerStub{},
		EnableEpochsHandler: sovEnableEpochsHandler,
		GasSchedule:         testscommon.NewGasScheduleNotifierMock(make(map[string]map[string]uint64)),
		WasmVMChangeLocker:  &sync.RWMutex{},
		VMOutputCacher:      txcache.NewDisabledCache(),
	}
}

func createAccountsDB() *state.AccountsDB {
	argsAccCreator := factory.ArgsAccountCreator{
		Hasher:              sovHasher,
		Marshaller:          sovMarshaller,
		EnableEpochsHandler: sovEnableEpochsHandler,
	}
	accCreator, _ := factory.NewAccountCreator(argsAccCreator)

	storageManagerArgs := storageMock.GetStorageManagerArgs()
	storageManagerArgs.Marshalizer = sovMarshaller
	storageManagerArgs.Hasher = sovHasher

	trieFactoryManager, _ := trie.CreateTrieStorageManager(storageManagerArgs, storageMock.GetStorageManagerOptions())
	tr, _ := trie.NewTrie(trieFactoryManager, sovMarshaller, sovHasher, sovEnableEpochsHandler, 5)
	ewlArgs := evictionWaitingList.MemoryEvictionWaitingListArgs{
		RootHashesSize: 100,
		HashesSize:     10000,
	}
	ewl, _ := evictionWaitingList.NewMemoryEvictionWaitingList(ewlArgs)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 10)

	args := state.ArgsAccountsDB{
		Trie:                  tr,
		Hasher:                sovHasher,
		Marshaller:            sovMarshaller,
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

func createESDTDataStorage(accountsDB state.AccountsAdapter) vmcommon.ESDTNFTStorageHandler {
	esdtDataStorage, _ := builtInFunctions.NewESDTDataStorage(builtInFunctions.ArgsNewESDTDataStorage{
		Accounts:              accountsDB,
		GlobalSettingsHandler: &testscommon.ESDTGlobalSettingsHandlerStub{},
		Marshalizer:           sovMarshaller,
		EnableEpochsHandler:   sovEnableEpochsHandler,
		ShardCoordinator:      sovShardCoord,
	})

	return esdtDataStorage
}

func createBuiltInFuncContainer(
	accountsDB state.AccountsAdapter,
	esdtDataStorage vmcommon.ESDTNFTStorageHandler,
) vmcommon.BuiltInFunctionContainer {
	builtInFuncContainer := builtInFunctions.NewBuiltInFunctionContainer()
	esdtMultiTransfer, _ := builtInFunctions.NewESDTNFTMultiTransferFunc(
		10,
		sovMarshaller,
		&mockVm.GlobalSettingsHandlerStub{},
		accountsDB,
		sovShardCoord,
		vmcommon.BaseOperationCost{},
		sovEnableEpochsHandler,
		&mockVm.ESDTRoleHandlerStub{},
		esdtDataStorage,
	)

	_ = esdtMultiTransfer.SetPayableChecker(&mockVm.PayableHandlerStub{})
	_ = builtInFuncContainer.Add(core.BuiltInFunctionMultiESDTNFTTransfer, esdtMultiTransfer)

	return builtInFuncContainer
}

func createSovBlockchainHook(
	accountsDB state.AccountsAdapter,
	builtInFuncContainer vmcommon.BuiltInFunctionContainer,
	esdtStorage vmcommon.ESDTNFTStorageHandler,
) process.BlockChainHookHandler {
	args := hooks.ArgBlockChainHook{
		Accounts:                 accountsDB,
		PubkeyConv:               sovPubKeyConv,
		StorageService:           genericMocks.NewChainStorerMock(0),
		DataPool:                 dataRetriever.NewPoolsHolderMock(),
		BlockChain:               &testscommon.ChainHandlerMock{},
		ShardCoordinator:         sovShardCoord,
		Marshalizer:              sovMarshaller,
		Uint64Converter:          uint64ByteSlice.NewBigEndianConverter(),
		BuiltInFunctions:         builtInFuncContainer,
		NFTStorageHandler:        esdtStorage,
		GlobalSettingsHandler:    &testscommon.ESDTGlobalSettingsHandlerStub{},
		CompiledSCPool:           &testscommon.TxCacherStub{},
		ConfigSCStorage:          config.StorageConfig{},
		EnableEpochs:             config.EnableEpochs{},
		EpochNotifier:            forking.NewGenericEpochNotifier(),
		EnableEpochsHandler:      sovEnableEpochsHandler,
		WorkingDir:               "",
		NilCompiledSCStore:       true,
		GasSchedule:              testscommon.NewGasScheduleNotifierMock(make(map[string]map[string]uint64)),
		Counter:                  &testscommon.BlockChainHookCounterStub{},
		MissingTrieNodesNotifier: syncer.NewMissingTrieNodesNotifier(),
	}

	blockChainHook, _ := hooks.NewBlockChainHookImpl(args)
	sovBlockchainHook, _ := hooks.NewSovereignBlockChainHook(blockChainHook)
	return sovBlockchainHook
}

func createTxTypeHandler(builtInFuncContainer vmcommon.BuiltInFunctionContainer) process.TxTypeHandler {
	esdtParser, _ := parsers.NewESDTTransferParser(sovMarshaller)
	txTypeHandler, _ := coordinator.NewTxTypeHandler(coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:     sovPubKeyConv,
		ShardCoordinator:    sovShardCoord,
		BuiltInFunctions:    builtInFuncContainer,
		ArgumentParser:      parsers.NewCallArgsParser(),
		ESDTTransferParser:  esdtParser,
		EnableEpochsHandler: sovEnableEpochsHandler,
	})

	return txTypeHandler
}

func requireTokenExists(
	t *testing.T,
	account vmcommon.AccountHandler,
	tokenName []byte,
	nonce uint64,
	expectedValue *big.Int,
) {
	tokenId := append(esdtKeyPrefix, tokenName...)
	esdtNFTTokenKey := computeESDTNFTTokenKey(tokenId, nonce)
	esdtData := &esdt.ESDigitalToken{}
	marshaledData, _, err := account.(vmcommon.UserAccountHandler).AccountDataHandler().RetrieveValue(esdtNFTTokenKey)
	require.Nil(t, err)

	err = sovMarshaller.Unmarshal(esdtData, marshaledData)
	require.Nil(t, err)
	require.Equal(t, expectedValue, esdtData.Value)
}

func computeESDTNFTTokenKey(esdtTokenKey []byte, nonce uint64) []byte {
	return append(esdtTokenKey, big.NewInt(0).SetUint64(nonce).Bytes()...)
}

func createNFTMetaData(value *big.Int, nonce uint64, creator []byte) []byte {
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

	nftMetaData, _ := sovMarshaller.Marshal(esdtData)
	return nftMetaData
}

func generateRandomBytes(len int) []byte {
	randomBytes := make([]byte, len)
	_, _ = rand.Read(randomBytes)
	return randomBytes
}

func TestNewSovereignSCRProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil ArgsParser should err", func(t *testing.T) {
		args := createSSCProcessArgs()
		args.ArgsParser = nil

		sovProc, err := NewSovereignSCRProcessor(args)
		require.Nil(t, sovProc)
		require.Equal(t, process.ErrNilArgumentParser, err)
	})
	t.Run("nil TxTypeHandler should err", func(t *testing.T) {
		args := createSSCProcessArgs()
		args.TxTypeHandler = nil

		sovProc, err := NewSovereignSCRProcessor(args)
		require.Nil(t, sovProc)
		require.Equal(t, process.ErrNilTxTypeHandler, err)
	})
	t.Run("nil SmartContractProcessor should err", func(t *testing.T) {
		args := createSSCProcessArgs()
		args.SmartContractProcessor = nil

		sovProc, err := NewSovereignSCRProcessor(args)
		require.Nil(t, sovProc)
		require.Equal(t, process.ErrNilSmartContractResultProcessor, err)
	})
	t.Run("nil SCProcessHelperHandler should err", func(t *testing.T) {
		args := createSSCProcessArgs()
		args.SCProcessHelperHandler = nil

		sovProc, err := NewSovereignSCRProcessor(args)
		require.Nil(t, sovProc)
		require.Equal(t, process.ErrNilSCProcessHelper, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createSSCProcessArgs()

		sovProc, err := NewSovereignSCRProcessor(args)
		require.NotNil(t, sovProc)
		require.Nil(t, err)
	})
}

func TestSovereignSCProcessor_ProcessSmartContractResultIncomingSCR(t *testing.T) {
	t.Parallel()

	arguments := createSovereignSmartContractProcessorArguments()
	sc, _ := NewSmartContractProcessorV2(arguments)
	scpHelper, _ := scrCommon.NewSCProcessHelper(scrCommon.SCProcessHelperArgs{
		Accounts:         arguments.AccountsDB,
		ShardCoordinator: arguments.ShardCoordinator,
		Marshalizer:      arguments.Marshalizer,
		Hasher:           arguments.Hasher,
		PubkeyConverter:  arguments.PubkeyConv,
	})
	sovProc, _ := NewSovereignSCRProcessor(SovereignSCProcessArgs{
		ArgsParser:             arguments.ArgsParser,
		TxTypeHandler:          arguments.TxTypeHandler,
		SmartContractProcessor: sc,
		SCProcessHelperHandler: scpHelper,
	})

	scAddress := generateRandomBytes(32)

	token1 := []byte("token1")
	nftTransferNonce := big.NewInt(4)
	nftTransferValue := big.NewInt(100)
	nftMetaData := createNFTMetaData(nftTransferValue, nftTransferNonce.Uint64(), scAddress)
	transferNFT :=
		hex.EncodeToString(token1) + "@" + // id
			hex.EncodeToString(nftTransferNonce.Bytes()) + "@" + // nonce != 0
			hex.EncodeToString(nftMetaData) + "@" // meta data

	token2 := []byte("token2")
	esdtTransferNonce := big.NewInt(0)
	esdtTransferVal := big.NewInt(50)
	transferESDT :=
		hex.EncodeToString(token2) + "@" + // id
			hex.EncodeToString(esdtTransferNonce.Bytes()) + "@" + // nonce = 0
			hex.EncodeToString(esdtTransferVal.Bytes()) // value

	scr := smartContractResult.SmartContractResult{
		SndAddr: core.ESDTSCAddress,
		RcvAddr: scAddress,
		Data:    []byte("MultiESDTNFTTransfer@02@" + transferNFT + transferESDT),
		Value:   big.NewInt(0),
	}
	_, err := sovProc.ProcessSmartContractResult(&scr)
	require.Nil(t, err)

	acc, err := arguments.AccountsDB.LoadAccount(scAddress)
	require.Nil(t, err)
	requireTokenExists(t, acc, token1, nftTransferNonce.Uint64(), nftTransferValue)
	requireTokenExists(t, acc, token2, esdtTransferNonce.Uint64(), esdtTransferVal)
}

func createSSCProcessArgs() SovereignSCProcessArgs {
	arguments := createSovereignSmartContractProcessorArguments()
	sc, _ := NewSmartContractProcessorV2(arguments)
	scpHelper, _ := scrCommon.NewSCProcessHelper(scrCommon.SCProcessHelperArgs{
		Accounts:         arguments.AccountsDB,
		ShardCoordinator: arguments.ShardCoordinator,
		Marshalizer:      arguments.Marshalizer,
		Hasher:           arguments.Hasher,
		PubkeyConverter:  arguments.PubkeyConv,
	})
	return SovereignSCProcessArgs{
		ArgsParser:             arguments.ArgsParser,
		TxTypeHandler:          arguments.TxTypeHandler,
		SmartContractProcessor: sc,
		SCProcessHelperHandler: scpHelper,
	}
}
