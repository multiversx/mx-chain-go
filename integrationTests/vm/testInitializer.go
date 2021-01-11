// +build cgo

package vm

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"testing"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/config"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/forking"
	"github.com/ElrondNetwork/elrond-go/core/parsers"
	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	dataTransaction "github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/data/trie/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/postprocess"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: Merge test utilities from this file with the ones from "arwen/utils.go"

var testMarshalizer = &marshal.GogoProtoMarshalizer{}
var testHasher = sha256.Sha256{}
var oneShardCoordinator = mock.NewMultiShardsCoordinatorMock(2)
var pubkeyConv, _ = pubkeyConverter.NewHexPubkeyConverter(32)

var log = logger.GetOrCreate("integrationtests")

const maxTrieLevelInMemory = uint(5)

// ArgEnableEpoch will specify the enable epoch values for certain flags
type ArgEnableEpoch struct {
	PenalizedTooMuchGasEnableEpoch uint32
	BuiltinEnableEpoch             uint32
	DeployEnableEpoch              uint32
	MetaProtectionEnableEpoch      uint32
	RelayedTxEnableEpoch           uint32
}

// VMTestContext -
type VMTestContext struct {
	TxProcessor      process.TransactionProcessor
	ScProcessor      process.SmartContractProcessor
	Accounts         state.AccountsAdapter
	BlockchainHook   vmcommon.BlockchainHook
	VMContainer      process.VirtualMachinesContainer
	TxFeeHandler     process.TransactionFeeHandler
	ShardCoordinator sharding.Coordinator
	ScForwarder      process.IntermediateTransactionHandler
	EconomicsData    process.EconomicsDataHandler
}

// Close -
func (vmTestContext *VMTestContext) Close() {
	_ = vmTestContext.VMContainer.Close()
}

// GetLatestError -
func (vmTestContext *VMTestContext) GetLatestError() error {
	return smartContract.GetLatestTestError(vmTestContext.ScProcessor)
}

// GetGasRemaining -
func (vmTestContext *VMTestContext) GetGasRemaining() uint64 {
	return smartContract.GetGasRemaining(vmTestContext.ScProcessor)
}

// GetIntermediateTransactions -
func (vmTestContext *VMTestContext) GetIntermediateTransactions(t *testing.T) []data.TransactionHandler {
	scForwarder := vmTestContext.ScForwarder
	mockIntermediate, ok := scForwarder.(*mock.IntermediateTransactionHandlerMock)
	require.True(t, ok)

	return mockIntermediate.GetIntermediateTransactions()
}

type accountFactory struct {
}

// CreateAccount -
func (af *accountFactory) CreateAccount(address []byte) (state.AccountHandler, error) {
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
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}
	trieStorage, _ := trie.NewTrieStorageManager(
		store,
		marsh,
		testHasher,
		config.DBConfig{
			FilePath:          "TrieStorage",
			Type:              "MemoryDB",
			BatchDelaySeconds: 30,
			MaxBatchSize:      6,
			MaxOpenFiles:      10,
		},
		ewl,
		generalCfg,
	)

	tr, _ := trie.NewTrie(trieStorage, marsh, testHasher, maxTrieLevelInMemory)
	adb, _ := state.NewAccountsDB(tr, testHasher, marsh, &accountFactory{})

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

func createEconomicsData(tb testing.TB, penalizedTooMuchGasEnableEpoch uint32) process.EconomicsDataHandler {
	maxGasLimitPerBlock := strconv.FormatUint(math.MaxUint64, 10)
	minGasPrice := strconv.FormatUint(1, 10)
	minGasLimit := strconv.FormatUint(1, 10)
	testProtocolSustainabilityAddress := "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp"
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
				LeaderPercentage:              0.1,
				DeveloperPercentage:           0.1,
				ProtocolSustainabilityAddress: testProtocolSustainabilityAddress,
				TopUpGradientPoint:            "100000",
			},
			FeeSettings: config.FeeSettings{
				MaxGasLimitPerBlock:     maxGasLimitPerBlock,
				MaxGasLimitPerMetaBlock: maxGasLimitPerBlock,
				MinGasPrice:             minGasPrice,
				MinGasLimit:             minGasLimit,
				GasPerDataByte:          "1",
				GasPriceModifier:        1.0,
			},
		},
		PenalizedTooMuchGasEnableEpoch: penalizedTooMuchGasEnableEpoch,
		EpochNotifier:                  &mock.EpochNotifierStub{},
	}
	economicsData, err := economics.NewEconomicsData(argsNewEconomicsData)
	require.Nil(tb, err)
	return economicsData
}

