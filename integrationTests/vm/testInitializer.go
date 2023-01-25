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

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	dataTransaction "github.com/multiversx/mx-chain-core-go/data/transaction"
	dataTx "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart/bootstrap/disabled"
	processDisabled "github.com/multiversx/mx-chain-go/genesis/process/disabled"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/postprocess"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/economics"
	"github.com/multiversx/mx-chain-go/process/factory/metachain"
	"github.com/multiversx/mx-chain-go/process/factory/shard"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/smartContract/builtInFunctions"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks/counters"
	syncDisabled "github.com/multiversx/mx-chain-go/process/sync/disabled"
	"github.com/multiversx/mx-chain-go/process/transaction"
	"github.com/multiversx/mx-chain-go/process/transactionLog"
	"github.com/multiversx/mx-chain-go/process/txsimulator"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/integrationtests"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/defaults"
	logger "github.com/multiversx/mx-chain-logger-go"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	vmcommonBuiltInFunctions "github.com/multiversx/mx-chain-vm-common-go/builtInFunctions"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
	wasmConfig "github.com/multiversx/mx-chain-vm-v1_4-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dnsAddr = []byte{0, 0, 0, 0, 0, 0, 0, 0, 5, 0, 137, 17, 46, 56, 127, 47, 62, 172, 4, 126, 190, 242, 221, 230, 209, 243, 105, 104, 242, 66, 49, 49}

// TODO: Merge test utilities from this file with the ones from "wasmvm/utils.go"

var globalEpochNotifier = forking.NewGenericEpochNotifier()
var pubkeyConv, _ = pubkeyConverter.NewHexPubkeyConverter(32)
var log = logger.GetOrCreate("integrationtests")

func getZeroGasAndFees() scheduled.GasAndFees {
	return scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		GasProvided:     0,
		GasPenalized:    0,
		GasRefunded:     0,
	}
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
	ChainHandler        *testscommon.ChainHandlerStub
	TxProcessor         process.TransactionProcessor
	ScProcessor         *smartContract.TestScProcessor
	Accounts            state.AccountsAdapter
	BlockchainHook      vmcommon.BlockchainHook
	VMContainer         process.VirtualMachinesContainer
	TxFeeHandler        process.TransactionFeeHandler
	ShardCoordinator    sharding.Coordinator
	ScForwarder         process.IntermediateTransactionHandler
	EconomicsData       process.EconomicsDataHandler
	Marshalizer         marshal.Marshalizer
	GasSchedule         core.GasScheduleNotifier
	VMConfiguration     *config.VirtualMachineConfig
	EpochNotifier       process.EpochNotifier
	EnableEpochsHandler common.EnableEpochsHandler
	SCQueryService      *smartContract.SCQueryService

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

