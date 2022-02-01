package vm

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"sync"
	"testing"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/config"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/scheduled"
	dataTransaction "github.com/ElrondNetwork/elrond-go-core/data/transaction"
	dataTx "github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/forking"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	processDisabled "github.com/ElrondNetwork/elrond-go/genesis/process/disabled"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/postprocess"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	syncDisabled "github.com/ElrondNetwork/elrond-go/process/sync/disabled"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/process/transactionLog"
	"github.com/ElrondNetwork/elrond-go/process/txsimulator"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager"
	"github.com/ElrondNetwork/elrond-go/state/storagePruningManager/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/ElrondNetwork/elrond-go/trie/hashesHolder"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	vmcommonBuiltInFunctions "github.com/ElrondNetwork/elrond-vm-common/builtInFunctions"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dnsAddr = []byte{0, 0, 0, 0, 0, 0, 0, 0, 5, 0, 137, 17, 46, 56, 127, 47, 62, 172, 4, 126, 190, 242, 221, 230, 209, 243, 105, 104, 242, 66, 49, 49}

// TODO: Merge test utilities from this file with the ones from "arwen/utils.go"

var testMarshalizer = &marshal.GogoProtoMarshalizer{}
var testHasher = sha256.NewSha256()
var oneShardCoordinator = mock.NewMultiShardsCoordinatorMock(2)
var globalEpochNotifier = forking.NewGenericEpochNotifier()
var pubkeyConv, _ = pubkeyConverter.NewHexPubkeyConverter(32)

var log = logger.GetOrCreate("integrationtests")

const maxTrieLevelInMemory = uint(5)

func getZeroGasAndFees() scheduled.GasAndFees {
	return scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		GasProvided:     0,
		GasPenalized:    0,
		GasRefunded:     0,
	}
}

// ArgEnableEpoch will specify enable epoch values for certain flags
type ArgEnableEpoch struct {
	PenalizedTooMuchGasEnableEpoch      uint32
	BuiltinEnableEpoch                  uint32
	DeployEnableEpoch                   uint32
	MetaProtectionEnableEpoch           uint32
	RelayedTxEnableEpoch                uint32
	UnbondTokensV2EnableEpoch           uint32
	BackwardCompSaveKeyValueEnableEpoch uint32
	CleanUpInformativeSCRsEnableEpoch   uint32
	ScheduledMiniBlocksEnableEpoch      uint32
}

// VMTestAccount -
type VMTestAccount struct {
	Balance      *big.Int
	Address      []byte
	Nonce        uint64
	TokenBalance *big.Int
}

// VMTestContext -
type VMTestContext struct {
	TxProcessor      process.TransactionProcessor
	ScProcessor      *smartContract.TestScProcessor
	Accounts         state.AccountsAdapter
	BlockchainHook   vmcommon.BlockchainHook
	VMContainer      process.VirtualMachinesContainer
	TxFeeHandler     process.TransactionFeeHandler
	ShardCoordinator sharding.Coordinator
	ScForwarder      process.IntermediateTransactionHandler
	EconomicsData    process.EconomicsDataHandler
	Marshalizer      marshal.Marshalizer
	GasSchedule      core.GasScheduleNotifier
	VMConfiguration  *config.VirtualMachineConfig
	EpochNotifier    process.EpochNotifier
	SCQueryService   *smartContract.SCQueryService

	Alice         VMTestAccount
	Bob           VMTestAccount
	ContractOwner VMTestAccount
	Contract      VMTestAccount

	TxCostHandler    external.TransactionCostHandler
	TxsLogsProcessor process.TransactionLogProcessor
}

// Close -
func (vmTestContext *VMTestContext) Close() {
	_ = vmTestContext.VMContainer.Close()
}

// GetLatestError -
func (vmTestContext *VMTestContext) GetLatestError() error {
	return vmTestContext.ScProcessor.GetLatestTestError()
}

// CreateBlockStarted -
func (vmTestContext *VMTestContext) CreateBlockStarted() {
	vmTestContext.TxFeeHandler.CreateBlockStarted(getZeroGasAndFees())
	vmTestContext.ScForwarder.CreateBlockStarted()
	vmTestContext.ScProcessor.CleanGasRefunded()
}

// GetGasRemaining -
func (vmTestContext *VMTestContext) GetGasRemaining() uint64 {
	return vmTestContext.ScProcessor.GetGasRemaining()
}

// GetIntermediateTransactions -
func (vmTestContext *VMTestContext) GetIntermediateTransactions(t *testing.T) []data.TransactionHandler {
	scForwarder := vmTestContext.ScForwarder
	mockIntermediate, ok := scForwarder.(*mock.IntermediateTransactionHandlerMock)
	require.True(t, ok)

	return mockIntermediate.GetIntermediateTransactions()
}

// CreateTransaction -
func (vmTestContext *VMTestContext) CreateTransaction(
	sender *VMTestAccount,
	receiver *VMTestAccount,
	value *big.Int,
	gasprice uint64,
	gasLimit uint64,
	data []byte,
) *dataTransaction.Transaction {
	return &dataTransaction.Transaction{
		SndAddr:  sender.Address,
		RcvAddr:  receiver.Address,
		Nonce:    sender.Nonce,
		Value:    big.NewInt(0).Set(value),
		GasPrice: gasprice,
		GasLimit: gasLimit,
		Data:     data,
	}
}

// CreateTransferTokenTx -
func (vmTestContext *VMTestContext) CreateTransferTokenTx(
	sender *VMTestAccount,
	receiver *VMTestAccount,
	value *big.Int,
	functionName string,
) *dataTransaction.Transaction {
	txData := txDataBuilder.NewBuilder()
	txData.Func(functionName)
	txData.Bytes(receiver.Address)
	txData.BigInt(value)
	txData.SetLast("00" + txData.GetLast())

	return &dataTransaction.Transaction{
		Nonce:    sender.Nonce,
		Value:    big.NewInt(0),
		RcvAddr:  vmTestContext.Contract.Address,
		SndAddr:  sender.Address,
		GasPrice: 1,
		GasLimit: 7000000,
		Data:     txData.ToBytes(),
		ChainID:  integrationTests.ChainID,
	}
}

// CreateAccount -
func (vmTestContext *VMTestContext) CreateAccount(account *VMTestAccount) {
	_, _ = CreateAccount(
		vmTestContext.Accounts,
		account.Address,
		account.Nonce,
		account.Balance,
	)
}

// GetIntValueFromSCWithTransientVM -
func (vmTestContext *VMTestContext) GetIntValueFromSCWithTransientVM(funcName string, args ...[]byte) *big.Int {
	vmOutput := vmTestContext.GetVMOutputWithTransientVM(funcName, args...)

	return big.NewInt(0).SetBytes(vmOutput.ReturnData[0])
}

