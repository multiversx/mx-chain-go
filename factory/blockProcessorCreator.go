package factory

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/core"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	metachainEpochStart "github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/postprocess"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/coordinator"
	"github.com/ElrondNetwork/elrond-go/process/factory/metachain"
	"github.com/ElrondNetwork/elrond-go/process/factory/shard"
	"github.com/ElrondNetwork/elrond-go/process/rewardTransaction"
	"github.com/ElrondNetwork/elrond-go/process/scToProtocol"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
	"github.com/ElrondNetwork/elrond-go/process/throttle"
	"github.com/ElrondNetwork/elrond-go/process/transaction"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
)

func (pcf *processComponentsFactory) newBlockProcessor(
	requestHandler process.RequestHandler,
	forkDetector process.ForkDetector,
	epochStartTrigger epochStart.TriggerHandler,
	bootStorer process.BootStorer,
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor,
	headerValidator process.HeaderConstructionValidator,
	blockTracker process.BlockTracker,
	pendingMiniBlocksHandler process.PendingMiniBlocksHandler,
) (process.BlockProcessor, error) {
	if pcf.shardCoordinator.SelfId() < pcf.shardCoordinator.NumberOfShards() {
		return pcf.newShardBlockProcessor(
			requestHandler,
			forkDetector,
			epochStartTrigger,
			bootStorer,
			headerValidator,
			blockTracker,
			pcf.smartContractParser,
		)
	}
	if pcf.shardCoordinator.SelfId() == core.MetachainShardId {
		return pcf.newMetaBlockProcessor(
			requestHandler,
			forkDetector,
			validatorStatisticsProcessor,
			epochStartTrigger,
			bootStorer,
			headerValidator,
			blockTracker,
			pendingMiniBlocksHandler,
		)
	}

	return nil, errors.New("could not create block processor")
}

