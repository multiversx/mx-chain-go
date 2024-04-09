package components

//type runTypeComponentsHolder struct {
//	blockChainHookHandlerCreator            hooks.BlockChainHookHandlerCreator
//	epochStartBootstrapperCreator           bootstrap.EpochStartBootstrapperCreator
//	bootstrapperFromStorageCreator          storageBootstrap.BootstrapperFromStorageCreator
//	bootstrapperCreator                     storageBootstrap.BootstrapperCreator
//	blockProcessorCreator                   processBlock.BlockProcessorCreator
//	forkDetectorCreator                     sync.ForkDetectorCreator
//	blockTrackerCreator                     track.BlockTrackerCreator
//	requestHandlerCreator                   requestHandlers.RequestHandlerCreator
//	headerValidatorCreator                  processBlock.HeaderValidatorCreator
//	scheduledTxsExecutionCreator            preprocess.ScheduledTxsExecutionCreator
//	transactionCoordinatorCreator           coordinator.TransactionCoordinatorCreator
//	validatorStatisticsProcessorCreator     peer.ValidatorStatisticsProcessorCreator
//	additionalStorageServiceCreator         process.AdditionalStorageServiceCreator
//	scProcessorCreator                      scrCommon.SCProcessorCreator
//	scResultPreProcessorCreator             preprocess.SmartContractResultPreProcessorCreator
//	consensusModel                          consensus.ConsensusModel
//	vmContainerMetaFactory                  factoryVm.VmContainerCreator
//	vmContainerShardFactory                 factoryVm.VmContainerCreator
//	accountsCreator                         state.AccountFactory
//	outGoingOperationsPoolHandler           sovereignBlock.OutGoingOperationsPool
//	dataCodecHandler                        sovereign.DataDecoderHandler
//	topicsCheckerHandler                    sovereign.TopicsCheckerHandler
//	shardCoordinatorCreator                 sharding.ShardCoordinatorFactory
//	nodesCoordinatorWithRaterFactoryCreator nodesCoord.NodesCoordinatorWithRaterFactory
//	requestersContainerFactoryCreator       requesterscontainer.RequesterContainerFactoryCreator
//	interceptorsContainerFactoryCreator     interceptorscontainer.InterceptorsContainerFactoryCreator
//	shardResolversContainerFactoryCreator   resolverscontainer.ShardResolversContainerFactoryCreator
//	txPreProcessorCreator                   preprocess.TxPreProcessorCreator
//	extraHeaderSigVerifierHandler           headerCheck.ExtraHeaderSigVerifierHolder
//	genesisBlockCreatorFactory              processGenesis.GenesisBlockCreatorFactory
//	genesisMetaBlockCheckerCreator          processGenesis.GenesisMetaBlockChecker
//}

//func GetRunTypeComponents(coreComponents factory.CoreComponentsHolder) (factory.RunTypeComponentsHolder, error) {
//	runTypeComponentsFactory, _ := runType.NewRunTypeComponentsFactory(coreComponents)
//	managedRunTypeComponents, err := runType.NewManagedRunTypeComponents(runTypeComponentsFactory)
//	if err != nil {
//		return nil, err
//	}
//	err = managedRunTypeComponents.Create()
//	if err != nil {
//		return nil, err
//	}
//
//	return managedRunTypeComponents, nil
//}