// GetCompositeTestError -
func (vmTestContext *VMTestContext) GetCompositeTestError() error {
	return vmTestContext.ScProcessor.GetCompositeTestError()
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
func (vmTestContext *VMTestContext) GetIntermediateTransactions(tb testing.TB) []data.TransactionHandler {
	scForwarder := vmTestContext.ScForwarder
	mockIntermediate, ok := scForwarder.(*mock.IntermediateTransactionHandlerMock)
	require.True(tb, ok)

	return mockIntermediate.GetIntermediateTransactions()
}

// CleanIntermediateTransactions -
func (vmTestContext *VMTestContext) CleanIntermediateTransactions(tb testing.TB) {
	scForwarder := vmTestContext.ScForwarder
	mockIntermediate, ok := scForwarder.(*mock.IntermediateTransactionHandlerMock)
	require.True(tb, ok)

	mockIntermediate.Clean()
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
	scAddressBytes := vmTestContext.Contract.Address
	feeHandler := &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return uint64(math.MaxUint64)
		},
	}

	argsNewSCQueryService := smartContract.ArgsNewSCQueryService{
		VmContainer:              vmTestContext.VMContainer,
		EconomicsFee:             feeHandler,
		BlockChainHook:           vmTestContext.BlockchainHook.(process.BlockChainHookHandler),
		BlockChain:               &testscommon.ChainHandlerStub{},
		WasmVMChangeLocker:       &sync.RWMutex{},
		Bootstrapper:             syncDisabled.NewDisabledBootstrapper(),
		AllowExternalQueriesChan: common.GetClosedUnbufferedChannel(),
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

// CreateEmptyAddress -
func CreateEmptyAddress() []byte {
	buff := make([]byte, integrationtests.TestHasher.Size())

	return buff
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

func createEconomicsData(enableEpochsConfig config.EnableEpochs) (process.EconomicsDataHandler, error) {
	maxGasLimitPerBlock := strconv.FormatUint(math.MaxUint64, 10)
	minGasPrice := strconv.FormatUint(1, 10)
	minGasLimit := strconv.FormatUint(1, 10)
	testProtocolSustainabilityAddress := "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp"

	builtInCost, _ := economics.NewBuiltInFunctionsCost(&economics.ArgsBuiltInFunctionCost{
		ArgsParser:  smartContract.NewArgumentParser(),
		GasSchedule: mock.NewGasScheduleNotifierMock(defaults.FillGasMapInternal(map[string]map[string]uint64{}, 1)),
	})

	realEpochNotifier := forking.NewGenericEpochNotifier()
	enableEpochsHandler, _ := enablers.NewEnableEpochsHandler(enableEpochsConfig, realEpochNotifier)

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
		EpochNotifier:               realEpochNotifier,
		EnableEpochsHandler:         enableEpochsHandler,
		BuiltInFunctionsCostHandler: builtInCost,
	}

	return economics.NewEconomicsData(argsNewEconomicsData)
}

// CreateTxProcessorWithOneSCExecutorMockVM -
func CreateTxProcessorWithOneSCExecutorMockVM(
	accnts state.AccountsAdapter,
	opGas uint64,
	enableEpochsConfig config.EnableEpochs,
	wasmVMChangeLocker common.Locker,
) (process.TransactionProcessor, error) {

	genericEpochNotifier := forking.NewGenericEpochNotifier()
	enableEpochsHandler, _ := enablers.NewEnableEpochsHandler(enableEpochsConfig, genericEpochNotifier)

	gasSchedule := make(map[string]map[string]uint64)
	defaults.FillGasMapInternal(gasSchedule, 1)
	gasScheduleNotifier := mock.NewGasScheduleNotifierMock(gasSchedule)

	builtInFuncs := vmcommonBuiltInFunctions.NewBuiltInFunctionContainer()
	datapool := dataRetrieverMock.NewPoolsHolderMock()
	args := hooks.ArgBlockChainHook{
		Accounts:              accnts,
		PubkeyConv:            pubkeyConv,
		StorageService:        &storageStubs.ChainStorerStub{},
		BlockChain:            &testscommon.ChainHandlerStub{},
		ShardCoordinator:      mock.NewMultiShardsCoordinatorMock(2),
		Marshalizer:           integrationtests.TestMarshalizer,
		Uint64Converter:       &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions:      builtInFuncs,
		NFTStorageHandler:     &testscommon.SimpleNFTStorageHandlerStub{},
		GlobalSettingsHandler: &testscommon.ESDTGlobalSettingsHandlerStub{},
		DataPool:              datapool,
		CompiledSCPool:        datapool.SmartContracts(),
		NilCompiledSCStore:    true,
		ConfigSCStorage:       *defaultStorageConfig(),
		EpochNotifier:         genericEpochNotifier,
		EnableEpochsHandler:   enableEpochsHandler,
		GasSchedule:           gasScheduleNotifier,
		Counter:               &testscommon.BlockChainHookCounterStub{},
	}

	blockChainHook, _ := hooks.NewBlockChainHookImpl(args)
	vm, _ := mock.NewOneSCExecutorMockVM(blockChainHook, integrationtests.TestHasher)
	vm.GasForOperation = opGas

	vmContainer := &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return vm, nil
		}}

	esdtTransferParser, _ := parsers.NewESDTTransferParser(integrationtests.TestMarshalizer)
	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:     pubkeyConv,
		ShardCoordinator:    mock.NewMultiShardsCoordinatorMock(2),
		BuiltInFunctions:    builtInFuncs,
		ArgumentParser:      parsers.NewCallArgsParser(),
		ESDTTransferParser:  esdtTransferParser,
		EnableEpochsHandler: enableEpochsHandler,
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argsTxTypeHandler)

	economicsData, err := createEconomicsData(enableEpochsConfig)
	if err != nil {
		return nil, err
	}

	argsNewSCProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:      vmContainer,
		ArgsParser:       smartContract.NewArgumentParser(),
		Hasher:           integrationtests.TestHasher,
		Marshalizer:      integrationtests.TestMarshalizer,
		AccountsDB:       accnts,
		BlockChainHook:   blockChainHook,
		BuiltInFunctions: builtInFuncs,
		PubkeyConv:       pubkeyConv,
		ShardCoordinator: mock.NewMultiShardsCoordinatorMock(2),
		ScrForwarder:     &mock.IntermediateTransactionHandlerMock{},
		BadTxForwarder:   &mock.IntermediateTransactionHandlerMock{},
		TxFeeHandler:     &testscommon.UnsignedTxHandlerStub{},
		EconomicsFee:     economicsData,
		TxTypeHandler:    txTypeHandler,
		GasHandler: &testscommon.GasHandlerStub{
			SetGasRefundedCalled: func(gasRefunded uint64, hash []byte) {},
		},
		GasSchedule:         gasScheduleNotifier,
		TxLogsProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler: enableEpochsHandler,
		VMOutputCacher:      txcache.NewDisabledCache(),
		WasmVMChangeLocker:  wasmVMChangeLocker,
	}
	scProcessor, _ := smartContract.NewSmartContractProcessor(argsNewSCProcessor)

	argsNewTxProcessor := transaction.ArgsNewTxProcessor{
		Accounts:            accnts,
		Hasher:              integrationtests.TestHasher,
		PubkeyConv:          pubkeyConv,
		Marshalizer:         integrationtests.TestMarshalizer,
		SignMarshalizer:     integrationtests.TestMarshalizer,
		ShardCoordinator:    mock.NewMultiShardsCoordinatorMock(2),
		ScProcessor:         scProcessor,
		TxFeeHandler:        &testscommon.UnsignedTxHandlerStub{},
		TxTypeHandler:       txTypeHandler,
		EconomicsFee:        economicsData,
		ReceiptForwarder:    &mock.IntermediateTransactionHandlerMock{},
		BadTxForwarder:      &mock.IntermediateTransactionHandlerMock{},
		ArgsParser:          smartContract.NewArgumentParser(),
		ScrForwarder:        &mock.IntermediateTransactionHandlerMock{},
		EnableEpochsHandler: enableEpochsHandler,
	}

	return transaction.NewTxProcessor(argsNewTxProcessor)
}

