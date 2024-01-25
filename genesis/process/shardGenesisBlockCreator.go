package process

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	disabledCommon "github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/common/enablers"
	"github.com/multiversx/mx-chain-go/common/forking"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	customErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/process/disabled"
	"github.com/multiversx/mx-chain-go/genesis/process/intermediate"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/factory/shard"
	disabledGuardian "github.com/multiversx/mx-chain-go/process/guardian/disabled"
	"github.com/multiversx/mx-chain-go/process/receipts"
	"github.com/multiversx/mx-chain-go/process/rewardTransaction"
	"github.com/multiversx/mx-chain-go/process/smartContract"
	"github.com/multiversx/mx-chain-go/process/smartContract/builtInFunctions"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks"
	"github.com/multiversx/mx-chain-go/process/smartContract/hooks/counters"
	"github.com/multiversx/mx-chain-go/process/smartContract/processorV2/scrProcFactory"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
	syncDisabled "github.com/multiversx/mx-chain-go/process/sync/disabled"
	"github.com/multiversx/mx-chain-go/process/transaction"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	"github.com/multiversx/mx-chain-go/update"
	hardForkProcess "github.com/multiversx/mx-chain-go/update/process"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-vm-common-go/parsers"
)

var log = logger.GetOrCreate("genesis/process")

var zero = big.NewInt(0)

type deployedScMetrics struct {
	numDelegation int
	numOtherTypes int
}

func createGenesisConfig() config.EnableEpochs {
	blsMultiSignerEnableEpoch := []config.MultiSignerConfig{
		{
			EnableEpoch: 0,
			Type:        "no-KOSK",
		},
	}

	return config.EnableEpochs{
		SCDeployEnableEpoch:                    unreachableEpoch,
		BuiltInFunctionsEnableEpoch:            0,
		RelayedTransactionsEnableEpoch:         unreachableEpoch,
		PenalizedTooMuchGasEnableEpoch:         unreachableEpoch,
		SwitchJailWaitingEnableEpoch:           unreachableEpoch,
		SwitchHysteresisForMinNodesEnableEpoch: unreachableEpoch,
		BelowSignedThresholdEnableEpoch:        unreachableEpoch,
		TransactionSignedWithTxHashEnableEpoch: unreachableEpoch,
		MetaProtectionEnableEpoch:              unreachableEpoch,
		AheadOfTimeGasUsageEnableEpoch:         unreachableEpoch,
		GasPriceModifierEnableEpoch:            unreachableEpoch,
		RepairCallbackEnableEpoch:              unreachableEpoch,
		MaxNodesChangeEnableEpoch: []config.MaxNodesChangeConfig{
			{
				EpochEnable:            unreachableEpoch,
				MaxNumNodes:            0,
				NodesToShufflePerShard: 0,
			},
		},
		BlockGasAndFeesReCheckEnableEpoch:                 unreachableEpoch,
		StakingV2EnableEpoch:                              unreachableEpoch,
		StakeEnableEpoch:                                  unreachableEpoch, // no need to enable this, we have builtin exceptions in staking system SC
		DoubleKeyProtectionEnableEpoch:                    0,
		ESDTEnableEpoch:                                   unreachableEpoch,
		GovernanceEnableEpoch:                             unreachableEpoch,
		DelegationManagerEnableEpoch:                      unreachableEpoch,
		DelegationSmartContractEnableEpoch:                unreachableEpoch,
		CorrectLastUnjailedEnableEpoch:                    unreachableEpoch,
		BalanceWaitingListsEnableEpoch:                    unreachableEpoch,
		ReturnDataToLastTransferEnableEpoch:               unreachableEpoch,
		SenderInOutTransferEnableEpoch:                    unreachableEpoch,
		RelayedTransactionsV2EnableEpoch:                  unreachableEpoch,
		UnbondTokensV2EnableEpoch:                         unreachableEpoch,
		SaveJailedAlwaysEnableEpoch:                       unreachableEpoch,
		ValidatorToDelegationEnableEpoch:                  unreachableEpoch,
		ReDelegateBelowMinCheckEnableEpoch:                unreachableEpoch,
		WaitingListFixEnableEpoch:                         unreachableEpoch,
		IncrementSCRNonceInMultiTransferEnableEpoch:       unreachableEpoch,
		ESDTMultiTransferEnableEpoch:                      unreachableEpoch,
		GlobalMintBurnDisableEpoch:                        unreachableEpoch,
		ESDTTransferRoleEnableEpoch:                       unreachableEpoch,
		BuiltInFunctionOnMetaEnableEpoch:                  unreachableEpoch,
		ComputeRewardCheckpointEnableEpoch:                unreachableEpoch,
		SCRSizeInvariantCheckEnableEpoch:                  unreachableEpoch,
		BackwardCompSaveKeyValueEnableEpoch:               unreachableEpoch,
		ESDTNFTCreateOnMultiShardEnableEpoch:              unreachableEpoch,
		MetaESDTSetEnableEpoch:                            unreachableEpoch,
		AddTokensToDelegationEnableEpoch:                  unreachableEpoch,
		MultiESDTTransferFixOnCallBackOnEnableEpoch:       unreachableEpoch,
		OptimizeGasUsedInCrossMiniBlocksEnableEpoch:       unreachableEpoch,
		CorrectFirstQueuedEpoch:                           unreachableEpoch,
		CorrectJailedNotUnstakedEmptyQueueEpoch:           unreachableEpoch,
		FixOOGReturnCodeEnableEpoch:                       unreachableEpoch,
		RemoveNonUpdatedStorageEnableEpoch:                unreachableEpoch,
		DeleteDelegatorAfterClaimRewardsEnableEpoch:       unreachableEpoch,
		OptimizeNFTStoreEnableEpoch:                       unreachableEpoch,
		CreateNFTThroughExecByCallerEnableEpoch:           unreachableEpoch,
		StopDecreasingValidatorRatingWhenStuckEnableEpoch: unreachableEpoch,
		FrontRunningProtectionEnableEpoch:                 unreachableEpoch,
		IsPayableBySCEnableEpoch:                          unreachableEpoch,
		CleanUpInformativeSCRsEnableEpoch:                 unreachableEpoch,
		StorageAPICostOptimizationEnableEpoch:             unreachableEpoch,
		TransformToMultiShardCreateEnableEpoch:            unreachableEpoch,
		ESDTRegisterAndSetAllRolesEnableEpoch:             unreachableEpoch,
		ScheduledMiniBlocksEnableEpoch:                    unreachableEpoch,
		FailExecutionOnEveryAPIErrorEnableEpoch:           unreachableEpoch,
		AddFailedRelayedTxToInvalidMBsDisableEpoch:        unreachableEpoch,
		SCRSizeInvariantOnBuiltInResultEnableEpoch:        unreachableEpoch,
		ManagedCryptoAPIsEnableEpoch:                      unreachableEpoch,
		CheckCorrectTokenIDForTransferRoleEnableEpoch:     unreachableEpoch,
		DisableExecByCallerEnableEpoch:                    unreachableEpoch,
		RefactorContextEnableEpoch:                        unreachableEpoch,
		CheckFunctionArgumentEnableEpoch:                  unreachableEpoch,
		CheckExecuteOnReadOnlyEnableEpoch:                 unreachableEpoch,
		MiniBlockPartialExecutionEnableEpoch:              unreachableEpoch,
		ESDTMetadataContinuousCleanupEnableEpoch:          unreachableEpoch,
		FixAsyncCallBackArgsListEnableEpoch:               unreachableEpoch,
		FixOldTokenLiquidityEnableEpoch:                   unreachableEpoch,
		SetSenderInEeiOutputTransferEnableEpoch:           unreachableEpoch,
		RefactorPeersMiniBlocksEnableEpoch:                unreachableEpoch,
		SCProcessorV2EnableEpoch:                          unreachableEpoch,
		DoNotReturnOldBlockInBlockchainHookEnableEpoch:    unreachableEpoch,
		MaxBlockchainHookCountersEnableEpoch:              unreachableEpoch,
		BLSMultiSignerEnableEpoch:                         blsMultiSignerEnableEpoch,
		SetGuardianEnableEpoch:                            unreachableEpoch,
		ScToScLogEventEnableEpoch:                         unreachableEpoch,
	}
}