// GetVMOutputWithTransientVM -
func (vmTestContext *VMTestContext) GetVMOutputWithTransientVM(funcName string, args ...[]byte) *vmcommon.VMOutput {
	gasSchedule := vmTestContext.GasSchedule
	accnts := vmTestContext.Accounts
	scAddressBytes := vmTestContext.Contract.Address
	vmConfig := vmTestContext.VMConfiguration
	arwenChangeLocker := &sync.RWMutex{}
	vmContainer, blockChainHook, _ := CreateVMAndBlockchainHookAndDataPool(accnts, gasSchedule, vmConfig, oneShardCoordinator, arwenChangeLocker)
	defer func() {
		_ = vmContainer.Close()
	}()

	feeHandler := &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return uint64(math.MaxUint64)
		},
	}

	argsNewSCQueryService := smartContract.ArgsNewSCQueryService{
		VmContainer:       vmContainer,
		EconomicsFee:      feeHandler,
		BlockChainHook:    blockChainHook,
		BlockChain:        &mock.BlockChainMock{},
		ArwenChangeLocker: &sync.RWMutex{},
		Bootstrapper:      syncDisabled.NewDisabledBootstrapper(),
	}
	scQueryService, _ := smartContract.NewSCQueryService(argsNewSCQueryService)

	vmOutput, err := scQueryService.ExecuteQuery(&process.SCQuery{
		ScAddress: scAddressBytes,
		FuncName:  funcName,
		Arguments: args,
	})

	if err != nil {
		fmt.Println("ERROR at GetVmOutput()", err)
		return nil
	}

	return vmOutput
}

type accountFactory struct {
}

// CreateAccount -
func (af *accountFactory) CreateAccount(address []byte) (vmcommon.AccountHandler, error) {
	return state.NewUserAccount(address)
}

// IsInterfaceNil returns true if there is no value under the interface
func (af *accountFactory) IsInterfaceNil() bool {
	return af == nil
}

// CreateEmptyAddress -
func CreateEmptyAddress() []byte {
	buff := make([]byte, testHasher.Size())

	return buff
}

// CreateMemUnit -
func CreateMemUnit() storage.Storer {
	capacity := uint32(10)
	shards := uint32(1)
	sizeInBytes := uint64(0)
	cache, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: capacity, Shards: shards, SizeInBytes: sizeInBytes})

	unit, _ := storageUnit.NewStorageUnit(cache, memorydb.New())
	return unit
}

// CreateInMemoryShardAccountsDB -
func CreateInMemoryShardAccountsDB() *state.AccountsDB {
	marsh := &marshal.GogoProtoMarshalizer{}
	store := CreateMemUnit()
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(100, memorydb.New(), marsh)
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:      1000,
		SnapshotsBufferLen:    10,
		MaxSnapshots:          2,
		SnapshotsGoroutineNum: 1,
	}
	args := trie.NewTrieStorageManagerArgs{
		DB:                store,
		MainStorer:        CreateMemUnit(),
		CheckpointsStorer: CreateMemUnit(),
		Marshalizer:       marsh,
		Hasher:            testHasher,
		SnapshotDbConfig: config.DBConfig{
			FilePath:          "TrieStorage",
			Type:              "MemoryDB",
			BatchDelaySeconds: 30,
			MaxBatchSize:      6,
			MaxOpenFiles:      10,
		},
		GeneralConfig:          generalCfg,
		CheckpointHashesHolder: hashesHolder.NewCheckpointHashesHolder(10000000, uint64(testHasher.Size())),
		EpochNotifier:          &epochNotifier.EpochNotifierStub{},
	}
	trieStorage, _ := trie.NewTrieStorageManager(args)

	tr, _ := trie.NewTrie(trieStorage, marsh, testHasher, maxTrieLevelInMemory)
	spm, _ := storagePruningManager.NewStoragePruningManager(ewl, 10)
	adb, _ := state.NewAccountsDB(tr, testHasher, marsh, &accountFactory{}, spm, common.Normal)

	return adb
}

// CreateAccount -
func CreateAccount(accnts state.AccountsAdapter, pubKey []byte, nonce uint64, balance *big.Int) ([]byte, error) {
	account, err := accnts.LoadAccount(pubKey)
	if err != nil {
		return nil, err
	}

	account.(state.UserAccountHandler).IncreaseNonce(nonce)
	_ = account.(state.UserAccountHandler).AddToBalance(balance)

	err = accnts.SaveAccount(account)
	if err != nil {
		return nil, err
	}

	hashCreated, err := accnts.Commit()
	if err != nil {
		return nil, err
	}

	return hashCreated, nil
}

func createEconomicsData(penalizedTooMuchGasEnableEpoch uint32) (process.EconomicsDataHandler, error) {
	maxGasLimitPerBlock := strconv.FormatUint(math.MaxUint64, 10)
	minGasPrice := strconv.FormatUint(1, 10)
	minGasLimit := strconv.FormatUint(1, 10)
	testProtocolSustainabilityAddress := "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp"

	builtInCost, _ := economics.NewBuiltInFunctionsCost(&economics.ArgsBuiltInFunctionCost{
		ArgsParser:  smartContract.NewArgumentParser(),
		GasSchedule: mock.NewGasScheduleNotifierMock(defaults.FillGasMapInternal(map[string]map[string]uint64{}, 1)),
	})

	argsNewEconomicsData := economics.ArgsNewEconomicsData{
		Economics: &config.EconomicsConfig{
			GlobalSettings: config.GlobalSettings{
				GenesisTotalSupply: "2000000000000000000000",
				MinimumInflation:   0,
				YearSettings: []*config.YearSetting{
					{
						Year:             0,
						MaximumInflation: 0.01,
					},
				},
			},
			RewardsSettings: config.RewardsSettings{
				RewardsConfigByEpoch: []config.EpochRewardSettings{
					{
						LeaderPercentage:                 0.1,
						ProtocolSustainabilityPercentage: 0.1,
						DeveloperPercentage:              0.1,
						ProtocolSustainabilityAddress:    testProtocolSustainabilityAddress,
						TopUpGradientPoint:               "100000",
					},
				},
			},
			FeeSettings: config.FeeSettings{
				GasLimitSettings: []config.GasLimitSetting{
					{
						MaxGasLimitPerBlock:         maxGasLimitPerBlock,
						MaxGasLimitPerMiniBlock:     maxGasLimitPerBlock,
						MaxGasLimitPerMetaBlock:     maxGasLimitPerBlock,
						MaxGasLimitPerMetaMiniBlock: maxGasLimitPerBlock,
						MaxGasLimitPerTx:            maxGasLimitPerBlock,
						MinGasLimit:                 minGasLimit,
					},
				},
				MinGasPrice:      minGasPrice,
				GasPerDataByte:   "1",
				GasPriceModifier: 1.0,
			},
		},
		PenalizedTooMuchGasEnableEpoch: penalizedTooMuchGasEnableEpoch,
		EpochNotifier:                  &epochNotifier.EpochNotifierStub{},
		BuiltInFunctionsCostHandler:    builtInCost,
	}

	return economics.NewEconomicsData(argsNewEconomicsData)
}