func (pcf *processComponentsFactory) newShardBlockProcessor(
	requestHandler process.RequestHandler,
	forkDetector process.ForkDetector,
	epochStartTrigger epochStart.TriggerHandler,
	bootStorer process.BootStorer,
	headerValidator process.HeaderConstructionValidator,
	blockTracker process.BlockTracker,
	smartContractParser genesis.InitialSmartContractParser,
) (process.BlockProcessor, error) {
	argsParser := smartContract.NewArgumentParser()

	mapDNSAddresses, err := smartContractParser.GetDeployedSCAddresses(genesis.DNSType)
	if err != nil {
		return nil, err
	}

	argsBuiltIn := builtInFunctions.ArgsCreateBuiltInFunctionContainer{
		GasMap:          pcf.gasSchedule,
		MapDNSAddresses: mapDNSAddresses,
		Marshalizer:     pcf.coreData.InternalMarshalizer(),
	}
	builtInFuncs, err := builtInFunctions.CreateBuiltInFunctionContainer(argsBuiltIn)
	if err != nil {
		return nil, err
	}

	argsHook := hooks.ArgBlockChainHook{
		Accounts:         pcf.state.AccountsAdapter(),
		PubkeyConv:       pcf.coreData.AddressPubKeyConverter(),
		StorageService:   pcf.data.StorageService(),
		BlockChain:       pcf.data.Blockchain(),
		ShardCoordinator: pcf.shardCoordinator,
		Marshalizer:      pcf.coreData.InternalMarshalizer(),
		Uint64Converter:  pcf.coreData.Uint64ByteSliceConverter(),
		BuiltInFunctions: builtInFuncs,
	}
	vmFactory, err := shard.NewVMContainerFactory(
		pcf.coreFactoryArgs.Config.VirtualMachineConfig,
		pcf.economicsData.MaxGasLimitPerBlock(pcf.shardCoordinator.SelfId()),
		pcf.gasSchedule,
		argsHook,
	)
	if err != nil {
		return nil, err
	}

	vmContainer, err := vmFactory.Create()
	if err != nil {
		return nil, err
	}

	interimProcFactory, err := shard.NewIntermediateProcessorsContainerFactory(
		pcf.shardCoordinator,
		pcf.coreData.InternalMarshalizer(),
		pcf.coreData.Hasher(),
		pcf.coreData.AddressPubKeyConverter(),
		pcf.data.StorageService(),
		pcf.data.Datapool(),
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
		PubkeyConverter:  pcf.coreData.AddressPubKeyConverter(),
		ShardCoordinator: pcf.shardCoordinator,
		BuiltInFuncNames: builtInFuncs.Keys(),
		ArgumentParser:   parsers.NewCallArgsParser(),
	}
	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	if err != nil {
		return nil, err
	}

	gasHandler, err := preprocess.NewGasComputation(pcf.economicsData, txTypeHandler)
	if err != nil {
		return nil, err
	}

	txFeeHandler, err := postprocess.NewFeeAccumulator()
	if err != nil {
		return nil, err
	}

	argsNewScProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:      vmContainer,
		ArgsParser:       argsParser,
		Hasher:           pcf.coreData.Hasher(),
		Marshalizer:      pcf.coreData.InternalMarshalizer(),
		AccountsDB:       pcf.state.AccountsAdapter(),
		TempAccounts:     vmFactory.BlockChainHookImpl(),
		PubkeyConv:       pcf.coreData.AddressPubKeyConverter(),
		Coordinator:      pcf.shardCoordinator,
		ScrForwarder:     scForwarder,
		TxFeeHandler:     txFeeHandler,
		EconomicsFee:     pcf.economicsData,
		GasHandler:       gasHandler,
		BuiltInFunctions: vmFactory.BlockChainHookImpl().GetBuiltInFunctions(),
		TxLogsProcessor:  pcf.txLogsProcessor,
		TxTypeHandler:    txTypeHandler,
	}
	scProcessor, err := smartContract.NewSmartContractProcessor(argsNewScProcessor)
	if err != nil {
		return nil, err
	}

	rewardsTxProcessor, err := rewardTransaction.NewRewardTxProcessor(
		pcf.state.AccountsAdapter(),
		pcf.coreData.AddressPubKeyConverter(),
		pcf.shardCoordinator,
	)
	if err != nil {
		return nil, err
	}

	transactionProcessor, err := transaction.NewTxProcessor(
		pcf.state.AccountsAdapter(),
		pcf.coreData.Hasher(),
		pcf.coreData.AddressPubKeyConverter(),
		pcf.coreData.InternalMarshalizer(),
		pcf.coreData.TxMarshalizer(),
		pcf.shardCoordinator,
		scProcessor,
		txFeeHandler,
		txTypeHandler,
		pcf.economicsData,
		receiptTxInterim,
		badTxInterim,
		argsParser,
		scForwarder,
	)
	if err != nil {
		return nil, errors.New("could not create transaction statisticsProcessor: " + err.Error())
	}

	blockSizeThrottler, err := throttle.NewBlockSizeThrottle(pcf.minSizeInBytes, pcf.maxSizeInBytes)
	if err != nil {
		return nil, err
	}

	blockSizeComputationHandler, err := preprocess.NewBlockSizeComputation(
		pcf.coreData.InternalMarshalizer(),
		blockSizeThrottler,
		pcf.maxSizeInBytes,
	)
	if err != nil {
		return nil, err
	}

	balanceComputationHandler, err := preprocess.NewBalanceComputation()
	if err != nil {
		return nil, err
	}

	preProcFactory, err := shard.NewPreProcessorsContainerFactory(
		pcf.shardCoordinator,
		pcf.data.StorageService(),
		pcf.coreData.InternalMarshalizer(),
		pcf.coreData.Hasher(),
		pcf.data.Datapool(),
		pcf.coreData.AddressPubKeyConverter(),
		pcf.state.AccountsAdapter(),
		requestHandler,
		transactionProcessor,
		scProcessor,
		scProcessor,
		rewardsTxProcessor,
		pcf.economicsData,
		gasHandler,
		blockTracker,
		blockSizeComputationHandler,
		balanceComputationHandler,
	)
	if err != nil {
		return nil, err
	}

	preProcContainer, err := preProcFactory.Create()
	if err != nil {
		return nil, err
	}

	txCoordinator, err := coordinator.NewTransactionCoordinator(
		pcf.coreData.Hasher(),
		pcf.coreData.InternalMarshalizer(),
		pcf.shardCoordinator,
		pcf.state.AccountsAdapter(),
		pcf.data.Datapool().MiniBlocks(),
		requestHandler,
		preProcContainer,
		interimProcContainer,
		gasHandler,
		txFeeHandler,
		blockSizeComputationHandler,
		balanceComputationHandler,
	)
	if err != nil {
		return nil, err
	}

	accountsDb := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accountsDb[state.UserAccountsState] = pcf.state.AccountsAdapter()

	argumentsBaseProcessor := block.ArgBaseProcessor{
		CoreComponents:         pcf.coreData,
		DataComponents:         pcf.data,
		Version:                pcf.version,
		AccountsDB:             accountsDb,
		ForkDetector:           forkDetector,
		ShardCoordinator:       pcf.shardCoordinator,
		NodesCoordinator:       pcf.nodesCoordinator,
		RequestHandler:         requestHandler,
		Core:                   pcf.coreServiceContainer,
		BlockChainHook:         vmFactory.BlockChainHookImpl(),
		TxCoordinator:          txCoordinator,
		Rounder:                pcf.rounder,
		EpochStartTrigger:      epochStartTrigger,
		HeaderValidator:        headerValidator,
		BootStorer:             bootStorer,
		BlockTracker:           blockTracker,
		FeeHandler:             txFeeHandler,
		StateCheckpointModulus: pcf.stateCheckpointModulus,
		BlockSizeThrottler:     blockSizeThrottler,
	}
	arguments := block.ArgShardProcessor{
		ArgBaseProcessor: argumentsBaseProcessor,
	}

	blockProcessor, err := block.NewShardProcessor(arguments)
	if err != nil {
		return nil, errors.New("could not create block statisticsProcessor: " + err.Error())
	}

	err = blockProcessor.SetAppStatusHandler(pcf.coreData.StatusHandler())
	if err != nil {
		return nil, err
	}

	return blockProcessor, nil
}