// CreateTxProcessorWithOneSCExecutorMockVM -
func CreateTxProcessorWithOneSCExecutorMockVM(
	tb testing.TB,
	accnts state.AccountsAdapter,
	opGas uint64,
	argEnableEpoch ArgEnableEpoch,
) process.TransactionProcessor {

	builtInFuncs := builtInFunctions.NewBuiltInFunctionContainer()
	datapool := testscommon.NewPoolsHolderMock()
	args := hooks.ArgBlockChainHook{
		Accounts:           accnts,
		PubkeyConv:         pubkeyConv,
		StorageService:     &mock.ChainStorerMock{},
		BlockChain:         &mock.BlockChainMock{},
		ShardCoordinator:   oneShardCoordinator,
		Marshalizer:        testMarshalizer,
		Uint64Converter:    &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions:   builtInFuncs,
		DataPool:           datapool,
		CompiledSCPool:     datapool.SmartContracts(),
		NilCompiledSCStore: true,
	}

	blockChainHook, _ := hooks.NewBlockChainHookImpl(args)
	vm, _ := mock.NewOneSCExecutorMockVM(blockChainHook, testHasher)
	vm.GasForOperation = opGas

	vmContainer := &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return vm, nil
		}}

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:  pubkeyConv,
		ShardCoordinator: oneShardCoordinator,
		BuiltInFuncNames: builtInFuncs.Keys(),
		ArgumentParser:   parsers.NewCallArgsParser(),
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	gasSchedule := make(map[string]map[string]uint64)
	defaults.FillGasMapInternal(gasSchedule, 1)

	economicsData := createEconomicsData(tb, argEnableEpoch.PenalizedTooMuchGasEnableEpoch)
	argsNewSCProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:    vmContainer,
		ArgsParser:     smartContract.NewArgumentParser(),
		Hasher:         testHasher,
		Marshalizer:    testMarshalizer,
		AccountsDB:     accnts,
		BlockChainHook: blockChainHook,
		PubkeyConv:     pubkeyConv,
		Coordinator:    oneShardCoordinator,
		ScrForwarder:   &mock.IntermediateTransactionHandlerMock{},
		BadTxForwarder: &mock.IntermediateTransactionHandlerMock{},
		TxFeeHandler:   &mock.UnsignedTxHandlerMock{},
		EconomicsFee:   economicsData,
		TxTypeHandler:  txTypeHandler,
		GasHandler: &mock.GasHandlerMock{
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		},
		GasSchedule:                    mock.NewGasScheduleNotifierMock(gasSchedule),
		BuiltInFunctions:               blockChainHook.GetBuiltInFunctions(),
		TxLogsProcessor:                &mock.TxLogsProcessorStub{},
		EpochNotifier:                  forking.NewGenericEpochNotifier(),
		PenalizedTooMuchGasEnableEpoch: argEnableEpoch.PenalizedTooMuchGasEnableEpoch,
		BuiltinEnableEpoch:             argEnableEpoch.BuiltinEnableEpoch,
		DeployEnableEpoch:              argEnableEpoch.DeployEnableEpoch,
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
		TxFeeHandler:                   &mock.UnsignedTxHandlerMock{},
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
	txProcessor, _ := transaction.NewTxProcessor(argsNewTxProcessor)

	return txProcessor
}

// CreateOneSCExecutorMockVM -
func CreateOneSCExecutorMockVM(accnts state.AccountsAdapter) vmcommon.VMExecutionHandler {
	datapool := testscommon.NewPoolsHolderMock()
	args := hooks.ArgBlockChainHook{
		Accounts:           accnts,
		PubkeyConv:         pubkeyConv,
		StorageService:     &mock.ChainStorerMock{},
		BlockChain:         &mock.BlockChainMock{},
		ShardCoordinator:   oneShardCoordinator,
		Marshalizer:        testMarshalizer,
		Uint64Converter:    &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions:   builtInFunctions.NewBuiltInFunctionContainer(),
		DataPool:           datapool,
		CompiledSCPool:     datapool.SmartContracts(),
		NilCompiledSCStore: true,
	}
	blockChainHook, _ := hooks.NewBlockChainHookImpl(args)
	vm, _ := mock.NewOneSCExecutorMockVM(blockChainHook, testHasher)

	return vm
}