// CreateTxProcessorWithOneSCExecutorMockVM -
func CreateTxProcessorWithOneSCExecutorMockVM(
	accnts state.AccountsAdapter,
	opGas uint64,
	argEnableEpoch ArgEnableEpoch,
	arwenChangeLocker common.Locker,
) (process.TransactionProcessor, error) {

	builtInFuncs := vmcommonBuiltInFunctions.NewBuiltInFunctionContainer()
	datapool := dataRetrieverMock.NewPoolsHolderMock()
	args := hooks.ArgBlockChainHook{
		Accounts:           accnts,
		PubkeyConv:         pubkeyConv,
		StorageService:     &mock.ChainStorerMock{},
		BlockChain:         &mock.BlockChainMock{},
		ShardCoordinator:   oneShardCoordinator,
		Marshalizer:        testMarshalizer,
		Uint64Converter:    &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions:   builtInFuncs,
		NFTStorageHandler:  &testscommon.SimpleNFTStorageHandlerStub{},
		DataPool:           datapool,
		CompiledSCPool:     datapool.SmartContracts(),
		NilCompiledSCStore: true,
		ConfigSCStorage:    *defaultStorageConfig(),
		EpochNotifier:      &epochNotifier.EpochNotifierStub{},
	}

	blockChainHook, _ := hooks.NewBlockChainHookImpl(args)
	vm, _ := mock.NewOneSCExecutorMockVM(blockChainHook, testHasher)
	vm.GasForOperation = opGas

	vmContainer := &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return vm, nil
		}}

	esdtTransferParser, _ := parsers.NewESDTTransferParser(testMarshalizer)
	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:    pubkeyConv,
		ShardCoordinator:   oneShardCoordinator,
		BuiltInFunctions:   builtInFuncs,
		ArgumentParser:     parsers.NewCallArgsParser(),
		EpochNotifier:      forking.NewGenericEpochNotifier(),
		ESDTTransferParser: esdtTransferParser,
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	gasSchedule := make(map[string]map[string]uint64)
	defaults.FillGasMapInternal(gasSchedule, 1)

	economicsData, err := createEconomicsData(argEnableEpoch.PenalizedTooMuchGasEnableEpoch)
	if err != nil {
		return nil, err
	}

	enableEpochs := config.EnableEpochs{
		PenalizedTooMuchGasEnableEpoch: argEnableEpoch.PenalizedTooMuchGasEnableEpoch,
		BuiltInFunctionsEnableEpoch:    argEnableEpoch.BuiltinEnableEpoch,
		SCDeployEnableEpoch:            argEnableEpoch.DeployEnableEpoch,
		// TODO: enable Scheduled in all tests
		ScheduledMiniBlocksEnableEpoch: 1000,
	}
	argsNewSCProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:      vmContainer,
		ArgsParser:       smartContract.NewArgumentParser(),
		Hasher:           testHasher,
		Marshalizer:      testMarshalizer,
		AccountsDB:       accnts,
		BlockChainHook:   blockChainHook,
		BuiltInFunctions: builtInFuncs,
		PubkeyConv:       pubkeyConv,
		ShardCoordinator: oneShardCoordinator,
		ScrForwarder:     &mock.IntermediateTransactionHandlerMock{},
		BadTxForwarder:   &mock.IntermediateTransactionHandlerMock{},
		TxFeeHandler:     &testscommon.UnsignedTxHandlerStub{},
		EconomicsFee:     economicsData,
		TxTypeHandler:    txTypeHandler,
		GasHandler: &testscommon.GasHandlerStub{
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		},
		GasSchedule:       mock.NewGasScheduleNotifierMock(gasSchedule),
		TxLogsProcessor:   &mock.TxLogsProcessorStub{},
		EpochNotifier:     forking.NewGenericEpochNotifier(),
		EnableEpochs:      enableEpochs,
		VMOutputCacher:    txcache.NewDisabledCache(),
		ArwenChangeLocker: arwenChangeLocker,
	}
	scProcessor, _ := smartContract.NewSmartContractProcessor(argsNewSCProcessor)

	argsNewTxProcessor := transaction.ArgsNewTxProcessor{
		Accounts:                       accnts,
		Hasher:                         testHasher,
		PubkeyConv:                     pubkeyConv,
		Marshalizer:                    testMarshalizer,
		SignMarshalizer:                testMarshalizer,
		ShardCoordinator:               oneShardCoordinator,
		ScProcessor:                    scProcessor,
		TxFeeHandler:                   &testscommon.UnsignedTxHandlerStub{},
		TxTypeHandler:                  txTypeHandler,
		EconomicsFee:                   economicsData,
		ReceiptForwarder:               &mock.IntermediateTransactionHandlerMock{},
		BadTxForwarder:                 &mock.IntermediateTransactionHandlerMock{},
		ArgsParser:                     smartContract.NewArgumentParser(),
		ScrForwarder:                   &mock.IntermediateTransactionHandlerMock{},
		EpochNotifier:                  forking.NewGenericEpochNotifier(),
		PenalizedTooMuchGasEnableEpoch: argEnableEpoch.PenalizedTooMuchGasEnableEpoch,
		MetaProtectionEnableEpoch:      argEnableEpoch.MetaProtectionEnableEpoch,
		RelayedTxEnableEpoch:           argEnableEpoch.RelayedTxEnableEpoch,
	}

	return transaction.NewTxProcessor(argsNewTxProcessor)
}

// CreateOneSCExecutorMockVM -
func CreateOneSCExecutorMockVM(accnts state.AccountsAdapter) vmcommon.VMExecutionHandler {
	datapool := dataRetrieverMock.NewPoolsHolderMock()
	args := hooks.ArgBlockChainHook{
		Accounts:           accnts,
		PubkeyConv:         pubkeyConv,
		StorageService:     &mock.ChainStorerMock{},
		BlockChain:         &mock.BlockChainMock{},
		ShardCoordinator:   oneShardCoordinator,
		Marshalizer:        testMarshalizer,
		Uint64Converter:    &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions:   vmcommonBuiltInFunctions.NewBuiltInFunctionContainer(),
		NFTStorageHandler:  &testscommon.SimpleNFTStorageHandlerStub{},
		DataPool:           datapool,
		CompiledSCPool:     datapool.SmartContracts(),
		NilCompiledSCStore: true,
		ConfigSCStorage:    *defaultStorageConfig(),
		EpochNotifier:      &epochNotifier.EpochNotifierStub{},
	}
	blockChainHook, _ := hooks.NewBlockChainHookImpl(args)
	vm, _ := mock.NewOneSCExecutorMockVM(blockChainHook, testHasher)

	return vm
}

