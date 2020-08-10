package factory

/*
func TestNode_StartConsensusGenesisBlockNotInitializedShouldErr(t *testing.T) {
	t.Parallel()

	n, _ := node.NewNode(
		node.WithBlockChain(&mock.ChainHandlerStub{
			GetGenesisHeaderHashCalled: func() []byte {
				return nil
			},
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return nil
			},
		}),
	)

	err := n.StartConsensus()

	assert.Equal(t, node.ErrGenesisBlockNotInitialized, err)
}

func TestStartConsensus_NilSyncTimer(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}

	n, _ := node.NewNode(
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
	)

	err := n.StartConsensus()
	assert.Equal(t, chronology.ErrNilSyncTimer, err)
}

func TestStartConsensus_ShardBootstrapperNilAccounts(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	rf := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
			return &mock.MiniBlocksResolverStub{}, nil
		},
		CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
			return &mock.HeaderResolverStub{}, nil
		},
	}

	store := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return nil
		},
	}

	n, _ := node.NewNode(
		node.WithDataPool(&testscommon.PoolsHolderStub{
			MiniBlocksCalled: func() storage.Cacher {
				return &testscommon.CacherStub{
					RegisterHandlerCalled: func(f func(key []byte, value interface{})) {

					},
				}
			},
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{
					RegisterHandlerCalled: func(handler func(header data.HeaderHandler, shardHeaderHash []byte)) {

					},
				}
			},
		}),
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncTimerStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithResolversFinder(rf),
		node.WithDataStore(store),
		node.WithHasher(&mock.HasherMock{}),
		node.WithInternalMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithForkDetector(&mock.ForkDetectorMock{
			CheckForkCalled: func() *process.ForkInfo {
				return &process.ForkInfo{}
			},
			ProbableHighestNonceCalled: func() uint64 {
				return 0
			},
		}),
		node.WithBootStorer(&mock.BoostrapStorerMock{
			GetHighestRoundCalled: func() int64 {
				return 0
			},
			GetCalled: func(round int64) (bootstrapData bootstrapStorage.BootstrapData, err error) {
				return bootstrapStorage.BootstrapData{}, errors.New("localErr")
			},
		}),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
		node.WithBlockProcessor(&mock.BlockProcessorStub{}),
		node.WithNodesCoordinator(&mock.NodesCoordinatorMock{}),
		node.WithRequestHandler(&mock.RequestHandlerStub{}),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
		node.WithBlockTracker(&mock.BlockTrackerStub{}),
		node.WithDataStore(&mock.ChainStorerMock{}),
		node.WithWatchdogTimer(&mock.WatchdogMock{}),
	)

	err := n.StartConsensus()
	assert.Equal(t, state.ErrNilAccountsAdapter, err)
}

func TestStartConsensus_ShardBootstrapperNilPoolHolder(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	rf := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
			return &mock.MiniBlocksResolverStub{}, nil
		},
	}

	store := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return &mock.StorerStub{}
		},
	}

	accountDb, _ := state.NewAccountsDB(&mock.TrieStub{}, &mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.AccountsFactoryStub{})

	n, _ := node.NewNode(
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncTimerStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithAccountsAdapter(accountDb),
		node.WithResolversFinder(rf),
		node.WithDataStore(store),
		node.WithBootStorer(&mock.BoostrapStorerMock{}),
		node.WithForkDetector(&mock.ForkDetectorMock{}),
		node.WithBlockProcessor(&mock.BlockProcessorStub{}),
		node.WithInternalMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithTxSignMarshalizer(&mock.MarshalizerMock{}),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
		node.WithNodesCoordinator(&mock.NodesCoordinatorMock{}),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
		node.WithBlockTracker(&mock.BlockTrackerStub{}),
		node.WithWatchdogTimer(&mock.WatchdogMock{}),
	)

	err := n.StartConsensus()
	assert.Equal(t, process.ErrNilPoolsHolder, err)
}

func TestStartConsensus_MetaBootstrapperNilPoolHolder(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	shardingCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	shardingCoordinator.CurrentShard = core.MetachainShardId
	store := &mock.ChainStorerMock{
		GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
			return nil
		},
	}
	n, _ := node.NewNode(
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncTimerStub{}),
		node.WithShardCoordinator(shardingCoordinator),
		node.WithDataStore(store),
		node.WithResolversFinder(&mock.ResolversFinderStub{
			IntraShardResolverCalled: func(baseTopic string) (dataRetriever.Resolver, error) {
				return &mock.MiniBlocksResolverStub{}, nil
			},
		}),
		node.WithBootStorer(&mock.BoostrapStorerMock{}),
		node.WithForkDetector(&mock.ForkDetectorMock{}),
		node.WithBlockTracker(&mock.BlockTrackerStub{}),
		node.WithBlockProcessor(&mock.BlockProcessorStub{}),
		node.WithInternalMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithTxSignMarshalizer(&mock.MarshalizerMock{}),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
		node.WithNodesCoordinator(&mock.NodesCoordinatorMock{}),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
		node.WithPendingMiniBlocksHandler(&mock.PendingMiniBlocksHandlerStub{}),
		node.WithWatchdogTimer(&mock.WatchdogMock{}),
	)

	err := n.StartConsensus()
	assert.Equal(t, process.ErrNilPoolsHolder, err)
}

func TestStartConsensus_MetaBootstrapperWrongNumberShards(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	shardingCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	shardingCoordinator.CurrentShard = 2
	n, _ := node.NewNode(
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncTimerStub{}),
		node.WithShardCoordinator(shardingCoordinator),
		node.WithDataStore(&mock.ChainStorerMock{}),
		node.WithDataPool(testscommon.NewPoolsHolderStub()),
		node.WithInternalMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithWatchdogTimer(&mock.WatchdogMock{}),
	)

	err := n.StartConsensus()
	assert.Equal(t, sharding.ErrShardIdOutOfRange, err)
}

func TestStartConsensus_ShardBootstrapperPubKeyToByteArrayError(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	rf := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
			return &mock.MiniBlocksResolverStub{}, nil
		},
		CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
			return &mock.HeaderResolverStub{}, nil
		},
	}
	accountDb, _ := state.NewAccountsDB(&mock.TrieStub{}, &mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.AccountsFactoryStub{})

	localErr := errors.New("err")
	n, _ := node.NewNode(
		node.WithDataPool(&testscommon.PoolsHolderStub{
			MiniBlocksCalled: func() storage.Cacher {
				return &testscommon.CacherStub{
					RegisterHandlerCalled: func(f func(key []byte, value interface{})) {

					},
				}
			},
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{
					RegisterHandlerCalled: func(handler func(header data.HeaderHandler, shardHeaderHash []byte)) {

					},
				}
			},
		}),
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncTimerStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithAccountsAdapter(accountDb),
		node.WithResolversFinder(rf),
		node.WithDataStore(&mock.ChainStorerMock{}),
		node.WithHasher(&mock.HasherMock{}),
		node.WithInternalMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithForkDetector(&mock.ForkDetectorMock{}),
		node.WithBlockBlackListHandler(&mock.TimeCacheStub{}),
		node.WithMessenger(&mock.MessengerStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return false
			},
		}),
		node.WithBootStorer(&mock.BoostrapStorerMock{
			GetHighestRoundCalled: func() int64 {
				return 0
			},
			GetCalled: func(round int64) (bootstrapData bootstrapStorage.BootstrapData, err error) {
				return bootstrapStorage.BootstrapData{}, errors.New("localErr")
			},
		}),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
		node.WithRequestedItemsHandler(&mock.TimeCacheStub{}),
		node.WithBlockProcessor(&mock.BlockProcessorStub{}),
		node.WithPubKey(&mock.PublicKeyMock{
			ToByteArrayHandler: func() (i []byte, err error) {
				return []byte("nil"), localErr
			},
		}),
		node.WithRequestHandler(&mock.RequestHandlerStub{}),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
		node.WithNodesCoordinator(&mock.NodesCoordinatorMock{}),
		node.WithBlockTracker(&mock.BlockTrackerStub{}),
		node.WithInternalMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithWatchdogTimer(&mock.WatchdogMock{}),
	)

	err := n.StartConsensus()
	assert.Equal(t, localErr, err)
}

func TestStartConsensus_ShardBootstrapperInvalidConsensusType(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	rf := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
			return &mock.MiniBlocksResolverStub{}, nil
		},
		CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
			return &mock.HeaderResolverStub{}, nil
		},
	}

	accountDb, _ := state.NewAccountsDB(&mock.TrieStub{}, &mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.AccountsFactoryStub{})

	n, _ := node.NewNode(
		node.WithDataPool(&testscommon.PoolsHolderStub{
			MiniBlocksCalled: func() storage.Cacher {
				return &testscommon.CacherStub{
					RegisterHandlerCalled: func(f func(key []byte, value interface{})) {

					},
				}
			},
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{
					RegisterHandlerCalled: func(handler func(header data.HeaderHandler, shardHeaderHash []byte)) {

					},
				}
			},
		}),
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncTimerStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithAccountsAdapter(accountDb),
		node.WithResolversFinder(rf),
		node.WithDataStore(&mock.ChainStorerMock{}),
		node.WithHasher(&mock.HasherMock{}),
		node.WithInternalMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithForkDetector(&mock.ForkDetectorMock{}),
		node.WithBlockBlackListHandler(&mock.TimeCacheStub{}),
		node.WithMessenger(&mock.MessengerStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return false
			},
		}),
		node.WithBootStorer(&mock.BoostrapStorerMock{
			GetHighestRoundCalled: func() int64 {
				return 0
			},
			GetCalled: func(round int64) (bootstrapData bootstrapStorage.BootstrapData, err error) {
				return bootstrapStorage.BootstrapData{}, errors.New("localErr")
			},
		}),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
		node.WithRequestedItemsHandler(&mock.TimeCacheStub{}),
		node.WithBlockProcessor(&mock.BlockProcessorStub{}),
		node.WithPubKey(&mock.PublicKeyMock{
			ToByteArrayHandler: func() (i []byte, err error) {
				return []byte("keyBytes"), nil
			},
		}),
		node.WithRequestHandler(&mock.RequestHandlerStub{}),
		node.WithNodesCoordinator(&mock.NodesCoordinatorMock{}),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
		node.WithBlockTracker(&mock.BlockTrackerStub{}),
		node.WithWatchdogTimer(&mock.WatchdogMock{}),
	)

	err := n.StartConsensus()
	assert.Equal(t, sposFactory.ErrInvalidConsensusType, err)
}

func TestStartConsensus_ShardBootstrapper(t *testing.T) {
	t.Parallel()

	chainHandler := &mock.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("hdrHash")
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	rf := &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
			return &mock.MiniBlocksResolverStub{}, nil
		},
		CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
			return &mock.HeaderResolverStub{}, nil
		},
	}

	accountDb, _ := state.NewAccountsDB(&mock.TrieStub{}, &mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.AccountsFactoryStub{})

	n, _ := node.NewNode(
		node.WithDataPool(&testscommon.PoolsHolderStub{
			MiniBlocksCalled: func() storage.Cacher {
				return &testscommon.CacherStub{
					RegisterHandlerCalled: func(f func(key []byte, value interface{})) {

					},
				}
			},
			HeadersCalled: func() dataRetriever.HeadersPool {
				return &mock.HeadersCacherStub{
					RegisterHandlerCalled: func(handler func(header data.HeaderHandler, shardHeaderHash []byte)) {

					},
				}
			},
		}),
		node.WithBlockChain(chainHandler),
		node.WithRounder(&mock.RounderMock{}),
		node.WithGenesisTime(time.Now().Local()),
		node.WithSyncer(&mock.SyncTimerStub{}),
		node.WithShardCoordinator(mock.NewOneShardCoordinatorMock()),
		node.WithAccountsAdapter(accountDb),
		node.WithResolversFinder(rf),
		node.WithDataStore(&mock.ChainStorerMock{}),
		node.WithHasher(&mock.HasherMock{}),
		node.WithInternalMarshalizer(&mock.MarshalizerMock{}, 0),
		node.WithForkDetector(&mock.ForkDetectorMock{
			CheckForkCalled: func() *process.ForkInfo {
				return &process.ForkInfo{}
			},
			ProbableHighestNonceCalled: func() uint64 {
				return 0
			},
		}),
		node.WithBlockBlackListHandler(&mock.TimeCacheStub{}),
		node.WithMessenger(&mock.MessengerStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return false
			},
			HasTopicValidatorCalled: func(name string) bool {
				return false
			},
			HasTopicCalled: func(name string) bool {
				return true
			},
			RegisterMessageProcessorCalled: func(topic string, handler p2p.MessageProcessor) error {
				return nil
			},
		}),
		node.WithBootStorer(&mock.BoostrapStorerMock{
			GetHighestRoundCalled: func() int64 {
				return 0
			},
			GetCalled: func(round int64) (bootstrapData bootstrapStorage.BootstrapData, err error) {
				return bootstrapStorage.BootstrapData{}, errors.New("localErr")
			},
		}),
		node.WithEpochStartTrigger(&mock.EpochStartTriggerStub{}),
		node.WithRequestedItemsHandler(&mock.TimeCacheStub{}),
		node.WithBlockProcessor(&mock.BlockProcessorStub{}),
		node.WithPubKey(&mock.PublicKeyMock{
			ToByteArrayHandler: func() (i []byte, err error) {
				return []byte("keyBytes"), nil
			},
		}),
		node.WithConsensusType("bls"),
		node.WithPrivKey(&mock.PrivateKeyStub{}),
		node.WithSingleSigner(&mock.SingleSignerMock{}),
		node.WithKeyGen(&mock.KeyGenMock{}),
		node.WithChainID([]byte("id")),
		node.WithHeaderSigVerifier(&mock.HeaderSigVerifierStub{}),
		node.WithMultiSigner(&mock.MultisignMock{}),
		node.WithValidatorStatistics(&mock.ValidatorStatisticsProcessorStub{}),
		node.WithNodesCoordinator(&mock.NodesCoordinatorMock{}),
		node.WithEpochStartEventNotifier(&mock.EpochStartNotifierStub{}),
		node.WithRequestHandler(&mock.RequestHandlerStub{}),
		node.WithUint64ByteSliceConverter(mock.NewNonceHashConverterMock()),
		node.WithBlockTracker(&mock.BlockTrackerStub{}),
		node.WithNetworkShardingCollector(&mock.NetworkShardingCollectorStub{}),
		node.WithInputAntifloodHandler(&mock.P2PAntifloodHandlerStub{}),
		node.WithHeaderIntegrityVerifier(&mock.HeaderIntegrityVerifierStub{}),
		node.WithPeerHonestyHandler(&testscommon.PeerHonestyHandlerStub{}),
		node.WithHardforkTrigger(&mock.HardforkTriggerStub{}),
		node.WithInterceptorsContainer(&mock.InterceptorsContainerStub{}),
		node.WithWatchdogTimer(&mock.WatchdogMock{}),
		node.WithPeerSignatureHandler(&mock.PeerSignatureHandler{}),
	)

	err := n.StartConsensus()
	assert.Nil(t, err)
}
*/