func createSovGenesisConfig() config.EnableEpochs {
	cfg := createGenesisConfig()
	cfg.ESDTMultiTransferEnableEpoch = 0

	return cfg
}

func createGenesisRoundConfig() *config.RoundConfig {
	return &config.RoundConfig{
		RoundActivations: map[string]config.ActivationRoundByName{
			"DisableAsyncCallV1": {
				Round: strconv.FormatUint(unreachableRound, 10),
			},
		},
	}
}

// CreateShardGenesisBlock will create a shard genesis block
func CreateShardGenesisBlock(
	arg ArgsGenesisBlockCreator,
	body *block.Body,
	nodesListSplitter genesis.NodesListSplitter,
	hardForkBlockProcessor update.HardForkBlockProcessor,
) (data.HeaderHandler, [][]byte, *genesis.IndexingData, error) {
	if mustDoHardForkImportProcess(arg) {
		return createShardGenesisBlockAfterHardFork(arg, body, hardForkBlockProcessor)
	}

	processors, err := createProcessorsForShardGenesisBlock(arg, createSovGenesisConfig(), createGenesisRoundConfig())
	if err != nil {
		return nil, nil, nil, err
	}

	return baseCreateShardGenesisBlock(arg, nodesListSplitter, processors)
}

func baseCreateShardGenesisBlock(
	arg ArgsGenesisBlockCreator,
	nodesListSplitter genesis.NodesListSplitter,
	processors *genesisProcessors,
) (data.HeaderHandler, [][]byte, *genesis.IndexingData, error) {

	indexingData := &genesis.IndexingData{
		DelegationTxs:      make([]data.TransactionHandler, 0),
		ScrsTxs:            make(map[string]data.TransactionHandler),
		StakingTxs:         make([]data.TransactionHandler, 0),
		DeploySystemScTxs:  make([]data.TransactionHandler, 0),
		DeployInitialScTxs: make([]data.TransactionHandler, 0),
	}

	deployMetrics := &deployedScMetrics{}

	scAddresses, scTxs, err := deployInitialSmartContracts(processors, arg, deployMetrics)
	if err != nil {
		return nil, nil, nil, err
	}
	indexingData.DeployInitialScTxs = scTxs

	numSetBalances, err := setBalancesToTrie(arg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while setting the balances to trie",
			err, arg.ShardCoordinator.SelfId())
	}

	numStaked, err := increaseStakersNonces(processors, arg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while incrementing nonces",
			err, arg.ShardCoordinator.SelfId())
	}

	delegationResult, delegationTxs, err := executeDelegation(processors, arg, nodesListSplitter)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while execution delegation",
			err, arg.ShardCoordinator.SelfId())
	}
	indexingData.DelegationTxs = delegationTxs

	numCrossShardDelegations, err := incrementNoncesForCrossShardDelegations(processors, arg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while incrementing crossshard nonce",
			err, arg.ShardCoordinator.SelfId())
	}

	scrsTxs := processors.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)
	indexingData.ScrsTxs = scrsTxs

	rootHash, err := arg.Accounts.Commit()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while commiting",
			err, arg.ShardCoordinator.SelfId())
	}

	log.Debug("shard block genesis",
		"shard ID", arg.ShardCoordinator.SelfId(),
		"num delegation SC deployed", deployMetrics.numDelegation,
		"num other SC deployed", deployMetrics.numOtherTypes,
		"num set balances", numSetBalances,
		"num staked directly", numStaked,
		"total staked on a delegation SC", delegationResult.NumTotalStaked,
		"total delegation nodes", delegationResult.NumTotalDelegated,
		"cross shard delegation calls", numCrossShardDelegations,
		"resulted roothash", rootHash,
	)

	round, nonce, epoch := getGenesisBlocksRoundNonceEpoch(arg)
	header := &block.Header{
		Epoch:           epoch,
		Round:           round,
		Nonce:           nonce,
		ShardID:         arg.ShardCoordinator.SelfId(),
		BlockBodyType:   block.StateBlock,
		PubKeysBitmap:   []byte{1},
		Signature:       rootHash,
		RootHash:        rootHash,
		PrevRandSeed:    rootHash,
		RandSeed:        rootHash,
		TimeStamp:       arg.GenesisTime,
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		ChainID:         []byte(arg.Core.ChainID()),
		SoftwareVersion: []byte(""),
	}

	err = processors.vmContainer.Close()
	if err != nil {
		return nil, nil, nil, err
	}

	err = processors.vmContainersFactory.Close()
	if err != nil {
		return nil, nil, nil, err
	}

	return header, scAddresses, indexingData, nil
}