// CreateVMAndBlockchainHookAndDataPool -
func CreateVMAndBlockchainHookAndDataPool(
	accnts state.AccountsAdapter,
	gasSchedule core.GasScheduleNotifier,
	vmConfig *config.VirtualMachineConfig,
	shardCoordinator sharding.Coordinator,
	arwenChangeLocker common.Locker,
) (process.VirtualMachinesContainer, *hooks.BlockChainHookImpl, dataRetriever.PoolsHolder) {
	if check.IfNil(gasSchedule) || gasSchedule.LatestGasSchedule() == nil {
		testGasSchedule := arwenConfig.MakeGasMapForTests()
		defaults.FillGasMapInternal(testGasSchedule, 1)
		gasSchedule = mock.NewGasScheduleNotifierMock(testGasSchedule)
	}

	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule: gasSchedule,
		MapDNSAddresses: map[string]struct{}{
			string(dnsAddr): {},
		},
		Marshalizer:      testMarshalizer,
		Accounts:         accnts,
		ShardCoordinator: shardCoordinator,
		EpochNotifier:    globalEpochNotifier,
	}
	builtInFuncs, nftStorageHandler, _ := builtInFunctions.CreateBuiltInFuncContainerAndNFTStorageHandler(argsBuiltIn)

	datapool := dataRetrieverMock.NewPoolsHolderMock()
	args := hooks.ArgBlockChainHook{
		Accounts:           accnts,
		PubkeyConv:         pubkeyConv,
		StorageService:     &mock.ChainStorerMock{},
		BlockChain:         &mock.BlockChainMock{},
		ShardCoordinator:   shardCoordinator,
		Marshalizer:        testMarshalizer,
		Uint64Converter:    &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions:   builtInFuncs,
		NFTStorageHandler:  nftStorageHandler,
		DataPool:           datapool,
		CompiledSCPool:     datapool.SmartContracts(),
		NilCompiledSCStore: true,
		ConfigSCStorage:    *defaultStorageConfig(),
		EpochNotifier:      &epochNotifier.EpochNotifierStub{},
	}

	esdtTransferParser, _ := parsers.NewESDTTransferParser(testMarshalizer)
	maxGasLimitPerBlock := uint64(0xFFFFFFFFFFFFFFFF)
	argsNewVMFactory := shard.ArgVMContainerFactory{
		Config:             *vmConfig,
		BlockGasLimit:      maxGasLimitPerBlock,
		GasSchedule:        gasSchedule,
		ArgBlockChainHook:  args,
		EpochNotifier:      globalEpochNotifier,
		EpochConfig:        config.EnableEpochs{},
		ArwenChangeLocker:  arwenChangeLocker,
		ESDTTransferParser: esdtTransferParser,
	}
	vmFactory, err := shard.NewVMContainerFactory(argsNewVMFactory)
	if err != nil {
		log.LogIfError(err)
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		panic(err)
	}

	blockChainHook, _ := vmFactory.BlockChainHookImpl().(*hooks.BlockChainHookImpl)
	_ = vmcommonBuiltInFunctions.SetPayableHandler(builtInFuncs, blockChainHook)

	return vmContainer, blockChainHook, datapool
}

// CreateVMAndBlockchainHookMeta -
func CreateVMAndBlockchainHookMeta(
	accnts state.AccountsAdapter,
	gasSchedule core.GasScheduleNotifier,
	shardCoordinator sharding.Coordinator,
	arg ArgEnableEpoch,
) (process.VirtualMachinesContainer, *hooks.BlockChainHookImpl) {
	if check.IfNil(gasSchedule) || gasSchedule.LatestGasSchedule() == nil {
		testGasSchedule := arwenConfig.MakeGasMapForTests()
		defaults.FillGasMapInternal(testGasSchedule, 1)
		gasSchedule = mock.NewGasScheduleNotifierMock(testGasSchedule)
	}

	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule: gasSchedule,
		MapDNSAddresses: map[string]struct{}{
			string(dnsAddr): {},
		},
		Marshalizer:      testMarshalizer,
		Accounts:         accnts,
		ShardCoordinator: shardCoordinator,
		EpochNotifier:    globalEpochNotifier,
	}
	builtInFuncs, nftStorageHandler, _ := builtInFunctions.CreateBuiltInFuncContainerAndNFTStorageHandler(argsBuiltIn)

	datapool := dataRetrieverMock.NewPoolsHolderMock()
	args := hooks.ArgBlockChainHook{
		Accounts:           accnts,
		PubkeyConv:         pubkeyConv,
		StorageService:     &mock.ChainStorerMock{},
		BlockChain:         &mock.BlockChainMock{},
		ShardCoordinator:   shardCoordinator,
		Marshalizer:        testMarshalizer,
		Uint64Converter:    &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions:   builtInFuncs,
		NFTStorageHandler:  nftStorageHandler,
		DataPool:           datapool,
		CompiledSCPool:     datapool.SmartContracts(),
		NilCompiledSCStore: true,
		EpochNotifier:      &epochNotifier.EpochNotifierStub{},
	}

	economicsData, err := createEconomicsData(0)
	if err != nil {
		log.LogIfError(err)
	}

	argVMContainer := metachain.ArgsNewVMContainerFactory{
		ArgBlockChainHook:   args,
		Economics:           economicsData,
		MessageSignVerifier: &mock.MessageSignVerifierMock{},
		GasSchedule:         gasSchedule,
		NodesConfigProvider: &mock.NodesSetupStub{},
		Hasher:              testHasher,
		Marshalizer:         testMarshalizer,
		SystemSCConfig:      createSystemSCConfig(),
		ValidatorAccountsDB: accnts,
		ChanceComputer:      &mock.NodesCoordinatorMock{},
		EpochNotifier:       &epochNotifier.EpochNotifierStub{},
		EpochConfig:         createEpochConfig(),
		ShardCoordinator:    mock.NewMultiShardsCoordinatorMock(1),
	}
	argVMContainer.EpochConfig.EnableEpochs.UnbondTokensV2EnableEpoch = arg.UnbondTokensV2EnableEpoch
	vmFactory, err := metachain.NewVMContainerFactory(argVMContainer)
	if err != nil {
		log.LogIfError(err)
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		panic(err)
	}

	blockChainHook, _ := vmFactory.BlockChainHookImpl().(*hooks.BlockChainHookImpl)
	_ = vmcommonBuiltInFunctions.SetPayableHandler(builtInFuncs, blockChainHook)

	return vmContainer, blockChainHook
}

func createEpochConfig() *config.EpochConfig {
	return &config.EpochConfig{
		EnableEpochs: config.EnableEpochs{
			StakingV2EnableEpoch:               0,
			StakeEnableEpoch:                   0,
			DoubleKeyProtectionEnableEpoch:     0,
			ESDTEnableEpoch:                    0,
			GovernanceEnableEpoch:              0,
			DelegationManagerEnableEpoch:       0,
			DelegationSmartContractEnableEpoch: 0,
		},
	}
}

func createSystemSCConfig() *config.SystemSmartContractsConfig {
	return &config.SystemSmartContractsConfig{
		ESDTSystemSCConfig: config.ESDTSystemSCConfig{
			BaseIssuingCost: "5000000000000000000",
			OwnerAddress:    "3132333435363738393031323334353637383930313233343536373839303233",
		},
		GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
			V1: config.GovernanceSystemSCConfigV1{
				ProposalCost:     "500",
				NumNodes:         100,
				MinQuorum:        50,
				MinPassThreshold: 50,
				MinVetoThreshold: 50,
			},
			Active: config.GovernanceSystemSCConfigActive{
				ProposalCost:     "500",
				MinQuorum:        "50",
				MinPassThreshold: "50",
				MinVetoThreshold: "50",
			},
			FirstWhitelistedAddress: "3132333435363738393031323334353637383930313233343536373839303234",
		},
		StakingSystemSCConfig: config.StakingSystemSCConfig{
			GenesisNodePrice:                     "2500000000000000000000",
			MinStakeValue:                        "100000000000000000000",
			MinUnstakeTokensValue:                "10000000000000000000",
			UnJailValue:                          "2500000000000000000",
			MinStepValue:                         "100000000000000000000",
			UnBondPeriod:                         250,
			UnBondPeriodInEpochs:                 1,
			NumRoundsWithoutBleed:                100,
			MaximumPercentageToBleed:             0.5,
			BleedPercentagePerRound:              0.00001,
			MaxNumberOfNodesForStake:             36,
			ActivateBLSPubKeyMessageVerification: false,
		},
		DelegationManagerSystemSCConfig: config.DelegationManagerSystemSCConfig{
			MinCreationDeposit:  "1250000000000000000000",
			MinStakeAmount:      "10000000000000000000",
			ConfigChangeAddress: "3132333435363738393031323334353637383930313233343536373839303234",
		},
		DelegationSystemSCConfig: config.DelegationSystemSCConfig{
			MinServiceFee: 1,
			MaxServiceFee: 20,
		},
	}
}