// CreateOneSCExecutorMockVM -
func CreateOneSCExecutorMockVM(accnts state.AccountsAdapter) vmcommon.VMExecutionHandler {
	datapool := dataRetrieverMock.NewPoolsHolderMock()
	args := hooks.ArgBlockChainHook{
		Accounts:              accnts,
		PubkeyConv:            pubkeyConv,
		StorageService:        &storageStubs.ChainStorerStub{},
		BlockChain:            &testscommon.ChainHandlerStub{},
		ShardCoordinator:      mock.NewMultiShardsCoordinatorMock(2),
		Marshalizer:           integrationtests.TestMarshalizer,
		Uint64Converter:       &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions:      vmcommonBuiltInFunctions.NewBuiltInFunctionContainer(),
		NFTStorageHandler:     &testscommon.SimpleNFTStorageHandlerStub{},
		GlobalSettingsHandler: &testscommon.ESDTGlobalSettingsHandlerStub{},
		DataPool:              datapool,
		CompiledSCPool:        datapool.SmartContracts(),
		NilCompiledSCStore:    true,
		ConfigSCStorage:       *defaultStorageConfig(),
		EpochNotifier:         &epochNotifier.EpochNotifierStub{},
		EnableEpochsHandler:   &testscommon.EnableEpochsHandlerStub{},
		GasSchedule:           createMockGasScheduleNotifier(),
		Counter:               &testscommon.BlockChainHookCounterStub{},
	}
	blockChainHook, _ := hooks.NewBlockChainHookImpl(args)
	vm, _ := mock.NewOneSCExecutorMockVM(blockChainHook, integrationtests.TestHasher)

	return vm
}

// CreateVMAndBlockchainHookAndDataPool -
func CreateVMAndBlockchainHookAndDataPool(
	accnts state.AccountsAdapter,
	gasSchedule core.GasScheduleNotifier,
	vmConfig *config.VirtualMachineConfig,
	shardCoordinator sharding.Coordinator,
	wasmVMChangeLocker common.Locker,
	epochNotifierInstance process.EpochNotifier,
	enableEpochsHandler common.EnableEpochsHandler,
	chainHandler data.ChainHandler,
) (process.VirtualMachinesContainer, *hooks.BlockChainHookImpl, dataRetriever.PoolsHolder) {
	if check.IfNil(gasSchedule) || gasSchedule.LatestGasSchedule() == nil {
		testGasSchedule := wasmConfig.MakeGasMapForTests()
		defaults.FillGasMapInternal(testGasSchedule, 1)
		gasSchedule = mock.NewGasScheduleNotifierMock(testGasSchedule)
	}

	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule: gasSchedule,
		MapDNSAddresses: map[string]struct{}{
			string(dnsAddr): {},
		},
		Marshalizer:               integrationtests.TestMarshalizer,
		Accounts:                  accnts,
		ShardCoordinator:          shardCoordinator,
		EpochNotifier:             epochNotifierInstance,
		EnableEpochsHandler:       enableEpochsHandler,
		MaxNumNodesInTransferRole: 100,
	}
	argsBuiltIn.AutomaticCrawlerAddresses = integrationTests.GenerateOneAddressPerShard(argsBuiltIn.ShardCoordinator)
	builtInFuncFactory, _ := builtInFunctions.CreateBuiltInFunctionsFactory(argsBuiltIn)

	esdtTransferParser, _ := parsers.NewESDTTransferParser(integrationtests.TestMarshalizer)
	counter, _ := counters.NewUsageCounter(esdtTransferParser)

	datapool := dataRetrieverMock.NewPoolsHolderMock()
	args := hooks.ArgBlockChainHook{
		Accounts:              accnts,
		PubkeyConv:            pubkeyConv,
		StorageService:        &storageStubs.ChainStorerStub{},
		BlockChain:            chainHandler,
		ShardCoordinator:      shardCoordinator,
		Marshalizer:           integrationtests.TestMarshalizer,
		Uint64Converter:       &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions:      builtInFuncFactory.BuiltInFunctionContainer(),
		NFTStorageHandler:     builtInFuncFactory.NFTStorageHandler(),
		GlobalSettingsHandler: builtInFuncFactory.ESDTGlobalSettingsHandler(),
		DataPool:              datapool,
		CompiledSCPool:        datapool.SmartContracts(),
		NilCompiledSCStore:    true,
		ConfigSCStorage:       *defaultStorageConfig(),
		EpochNotifier:         epochNotifierInstance,
		EnableEpochsHandler:   enableEpochsHandler,
		GasSchedule:           gasSchedule,
		Counter:               counter,
	}

	hasher := &hashingMocks.HasherMock{}
	maxGasLimitPerBlock := uint64(0xFFFFFFFFFFFFFFFF)
	blockChainHookImpl, _ := hooks.NewBlockChainHookImpl(args)
	argsNewVMFactory := shard.ArgVMContainerFactory{
		Config:              *vmConfig,
		BlockGasLimit:       maxGasLimitPerBlock,
		GasSchedule:         gasSchedule,
		BlockChainHook:      blockChainHookImpl,
		BuiltInFunctions:    args.BuiltInFunctions,
		EpochNotifier:       epochNotifierInstance,
		EnableEpochsHandler: enableEpochsHandler,
		WasmVMChangeLocker:  wasmVMChangeLocker,
		ESDTTransferParser:  esdtTransferParser,
		Hasher:              hasher,
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
	_ = builtInFuncFactory.SetPayableHandler(blockChainHook)

	return vmContainer, blockChainHook, datapool
}