func createShardGenesisBlockAfterHardFork(
	arg ArgsGenesisBlockCreator,
	body *block.Body,
	hardForkBlockProcessor update.HardForkBlockProcessor,
) (data.HeaderHandler, [][]byte, *genesis.IndexingData, error) {
	if check.IfNil(hardForkBlockProcessor) {
		return nil, nil, nil, update.ErrNilHardForkBlockProcessor
	}

	hdrHandler, err := hardForkBlockProcessor.CreateBlock(
		body,
		arg.Core.ChainID(),
		arg.HardForkConfig.StartRound,
		arg.HardForkConfig.StartNonce,
		arg.HardForkConfig.StartEpoch,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	err = hdrHandler.SetTimeStamp(arg.GenesisTime)
	if err != nil {
		return nil, nil, nil, err
	}

	err = arg.Accounts.RecreateTrie(hdrHandler.GetRootHash())
	if err != nil {
		return nil, nil, nil, err
	}

	indexingData := &genesis.IndexingData{
		DelegationTxs:      make([]data.TransactionHandler, 0),
		ScrsTxs:            make(map[string]data.TransactionHandler),
		StakingTxs:         make([]data.TransactionHandler, 0),
		DeploySystemScTxs:  make([]data.TransactionHandler, 0),
		DeployInitialScTxs: make([]data.TransactionHandler, 0),
	}

	return hdrHandler, make([][]byte, 0), indexingData, nil
}

func createArgsShardBlockCreatorAfterHardFork(
	arg ArgsGenesisBlockCreator,
	selfShardID uint32,
) (hardForkProcess.ArgsNewShardBlockCreatorAfterHardFork, error) {
	tmpArg := arg
	tmpArg.Accounts = arg.importHandler.GetAccountsDBForShard(arg.ShardCoordinator.SelfId())
	processors, err := createProcessorsForShardGenesisBlock(tmpArg, arg.EpochConfig.EnableEpochs, arg.RoundConfig)
	if err != nil {
		return hardForkProcess.ArgsNewShardBlockCreatorAfterHardFork{}, err
	}

	argsPendingTxProcessor := hardForkProcess.ArgsPendingTransactionProcessor{
		Accounts:         tmpArg.Accounts,
		TxProcessor:      processors.txProcessor,
		RwdTxProcessor:   processors.rwdProcessor,
		ScrTxProcessor:   processors.scrProcessor,
		PubKeyConv:       arg.Core.AddressPubKeyConverter(),
		ShardCoordinator: arg.ShardCoordinator,
	}
	pendingTxProcessor, err := hardForkProcess.NewPendingTransactionProcessor(argsPendingTxProcessor)
	if err != nil {
		return hardForkProcess.ArgsNewShardBlockCreatorAfterHardFork{}, err
	}

	receiptsRepository, err := receipts.NewReceiptsRepository(receipts.ArgsNewReceiptsRepository{
		Marshaller: arg.Core.InternalMarshalizer(),
		Hasher:     arg.Core.Hasher(),
		Store:      arg.Data.StorageService(),
	})
	if err != nil {
		return hardForkProcess.ArgsNewShardBlockCreatorAfterHardFork{}, err
	}

	argsShardBlockCreatorAfterHardFork := hardForkProcess.ArgsNewShardBlockCreatorAfterHardFork{
		Hasher:             arg.Core.Hasher(),
		ImportHandler:      arg.importHandler,
		Marshalizer:        arg.Core.InternalMarshalizer(),
		PendingTxProcessor: pendingTxProcessor,
		ShardCoordinator:   arg.ShardCoordinator,
		Storage:            arg.Data.StorageService(),
		TxCoordinator:      processors.txCoordinator,
		ReceiptsRepository: receiptsRepository,
		SelfShardID:        selfShardID,
	}

	return argsShardBlockCreatorAfterHardFork, nil
}

// setBalancesToTrie adds balances to trie
func setBalancesToTrie(arg ArgsGenesisBlockCreator) (int, error) {
	initialAccounts, err := arg.AccountsParser.InitialAccountsSplitOnAddressesShards(arg.ShardCoordinator)
	if err != nil {
		return 0, err
	}

	initialAccountsForShard := initialAccounts[arg.ShardCoordinator.SelfId()]

	for _, accnt := range initialAccountsForShard {
		err = setBalanceToTrie(arg, accnt)
		if err != nil {
			return 0, err
		}
	}

	return len(initialAccountsForShard), nil
}

func setBalanceToTrie(arg ArgsGenesisBlockCreator, accnt genesis.InitialAccountHandler) error {
	accWrp, err := arg.Accounts.LoadAccount(accnt.AddressBytes())
	if err != nil {
		return err
	}

	account, ok := accWrp.(state.UserAccountHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	err = account.AddToBalance(accnt.GetBalanceValue())
	if err != nil {
		return err
	}

	return arg.Accounts.SaveAccount(account)
}

func createProcessorsForShardGenesisBlock(arg ArgsGenesisBlockCreator, enableEpochsConfig config.EnableEpochs, roundConfig *config.RoundConfig) (*genesisProcessors, error) {
	genesisWasmVMLocker := &sync.RWMutex{} // use a local instance as to not run in concurrent issues when doing bootstrap
	epochNotifier := forking.NewGenericEpochNotifier()
	enableEpochsHandler, err := enablers.NewEnableEpochsHandler(enableEpochsConfig, epochNotifier)
	if err != nil {
		return nil, err
	}

	roundNotifier := forking.NewGenericRoundNotifier()
	enableRoundsHandler, err := enablers.NewEnableRoundsHandler(*roundConfig, roundNotifier)
	if err != nil {
		return nil, err
	}

	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasSchedule:               arg.GasSchedule,
		MapDNSAddresses:           make(map[string]struct{}),
		MapDNSV2Addresses:         make(map[string]struct{}),
		EnableUserNameChange:      false,
		Marshalizer:               arg.Core.InternalMarshalizer(),
		Accounts:                  arg.Accounts,
		ShardCoordinator:          arg.ShardCoordinator,
		EpochNotifier:             epochNotifier,
		EnableEpochsHandler:       enableEpochsHandler,
		AutomaticCrawlerAddresses: [][]byte{make([]byte, 32)},
		MaxNumNodesInTransferRole: math.MaxUint32,
		GuardedAccountHandler:     disabledGuardian.NewDisabledGuardedAccountHandler(),
	}
	builtInFuncFactory, err := builtInFunctions.CreateBuiltInFunctionsFactory(argsBuiltIn)
	if err != nil {
		return nil, err
	}

	argsHook := hooks.ArgBlockChainHook{
		Accounts:                 arg.Accounts,
		PubkeyConv:               arg.Core.AddressPubKeyConverter(),
		StorageService:           arg.Data.StorageService(),
		BlockChain:               arg.Data.Blockchain(),
		ShardCoordinator:         arg.ShardCoordinator,
		Marshalizer:              arg.Core.InternalMarshalizer(),
		Uint64Converter:          arg.Core.Uint64ByteSliceConverter(),
		BuiltInFunctions:         builtInFuncFactory.BuiltInFunctionContainer(),
		NFTStorageHandler:        builtInFuncFactory.NFTStorageHandler(),
		GlobalSettingsHandler:    builtInFuncFactory.ESDTGlobalSettingsHandler(),
		DataPool:                 arg.Data.Datapool(),
		CompiledSCPool:           arg.Data.Datapool().SmartContracts(),
		EpochNotifier:            epochNotifier,
		EnableEpochsHandler:      enableEpochsHandler,
		NilCompiledSCStore:       true,
		GasSchedule:              arg.GasSchedule,
		Counter:                  counters.NewDisabledCounter(),
		MissingTrieNodesNotifier: syncer.NewMissingTrieNodesNotifier(),
	}
	esdtTransferParser, err := parsers.NewESDTTransferParser(arg.Core.InternalMarshalizer())
	if err != nil {
		return nil, err
	}

	blockChainHookImpl, err := hooks.CreateBlockChainHook(arg.ChainRunType, argsHook)
	if err != nil {
		return nil, err
	}

	argsNewVMFactory := shard.ArgVMContainerFactory{
		Config:              arg.VirtualMachineConfig,
		BlockGasLimit:       math.MaxUint64,
		GasSchedule:         arg.GasSchedule,
		BlockChainHook:      blockChainHookImpl,
		EpochNotifier:       epochNotifier,
		EnableEpochsHandler: enableEpochsHandler,
		WasmVMChangeLocker:  genesisWasmVMLocker,
		ESDTTransferParser:  esdtTransferParser,
		BuiltInFunctions:    argsHook.BuiltInFunctions,
		Hasher:              arg.Core.Hasher(),
	}
	vmFactoryImpl, err := shard.NewVMContainerFactory(argsNewVMFactory)
	if err != nil {
		return nil, err
	}

	vmContainer, err := vmFactoryImpl.Create()
	if err != nil {
		return nil, err
	}

	err = blockChainHookImpl.SetVMContainer(vmContainer)
	if err != nil {
		return nil, err
	}

	err = builtInFuncFactory.SetPayableHandler(vmFactoryImpl.BlockChainHookImpl())
	if err != nil {
		return nil, err
	}

	genesisFeeHandler := &disabled.FeeHandler{}
	argsFactory := shard.ArgsNewIntermediateProcessorsContainerFactory{
		ShardCoordinator:        arg.ShardCoordinator,
		Marshalizer:             arg.Core.InternalMarshalizer(),
		Hasher:                  arg.Core.Hasher(),
		PubkeyConverter:         arg.Core.AddressPubKeyConverter(),
		Store:                   arg.Data.StorageService(),
		PoolsHolder:             arg.Data.Datapool(),
		EconomicsFee:            genesisFeeHandler,
		EnableEpochsHandler:     enableEpochsHandler,
		TxExecutionOrderHandler: arg.TxExecutionOrderHandler,
	}
	interimProcFactory, err := shard.NewIntermediateProcessorsContainerFactory(argsFactory)
	if err != nil {
		return nil, err
	}

	interimProcContainer, err := interimProcFactory.Create()
	if err != nil {
		return nil, err
	}

	scForwarder, err := interimProcContainer.Get(dataBlock.SmartContractResultBlock)
	if err != nil {
		return nil, err
	}
	scForwarder.GetCreatedInShardMiniBlock()

	receiptTxInterim, err := interimProcContainer.Get(dataBlock.ReceiptBlock)
	if err != nil {
		return nil, err
	}

	badTxInterim, err := interimProcContainer.Get(dataBlock.InvalidBlock)
	if err != nil {
		return nil, err
	}

	temporaryBlock := &dataBlock.Header{
		Epoch:     arg.StartEpochNum,
		TimeStamp: arg.GenesisTime,
	}
	epochNotifier.CheckEpoch(temporaryBlock)

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:     arg.Core.AddressPubKeyConverter(),
		ShardCoordinator:    arg.ShardCoordinator,
		BuiltInFunctions:    builtInFuncFactory.BuiltInFunctionContainer(),
		ArgumentParser:      parsers.NewCallArgsParser(),
		ESDTTransferParser:  esdtTransferParser,
		EnableEpochsHandler: enableEpochsHandler,
	}
	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	if err != nil {
		return nil, err
	}

	gasHandler, err := preprocess.NewGasComputation(arg.Economics, txTypeHandler, enableEpochsHandler)
	if err != nil {
		return nil, err
	}

	argsNewScProcessor := scrCommon.ArgsNewSmartContractProcessor{
		VmContainer:         vmContainer,
		ArgsParser:          smartContract.NewArgumentParser(),
		Hasher:              arg.Core.Hasher(),
		Marshalizer:         arg.Core.InternalMarshalizer(),
		AccountsDB:          arg.Accounts,
		BlockChainHook:      vmFactoryImpl.BlockChainHookImpl(),
		BuiltInFunctions:    builtInFuncFactory.BuiltInFunctionContainer(),
		PubkeyConv:          arg.Core.AddressPubKeyConverter(),
		ShardCoordinator:    arg.ShardCoordinator,
		ScrForwarder:        scForwarder,
		TxFeeHandler:        genesisFeeHandler,
		EconomicsFee:        genesisFeeHandler,
		TxTypeHandler:       txTypeHandler,
		GasHandler:          gasHandler,
		GasSchedule:         arg.GasSchedule,
		TxLogsProcessor:     arg.TxLogsProcessor,
		BadTxForwarder:      badTxInterim,
		EnableRoundsHandler: enableRoundsHandler,
		EnableEpochsHandler: enableEpochsHandler,
		IsGenesisProcessing: true,
		VMOutputCacher:      txcache.NewDisabledCache(),
		WasmVMChangeLocker:  genesisWasmVMLocker,
	}

	scProcessorProxy, err := scrProcFactory.CreateSCRProcessor(arg.ChainRunType, argsNewScProcessor, epochNotifier)
	if err != nil {
		return nil, err
	}

	rewardsTxProcessor, err := rewardTransaction.NewRewardTxProcessor(
		arg.Accounts,
		arg.Core.AddressPubKeyConverter(),
		arg.ShardCoordinator,
	)
	if err != nil {
		return nil, err
	}

	argsNewTxProcessor := transaction.ArgsNewTxProcessor{
		Accounts:            arg.Accounts,
		Hasher:              arg.Core.Hasher(),
		PubkeyConv:          arg.Core.AddressPubKeyConverter(),
		Marshalizer:         arg.Core.InternalMarshalizer(),
		SignMarshalizer:     arg.Core.TxMarshalizer(),
		ShardCoordinator:    arg.ShardCoordinator,
		ScProcessor:         scProcessorProxy,
		TxFeeHandler:        genesisFeeHandler,
		TxTypeHandler:       txTypeHandler,
		EconomicsFee:        genesisFeeHandler,
		ReceiptForwarder:    receiptTxInterim,
		BadTxForwarder:      badTxInterim,
		ArgsParser:          smartContract.NewArgumentParser(),
		ScrForwarder:        scForwarder,
		EnableRoundsHandler: enableRoundsHandler,
		EnableEpochsHandler: enableEpochsHandler,
		TxVersionChecker:    arg.Core.TxVersionChecker(),
		GuardianChecker:     disabledGuardian.NewDisabledGuardedAccountHandler(),
		TxLogsProcessor:     arg.TxLogsProcessor,
	}
	transactionProcessor, err := transaction.NewTxProcessor(argsNewTxProcessor)
	if err != nil {
		return nil, errors.New("could not create transaction processor: " + err.Error())
	}

	disabledRequestHandler := &disabled.RequestHandler{}
	disabledBlockTracker := &disabled.BlockTracker{}
	disabledBlockSizeComputationHandler := &disabled.BlockSizeComputationHandler{}
	disabledBalanceComputationHandler := &disabled.BalanceComputationHandler{}
	disabledScheduledTxsExecutionHandler := &disabled.ScheduledTxsExecutionHandler{}
	disabledProcessedMiniBlocksTracker := &disabled.ProcessedMiniBlocksTracker{}

	argsPreProc := shard.ArgPreProcessorsContainerFactory{
		ShardCoordinator:             arg.ShardCoordinator,
		Store:                        arg.Data.StorageService(),
		Marshaller:                   arg.Core.InternalMarshalizer(),
		Hasher:                       arg.Core.Hasher(),
		DataPool:                     arg.Data.Datapool(),
		PubkeyConverter:              arg.Core.AddressPubKeyConverter(),
		Accounts:                     arg.Accounts,
		RequestHandler:               disabledRequestHandler,
		TxProcessor:                  transactionProcessor,
		ScProcessor:                  scProcessorProxy,
		ScResultProcessor:            scProcessorProxy,
		RewardsTxProcessor:           rewardsTxProcessor,
		EconomicsFee:                 arg.Economics,
		GasHandler:                   gasHandler,
		BlockTracker:                 disabledBlockTracker,
		BlockSizeComputation:         disabledBlockSizeComputationHandler,
		BalanceComputation:           disabledBalanceComputationHandler,
		EnableEpochsHandler:          enableEpochsHandler,
		TxTypeHandler:                txTypeHandler,
		ScheduledTxsExecutionHandler: disabledScheduledTxsExecutionHandler,
		ProcessedMiniBlocksTracker:   disabledProcessedMiniBlocksTracker,
		TxExecutionOrderHandler:      arg.TxExecutionOrderHandler,
		TxPreProcessorCreator:        arg.TxPreprocessorCreator,
		ChainRunType:                 arg.ChainRunType,
	}
	preProcFactory, err := shard.NewPreProcessorsContainerFactory(argsPreProc)
	if err != nil {
		return nil, err
	}

	preProcContainer, err := preProcFactory.Create()
	if err != nil {
		return nil, err
	}

	argsDetector := coordinator.ArgsPrintDoubleTransactionsDetector{
		Marshaller:          arg.Core.InternalMarshalizer(),
		Hasher:              arg.Core.Hasher(),
		EnableEpochsHandler: enableEpochsHandler,
	}
	doubleTransactionsDetector, err := coordinator.NewPrintDoubleTransactionsDetector(argsDetector)
	if err != nil {
		return nil, err
	}

	argsTransactionCoordinator := coordinator.ArgTransactionCoordinator{
		Hasher:                       arg.Core.Hasher(),
		Marshalizer:                  arg.Core.InternalMarshalizer(),
		ShardCoordinator:             arg.ShardCoordinator,
		Accounts:                     arg.Accounts,
		MiniBlockPool:                arg.Data.Datapool().MiniBlocks(),
		RequestHandler:               disabledRequestHandler,
		PreProcessors:                preProcContainer,
		InterProcessors:              interimProcContainer,
		GasHandler:                   gasHandler,
		FeeHandler:                   genesisFeeHandler,
		BlockSizeComputation:         disabledBlockSizeComputationHandler,
		BalanceComputation:           disabledBalanceComputationHandler,
		EconomicsFee:                 genesisFeeHandler,
		TxTypeHandler:                txTypeHandler,
		TransactionsLogProcessor:     arg.TxLogsProcessor,
		EnableEpochsHandler:          enableEpochsHandler,
		ScheduledTxsExecutionHandler: disabledScheduledTxsExecutionHandler,
		DoubleTransactionsDetector:   doubleTransactionsDetector,
		ProcessedMiniBlocksTracker:   disabledProcessedMiniBlocksTracker,
		TxExecutionOrderHandler:      arg.TxExecutionOrderHandler,
	}
	txCoordinator, err := createTransactionCoordinator(argsTransactionCoordinator, arg.ChainRunType)
	if err != nil {
		return nil, err
	}

	apiBlockchain, err := blockchain.NewBlockChain(disabledCommon.NewAppStatusHandler())
	if err != nil {
		return nil, err
	}

	argsNewSCQueryService := smartContract.ArgsNewSCQueryService{
		VmContainer:              vmContainer,
		EconomicsFee:             arg.Economics,
		BlockChainHook:           vmFactoryImpl.BlockChainHookImpl(),
		MainBlockChain:           arg.Data.Blockchain(),
		APIBlockChain:            apiBlockchain,
		WasmVMChangeLocker:       genesisWasmVMLocker,
		Bootstrapper:             syncDisabled.NewDisabledBootstrapper(),
		AllowExternalQueriesChan: common.GetClosedUnbufferedChannel(),
		HistoryRepository:        arg.HistoryRepository,
		ShardCoordinator:         arg.ShardCoordinator,
		StorageService:           arg.Data.StorageService(),
		Marshaller:               arg.Core.InternalMarshalizer(),
		Hasher:                   arg.Core.Hasher(),
		Uint64ByteSliceConverter: arg.Core.Uint64ByteSliceConverter(),
	}
	queryService, err := smartContract.NewSCQueryService(argsNewSCQueryService)
	if err != nil {
		return nil, err
	}

	return &genesisProcessors{
		txCoordinator:       txCoordinator,
		systemSCs:           nil,
		txProcessor:         transactionProcessor,
		scProcessor:         scProcessorProxy,
		scrProcessor:        scProcessorProxy,
		rwdProcessor:        rewardsTxProcessor,
		blockchainHook:      vmFactoryImpl.BlockChainHookImpl(),
		queryService:        queryService,
		vmContainersFactory: vmFactoryImpl,
		vmContainer:         vmContainer,
	}, nil
}