func createDefaultVMConfig() *config.VirtualMachineConfig {
	return &config.VirtualMachineConfig{
		ArwenVersions: []config.ArwenVersionByEpoch{
			{StartEpoch: 0, Version: "*"},
		},
	}
}

// ResultsCreateTxProcessor is the struct that will hold all needed processor instances
type ResultsCreateTxProcessor struct {
	TxProc             process.TransactionProcessor
	SCProc             *smartContract.TestScProcessor
	IntermediateTxProc process.IntermediateTransactionHandler
	EconomicsHandler   process.EconomicsDataHandler
	CostHandler        external.TransactionCostHandler
	TxLogProc          process.TransactionLogProcessor
}

// CreateTxProcessorWithOneSCExecutorWithVMs -
func CreateTxProcessorWithOneSCExecutorWithVMs(
	accnts state.AccountsAdapter,
	vmContainer process.VirtualMachinesContainer,
	blockChainHook *hooks.BlockChainHookImpl,
	feeAccumulator process.TransactionFeeHandler,
	shardCoordinator sharding.Coordinator,
	argEnableEpoch ArgEnableEpoch,
	arwenChangeLocker common.Locker,
	poolsHolder dataRetriever.PoolsHolder,
) (*ResultsCreateTxProcessor, error) {
	if check.IfNil(poolsHolder) {
		poolsHolder = dataRetrieverMock.NewPoolsHolderMock()
	}

	esdtTransferParser, _ := parsers.NewESDTTransferParser(testMarshalizer)
	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:    pubkeyConv,
		ShardCoordinator:   shardCoordinator,
		BuiltInFunctions:   blockChainHook.GetBuiltinFunctionsContainer(),
		ArgumentParser:     parsers.NewCallArgsParser(),
		EpochNotifier:      forking.NewGenericEpochNotifier(),
		ESDTTransferParser: esdtTransferParser,
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argsTxTypeHandler)

	gasSchedule := make(map[string]map[string]uint64)
	defaults.FillGasMapInternal(gasSchedule, 1)
	economicsData, err := createEconomicsData(argEnableEpoch.PenalizedTooMuchGasEnableEpoch)
	if err != nil {
		return nil, err
	}

	gasComp, err := preprocess.NewGasComputation(economicsData, txTypeHandler, forking.NewGenericEpochNotifier(), argEnableEpoch.DeployEnableEpoch)
	if err != nil {
		return nil, err
	}

	enableEpochs := config.EnableEpochs{
		PenalizedTooMuchGasEnableEpoch:      argEnableEpoch.PenalizedTooMuchGasEnableEpoch,
		BuiltInFunctionsEnableEpoch:         argEnableEpoch.BuiltinEnableEpoch,
		SCDeployEnableEpoch:                 argEnableEpoch.DeployEnableEpoch,
		BackwardCompSaveKeyValueEnableEpoch: argEnableEpoch.BackwardCompSaveKeyValueEnableEpoch,
		CleanUpInformativeSCRsEnableEpoch:   argEnableEpoch.CleanUpInformativeSCRsEnableEpoch,
	}
	logProc, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		SaveInStorageEnabled: false,
		Marshalizer:          testMarshalizer,
	})
	intermediateTxHandler := &mock.IntermediateTransactionHandlerMock{}
	argsNewSCProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:       vmContainer,
		ArgsParser:        smartContract.NewArgumentParser(),
		Hasher:            testHasher,
		Marshalizer:       testMarshalizer,
		AccountsDB:        accnts,
		BlockChainHook:    blockChainHook,
		BuiltInFunctions:  blockChainHook.GetBuiltinFunctionsContainer(),
		PubkeyConv:        pubkeyConv,
		ShardCoordinator:  shardCoordinator,
		ScrForwarder:      intermediateTxHandler,
		BadTxForwarder:    intermediateTxHandler,
		TxFeeHandler:      feeAccumulator,
		EconomicsFee:      economicsData,
		TxTypeHandler:     txTypeHandler,
		GasHandler:        gasComp,
		GasSchedule:       mock.NewGasScheduleNotifierMock(gasSchedule),
		TxLogsProcessor:   logProc,
		EpochNotifier:     forking.NewGenericEpochNotifier(),
		ArwenChangeLocker: arwenChangeLocker,
		VMOutputCacher:    txcache.NewDisabledCache(),
		EnableEpochs:      enableEpochs,
	}

	scProcessor, err := smartContract.NewSmartContractProcessor(argsNewSCProcessor)
	if err != nil {
		return nil, err
	}
	testScProcessor := smartContract.NewTestScProcessor(scProcessor)

	argsNewTxProcessor := transaction.ArgsNewTxProcessor{
		Accounts:                       accnts,
		Hasher:                         testHasher,
		PubkeyConv:                     pubkeyConv,
		Marshalizer:                    testMarshalizer,
		SignMarshalizer:                testMarshalizer,
		ShardCoordinator:               shardCoordinator,
		ScProcessor:                    scProcessor,
		TxFeeHandler:                   feeAccumulator,
		TxTypeHandler:                  txTypeHandler,
		EconomicsFee:                   economicsData,
		ReceiptForwarder:               intermediateTxHandler,
		BadTxForwarder:                 intermediateTxHandler,
		ArgsParser:                     smartContract.NewArgumentParser(),
		ScrForwarder:                   intermediateTxHandler,
		EpochNotifier:                  forking.NewGenericEpochNotifier(),
		PenalizedTooMuchGasEnableEpoch: argEnableEpoch.PenalizedTooMuchGasEnableEpoch,
		RelayedTxEnableEpoch:           argEnableEpoch.RelayedTxEnableEpoch,
		MetaProtectionEnableEpoch:      argEnableEpoch.MetaProtectionEnableEpoch,
	}
	txProcessor, err := transaction.NewTxProcessor(argsNewTxProcessor)
	if err != nil {
		return nil, err
	}

	// create transaction simulator
	readOnlyAccountsDB, err := txsimulator.NewReadOnlyAccountsDB(accnts)
	if err != nil {
		return nil, err
	}

	interimProcFactory, err := shard.NewIntermediateProcessorsContainerFactory(
		shardCoordinator,
		testMarshalizer,
		testHasher,
		pubkeyConv,
		disabled.NewChainStorer(),
		poolsHolder,
		&processDisabled.FeeHandler{},
	)
	if err != nil {
		return nil, err
	}

	interimProcContainer, err := interimProcFactory.Create()
	if err != nil {
		return nil, err
	}

	scForwarder, err := interimProcContainer.Get(block.SmartContractResultBlock)
	if err != nil {
		return nil, err
	}

	argsNewSCProcessor.ScrForwarder = scForwarder

	receiptTxInterim, err := interimProcContainer.Get(block.ReceiptBlock)
	if err != nil {
		return nil, err
	}
	argsNewTxProcessor.ReceiptForwarder = receiptTxInterim

	badTxInterim, err := interimProcContainer.Get(block.InvalidBlock)
	if err != nil {
		return nil, err
	}
	argsNewSCProcessor.BadTxForwarder = badTxInterim
	argsNewTxProcessor.BadTxForwarder = badTxInterim

	argsNewSCProcessor.TxFeeHandler = &processDisabled.FeeHandler{}
	argsNewTxProcessor.TxFeeHandler = &processDisabled.FeeHandler{}

	argsNewSCProcessor.AccountsDB = readOnlyAccountsDB

	vmOutputCacher, _ := storageUnit.NewCache(storageUnit.CacheConfig{
		Type:     storageUnit.LRUCache,
		Capacity: 10000,
	})
	txSimulatorProcessorArgs := txsimulator.ArgsTxSimulator{
		AddressPubKeyConverter: pubkeyConv,
		ShardCoordinator:       shardCoordinator,
		VMOutputCacher:         vmOutputCacher,
		Marshalizer:            testMarshalizer,
		Hasher:                 testHasher,
	}

	argsNewSCProcessor.VMOutputCacher = txSimulatorProcessorArgs.VMOutputCacher

	scProcessorTxSim, err := smartContract.NewSmartContractProcessor(argsNewSCProcessor)
	if err != nil {
		return nil, err
	}
	argsNewTxProcessor.ScProcessor = scProcessorTxSim

	argsNewTxProcessor.Accounts = readOnlyAccountsDB

	txSimulatorProcessorArgs.TransactionProcessor, err = transaction.NewTxProcessor(argsNewTxProcessor)
	if err != nil {
		return nil, err
	}

	txSimulatorProcessorArgs.IntermediateProcContainer = interimProcContainer

	txSimulator, err := txsimulator.NewTransactionSimulator(txSimulatorProcessorArgs)
	if err != nil {
		return nil, err
	}

	txCostEstimator, err := transaction.NewTransactionCostEstimator(
		txTypeHandler,
		economicsData,
		txSimulator,
		argsNewTxProcessor.Accounts,
		shardCoordinator,
	)
	if err != nil {
		return nil, err
	}

	return &ResultsCreateTxProcessor{
		TxProc:             txProcessor,
		SCProc:             testScProcessor,
		IntermediateTxProc: intermediateTxHandler,
		EconomicsHandler:   economicsData,
		CostHandler:        txCostEstimator,
		TxLogProc:          logProc,
	}, nil
}

