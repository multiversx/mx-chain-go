package consensusComponents

// ------------ Test TestConsensusComponents --------------------
//func TestConsensusComponents_Close_ShouldWork(t *testing.T) {
//	t.Skip()
//	generalConfig, _ := core.LoadMainConfig(ConfigPath)
//	ratingsConfig, _ := core.LoadRatingsConfig(RatingsPath)
//	economicsConfig, _ := core.LoadEconomicsConfig(EconomicsPath)
//	prefsConfig, _ := core.LoadPreferencesConfig(PrefsPath)
//	p2pConfig, _ := core.LoadP2PConfig(P2pPath)
//	externalConfig, _ := core.LoadExternalConfig(ExternalPath)
//	systemSCConfig, _ := core.LoadSystemSmartContractsConfig(SystemSCConfigPath)
//
//	nrBefore := runtime.NumGoroutine()
//	PrintStack()
//
//	managedCoreComponents, _ := CreateCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
//	managedCryptoComponents, _ := CreateCryptoComponents(*generalConfig, *systemSCConfig, managedCoreComponents)
//	managedNetworkComponents, _ := CreateNetworkComponents(*generalConfig, *p2pConfig, *ratingsConfig, managedCoreComponents)
//	managedBootstrapComponents, _ := createBootstrapComponents(
//		*generalConfig,
//		prefsConfig.Preferences,
//		managedCoreComponents,
//		managedCryptoComponents,
//		managedNetworkComponents)
//	epochStartNotifier := notifier.NewEpochStartSubscriptionHandler()
//	managedDataComponents, _ := CreateDataComponents(*generalConfig, *economicsConfig, epochStartNotifier, managedCoreComponents)
//	time.Sleep(2 * time.Second)
//
//	nodesSetup := managedCoreComponents.GenesisNodesSetup()
//	chanStopNodeProcess := make(chan endProcess.ArgEndProcess, 1)
//	genesisShardCoordinator, nodesCoordinator, nodesShufflerOut, ratingsData, rater := createCoordinators(
//		generalConfig,
//		prefsConfig,
//		ratingsConfig,
//		nodesSetup,
//		epochStartNotifier,
//		chanStopNodeProcess,
//		managedCoreComponents,
//		managedCryptoComponents,
//		managedDataComponents,
//		managedBootstrapComponents)
//
//	managedStateComponents, _ := createStateComponents(*generalConfig,managedCoreComponents,managedBootstrapComponents,)
//	managedStatusComponents, err := createStatusComponents(
//		*generalConfig,
//		*externalConfig,
//		genesisShardCoordinator,
//		nodesCoordinator,
//		epochStartNotifier,
//		managedCoreComponents,
//		managedDataComponents,
//		managedNetworkComponents)
//	require.Nil(t, err)
//	require.NotNil(t, managedStatusComponents)
//
//	time.Sleep(5 * time.Second)
//
//	managedProcessComponents, err := createProcessComponents(
//		generalConfig,
//		economicsConfig,
//		ratingsConfig,
//		systemSCConfig,
//		nodesSetup,
//		nodesCoordinator,
//		epochStartNotifier,
//		genesisShardCoordinator,
//		ratingsData,
//		rater,
//		managedCoreComponents,
//		managedCryptoComponents,
//		managedDataComponents,
//		managedStateComponents,
//		managedNetworkComponents,
//		managedBootstrapComponents,
//		managedStatusComponents,
//		chanStopNodeProcess)
//
//	require.Nil(t, err)
//	require.NotNil(t, managedProcessComponents)
//
//	managedStatusComponents.SetForkDetector(managedProcessComponents.ForkDetector())
//	err = managedStatusComponents.StartPolling()
//	require.Nil(t, err)
//
//	hardForkTrigger, err := node.CreateHardForkTrigger(
//		generalConfig,
//		genesisShardCoordinator,
//		nodesCoordinator,
//		managedCoreComponents,
//		managedStateComponents,
//		managedDataComponents,
//		managedCryptoComponents,
//		managedProcessComponents,
//		managedNetworkComponents,
//		whiteListRequest,
//		whiteListerVerifiedTxs,
//		chanStopNodeProcess,
//		epochStartNotifier,
//		importStartHandler,
//		nodesSetup,
//		"workingDir",
//	)
//	require.Nil(t, err)
//
//	err = hardForkTrigger.AddCloser(nodesShufflerOut)
//	require.Nil(t, err)
//
//	elasticIndexer := managedStatusComponents.ElasticIndexer()
//	if !elasticIndexer.IsNilIndexer() {
//		elasticIndexer.SetTxLogsProcessor(managedProcessComponents.TxLogsProcessor())
//		managedProcessComponents.TxLogsProcessor().EnableLogToBeSavedInCache()
//	}
//
//	currentNode, err := node.CreateNode(
//		generalConfig,
//		prefsConfig,
//		nodesSetup,
//		managedBootstrapComponents,
//		managedCoreComponents,
//		managedCryptoComponents,
//		managedDataComponents,
//		managedNetworkComponents,
//		managedProcessComponents,
//		managedStateComponents,
//		managedStatusComponents,
//		ctx.GlobalUint64(bootstrapRoundIndex.Name),
//		"version",
//		requestedItemsHandler,
//		whiteListRequest,
//		whiteListerVerifiedTxs,
//		chanStopNodeProcess,
//		hardForkTrigger,
//		historyRepository,
//	)
//	if err != nil {
//		return err
//	}
//
//	log.Trace("creating software checker structure")
//	softwareVersionChecker, err := factory.CreateSoftwareVersionChecker(
//		managedCoreComponents.StatusHandler(),
//		cfgs.generalConfig.SoftwareVersionConfig,
//	)
//	if err != nil {
//		log.Debug("nil software version checker", "error", err.Error())
//	} else {
//		softwareVersionChecker.StartCheckSoftwareVersion()
//	}
//
//	if shardCoordinator.SelfId() == core.MetachainShardId {
//		log.Trace("activating nodesCoordinator's validators indexing")
//		indexValidatorsListIfNeeded(
//			elasticIndexer,
//			nodesCoordinator,
//			managedProcessComponents.EpochStartTrigger().Epoch(),
//			log,
//		)
//	}
//
//	log.Trace("creating api resolver structure")
//	apiResolver, err := mainFactory.CreateApiResolver(
//		cfgs.generalConfig,
//		managedStateComponents.AccountsAdapter(),
//		managedStateComponents.PeerAccounts(),
//		managedCoreComponents.AddressPubKeyConverter(),
//		managedDataComponents.StorageService(),
//		managedDataComponents.Blockchain(),
//		managedCoreComponents.InternalMarshalizer(),
//		managedCoreComponents.Hasher(),
//		managedCoreComponents.Uint64ByteSliceConverter(),
//		genesisShardCoordinator,
//		managedCoreComponents.StatusHandlerUtils().Metrics(),
//		GasSchedule,
//		economicsData,
//		managedCryptoComponents.MessageSignVerifier(),
//		nodesSetup,
//		systemSCConfig,
//	)
//	if err != nil {
//		return err
//	}
//
//	restAPIServerDebugMode := ctx.GlobalBool(restApiDebug.Name)
//
//	argNodeFacade := facade.ArgNodeFacade{
//		Node:                   currentNode,
//		ApiResolver:            apiResolver,
//		TxSimulatorProcessor:   managedProcessComponents.TransactionSimulatorProcessor(),
//		RestAPIServerDebugMode: restAPIServerDebugMode,
//		WsAntifloodConfig:      cfgs.generalConfig.Antiflood.WebServer,
//		FacadeConfig: config.FacadeConfig{
//			RestApiInterface: ctx.GlobalString(restApiInterface.Name),
//			PprofEnabled:     ctx.GlobalBool(profileMode.Name),
//		},
//		ApiRoutesConfig: *cfgs.apiRoutesConfig,
//		AccountsState:   managedStateComponents.AccountsAdapter(),
//		PeerState:       managedStateComponents.PeerAccounts(),
//	}
//
//	ef, err := facade.NewNodeFacade(argNodeFacade)
//	if err != nil {
//		return fmt.Errorf("%w while creating NodeFacade", err)
//	}
//
//	ef.SetSyncer(managedCoreComponents.SyncTimer())
//	ef.SetTpsBenchmark(managedStatusComponents.TpsBenchmark())
//
//	ef.StartBackgroundServices()
//
//	consensusArgs := mainFactory.ConsensusComponentsFactoryArgs{
//		Config:              *cfgs.generalConfig,
//		ConsensusGroupSize:  int(nodesSetup.GetShardConsensusGroupSize()),
//		BootstrapRoundIndex: ctx.GlobalUint64(bootstrapRoundIndex.Name),
//		hardforkTrigger:     hardForkTrigger,
//		CoreComponents:      managedCoreComponents,
//		NetworkComponents:   managedNetworkComponents,
//		CryptoComponents:    managedCryptoComponents,
//		DataComponents:      managedDataComponents,
//		ProcessComponents:   managedProcessComponents,
//		StateComponents:     managedStateComponents,
//		StatusComponents:    managedStatusComponents,
//	}
//
//	consensusFactory, err := mainFactory.NewConsensusComponentsFactory(consensusArgs)
//	if err != nil {
//		return fmt.Errorf("NewConsensusComponentsFactory failed: %w", err)
//	}
//
//	managedConsensusComponents, err := mainFactory.NewManagedConsensusComponents(consensusFactory)
//	if err != nil {
//		return err
//	}
//
//	err = managedConsensusComponents.Create()
//	if err != nil {
//		log.Error("starting node failed", "epoch", currentEpoch, "error", err.Error())
//		return err
//	}
//
//
//	time.Sleep(5 * time.Second)
//
//	_ = managedProcessComponents.Close()
//	_ = managedStatusComponents.Close()
//	_ = managedStateComponents.Close()
//	_ = managedDataComponents.Close()
//	_ = managedBootstrapComponents.Close()
//	_ = managedNetworkComponents.Close()
//	_ = managedCryptoComponents.Close()
//	_ = managedCoreComponents.Close()
//
//	time.Sleep(5 * time.Second)
//
//	nrAfter := runtime.NumGoroutine()
//	if nrBefore != nrAfter {
//		PrintStack()
//	}
//
//	require.Equal(t, nrBefore, nrAfter)
//}