func createTransactionCoordinator(
	argsTransactionCoordinator coordinator.ArgTransactionCoordinator,
	chainRunType common.ChainRunType,
) (process.TransactionCoordinator, error) {
	transactionCoordinator, err := coordinator.NewTransactionCoordinator(argsTransactionCoordinator)
	if err != nil {
		return nil, err
	}

	switch chainRunType {
	case common.ChainRunTypeRegular:
		return transactionCoordinator, nil
	case common.ChainRunTypeSovereign:
		return coordinator.NewSovereignChainTransactionCoordinator(transactionCoordinator)
	default:
		return nil, fmt.Errorf("%w type %v", customErrors.ErrUnimplementedChainRunType, chainRunType)
	}
}

func deployInitialSmartContracts(
	processors *genesisProcessors,
	arg ArgsGenesisBlockCreator,
	deployMetrics *deployedScMetrics,
) ([][]byte, []data.TransactionHandler, error) {
	smartContracts, err := arg.SmartContractParser.InitialSmartContractsSplitOnOwnersShards(arg.ShardCoordinator)
	if err != nil {
		return nil, nil, err
	}

	allScTxs := make([]data.TransactionHandler, 0)
	var scAddresses = make([][]byte, 0)
	currentShardSmartContracts := smartContracts[arg.ShardCoordinator.SelfId()]
	for _, sc := range currentShardSmartContracts {
		var scResulted [][]byte
		scResulted, scTxs, errDeploy := deployInitialSmartContract(processors, sc, arg, deployMetrics)
		if errDeploy != nil {
			return nil, nil, fmt.Errorf("%w for owner %s and filename %s",
				errDeploy, sc.GetOwner(), sc.GetFilename())
		}

		scAddresses = append(scAddresses, scResulted...)
		allScTxs = append(allScTxs, scTxs...)
	}

	return scAddresses, allScTxs, nil
}