func (pcf *processComponentsFactory) newMetaBlockProcessor(
	requestHandler process.RequestHandler,
	forkDetector process.ForkDetector,
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor,
	epochStartTrigger epochStart.TriggerHandler,
	bootStorer process.BootStorer,
	headerValidator process.HeaderConstructionValidator,
	blockTracker process.BlockTracker,
	pendingMiniBlocksHandler process.PendingMiniBlocksHandler,
) (process.BlockProcessor, error) {

	builtInFuncs := builtInFunctions.NewBuiltInFunctionContainer()
	argsHook := hooks.ArgBlockChainHook{
		Accounts:         pcf.state.AccountsAdapter(),
		PubkeyConv:       pcf.coreData.AddressPubKeyConverter(),
		StorageService:   pcf.data.StorageService(),
		BlockChain:       pcf.data.Blockchain(),
		ShardCoordinator: pcf.shardCoordinator,
		Marshalizer:      pcf.coreData.InternalMarshalizer(),
		Uint64Converter:  pcf.coreData.Uint64ByteSliceConverter(),
		BuiltInFunctions: builtInFuncs, // no built-in functions for meta.
	}
	vmFactory, err := metachain.NewVMContainerFactory(
		argsHook,
		pcf.economicsData,
		pcf.crypto.MessageSignVerifier(),
		pcf.gasSchedule,
		pcf.nodesConfig,
		pcf.coreData.Hasher(),
		pcf.coreData.InternalMarshalizer(),
		pcf.systemSCConfig,
		pcf.state.PeerAccounts(),
	)
	if err != nil {
		return nil, err
	}

	argsParser := smartContract.NewArgumentParser()

	vmContainer, err := vmFactory.Create()
	if err != nil {
		return nil, err
	}

	interimProcFactory, err := metachain.NewIntermediateProcessorsContainerFactory(
		pcf.shardCoordinator,
		pcf.coreData.InternalMarshalizer(),
		pcf.coreData.Hasher(),
		pcf.coreData.AddressPubKeyConverter(),
		pcf.data.StorageService(),
		pcf.data.Datapool(),
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

	argsTxTypeHandler := coordinator.ArgNewTxTypeHandler{
		PubkeyConverter:  pcf.coreData.AddressPubKeyConverter(),
		ShardCoordinator: pcf.shardCoordinator,
		BuiltInFuncNames: builtInFuncs.Keys(),
		ArgumentParser:   parsers.NewCallArgsParser(),
	}
	txTypeHandler, err := coordinator.NewTxTypeHandler(argsTxTypeHandler)
	if err != nil {
		return nil, err
	}

	gasHandler, err := preprocess.NewGasComputation(pcf.economicsData, txTypeHandler)
	if err != nil {
		return nil, err
	}

	txFeeHandler, err := postprocess.NewFeeAccumulator()
	if err != nil {
		return nil, err
	}

	argsNewScProcessor := smartContract.ArgsNewSmartContractProcessor{
		VmContainer:      vmContainer,
		ArgsParser:       argsParser,
		Hasher:           pcf.coreData.Hasher(),
		Marshalizer:      pcf.coreData.InternalMarshalizer(),
		AccountsDB:       pcf.state.AccountsAdapter(),
		TempAccounts:     vmFactory.BlockChainHookImpl(),
		PubkeyConv:       pcf.coreData.AddressPubKeyConverter(),
		Coordinator:      pcf.shardCoordinator,
		ScrForwarder:     scForwarder,
		TxFeeHandler:     txFeeHandler,
		EconomicsFee:     pcf.economicsData,
		TxTypeHandler:    txTypeHandler,
		GasHandler:       gasHandler,
		BuiltInFunctions: vmFactory.BlockChainHookImpl().GetBuiltInFunctions(),
		TxLogsProcessor:  pcf.txLogsProcessor,
	}
	scProcessor, err := smartContract.NewSmartContractProcessor(argsNewScProcessor)
	if err != nil {
		return nil, err
	}

	transactionProcessor, err := transaction.NewMetaTxProcessor(
		pcf.coreData.Hasher(),
		pcf.coreData.InternalMarshalizer(),
		pcf.state.AccountsAdapter(),
		pcf.coreData.AddressPubKeyConverter(),
		pcf.shardCoordinator,
		scProcessor,
		txTypeHandler,
		pcf.economicsData,
	)
	if err != nil {
		return nil, errors.New("could not create transaction processor: " + err.Error())
	}

	blockSizeThrottler, err := throttle.NewBlockSizeThrottle(pcf.minSizeInBytes, pcf.maxSizeInBytes)
	if err != nil {
		return nil, err
	}

	blockSizeComputationHandler, err := preprocess.NewBlockSizeComputation(
		pcf.coreData.InternalMarshalizer(),
		blockSizeThrottler,
		pcf.maxSizeInBytes,
	)
	if err != nil {
		return nil, err
	}

	balanceComputationHandler, err := preprocess.NewBalanceComputation()
	if err != nil {
		return nil, err
	}

	preProcFactory, err := metachain.NewPreProcessorsContainerFactory(
		pcf.shardCoordinator,
		pcf.data.StorageService(),
		pcf.coreData.InternalMarshalizer(),
		pcf.coreData.Hasher(),
		pcf.data.Datapool(),
		pcf.state.AccountsAdapter(),
		requestHandler,
		transactionProcessor,
		scProcessor,
		pcf.economicsData,
		gasHandler,
		blockTracker,
		pcf.coreData.AddressPubKeyConverter(),
		blockSizeComputationHandler,
		balanceComputationHandler,
	)
	if err != nil {
		return nil, err
	}

	preProcContainer, err := preProcFactory.Create()
	if err != nil {
		return nil, err
	}

	txCoordinator, err := coordinator.NewTransactionCoordinator(
		pcf.coreData.Hasher(),
		pcf.coreData.InternalMarshalizer(),
		pcf.shardCoordinator,
		pcf.state.AccountsAdapter(),
		pcf.data.Datapool().MiniBlocks(),
		requestHandler,
		preProcContainer,
		interimProcContainer,
		gasHandler,
		txFeeHandler,
		blockSizeComputationHandler,
		balanceComputationHandler,
	)
	if err != nil {
		return nil, err
	}

	scDataGetter, err := smartContract.NewSCQueryService(vmContainer, pcf.economicsData)
	if err != nil {
		return nil, err
	}

	argsStaking := scToProtocol.ArgStakingToPeer{
		PubkeyConv:  pcf.coreData.ValidatorPubKeyConverter(),
		Hasher:      pcf.coreData.Hasher(),
		Marshalizer: pcf.coreData.InternalMarshalizer(),
		PeerState:   pcf.state.PeerAccounts(),
		BaseState:   pcf.state.AccountsAdapter(),
		ArgParser:   argsParser,
		CurrTxs:     pcf.data.Datapool().CurrentBlockTxs(),
		ScQuery:     scDataGetter,
		RatingsData: pcf.ratingsData,
	}
	smartContractToProtocol, err := scToProtocol.NewStakingToPeer(argsStaking)
	if err != nil {
		return nil, err
	}

	genesisHdr := pcf.data.Blockchain().GetGenesisHeader()
	argsEpochStartData := metachainEpochStart.ArgsNewEpochStartData{
		Marshalizer:       pcf.coreData.InternalMarshalizer(),
		Hasher:            pcf.coreData.Hasher(),
		Store:             pcf.data.StorageService(),
		DataPool:          pcf.data.Datapool(),
		BlockTracker:      blockTracker,
		ShardCoordinator:  pcf.shardCoordinator,
		EpochStartTrigger: epochStartTrigger,
		RequestHandler:    requestHandler,
		GenesisEpoch:      genesisHdr.GetEpoch(),
	}
	epochStartDataCreator, err := metachainEpochStart.NewEpochStartData(argsEpochStartData)
	if err != nil {
		return nil, err
	}

	argsEpochEconomics := metachainEpochStart.ArgsNewEpochEconomics{
		Marshalizer:        pcf.coreData.InternalMarshalizer(),
		Hasher:             pcf.coreData.Hasher(),
		Store:              pcf.data.StorageService(),
		ShardCoordinator:   pcf.shardCoordinator,
		RewardsHandler:     pcf.economicsData,
		RoundTime:          pcf.rounder,
		GenesisNonce:       genesisHdr.GetNonce(),
		GenesisEpoch:       genesisHdr.GetEpoch(),
		GenesisTotalSupply: pcf.economicsData.GenesisTotalSupply(),
	}
	epochEconomics, err := metachainEpochStart.NewEndOfEpochEconomicsDataCreator(argsEpochEconomics)
	if err != nil {
		return nil, err
	}

	rewardsStorage := pcf.data.StorageService().GetStorer(dataRetriever.RewardTransactionUnit)
	miniBlockStorage := pcf.data.StorageService().GetStorer(dataRetriever.MiniBlockUnit)
	argsEpochRewards := metachainEpochStart.ArgsNewRewardsCreator{
		ShardCoordinator:    pcf.shardCoordinator,
		PubkeyConverter:     pcf.coreData.AddressPubKeyConverter(),
		RewardsStorage:      rewardsStorage,
		MiniBlockStorage:    miniBlockStorage,
		Hasher:              pcf.coreData.Hasher(),
		Marshalizer:         pcf.coreData.InternalMarshalizer(),
		DataPool:            pcf.data.Datapool(),
		CommunityAddress:    pcf.economicsData.CommunityAddress(),
		NodesConfigProvider: pcf.nodesCoordinator,
	}
	epochRewards, err := metachainEpochStart.NewEpochStartRewardsCreator(argsEpochRewards)
	if err != nil {
		return nil, err
	}

	argsEpochValidatorInfo := metachainEpochStart.ArgsNewValidatorInfoCreator{
		ShardCoordinator: pcf.shardCoordinator,
		MiniBlockStorage: miniBlockStorage,
		Hasher:           pcf.coreData.Hasher(),
		Marshalizer:      pcf.coreData.InternalMarshalizer(),
		DataPool:         pcf.data.Datapool(),
	}
	validatorInfoCreator, err := metachainEpochStart.NewValidatorInfoCreator(argsEpochValidatorInfo)
	if err != nil {
		return nil, err
	}

	accountsDb := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accountsDb[state.UserAccountsState] = pcf.state.AccountsAdapter()
	accountsDb[state.PeerAccountsState] = pcf.state.PeerAccounts()

	argumentsBaseProcessor := block.ArgBaseProcessor{
		CoreComponents:         pcf.coreData,
		DataComponents:         pcf.data,
		Version:                pcf.version,
		AccountsDB:             accountsDb,
		ForkDetector:           forkDetector,
		ShardCoordinator:       pcf.shardCoordinator,
		NodesCoordinator:       pcf.nodesCoordinator,
		RequestHandler:         requestHandler,
		Core:                   pcf.coreServiceContainer,
		BlockChainHook:         vmFactory.BlockChainHookImpl(),
		TxCoordinator:          txCoordinator,
		EpochStartTrigger:      epochStartTrigger,
		Rounder:                pcf.rounder,
		HeaderValidator:        headerValidator,
		BootStorer:             bootStorer,
		BlockTracker:           blockTracker,
		FeeHandler:             txFeeHandler,
		StateCheckpointModulus: pcf.stateCheckpointModulus,
		BlockSizeThrottler:     blockSizeThrottler,
	}
	arguments := block.ArgMetaProcessor{
		ArgBaseProcessor:             argumentsBaseProcessor,
		SCDataGetter:                 scDataGetter,
		SCToProtocol:                 smartContractToProtocol,
		PendingMiniBlocksHandler:     pendingMiniBlocksHandler,
		EpochStartDataCreator:        epochStartDataCreator,
		EpochEconomics:               epochEconomics,
		EpochRewardsCreator:          epochRewards,
		EpochValidatorInfoCreator:    validatorInfoCreator,
		ValidatorStatisticsProcessor: validatorStatisticsProcessor,
	}

	metaProcessor, err := block.NewMetaProcessor(arguments)
	if err != nil {
		return nil, errors.New("could not create block processor: " + err.Error())
	}

	err = metaProcessor.SetAppStatusHandler(pcf.coreData.StatusHandler())
	if err != nil {
		return nil, err
	}

	return metaProcessor, nil
}