// CreateVMAndBlockchainHookMeta -
func CreateVMAndBlockchainHookMeta(
	accnts state.AccountsAdapter,
	gasSchedule core.GasScheduleNotifier,
	shardCoordinator sharding.Coordinator,
	enableEpochsConfig config.EnableEpochs,
) (process.VirtualMachinesContainer, *hooks.BlockChainHookImpl) {
	if check.IfNil(gasSchedule) || gasSchedule.LatestGasSchedule() == nil {
		testGasSchedule := wasmConfig.MakeGasMapForTests()
		defaults.FillGasMapInternal(testGasSchedule, 1)
		gasSchedule = mock.NewGasScheduleNotifierMock(testGasSchedule)
	}

	enableEpochsHandler, _ := enablers.NewEnableEpochsHandler(enableEpochsConfig, globalEpochNotifier)
	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule: gasSchedule,
		MapDNSAddresses: map[string]struct{}{
			string(dnsAddr): {},
		},
		Marshalizer:               integrationtests.TestMarshalizer,
		Accounts:                  accnts,
		ShardCoordinator:          shardCoordinator,
		EpochNotifier:             globalEpochNotifier,
		EnableEpochsHandler:       enableEpochsHandler,
		MaxNumNodesInTransferRole: 100,
	}
	argsBuiltIn.AutomaticCrawlerAddresses = integrationTests.GenerateOneAddressPerShard(argsBuiltIn.ShardCoordinator)
	builtInFuncFactory, _ := builtInFunctions.CreateBuiltInFunctionsFactory(argsBuiltIn)

	datapool := dataRetrieverMock.NewPoolsHolderMock()
	args := hooks.ArgBlockChainHook{
		Accounts:              accnts,
		PubkeyConv:            pubkeyConv,
		StorageService:        &storageStubs.ChainStorerStub{},
		BlockChain:            &testscommon.ChainHandlerStub{},
		ShardCoordinator:      shardCoordinator,
		Marshalizer:           integrationtests.TestMarshalizer,
		Uint64Converter:       &mock.Uint64ByteSliceConverterMock{},
		BuiltInFunctions:      builtInFuncFactory.BuiltInFunctionContainer(),
		NFTStorageHandler:     builtInFuncFactory.NFTStorageHandler(),
		GlobalSettingsHandler: builtInFuncFactory.ESDTGlobalSettingsHandler(),
		DataPool:              datapool,
		CompiledSCPool:        datapool.SmartContracts(),
		NilCompiledSCStore:    true,
		EpochNotifier:         globalEpochNotifier,
		EnableEpochsHandler:   enableEpochsHandler,
		GasSchedule:           gasSchedule,
		Counter:               &testscommon.BlockChainHookCounterStub{},
	}

	economicsData, err := createEconomicsData(config.EnableEpochs{})
	if err != nil {
		log.LogIfError(err)
	}

	blockChainHookImpl, _ := hooks.NewBlockChainHookImpl(args)
	argVMContainer := metachain.ArgsNewVMContainerFactory{
		BlockChainHook:      blockChainHookImpl,
		PubkeyConv:          args.PubkeyConv,
		Economics:           economicsData,
		MessageSignVerifier: &mock.MessageSignVerifierMock{},
		GasSchedule:         gasSchedule,
		NodesConfigProvider: &mock.NodesSetupStub{},
		Hasher:              integrationtests.TestHasher,
		Marshalizer:         integrationtests.TestMarshalizer,
		SystemSCConfig:      createSystemSCConfig(),
		ValidatorAccountsDB: accnts,
		ChanceComputer:      &shardingMocks.NodesCoordinatorMock{},
		ShardCoordinator:    mock.NewMultiShardsCoordinatorMock(1),
		EnableEpochsHandler: enableEpochsHandler,
	}
	vmFactory, err := metachain.NewVMContainerFactory(argVMContainer)
	if err != nil {
		log.LogIfError(err)
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		panic(err)
	}

	blockChainHook, _ := vmFactory.BlockChainHookImpl().(*hooks.BlockChainHookImpl)
	_ = builtInFuncFactory.SetPayableHandler(blockChainHook)

	return vmContainer, blockChainHook
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
		WasmVMVersions: []config.WasmVMVersionByEpoch{
			{
				StartEpoch: 0,
				Version:    "*",
			},
		},
		TimeOutForSCExecutionInMilliseconds: 10000, // 10 seconds
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
	enableEpochsConfig config.EnableEpochs,
	wasmVMChangeLocker common.Locker,
	poolsHolder dataRetriever.PoolsHolder,
	epochNotifierInstance process.EpochNotifier,
) (*ResultsCreateTxProcessor, error) {
	if check.IfNil(poolsHolder) {
		poolsHolder = dataRetrieverMock.NewPoolsHolderMock()
	}

	enableEpochsHandler, _ := enablers.NewEnableEpochsHandler(enableEpochsConfig, epochNotifierInstance)

	esdtTransferParser, _ := parsers.NewESDTTransferParser(integrationtests.TestMarshalizer)
	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:     pubkeyConv,
		ShardCoordinator:    shardCoordinator,
		BuiltInFunctions:    blockChainHook.GetBuiltinFunctionsContainer(),
		ArgumentParser:      parsers.NewCallArgsParser(),
		ESDTTransferParser:  esdtTransferParser,
		EnableEpochsHandler: enableEpochsHandler,
	}
	txTypeHandler, _ := coordinator.NewTxTypeHandler(argsTxTypeHandler)

	gasSchedule := make(map[string]map[string]uint64)
	defaults.FillGasMapInternal(gasSchedule, 1)
	economicsData, err := createEconomicsData(enableEpochsConfig)
	if err != nil {
		return nil, err
	}

	gasComp, err := preprocess.NewGasComputation(economicsData, txTypeHandler, enableEpochsHandler)
	if err != nil {
		return nil, err
	}

	logProc, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		SaveInStorageEnabled: false,
		Marshalizer:          integrationtests.TestMarshalizer,
	})

	intermediateTxHandler := &mock.IntermediateTransactionHandlerMock{}
	argsNewSCProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:         vmContainer,
		ArgsParser:          smartContract.NewArgumentParser(),
		Hasher:              integrationtests.TestHasher,
		Marshalizer:         integrationtests.TestMarshalizer,
		AccountsDB:          accnts,
		BlockChainHook:      blockChainHook,
		BuiltInFunctions:    blockChainHook.GetBuiltinFunctionsContainer(),
		PubkeyConv:          pubkeyConv,
		ShardCoordinator:    shardCoordinator,
		ScrForwarder:        intermediateTxHandler,
		BadTxForwarder:      intermediateTxHandler,
		TxFeeHandler:        feeAccumulator,
		EconomicsFee:        economicsData,
		TxTypeHandler:       txTypeHandler,
		GasHandler:          gasComp,
		GasSchedule:         mock.NewGasScheduleNotifierMock(gasSchedule),
		TxLogsProcessor:     logProc,
		EnableEpochsHandler: enableEpochsHandler,
		WasmVMChangeLocker:  wasmVMChangeLocker,
		VMOutputCacher:      txcache.NewDisabledCache(),
	}

	scProcessor, err := smartContract.NewSmartContractProcessor(argsNewSCProcessor)
	if err != nil {
		return nil, err
	}
	testScProcessor := smartContract.NewTestScProcessor(scProcessor)

	argsNewTxProcessor := transaction.ArgsNewTxProcessor{
		Accounts:            accnts,
		Hasher:              integrationtests.TestHasher,
		PubkeyConv:          pubkeyConv,
		Marshalizer:         integrationtests.TestMarshalizer,
		SignMarshalizer:     integrationtests.TestMarshalizer,
		ShardCoordinator:    shardCoordinator,
		ScProcessor:         scProcessor,
		TxFeeHandler:        feeAccumulator,
		TxTypeHandler:       txTypeHandler,
		EconomicsFee:        economicsData,
		ReceiptForwarder:    intermediateTxHandler,
		BadTxForwarder:      intermediateTxHandler,
		ArgsParser:          smartContract.NewArgumentParser(),
		ScrForwarder:        intermediateTxHandler,
		EnableEpochsHandler: enableEpochsHandler,
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
		integrationtests.TestMarshalizer,
		integrationtests.TestHasher,
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

	vmOutputCacher, _ := storageunit.NewCache(storageunit.CacheConfig{
		Type:     storageunit.LRUCache,
		Capacity: 10000,
	})
	txSimulatorProcessorArgs := txsimulator.ArgsTxSimulator{
		AddressPubKeyConverter: pubkeyConv,
		ShardCoordinator:       shardCoordinator,
		VMOutputCacher:         vmOutputCacher,
		Marshalizer:            integrationtests.TestMarshalizer,
		Hasher:                 integrationtests.TestHasher,
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
		argsNewSCProcessor.EnableEpochsHandler,
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
	// test code hash
	assert.Equal(t, integrationtests.TestHasher.Compute(string(scCodeBytes)), destinationRecovShardAccount.GetCodeHash())
	// test code
	assert.Equal(t, scCodeBytes, accnts.GetCode(destinationRecovShardAccount.GetCodeHash()))
	// in this test we know we have a as a variable inside the contract, we can ask directly its value
	// using trackableDataTrie functionality
	assert.NotNil(t, destinationRecovShardAccount.GetRootHash())

	for variable, requiredVal := range dataValues {
		contractVariableData, _, err := destinationRecovShardAccount.RetrieveValue([]byte(variable))
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
	enableEpochsConfig config.EnableEpochs,
) (*VMTestContext, error) {
	feeAccumulator, _ := postprocess.NewFeeAccumulator()
	accounts := integrationtests.CreateInMemoryShardAccountsDB()
	_, _ = CreateAccount(accounts, senderAddressBytes, senderNonce, senderBalance)
	vmConfig := createDefaultVMConfig()
	wasmVMChangeLocker := &sync.RWMutex{}
	epochNotifierInstance := forking.NewGenericEpochNotifier()
	enableEpochsHandler, _ := enablers.NewEnableEpochsHandler(enableEpochsConfig, epochNotifierInstance)
	chainHandler := &testscommon.ChainHandlerStub{}

	vmContainer, blockchainHook, pool := CreateVMAndBlockchainHookAndDataPool(
		accounts,
		nil,
		vmConfig,
		mock.NewMultiShardsCoordinatorMock(2),
		wasmVMChangeLocker,
		epochNotifierInstance,
		enableEpochsHandler,
		chainHandler,
	)
	res, err := CreateTxProcessorWithOneSCExecutorWithVMs(
		accounts,
		vmContainer,
		blockchainHook,
		feeAccumulator,
		mock.NewMultiShardsCoordinatorMock(2),
		enableEpochsConfig,
		wasmVMChangeLocker,
		pool,
		epochNotifierInstance,
	)
	if err != nil {
		return nil, err
	}

	return &VMTestContext{
		TxProcessor:         res.TxProc,
		ScProcessor:         res.SCProc,
		Accounts:            accounts,
		BlockchainHook:      blockchainHook,
		VMContainer:         vmContainer,
		TxFeeHandler:        feeAccumulator,
		ScForwarder:         res.IntermediateTxProc,
		EpochNotifier:       epochNotifierInstance,
		EnableEpochsHandler: enableEpochsHandler,
		ChainHandler:        chainHandler,
	}, nil
}

func createMockGasScheduleNotifier() *mock.GasScheduleNotifierMock {
	return createMockGasScheduleNotifierWithCustomGasSchedule(func(gasMap wasmConfig.GasScheduleMap) {})
}

func createMockGasScheduleNotifierWithCustomGasSchedule(updateGasSchedule func(gasMap wasmConfig.GasScheduleMap)) *mock.GasScheduleNotifierMock {
	testGasSchedule := wasmConfig.MakeGasMapForTests()
	defaults.FillGasMapInternal(testGasSchedule, 1)
	updateGasSchedule(testGasSchedule)
	return mock.NewGasScheduleNotifierMock(testGasSchedule)
}

// CreatePreparedTxProcessorWithVMs -
func CreatePreparedTxProcessorWithVMs(enableEpochs config.EnableEpochs) (*VMTestContext, error) {
	return CreatePreparedTxProcessorWithVMsAndCustomGasSchedule(enableEpochs, func(gasMap wasmConfig.GasScheduleMap) {})
}

// CreatePreparedTxProcessorWithVMsAndCustomGasSchedule -
func CreatePreparedTxProcessorWithVMsAndCustomGasSchedule(
	enableEpochs config.EnableEpochs,
	updateGasSchedule func(gasMap wasmConfig.GasScheduleMap)) (*VMTestContext, error) {
	return CreatePreparedTxProcessorWithVMsWithShardCoordinatorDBAndGas(
		enableEpochs,
		mock.NewMultiShardsCoordinatorMock(2),
		integrationtests.CreateMemUnit(),
		createMockGasScheduleNotifierWithCustomGasSchedule(updateGasSchedule),
	)
}

// CreatePreparedTxProcessorWithVMsWithShardCoordinator -
func CreatePreparedTxProcessorWithVMsWithShardCoordinator(enableEpochsConfig config.EnableEpochs, shardCoordinator sharding.Coordinator) (*VMTestContext, error) {
	return CreatePreparedTxProcessorWithVMsWithShardCoordinatorDBAndGas(
		enableEpochsConfig,
		shardCoordinator,
		integrationtests.CreateMemUnit(),
		createMockGasScheduleNotifier(),
	)
}

// CreatePreparedTxProcessorWithVMsWithShardCoordinatorDBAndGas -
func CreatePreparedTxProcessorWithVMsWithShardCoordinatorDBAndGas(
	enableEpochsConfig config.EnableEpochs,
	shardCoordinator sharding.Coordinator,
	db storage.Storer,
	gasScheduleNotifier core.GasScheduleNotifier,
) (*VMTestContext, error) {
	feeAccumulator, _ := postprocess.NewFeeAccumulator()
	accounts := integrationtests.CreateAccountsDB(db)
	vmConfig := createDefaultVMConfig()
	wasmVMChangeLocker := &sync.RWMutex{}

	epochNotifierInstance := forking.NewGenericEpochNotifier()
	enableEpochsHandler, _ := enablers.NewEnableEpochsHandler(enableEpochsConfig, epochNotifierInstance)
	chainHandler := &testscommon.ChainHandlerStub{}

	vmContainer, blockchainHook, pool := CreateVMAndBlockchainHookAndDataPool(
		accounts,
		gasScheduleNotifier,
		vmConfig,
		shardCoordinator,
		wasmVMChangeLocker,
		epochNotifierInstance,
		enableEpochsHandler,
		chainHandler,
	)
	res, err := CreateTxProcessorWithOneSCExecutorWithVMs(
		accounts,
		vmContainer,
		blockchainHook,
		feeAccumulator,
		shardCoordinator,
		enableEpochsConfig,
		wasmVMChangeLocker,
		pool,
		epochNotifierInstance,
	)
	if err != nil {
		return nil, err
	}

	return &VMTestContext{
		TxProcessor:         res.TxProc,
		ScProcessor:         res.SCProc,
		Accounts:            accounts,
		BlockchainHook:      blockchainHook,
		VMContainer:         vmContainer,
		TxFeeHandler:        feeAccumulator,
		ScForwarder:         res.IntermediateTxProc,
		ShardCoordinator:    shardCoordinator,
		EconomicsData:       res.EconomicsHandler,
		TxCostHandler:       res.CostHandler,
		TxsLogsProcessor:    res.TxLogProc,
		GasSchedule:         gasScheduleNotifier,
		EpochNotifier:       epochNotifierInstance,
		EnableEpochsHandler: enableEpochsHandler,
		ChainHandler:        chainHandler,
	}, nil
}

// CreateTxProcessorWasmVMWithGasSchedule -
func CreateTxProcessorWasmVMWithGasSchedule(
	senderNonce uint64,
	senderAddressBytes []byte,
	senderBalance *big.Int,
	gasScheduleMap map[string]map[string]uint64,
	enableEpochsConfig config.EnableEpochs,
) (*VMTestContext, error) {
	feeAccumulator, _ := postprocess.NewFeeAccumulator()
	accounts := integrationtests.CreateInMemoryShardAccountsDB()
	_, _ = CreateAccount(accounts, senderAddressBytes, senderNonce, senderBalance)
	vmConfig := createDefaultVMConfig()
	wasmVMChangeLocker := &sync.RWMutex{}

	gasScheduleNotifier := mock.NewGasScheduleNotifierMock(gasScheduleMap)
	epochNotifierInstance := forking.NewGenericEpochNotifier()
	enableEpochsHandler, _ := enablers.NewEnableEpochsHandler(enableEpochsConfig, epochNotifierInstance)
	chainHandler := &testscommon.ChainHandlerStub{}
	vmContainer, blockchainHook, pool := CreateVMAndBlockchainHookAndDataPool(
		accounts,
		gasScheduleNotifier,
		vmConfig,
		mock.NewMultiShardsCoordinatorMock(2),
		wasmVMChangeLocker,
		epochNotifierInstance,
		enableEpochsHandler,
		chainHandler,
	)
	res, err := CreateTxProcessorWithOneSCExecutorWithVMs(
		accounts,
		vmContainer,
		blockchainHook,
		feeAccumulator,
		mock.NewMultiShardsCoordinatorMock(2),
		enableEpochsConfig,
		wasmVMChangeLocker,
		pool,
		epochNotifierInstance,
	)
	if err != nil {
		return nil, err
	}

	return &VMTestContext{
		TxProcessor:         res.TxProc,
		ScProcessor:         res.SCProc,
		Accounts:            accounts,
		BlockchainHook:      blockchainHook,
		VMContainer:         vmContainer,
		TxFeeHandler:        feeAccumulator,
		ScForwarder:         res.IntermediateTxProc,
		GasSchedule:         gasScheduleNotifier,
		EpochNotifier:       epochNotifierInstance,
		EnableEpochsHandler: enableEpochsHandler,
		ChainHandler:        chainHandler,
	}, nil
}

// CreateTxProcessorWasmVMWithVMConfig -
func CreateTxProcessorWasmVMWithVMConfig(
	enableEpochsConfig config.EnableEpochs,
	vmConfig *config.VirtualMachineConfig,
	gasSchedule map[string]map[string]uint64,
) (*VMTestContext, error) {
	feeAccumulator, _ := postprocess.NewFeeAccumulator()
	accounts := integrationtests.CreateInMemoryShardAccountsDB()
	wasmVMChangeLocker := &sync.RWMutex{}
	gasScheduleNotifier := mock.NewGasScheduleNotifierMock(gasSchedule)
	epochNotifierInstance := forking.NewGenericEpochNotifier()
	enableEpochsHandler, _ := enablers.NewEnableEpochsHandler(enableEpochsConfig, epochNotifierInstance)
	chainHandler := &testscommon.ChainHandlerStub{}
	vmContainer, blockchainHook, pool := CreateVMAndBlockchainHookAndDataPool(
		accounts,
		gasScheduleNotifier,
		vmConfig,
		mock.NewMultiShardsCoordinatorMock(2),
		wasmVMChangeLocker,
		epochNotifierInstance,
		enableEpochsHandler,
		chainHandler,
	)
	res, err := CreateTxProcessorWithOneSCExecutorWithVMs(
		accounts,
		vmContainer,
		blockchainHook,
		feeAccumulator,
		mock.NewMultiShardsCoordinatorMock(2),
		enableEpochsConfig,
		wasmVMChangeLocker,
		pool,
		epochNotifierInstance,
	)
	if err != nil {
		return nil, err
	}

	return &VMTestContext{
		TxProcessor:         res.TxProc,
		ScProcessor:         res.SCProc,
		Accounts:            accounts,
		BlockchainHook:      blockchainHook,
		VMContainer:         vmContainer,
		TxFeeHandler:        feeAccumulator,
		ScForwarder:         res.IntermediateTxProc,
		GasSchedule:         gasScheduleNotifier,
		VMConfiguration:     vmConfig,
		EpochNotifier:       epochNotifierInstance,
		EnableEpochsHandler: enableEpochsHandler,
		ChainHandler:        chainHandler,
	}, nil
}

// CreatePreparedTxProcessorAndAccountsWithMockedVM -
func CreatePreparedTxProcessorAndAccountsWithMockedVM(
	vmOpGas uint64,
	senderNonce uint64,
	senderAddressBytes []byte,
	senderBalance *big.Int,
	enableEpochs config.EnableEpochs,
	wasmVMChangeLocker common.Locker,
) (process.TransactionProcessor, state.AccountsAdapter, error) {

	accnts := integrationtests.CreateInMemoryShardAccountsDB()
	_, _ = CreateAccount(accnts, senderAddressBytes, senderNonce, senderBalance)

	txProcessor, err := CreateTxProcessorWithOneSCExecutorMockVM(accnts, vmOpGas, enableEpochs, wasmVMChangeLocker)
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
	epochNotifierInstance := forking.NewGenericEpochNotifier()
	enableEpochsHandler, _ := enablers.NewEnableEpochsHandler(config.EnableEpochs{}, epochNotifierInstance)
	vmContainer, blockChainHook, _ := CreateVMAndBlockchainHookAndDataPool(
		accnts,
		gasScheduleNotifier,
		vmConfig,
		mock.NewMultiShardsCoordinatorMock(2),
		&sync.RWMutex{},
		epochNotifierInstance,
		enableEpochsHandler,
		&testscommon.ChainHandlerStub{},
	)
	defer func() {
		_ = vmContainer.Close()
	}()

	feeHandler := &mock.FeeHandlerStub{
		MaxGasLimitPerBlockCalled: func() uint64 {
			return uint64(math.MaxUint64)
		},
	}

	argsNewSCQueryService := smartContract.ArgsNewSCQueryService{
		VmContainer:              vmContainer,
		EconomicsFee:             feeHandler,
		BlockChainHook:           blockChainHook,
		BlockChain:               &testscommon.ChainHandlerStub{},
		WasmVMChangeLocker:       &sync.RWMutex{},
		Bootstrapper:             syncDisabled.NewDisabledBootstrapper(),
		AllowExternalQueriesChan: common.GetClosedUnbufferedChannel(),
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
	vmContainer, blockChainHook, _ := CreateVMAndBlockchainHookAndDataPool(
		testContext.Accounts,
		gasScheduleNotifier,
		vmConfig,
		mock.NewMultiShardsCoordinatorMock(2),
		&sync.RWMutex{},
		testContext.EpochNotifier,
		testContext.EnableEpochsHandler,
		&testscommon.ChainHandlerStub{},
	)
	defer func() {
		_ = vmContainer.Close()
	}()

	argsNewSCQueryService := smartContract.ArgsNewSCQueryService{
		VmContainer:    vmContainer,
		EconomicsFee:   testContext.EconomicsData,
		BlockChainHook: blockChainHook,
		BlockChain: &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.Header{
					ShardID: testContext.ShardCoordinator.SelfId(),
				}
			},
		},
		WasmVMChangeLocker:       &sync.RWMutex{},
		Bootstrapper:             syncDisabled.NewDisabledBootstrapper(),
		AllowExternalQueriesChan: common.GetClosedUnbufferedChannel(),
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
func CreatePreparedTxProcessorWithVMsMultiShard(selfShardID uint32, enableEpochsConfig config.EnableEpochs) (*VMTestContext, error) {
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(3, selfShardID)

	feeAccumulator, _ := postprocess.NewFeeAccumulator()
	accounts := integrationtests.CreateInMemoryShardAccountsDB()

	wasmVMChangeLocker := &sync.RWMutex{}
	var vmContainer process.VirtualMachinesContainer
	var blockchainHook *hooks.BlockChainHookImpl
	epochNotifierInstance := forking.NewGenericEpochNotifier()
	enableEpochsHandler, _ := enablers.NewEnableEpochsHandler(enableEpochsConfig, epochNotifierInstance)
	chainHandler := &testscommon.ChainHandlerStub{}
	if selfShardID == core.MetachainShardId {
		vmContainer, blockchainHook = CreateVMAndBlockchainHookMeta(accounts, nil, shardCoordinator, enableEpochsConfig)
	} else {
		vmConfig := createDefaultVMConfig()
		vmContainer, blockchainHook, _ = CreateVMAndBlockchainHookAndDataPool(
			accounts,
			nil,
			vmConfig,
			shardCoordinator,
			wasmVMChangeLocker,
			epochNotifierInstance,
			enableEpochsHandler,
			chainHandler,
		)
	}

	res, err := CreateTxProcessorWithOneSCExecutorWithVMs(
		accounts,
		vmContainer,
		blockchainHook,
		feeAccumulator,
		shardCoordinator,
		enableEpochsConfig,
		wasmVMChangeLocker,
		nil,
		epochNotifierInstance,
	)
	if err != nil {
		return nil, err
	}

	return &VMTestContext{
		TxProcessor:         res.TxProc,
		ScProcessor:         res.SCProc,
		Accounts:            accounts,
		BlockchainHook:      blockchainHook,
		VMContainer:         vmContainer,
		TxFeeHandler:        feeAccumulator,
		ShardCoordinator:    shardCoordinator,
		ScForwarder:         res.IntermediateTxProc,
		EconomicsData:       res.EconomicsHandler,
		Marshalizer:         integrationtests.TestMarshalizer,
		TxsLogsProcessor:    res.TxLogProc,
		EpochNotifier:       epochNotifierInstance,
		EnableEpochsHandler: enableEpochsHandler,
		ChainHandler:        chainHandler,
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