func deployInitialSmartContract(
	processors *genesisProcessors,
	sc genesis.InitialSmartContractHandler,
	arg ArgsGenesisBlockCreator,
	deployMetrics *deployedScMetrics,
) ([][]byte, []data.TransactionHandler, error) {

	txExecutor, err := intermediate.NewTxExecutionProcessor(processors.txProcessor, arg.Accounts)
	if err != nil {
		return nil, nil, err
	}

	var deployProc genesis.DeployProcessor

	switch sc.GetType() {
	case genesis.DNSType:
		deployMetrics.numOtherTypes++
		argDeployLibrary := intermediate.ArgDeployLibrarySC{
			Executor:         txExecutor,
			PubkeyConv:       arg.Core.AddressPubKeyConverter(),
			BlockchainHook:   processors.blockchainHook,
			ShardCoordinator: arg.ShardCoordinator,
		}
		deployProc, err = intermediate.NewDeployLibrarySC(argDeployLibrary)
		if err != nil {
			return nil, nil, err
		}
	case genesis.DelegationType:
		deployMetrics.numDelegation++
		fallthrough
	default:
		argDeploy := intermediate.ArgDeployProcessor{
			Executor:       txExecutor,
			PubkeyConv:     arg.Core.AddressPubKeyConverter(),
			BlockchainHook: processors.blockchainHook,
			QueryService:   processors.queryService,
		}
		deployProc, err = intermediate.NewDeployProcessor(argDeploy)
		if err != nil {
			return nil, nil, err
		}
	}

	dpResult, err := deployProc.Deploy(sc)
	return dpResult, txExecutor.GetExecutedTransactions(), err
}