// CreateVMAndBlockchainHook -
func CreateVMAndBlockchainHook(
	accnts state.AccountsAdapter,
	gasSchedule map[string]map[string]uint64,
	outOfProcess bool,
	shardCoordinator sharding.Coordinator,
) (process.VirtualMachinesContainer, *hooks.BlockChainHookImpl) {
	actualGasSchedule := gasSchedule
	if gasSchedule == nil {
		actualGasSchedule = arwenConfig.MakeGasMapForTests()
		defaults.FillGasMapInternal(actualGasSchedule, 1)
	}

	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule:     mock.NewGasScheduleNotifierMock(actualGasSchedule),
		MapDNSAddresses: make(map[string]struct{}),
		Marshalizer:     testMarshalizer,
		Accounts:        accnts,
	}
	builtInFuncFactory, _ := builtInFunctions.NewBuiltInFunctionsFactory(argsBuiltIn)
	builtInFuncs, _ := builtInFuncFactory.CreateBuiltInFunctionContainer()

	datapool := testscommon.NewPoolsHolderMock()
	args := hooks.ArgBlockChainHook{
		Accounts:           accnts,
		PubkeyConv:         pubkeyConv,
		StorageService:     &mock.ChainStorerMock{},
		BlockChain:         &mock.BlockChainMock{},
		ShardCoordinator:   shardCoordinator,
		Marshalizer:        testMarshalizer,
		Uint64Converter:    &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions:   builtInFuncs,
		DataPool:           datapool,
		CompiledSCPool:     datapool.SmartContracts(),
		NilCompiledSCStore: true,
	}

	maxGasLimitPerBlock := uint64(0xFFFFFFFFFFFFFFFF)
	vmFactory, err := shard.NewVMContainerFactory(
		config.VirtualMachineConfig{
			OutOfProcessEnabled: outOfProcess,
			OutOfProcessConfig:  config.VirtualMachineOutOfProcessConfig{MaxLoopTime: 1000},
		},
		maxGasLimitPerBlock,
		mock.NewGasScheduleNotifierMock(actualGasSchedule),
		args,
		0,
		0,
	)
	if err != nil {
		log.LogIfError(err)
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		panic(err)
	}

	blockChainHook, _ := vmFactory.BlockChainHookImpl().(*hooks.BlockChainHookImpl)
	_ = builtInFunctions.SetPayableHandler(builtInFuncs, blockChainHook)

	return vmContainer, blockChainHook
}