// TestDeployedContractContents -
func TestDeployedContractContents(
	t *testing.T,
	destinationAddressBytes []byte,
	accnts state.AccountsAdapter,
	requiredBalance *big.Int,
	scCode string,
	dataValues map[string]*big.Int,
) {

	scCodeBytes, _ := hex.DecodeString(scCode)
	destinationRecovAccount, _ := accnts.GetExistingAccount(destinationAddressBytes)
	destinationRecovShardAccount, ok := destinationRecovAccount.(state.UserAccountHandler)

	assert.True(t, ok)
	assert.NotNil(t, destinationRecovShardAccount)
	assert.Equal(t, uint64(0), destinationRecovShardAccount.GetNonce())
	assert.Equal(t, requiredBalance, destinationRecovShardAccount.GetBalance())
	// test codehash
	assert.Equal(t, testHasher.Compute(string(scCodeBytes)), destinationRecovShardAccount.GetCodeHash())
	// test code
	assert.Equal(t, scCodeBytes, accnts.GetCode(destinationRecovShardAccount.GetCodeHash()))
	// in this test we know we have a as a variable inside the contract, we can ask directly its value
	// using trackableDataTrie functionality
	assert.NotNil(t, destinationRecovShardAccount.GetRootHash())

	for variable, requiredVal := range dataValues {
		contractVariableData, err := destinationRecovShardAccount.DataTrieTracker().RetrieveValue([]byte(variable))
		assert.Nil(t, err)
		assert.NotNil(t, contractVariableData)

		contractVariableValue := big.NewInt(0).SetBytes(contractVariableData)
		assert.Equal(t, requiredVal, contractVariableValue)
	}
}

// AccountExists -
func AccountExists(accnts state.AccountsAdapter, addressBytes []byte) bool {
	accnt, _ := accnts.GetExistingAccount(addressBytes)

	return accnt != nil
}

// CreatePreparedTxProcessorAndAccountsWithVMs -
func CreatePreparedTxProcessorAndAccountsWithVMs(
	senderNonce uint64,
	senderAddressBytes []byte,
	senderBalance *big.Int,
	argEnableEpoch ArgEnableEpoch,
) (*VMTestContext, error) {
	feeAccumulator, _ := postprocess.NewFeeAccumulator()
	accounts := CreateInMemoryShardAccountsDB()
	_, _ = CreateAccount(accounts, senderAddressBytes, senderNonce, senderBalance)
	vmConfig := createDefaultVMConfig()
	arwenChangeLocker := &sync.RWMutex{}
	vmContainer, blockchainHook, pool := CreateVMAndBlockchainHookAndDataPool(accounts, nil, vmConfig, oneShardCoordinator, arwenChangeLocker)
	res, err := CreateTxProcessorWithOneSCExecutorWithVMs(
		accounts,
		vmContainer,
		blockchainHook,
		feeAccumulator,
		oneShardCoordinator,
		argEnableEpoch,
		arwenChangeLocker,
		pool,
	)
	if err != nil {
		return nil, err
	}

	return &VMTestContext{
		TxProcessor:    res.TxProc,
		ScProcessor:    res.SCProc,
		Accounts:       accounts,
		BlockchainHook: blockchainHook,
		VMContainer:    vmContainer,
		TxFeeHandler:   feeAccumulator,
		ScForwarder:    res.IntermediateTxProc,
	}, nil
}

// CreatePreparedTxProcessorWithVMs -
func CreatePreparedTxProcessorWithVMs(argEnableEpoch ArgEnableEpoch) (*VMTestContext, error) {
	return CreatePreparedTxProcessorWithVMsWithShardCoordinator(argEnableEpoch, oneShardCoordinator)
}

// CreatePreparedTxProcessorWithVMsWithShardCoordinator -
func CreatePreparedTxProcessorWithVMsWithShardCoordinator(argEnableEpoch ArgEnableEpoch, shardCoordinator sharding.Coordinator) (*VMTestContext, error) {
	feeAccumulator, _ := postprocess.NewFeeAccumulator()
	accounts := CreateInMemoryShardAccountsDB()
	vmConfig := createDefaultVMConfig()
	arwenChangeLocker := &sync.RWMutex{}

	testGasSchedule := arwenConfig.MakeGasMapForTests()
	defaults.FillGasMapInternal(testGasSchedule, 1)
	gasSchedule := mock.NewGasScheduleNotifierMock(testGasSchedule)

	vmContainer, blockchainHook, pool := CreateVMAndBlockchainHookAndDataPool(accounts, gasSchedule, vmConfig, shardCoordinator, arwenChangeLocker)
	res, err := CreateTxProcessorWithOneSCExecutorWithVMs(
		accounts,
		vmContainer,
		blockchainHook,
		feeAccumulator,
		shardCoordinator,
		argEnableEpoch,
		arwenChangeLocker,
		pool,
	)
	if err != nil {
		return nil, err
	}

	return &VMTestContext{
		TxProcessor:      res.TxProc,
		ScProcessor:      res.SCProc,
		Accounts:         accounts,
		BlockchainHook:   blockchainHook,
		VMContainer:      vmContainer,
		TxFeeHandler:     feeAccumulator,
		ScForwarder:      res.IntermediateTxProc,
		ShardCoordinator: shardCoordinator,
		EconomicsData:    res.EconomicsHandler,
		TxCostHandler:    res.CostHandler,
		TxsLogsProcessor: res.TxLogProc,
		GasSchedule:      gasSchedule,
	}, nil
}

