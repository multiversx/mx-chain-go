package process

import (
	"errors"
	"fmt"
	"math"
	"math/big"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/forking"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/genesis/process/disabled"
	"github.com/ElrondNetwork/elrond-go/genesis/process/intermediate"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/rewardTransaction"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	hardForkProcess "github.com/ElrondNetwork/elrond-go/update/process"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
)

var log = logger.GetOrCreate("genesis/process")

var zero = big.NewInt(0)

type deployedScMetrics struct {
	numDelegation int
	numOtherTypes int
}

// CreateShardGenesisBlock will create a shard genesis block
func CreateShardGenesisBlock(arg ArgsGenesisBlockCreator, nodesListSplitter genesis.NodesListSplitter, selfShardID uint32) (data.HeaderHandler, [][]byte, error) {
	if mustDoHardForkImportProcess(arg) {
		return createShardGenesisAfterHardFork(arg, selfShardID)
	}

	genesisOverrideConfig := config.GeneralSettingsConfig{
		BuiltInFunctionsEnableEpoch:    0,
		SCDeployEnableEpoch:            0,
		RelayedTransactionsEnableEpoch: 0,
		PenalizedTooMuchGasEnableEpoch: 0,
	}

	processors, err := createProcessorsForShard(arg, genesisOverrideConfig)
	if err != nil {
		return nil, nil, err
	}

	deployMetrics := &deployedScMetrics{}

	scAddresses, err := deployInitialSmartContracts(processors, arg, deployMetrics)
	if err != nil {
		return nil, nil, err
	}

	numSetBalances, err := setBalancesToTrie(arg)
	if err != nil {
		return nil, nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while setting the balances to trie",
			err, arg.ShardCoordinator.SelfId())
	}

	numStaked, err := increaseStakersNonces(processors, arg)
	if err != nil {
		return nil, nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while incrementing nonces",
			err, arg.ShardCoordinator.SelfId())
	}

	delegationResult, err := executeDelegation(processors, arg, nodesListSplitter)
	if err != nil {
		return nil, nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while execution delegation",
			err, arg.ShardCoordinator.SelfId())
	}

	numCrossShardDelegations, err := incrementNoncesForCrossShardDelegations(processors, arg)
	if err != nil {
		return nil, nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while incrementing crossshard nonce",
			err, arg.ShardCoordinator.SelfId())
	}

	rootHash, err := arg.Accounts.Commit()
	if err != nil {
		return nil, nil, fmt.Errorf("%w encountered when creating genesis block for shard %d while commiting",
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
		ChainID:         []byte(arg.ChainID),
		SoftwareVersion: []byte(""),
	}

	err = processors.vmContainer.Close()
	if err != nil {
		return nil, nil, err
	}

	return header, scAddresses, nil
}