func increaseStakersNonces(processors *genesisProcessors, arg ArgsGenesisBlockCreator) (int, error) {
	txExecutor, err := intermediate.NewTxExecutionProcessor(processors.txProcessor, arg.Accounts)
	if err != nil {
		return 0, err
	}

	initialAddresses, err := arg.AccountsParser.InitialAccountsSplitOnAddressesShards(arg.ShardCoordinator)
	if err != nil {
		return 0, err
	}

	stakersCounter := 0
	initalAddressesInCurrentShard := initialAddresses[arg.ShardCoordinator.SelfId()]
	for _, ia := range initalAddressesInCurrentShard {
		if ia.GetStakingValue().Cmp(zero) < 1 {
			continue
		}

		numNodesStaked := big.NewInt(0).Set(ia.GetStakingValue())
		numNodesStaked.Div(numNodesStaked, arg.GenesisNodePrice)

		stakersCounter++
		err = txExecutor.AddNonce(ia.AddressBytes(), numNodesStaked.Uint64())
		if err != nil {
			return 0, fmt.Errorf("%w when adding nonce for address %s", err, ia.GetAddress())
		}
	}

	return stakersCounter, nil
}

func executeDelegation(
	processors *genesisProcessors,
	arg ArgsGenesisBlockCreator,
	nodesListSplitter genesis.NodesListSplitter,
) (genesis.DelegationResult, []data.TransactionHandler, error) {
	txExecutor, err := intermediate.NewTxExecutionProcessor(processors.txProcessor, arg.Accounts)
	if err != nil {
		return genesis.DelegationResult{}, nil, err
	}

	argDP := intermediate.ArgStandardDelegationProcessor{
		Executor:            txExecutor,
		ShardCoordinator:    arg.ShardCoordinator,
		AccountsParser:      arg.AccountsParser,
		SmartContractParser: arg.SmartContractParser,
		NodesListSplitter:   nodesListSplitter,
		QueryService:        processors.queryService,
		NodePrice:           arg.GenesisNodePrice,
	}

	delegationProcessor, err := intermediate.NewStandardDelegationProcessor(argDP)
	if err != nil {
		return genesis.DelegationResult{}, nil, err
	}

	return delegationProcessor.ExecuteDelegation()
}