// CreateTxProcessorArwenVMWithGasSchedule -
func CreateTxProcessorArwenVMWithGasSchedule(
	senderNonce uint64,
	senderAddressBytes []byte,
	senderBalance *big.Int,
	gasScheduleMap map[string]map[string]uint64,
	argEnableEpoch ArgEnableEpoch,
) (*VMTestContext, error) {
	feeAccumulator, _ := postprocess.NewFeeAccumulator()
	accounts := CreateInMemoryShardAccountsDB()
	_, _ = CreateAccount(accounts, senderAddressBytes, senderNonce, senderBalance)
	vmConfig := createDefaultVMConfig()
	arwenChangeLocker := &sync.RWMutex{}

	gasScheduleNotifier := mock.NewGasScheduleNotifierMock(gasScheduleMap)

	vmContainer, blockchainHook, pool := CreateVMAndBlockchainHookAndDataPool(accounts, gasScheduleNotifier, vmConfig, oneShardCoordinator, arwenChangeLocker)
	res, err := CreateTxProcessorWithOneSCExecutorWithVMs(
		accounts,
		vmContainer,
		blockchainHook,
		feeAccumulator,
		oneShardCoordinator,
		argEnableEpoch,
		arwenChangeLocker,
		pool,
	)
	if err != nil {
		return nil, err
	}

	return &VMTestContext{
		TxProcessor:    res.TxProc,
		ScProcessor:    res.SCProc,
		Accounts:       accounts,
		BlockchainHook: blockchainHook,
		VMContainer:    vmContainer,
		TxFeeHandler:   feeAccumulator,
		ScForwarder:    res.IntermediateTxProc,
		GasSchedule:    gasScheduleNotifier,
	}, nil
}

// CreateTxProcessorArwenWithVMConfig -
func CreateTxProcessorArwenWithVMConfig(
	argEnableEpoch ArgEnableEpoch,
	vmConfig *config.VirtualMachineConfig,
	gasSchedule map[string]map[string]uint64,
) (*VMTestContext, error) {
	feeAccumulator, _ := postprocess.NewFeeAccumulator()
	accounts := CreateInMemoryShardAccountsDB()
	arwenChangeLocker := &sync.RWMutex{}
	gasScheduleNotifier := mock.NewGasScheduleNotifierMock(gasSchedule)
	vmContainer, blockchainHook, pool := CreateVMAndBlockchainHookAndDataPool(accounts, gasScheduleNotifier, vmConfig, oneShardCoordinator, arwenChangeLocker)
	res, err := CreateTxProcessorWithOneSCExecutorWithVMs(
		accounts,
		vmContainer,
		blockchainHook,
		feeAccumulator,
		oneShardCoordinator,
		argEnableEpoch,
		arwenChangeLocker,
		pool,
	)
	if err != nil {
		return nil, err
	}

	return &VMTestContext{
		TxProcessor:     res.TxProc,
		ScProcessor:     res.SCProc,
		Accounts:        accounts,
		BlockchainHook:  blockchainHook,
		VMContainer:     vmContainer,
		TxFeeHandler:    feeAccumulator,
		ScForwarder:     res.IntermediateTxProc,
		GasSchedule:     gasScheduleNotifier,
		VMConfiguration: vmConfig,
		EpochNotifier:   globalEpochNotifier,
	}, nil
}

// CreatePreparedTxProcessorAndAccountsWithMockedVM -
func CreatePreparedTxProcessorAndAccountsWithMockedVM(
	vmOpGas uint64,
	senderNonce uint64,
	senderAddressBytes []byte,
	senderBalance *big.Int,
	argEnableEpoch ArgEnableEpoch,
	arwenChangeLocker common.Locker,
) (process.TransactionProcessor, state.AccountsAdapter, error) {

	accnts := CreateInMemoryShardAccountsDB()
	_, _ = CreateAccount(accnts, senderAddressBytes, senderNonce, senderBalance)

	txProcessor, err := CreateTxProcessorWithOneSCExecutorMockVM(accnts, vmOpGas, argEnableEpoch, arwenChangeLocker)
	if err != nil {
		return nil, nil, err
	}

	return txProcessor, accnts, err
}

// CreateTx -
func CreateTx(
	senderAddressBytes []byte,
	receiverAddressBytes []byte,
	senderNonce uint64,
	value *big.Int,
	gasPrice uint64,
	gasLimit uint64,
	scCodeOrFunc string,
) *dataTransaction.Transaction {

	txData := scCodeOrFunc
	tx := &dataTransaction.Transaction{
		Nonce:    senderNonce,
		Value:    new(big.Int).Set(value),
		SndAddr:  senderAddressBytes,
		RcvAddr:  receiverAddressBytes,
		Data:     []byte(txData),
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}

	return tx
}

// CreateDeployTx -
func CreateDeployTx(
	senderAddressBytes []byte,
	senderNonce uint64,
	value *big.Int,
	gasPrice uint64,
	gasLimit uint64,
	scCodeAndVMType string,
) *dataTransaction.Transaction {

	return &dataTransaction.Transaction{
		Nonce:    senderNonce,
		Value:    new(big.Int).Set(value),
		SndAddr:  senderAddressBytes,
		RcvAddr:  CreateEmptyAddress(),
		Data:     []byte(scCodeAndVMType),
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}
}

// TestAccount -
func TestAccount(
	t *testing.T,
	accnts state.AccountsAdapter,
	senderAddressBytes []byte,
	expectedNonce uint64,
	expectedBalance *big.Int,
) *big.Int {

	senderRecovAccount, _ := accnts.GetExistingAccount(senderAddressBytes)
	if senderRecovAccount == nil {
		return big.NewInt(0)
	}

	senderRecovShardAccount := senderRecovAccount.(state.UserAccountHandler)

	assert.Equal(t, expectedNonce, senderRecovShardAccount.GetNonce())
	assert.Equal(t, expectedBalance, senderRecovShardAccount.GetBalance())
	return senderRecovShardAccount.GetBalance()
}

// ComputeExpectedBalance -
func ComputeExpectedBalance(
	existing *big.Int,
	transferred *big.Int,
	gasLimit uint64,
	gasPrice uint64,
) *big.Int {

	expectedSenderBalance := big.NewInt(0).Sub(existing, transferred)
	gasFunds := big.NewInt(0).Mul(big.NewInt(0).SetUint64(gasLimit), big.NewInt(0).SetUint64(gasPrice))
	expectedSenderBalance.Sub(expectedSenderBalance, gasFunds)

	return expectedSenderBalance
}