func createShardGenesisAfterHardFork(arg ArgsGenesisBlockCreator, selfShardId uint32) (data.HeaderHandler, [][]byte, error) {
	tmpArg := arg
	tmpArg.Accounts = arg.importHandler.GetAccountsDBForShard(arg.ShardCoordinator.SelfId())
	processors, err := createProcessorsForShard(tmpArg, *arg.GeneralConfig)
	if err != nil {
		return nil, nil, err
	}

	argsPendingTxProcessor := hardForkProcess.ArgsPendingTransactionProcessor{
		Accounts:         tmpArg.Accounts,
		TxProcessor:      processors.txProcessor,
		RwdTxProcessor:   processors.rwdProcessor,
		ScrTxProcessor:   processors.scrProcessor,
		PubKeyConv:       arg.PubkeyConv,
		ShardCoordinator: arg.ShardCoordinator,
	}
	pendingTxProcessor, err := hardForkProcess.NewPendingTransactionProcessor(argsPendingTxProcessor)
	if err != nil {
		return nil, nil, err
	}

	argsShardBlockAfterHardFork := hardForkProcess.ArgsNewShardBlockCreatorAfterHardFork{
		ShardCoordinator:   arg.ShardCoordinator,
		TxCoordinator:      processors.txCoordinator,
		PendingTxProcessor: pendingTxProcessor,
		ImportHandler:      arg.importHandler,
		Marshalizer:        arg.Marshalizer,
		Hasher:             arg.Hasher,
		DataPool:           arg.DataPool,
		Storage:            arg.Store,
		SelfShardID:        selfShardId,
	}
	shardBlockCreator, err := hardForkProcess.NewShardBlockCreatorAfterHardFork(argsShardBlockAfterHardFork)
	if err != nil {
		return nil, nil, err
	}

	hdrHandler, _, err := shardBlockCreator.CreateNewBlock(
		arg.ChainID,
		arg.HardForkConfig.StartRound,
		arg.HardForkConfig.StartNonce,
		arg.HardForkConfig.StartEpoch,
	)
	if err != nil {
		return nil, nil, err
	}
	hdrHandler.SetTimeStamp(arg.GenesisTime)

	err = arg.Accounts.RecreateTrie(hdrHandler.GetRootHash())
	if err != nil {
		return nil, nil, err
	}

	return hdrHandler, make([][]byte, 0), nil
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

func createProcessorsForShard(arg ArgsGenesisBlockCreator, generalConfig config.GeneralSettingsConfig) (*genesisProcessors, error) {
	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasMap:               arg.GasMap,
		MapDNSAddresses:      make(map[string]struct{}),
		EnableUserNameChange: false,
		Marshalizer:          arg.Marshalizer,
		Accounts:             arg.Accounts,
	}
	builtInFuncs, err := builtInFunctions.CreateBuiltInFunctionContainer(argsBuiltIn)
	if err != nil {
		return nil, err
	}

	argsHook := hooks.ArgBlockChainHook{
		Accounts:         arg.Accounts,
		PubkeyConv:       arg.PubkeyConv,
		StorageService:   arg.Store,
		BlockChain:       arg.Blkc,
		ShardCoordinator: arg.ShardCoordinator,
		Marshalizer:      arg.Marshalizer,
		Uint64Converter:  arg.Uint64ByteSliceConverter,
		BuiltInFunctions: builtInFuncs,
		DataPool:         arg.DataPool,
	}
	vmFactoryImpl, err := shard.NewVMContainerFactory(
		arg.VirtualMachineConfig,
		math.MaxUint64,
		arg.GasMap,
		argsHook,
		arg.GeneralConfig.SCDeployEnableEpoch,
	)
	if err != nil {
		return nil, err
	}

	vmContainer, err := vmFactoryImpl.Create()
	if err != nil {
		return nil, err
	}

	err = builtInFunctions.SetPayableHandler(builtInFuncs, vmFactoryImpl.BlockChainHookImpl())
	if err != nil {
		return nil, err
	}

	interimProcFactory, err := shard.NewIntermediateProcessorsContainerFactory(
		arg.ShardCoordinator,
		arg.Marshalizer,
		arg.Hasher,
		arg.PubkeyConv,
		arg.Store,
		arg.DataPool,
	)
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

	receiptTxInterim, err := interimProcContainer.Get(dataBlock.ReceiptBlock)
	if err != nil {
		return nil, err
	}

	badTxInterim, err := interimProcContainer.Get(dataBlock.InvalidBlock)
	if err != nil {
		return nil, err
	}

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:  arg.PubkeyConv,
		ShardCoordinator: arg.ShardCoordinator,
		BuiltInFuncNames: builtInFuncs.Keys(),
		ArgumentParser:   parsers.NewCallArgsParser(),
	}
	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	if err != nil {
		return nil, err
	}

	gasHandler, err := preprocess.NewGasComputation(arg.Economics, txTypeHandler)
	if err != nil {
		return nil, err
	}

	epochNotifier := forking.NewGenericEpochNotifier()
	epochNotifier.CheckEpoch(arg.StartEpochNum)

	genesisFeeHandler := &disabled.FeeHandler{}
	argsNewScProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:                    vmContainer,
		ArgsParser:                     smartContract.NewArgumentParser(),
		Hasher:                         arg.Hasher,
		Marshalizer:                    arg.Marshalizer,
		AccountsDB:                     arg.Accounts,
		BlockChainHook:                 vmFactoryImpl.BlockChainHookImpl(),
		PubkeyConv:                     arg.PubkeyConv,
		Coordinator:                    arg.ShardCoordinator,
		ScrForwarder:                   scForwarder,
		TxFeeHandler:                   genesisFeeHandler,
		EconomicsFee:                   genesisFeeHandler,
		TxTypeHandler:                  txTypeHandler,
		GasHandler:                     gasHandler,
		GasSchedule:                    arg.GasMap,
		BuiltInFunctions:               vmFactoryImpl.BlockChainHookImpl().GetBuiltInFunctions(),
		TxLogsProcessor:                arg.TxLogsProcessor,
		BadTxForwarder:                 badTxInterim,
		EpochNotifier:                  epochNotifier,
		BuiltinEnableEpoch:             generalConfig.BuiltInFunctionsEnableEpoch,
		DeployEnableEpoch:              generalConfig.SCDeployEnableEpoch,
		PenalizedTooMuchGasEnableEpoch: generalConfig.PenalizedTooMuchGasEnableEpoch,
	}
	scProcessor, err := smartContract.NewSmartContractProcessor(argsNewScProcessor)
	if err != nil {
		return nil, err
	}

	rewardsTxProcessor, err := rewardTransaction.NewRewardTxProcessor(
		arg.Accounts,
		arg.PubkeyConv,
		arg.ShardCoordinator,
	)
	if err != nil {
		return nil, err
	}

	argsNewTxProcessor := transaction.ArgsNewTxProcessor{
		Accounts:                       arg.Accounts,
		Hasher:                         arg.Hasher,
		PubkeyConv:                     arg.PubkeyConv,
		Marshalizer:                    arg.Marshalizer,
		SignMarshalizer:                arg.SignMarshalizer,
		ShardCoordinator:               arg.ShardCoordinator,
		ScProcessor:                    scProcessor,
		TxFeeHandler:                   genesisFeeHandler,
		TxTypeHandler:                  txTypeHandler,
		EconomicsFee:                   genesisFeeHandler,
		ReceiptForwarder:               receiptTxInterim,
		BadTxForwarder:                 badTxInterim,
		ArgsParser:                     smartContract.NewArgumentParser(),
		ScrForwarder:                   scForwarder,
		EpochNotifier:                  epochNotifier,
		RelayedTxEnableEpoch:           generalConfig.RelayedTransactionsEnableEpoch,
		PenalizedTooMuchGasEnableEpoch: generalConfig.PenalizedTooMuchGasEnableEpoch,
		MetaProtectionEnableEpoch:      generalConfig.MetaProtectionEnableEpoch,
	}
	transactionProcessor, err := transaction.NewTxProcessor(argsNewTxProcessor)
	if err != nil {
		return nil, errors.New("could not create transaction statisticsProcessor: " + err.Error())
	}

	disabledRequestHandler := &disabled.RequestHandler{}
	disabledBlockTracker := &disabled.BlockTracker{}
	disabledBlockSizeComputationHandler := &disabled.BlockSizeComputationHandler{}
	disabledBalanceComputationHandler := &disabled.BalanceComputationHandler{}

	preProcFactory, err := shard.NewPreProcessorsContainerFactory(
		arg.ShardCoordinator,
		arg.Store,
		arg.Marshalizer,
		arg.Hasher,
		arg.DataPool,
		arg.PubkeyConv,
		arg.Accounts,
		disabledRequestHandler,
		transactionProcessor,
		scProcessor,
		scProcessor,
		rewardsTxProcessor,
		arg.Economics,
		gasHandler,
		disabledBlockTracker,
		disabledBlockSizeComputationHandler,
		disabledBalanceComputationHandler,
	)
	if err != nil {
		return nil, err
	}

	preProcContainer, err := preProcFactory.Create()
	if err != nil {
		return nil, err
	}

	txCoordinator, err := coordinator.NewTransactionCoordinator(
		arg.Hasher,
		arg.Marshalizer,
		arg.ShardCoordinator,
		arg.Accounts,
		arg.DataPool.MiniBlocks(),
		disabledRequestHandler,
		preProcContainer,
		interimProcContainer,
		gasHandler,
		genesisFeeHandler,
		disabledBlockSizeComputationHandler,
		disabledBalanceComputationHandler,
	)
	if err != nil {
		return nil, err
	}

	queryService, err := smartContract.NewSCQueryService(
		vmContainer,
		arg.Economics,
		vmFactoryImpl.BlockChainHookImpl(),
		arg.Blkc,
	)
	if err != nil {
		return nil, err
	}

	return &genesisProcessors{
		txCoordinator:  txCoordinator,
		systemSCs:      nil,
		txProcessor:    transactionProcessor,
		scProcessor:    scProcessor,
		scrProcessor:   scProcessor,
		rwdProcessor:   rewardsTxProcessor,
		blockchainHook: vmFactoryImpl.BlockChainHookImpl(),
		queryService:   queryService,
		vmContainer:    vmContainer,
	}, nil
}