// CreateTxProcessorWithOneSCExecutorWithVMs -
func CreateTxProcessorWithOneSCExecutorWithVMs(
	tb testing.TB,
	accnts state.AccountsAdapter,
	vmContainer process.VirtualMachinesContainer,
	blockChainHook *hooks.BlockChainHookImpl,
	feeAccumulator process.TransactionFeeHandler,
	shardCoordinator sharding.Coordinator,
	argEnableEpoch ArgEnableEpoch,
) (process.TransactionProcessor, process.SmartContractProcessor, process.IntermediateTransactionHandler, process.EconomicsDataHandler) {
	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:  pubkeyConv,
		ShardCoordinator: shardCoordinator,
		BuiltInFuncNames: blockChainHook.GetBuiltInFunctions().Keys(),
		ArgumentParser:   parsers.NewCallArgsParser(),
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argsTxTypeHandler)

	gasSchedule := make(map[string]map[string]uint64)
	defaults.FillGasMapInternal(gasSchedule, 1)
	economicsData := createEconomicsData(tb, argEnableEpoch.PenalizedTooMuchGasEnableEpoch)

	gasComp, _ := preprocess.NewGasComputation(economicsData, txTypeHandler)

	intermediateTxHandler := &mock.IntermediateTransactionHandlerMock{}
	argsNewSCProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:                    vmContainer,
		ArgsParser:                     smartContract.NewArgumentParser(),
		Hasher:                         testHasher,
		Marshalizer:                    testMarshalizer,
		AccountsDB:                     accnts,
		BlockChainHook:                 blockChainHook,
		PubkeyConv:                     pubkeyConv,
		Coordinator:                    shardCoordinator,
		ScrForwarder:                   intermediateTxHandler,
		BadTxForwarder:                 intermediateTxHandler,
		TxFeeHandler:                   feeAccumulator,
		EconomicsFee:                   economicsData,
		TxTypeHandler:                  txTypeHandler,
		GasHandler:                     gasComp,
		GasSchedule:                    mock.NewGasScheduleNotifierMock(gasSchedule),
		BuiltInFunctions:               blockChainHook.GetBuiltInFunctions(),
		TxLogsProcessor:                &mock.TxLogsProcessorStub{},
		EpochNotifier:                  forking.NewGenericEpochNotifier(),
		PenalizedTooMuchGasEnableEpoch: argEnableEpoch.PenalizedTooMuchGasEnableEpoch,
		DeployEnableEpoch:              argEnableEpoch.DeployEnableEpoch,
		BuiltinEnableEpoch:             argEnableEpoch.BuiltinEnableEpoch,
	}

	scProcessor, _ := smartContract.NewSmartContractProcessor(argsNewSCProcessor)

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
	txProcessor, _ := transaction.NewTxProcessor(argsNewTxProcessor)

	return txProcessor, scProcessor, intermediateTxHandler, economicsData
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
	//test codehash
	assert.Equal(t, testHasher.Compute(string(scCodeBytes)), destinationRecovShardAccount.GetCodeHash())
	//test code
	assert.Equal(t, scCodeBytes, destinationRecovShardAccount.GetCode())
	//in this test we know we have a as a variable inside the contract, we can ask directly its value
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
	tb testing.TB,
	senderNonce uint64,
	senderAddressBytes []byte,
	senderBalance *big.Int,
	argEnableEpoch ArgEnableEpoch,
) VMTestContext {
	feeAccumulator, _ := postprocess.NewFeeAccumulator()
	accounts := CreateInMemoryShardAccountsDB()
	_, _ = CreateAccount(accounts, senderAddressBytes, senderNonce, senderBalance)
	vmContainer, blockchainHook := CreateVMAndBlockchainHook(accounts, nil, false, oneShardCoordinator)
	txProcessor, scProcessor, scForwarder, _ := CreateTxProcessorWithOneSCExecutorWithVMs(
		tb,
		accounts,
		vmContainer,
		blockchainHook,
		feeAccumulator,
		oneShardCoordinator,
		argEnableEpoch,
	)

	return VMTestContext{
		TxProcessor:    txProcessor,
		ScProcessor:    scProcessor,
		Accounts:       accounts,
		BlockchainHook: blockchainHook,
		VMContainer:    vmContainer,
		TxFeeHandler:   feeAccumulator,
		ScForwarder:    scForwarder,
	}
}

// CreatePreparedTxProcessorWithVMs -
func CreatePreparedTxProcessorWithVMs(tb testing.TB, argEnableEpoch ArgEnableEpoch) VMTestContext {
	feeAccumulator, _ := postprocess.NewFeeAccumulator()
	accounts := CreateInMemoryShardAccountsDB()
	vmContainer, blockchainHook := CreateVMAndBlockchainHook(accounts, nil, false, oneShardCoordinator)
	txProcessor, scProcessor, scForwarder, economicsData := CreateTxProcessorWithOneSCExecutorWithVMs(
		tb,
		accounts,
		vmContainer,
		blockchainHook,
		feeAccumulator,
		oneShardCoordinator,
		argEnableEpoch,
	)

	return VMTestContext{
		TxProcessor:      txProcessor,
		ScProcessor:      scProcessor,
		Accounts:         accounts,
		BlockchainHook:   blockchainHook,
		VMContainer:      vmContainer,
		TxFeeHandler:     feeAccumulator,
		ScForwarder:      scForwarder,
		ShardCoordinator: oneShardCoordinator,
		EconomicsData:    economicsData,
	}
}