// GetIntValueFromSC -
func GetIntValueFromSC(
	gasSchedule map[string]map[string]uint64,
	accnts state.AccountsAdapter,
	scAddressBytes []byte,
	funcName string,
	args ...[]byte,
) *big.Int {

	vmOutput := GetVmOutput(gasSchedule, accnts, scAddressBytes, funcName, args...)

	return big.NewInt(0).SetBytes(vmOutput.ReturnData[0])
}

// GetVmOutput -
func GetVmOutput(gasSchedule map[string]map[string]uint64, accnts state.AccountsAdapter, scAddressBytes []byte, funcName string, args ...[]byte) *vmcommon.VMOutput {
	vmConfig := createDefaultVMConfig()
	gasScheduleNotifier := mock.NewGasScheduleNotifierMock(gasSchedule)
	vmContainer, blockChainHook, _ := CreateVMAndBlockchainHookAndDataPool(accnts, gasScheduleNotifier, vmConfig, oneShardCoordinator, &sync.RWMutex{})
	defer func() {
		_ = vmContainer.Close()
	}()

	feeHandler := &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return uint64(math.MaxUint64)
		},
	}

	argsNewSCQueryService := smartContract.ArgsNewSCQueryService{
		VmContainer:       vmContainer,
		EconomicsFee:      feeHandler,
		BlockChainHook:    blockChainHook,
		BlockChain:        &mock.BlockChainMock{},
		ArwenChangeLocker: &sync.RWMutex{},
		Bootstrapper:      syncDisabled.NewDisabledBootstrapper(),
	}
	scQueryService, _ := smartContract.NewSCQueryService(argsNewSCQueryService)

	vmOutput, err := scQueryService.ExecuteQuery(&process.SCQuery{
		ScAddress: scAddressBytes,
		FuncName:  funcName,
		Arguments: args,
	})

	if err != nil {
		fmt.Println("ERROR at GetVmOutput()", err)
		return nil
	}

	return vmOutput
}

// ComputeGasLimit -
func ComputeGasLimit(gasSchedule map[string]map[string]uint64, testContext *VMTestContext, tx *dataTx.Transaction) uint64 {
	vmConfig := createDefaultVMConfig()
	gasScheduleNotifier := mock.NewGasScheduleNotifierMock(gasSchedule)
	vmContainer, blockChainHook, _ := CreateVMAndBlockchainHookAndDataPool(testContext.Accounts, gasScheduleNotifier, vmConfig, oneShardCoordinator, &sync.RWMutex{})
	defer func() {
		_ = vmContainer.Close()
	}()

	argsNewSCQueryService := smartContract.ArgsNewSCQueryService{
		VmContainer:    vmContainer,
		EconomicsFee:   testContext.EconomicsData,
		BlockChainHook: blockChainHook,
		BlockChain: &mock.BlockChainMock{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.Header{
					ShardID: testContext.ShardCoordinator.SelfId(),
				}
			},
		},
		ArwenChangeLocker: &sync.RWMutex{},
		Bootstrapper:      syncDisabled.NewDisabledBootstrapper(),
	}
	scQueryService, _ := smartContract.NewSCQueryService(argsNewSCQueryService)

	gasLimit, err := scQueryService.ComputeScCallGasLimit(tx)
	if err != nil {
		log.Error("ComputeScCallGasLimit()", "error", err)
		return 0
	}

	return gasLimit
}

// CreateTransferTokenTx -
func CreateTransferTokenTx(
	nonce uint64,
	functionName string,
	value *big.Int,
	scAddrress []byte,
	sndAddress []byte,
	rcvAddress []byte,
) *dataTransaction.Transaction {
	return &dataTransaction.Transaction{
		Nonce:    nonce,
		Value:    big.NewInt(0),
		RcvAddr:  scAddrress,
		SndAddr:  sndAddress,
		GasPrice: 1,
		GasLimit: 7000000,
		Data:     []byte(functionName + "@" + hex.EncodeToString(rcvAddress) + "@00" + hex.EncodeToString(value.Bytes())),
		ChainID:  integrationTests.ChainID,
	}
}

// CreateTransaction -
func CreateTransaction(
	nonce uint64,
	value *big.Int,
	sndAddress []byte,
	rcvAddress []byte,
	gasprice uint64,
	gasLimit uint64,
	data []byte,
) *dataTransaction.Transaction {
	return &dataTransaction.Transaction{
		Nonce:    nonce,
		Value:    big.NewInt(0).Set(value),
		RcvAddr:  rcvAddress,
		SndAddr:  sndAddress,
		GasPrice: gasprice,
		GasLimit: gasLimit,
		Data:     data,
	}
}

// GetNodeIndex -
func GetNodeIndex(nodeList []*integrationTests.TestProcessorNode, node *integrationTests.TestProcessorNode) (int, error) {
	for i := range nodeList {
		if node == nodeList[i] {
			return i, nil
		}
	}

	return 0, errors.New("no such node in list")
}

// CreatePreparedTxProcessorWithVMsMultiShard -
func CreatePreparedTxProcessorWithVMsMultiShard(selfShardID uint32, argEnableEpoch ArgEnableEpoch) (*VMTestContext, error) {
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(3, selfShardID)

	feeAccumulator, _ := postprocess.NewFeeAccumulator()
	accounts := CreateInMemoryShardAccountsDB()

	arwenChangeLocker := &sync.RWMutex{}
	var vmContainer process.VirtualMachinesContainer
	var blockchainHook *hooks.BlockChainHookImpl
	if selfShardID == core.MetachainShardId {
		vmContainer, blockchainHook = CreateVMAndBlockchainHookMeta(accounts, nil, shardCoordinator, argEnableEpoch)
	} else {
		vmConfig := createDefaultVMConfig()
		vmContainer, blockchainHook, _ = CreateVMAndBlockchainHookAndDataPool(accounts, nil, vmConfig, shardCoordinator, arwenChangeLocker)
	}

	res, err := CreateTxProcessorWithOneSCExecutorWithVMs(
		accounts,
		vmContainer,
		blockchainHook,
		feeAccumulator,
		shardCoordinator,
		argEnableEpoch,
		arwenChangeLocker,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &VMTestContext{
		TxProcessor:      res.TxProc,
		ScProcessor:      res.SCProc,
		Accounts:         accounts,
		BlockchainHook:   blockchainHook,
		VMContainer:      vmContainer,
		TxFeeHandler:     feeAccumulator,
		ShardCoordinator: shardCoordinator,
		ScForwarder:      res.IntermediateTxProc,
		EconomicsData:    res.EconomicsHandler,
		Marshalizer:      testMarshalizer,
		TxsLogsProcessor: res.TxLogProc,
	}, nil
}

func defaultStorageConfig() *config.StorageConfig {
	return &config.StorageConfig{
		Cache: config.CacheConfig{
			Name:     "SmartContractsStorage",
			Type:     "LRU",
			Capacity: 100,
		},
		DB: config.DBConfig{
			FilePath:          "SmartContractsStorage",
			Type:              "LvlDBSerial",
			BatchDelaySeconds: 2,
			MaxBatchSize:      100,
		},
	}
}