func deployInitialSmartContracts(
	processors *genesisProcessors,
	arg ArgsGenesisBlockCreator,
	deployMetrics *deployedScMetrics,
) ([][]byte, error) {
	smartContracts, err := arg.SmartContractParser.InitialSmartContractsSplitOnOwnersShards(arg.ShardCoordinator)
	if err != nil {
		return nil, err
	}

	var scAddresses = make([][]byte, 0)
	currentShardSmartContracts := smartContracts[arg.ShardCoordinator.SelfId()]
	for _, sc := range currentShardSmartContracts {
		var scResulted [][]byte
		scResulted, err = deployInitialSmartContract(processors, sc, arg, deployMetrics)
		if err != nil {
			return nil, fmt.Errorf("%w for owner %s and filename %s",
				err, sc.GetOwner(), sc.GetFilename())
		}

		scAddresses = append(scAddresses, scResulted...)
	}

	return scAddresses, nil
}

func deployInitialSmartContract(
	processors *genesisProcessors,
	sc genesis.InitialSmartContractHandler,
	arg ArgsGenesisBlockCreator,
	deployMetrics *deployedScMetrics,
) ([][]byte, error) {

	txExecutor, err := intermediate.NewTxExecutionProcessor(processors.txProcessor, arg.Accounts)
	if err != nil {
		return nil, err
	}

	var deployProc genesis.DeployProcessor

	switch sc.GetType() {
	case genesis.DNSType:
		deployMetrics.numOtherTypes++
		argDeployLibrary := intermediate.ArgDeployLibrarySC{
			Executor:         txExecutor,
			PubkeyConv:       arg.PubkeyConv,
			BlockchainHook:   processors.blockchainHook,
			ShardCoordinator: arg.ShardCoordinator,
		}
		deployProc, err = intermediate.NewDeployLibrarySC(argDeployLibrary)
		if err != nil {
			return nil, err
		}
	case genesis.DelegationType:
		deployMetrics.numDelegation++
		fallthrough
	default:
		argDeploy := intermediate.ArgDeployProcessor{
			Executor:       txExecutor,
			PubkeyConv:     arg.PubkeyConv,
			BlockchainHook: processors.blockchainHook,
			QueryService:   processors.queryService,
		}
		deployProc, err = intermediate.NewDeployProcessor(argDeploy)
		if err != nil {
			return nil, err
		}
	}

	return deployProc.Deploy(sc)
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
) (genesis.DelegationResult, error) {
	txExecutor, err := intermediate.NewTxExecutionProcessor(processors.txProcessor, arg.Accounts)
	if err != nil {
		return genesis.DelegationResult{}, err
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
		return genesis.DelegationResult{}, err
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
		if arg.ShardCoordinator.SameShard(ia.AddressBytes(), dh.AddressBytes()) {
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