func incrementNoncesForCrossShardDelegations(processors *genesisProcessors, arg ArgsGenesisBlockCreator) (int, error) {
	txExecutor, err := intermediate.NewTxExecutionProcessor(processors.txProcessor, arg.Accounts)
	if err != nil {
		return 0, err
	}

	initialAddresses, err := arg.AccountsParser.InitialAccountsSplitOnAddressesShards(arg.ShardCoordinator)
	if err != nil {
		return 0, err
	}

	counter := 0
	initalAddressesInCurrentShard := initialAddresses[arg.ShardCoordinator.SelfId()]
	for _, ia := range initalAddressesInCurrentShard {
		dh := ia.GetDelegationHandler()
		if check.IfNil(dh) {
			continue
		}
		sameShard := arg.ShardCoordinator.SameShard(ia.AddressBytes(), dh.AddressBytes())
		if len(dh.AddressBytes()) == 0 {
			// backwards compatibility, do not make "" address be considered empty and thus, belonging to the same shard
			if arg.ShardCoordinator.ComputeId(ia.AddressBytes()) != 0 {
				sameShard = false
			}
		}
		if sameShard {
			continue
		}

		counter++
		err = txExecutor.AddNonce(ia.AddressBytes(), 1)
		if err != nil {
			return 0, fmt.Errorf("%w when adding nonce for address %s", err, ia.GetAddress())
		}
	}

	return counter, nil
}