// CreateTxProcessorArwenVMWithGasSchedule -
func CreateTxProcessorArwenVMWithGasSchedule(
	tb testing.TB,
	senderNonce uint64,
	senderAddressBytes []byte,
	senderBalance *big.Int,
	gasSchedule map[string]map[string]uint64,
	outOfProcess bool,
	argEnableEpoch ArgEnableEpoch,
) VMTestContext {
	feeAccumulator, _ := postprocess.NewFeeAccumulator()
	accounts := CreateInMemoryShardAccountsDB()
	_, _ = CreateAccount(accounts, senderAddressBytes, senderNonce, senderBalance)
	vmContainer, blockchainHook := CreateVMAndBlockchainHook(accounts, gasSchedule, outOfProcess, oneShardCoordinator)
	txProcessor, scProcessor, scForwarder, _ := CreateTxProcessorWithOneSCExecutorWithVMs(
		tb,
		accounts,
		vmContainer,
		blockchainHook,
		feeAccumulator,
		oneShardCoordinator,
		argEnableEpoch,
	)

	return VMTestContext{
		TxProcessor:    txProcessor,
		ScProcessor:    scProcessor,
		Accounts:       accounts,
		BlockchainHook: blockchainHook,
		VMContainer:    vmContainer,
		TxFeeHandler:   feeAccumulator,
		ScForwarder:    scForwarder,
	}
}

// CreatePreparedTxProcessorAndAccountsWithMockedVM -
func CreatePreparedTxProcessorAndAccountsWithMockedVM(
	tb testing.TB,
	vmOpGas uint64,
	senderNonce uint64,
	senderAddressBytes []byte,
	senderBalance *big.Int,
	argEnableEpoch ArgEnableEpoch,
) (process.TransactionProcessor, state.AccountsAdapter) {

	accnts := CreateInMemoryShardAccountsDB()
	_, _ = CreateAccount(accnts, senderAddressBytes, senderNonce, senderBalance)

	txProcessor := CreateTxProcessorWithOneSCExecutorMockVM(tb, accnts, vmOpGas, argEnableEpoch)
	assert.NotNil(tb, txProcessor)

	return txProcessor, accnts
}

// CreateTx -
func CreateTx(
	tb testing.TB,
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
	assert.NotNil(tb, tx)

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
func GetIntValueFromSC(gasSchedule map[string]map[string]uint64, accnts state.AccountsAdapter, scAddressBytes []byte, funcName string, args ...[]byte) *big.Int {
	vmContainer, blockChainHook := CreateVMAndBlockchainHook(accnts, gasSchedule, false, oneShardCoordinator)
	defer func() {
		_ = vmContainer.Close()
	}()

	feeHandler := &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return uint64(math.MaxUint64)
		},
	}

	scQueryService, _ := smartContract.NewSCQueryService(vmContainer, feeHandler, blockChainHook, &mock.BlockChainMock{})

	vmOutput, err := scQueryService.ExecuteQuery(&process.SCQuery{
		ScAddress: scAddressBytes,
		FuncName:  funcName,
		Arguments: args,
	})

	if err != nil {
		fmt.Println("ERROR at GetIntValueFromSC()", err)
		return big.NewInt(-1)
	}

	return big.NewInt(0).SetBytes(vmOutput.ReturnData[0])
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
		GasLimit: 5000000,
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
func CreatePreparedTxProcessorWithVMsMultiShard(tb testing.TB, selfShardID uint32, argEnableEpoch ArgEnableEpoch) *VMTestContext {
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(3, selfShardID)

	feeAccumulator, _ := postprocess.NewFeeAccumulator()
	accounts := CreateInMemoryShardAccountsDB()
	vmContainer, blockchainHook := CreateVMAndBlockchainHook(accounts, nil, false, shardCoordinator)
	txProcessor, scProcessor, scrForwarder, economicsData := CreateTxProcessorWithOneSCExecutorWithVMs(
		tb,
		accounts,
		vmContainer,
		blockchainHook,
		feeAccumulator,
		shardCoordinator,
		argEnableEpoch,
	)

	return &VMTestContext{
		TxProcessor:      txProcessor,
		ScProcessor:      scProcessor,
		Accounts:         accounts,
		BlockchainHook:   blockchainHook,
		VMContainer:      vmContainer,
		TxFeeHandler:     feeAccumulator,
		ShardCoordinator: shardCoordinator,
		ScForwarder:      scrForwarder,
		EconomicsData:    economicsData,
	}
}
