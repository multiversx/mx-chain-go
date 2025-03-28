package sharding

import (
	"encoding/hex"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/sharding/mock"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
)

var (
	pubKeys = []string{
		"41378f754e2c7b2745208c3ed21b151d297acdc84c3aca00b9e292cf28ec2d444771070157ea7760ed83c26f4fed387d0077e00b563a95825dac2cbc349fc0025ccf774e37b0a98ad9724d30e90f8c29b4091ccb738ed9ffc0573df776ee9ea30b3c038b55e532760ea4a8f152f2a52848020e5cee1cc537f2c2323399723081",
		"52f3bf5c01771f601ec2137e267319ab6716ef6ff5dfddaea48b42d955f631167f2ce19296a202bb8fd174f4e94f8c85f619df85a7f9f8de0f3768e5e6d8c48187b767deccf9829be246aa331aa86d182eb8fa28ea8a3e45d357ed1647a9be020a5569d686253a6f89e9123c7f21f302e82f67d3e3cd69cf267b9910a663ef32",
		"5e91c426c5c8f5f805f86de1e0653e2ec33853772e583b88e9f0f201089d03d8570759c3c3ab610ce573493c33ba0adf954c8939dba5d5ef7f2be4e87145d8153fc5b4fb91cecb8d9b1f62e080743fbf69c8c3096bf07980bb82cb450ba9b902673373d5b671ea73620cc5bc4d36f7a0f5ca3684d4c8aa5c1b425ab2a8673140",
		"73972bf46dca59fba211c58f11b530f8e9d6392c499655ce760abc6458fd9c6b54b9676ee4b95aa32f6c254c9aad2f63a6195cd65d837a4320d7b8e915ba3a7123c8f4983b201035573c0752bb54e9021eb383b40d302447b62ea7a3790c89c47f5ab81d183f414e87611a31ff635ad22e969495356d5bc44eec7917aaad4c5e",
		"7391ccce066ab5674304b10220643bc64829afa626a165f1e7a6618e260fa68f8e79018ac5964f7a1b8dd419645049042e34ebe7f2772def71e6176ce9daf50a57c17ee2a7445b908fe47e8f978380fcc2654a19925bf73db2402b09dde515148081f8ca7c331fbedec689de1b7bfce6bf106e4433557c29752c12d0a009f47a",
		"24dea9b5c79174c558c38316b2df25b956c53f0d0128b7427d219834867cc1b0868b7faff0205fe23e5ffdf276acfad6423890c782c7be7b98a31d501e4276a015a54d9849109322130fc9a9cb61d183318d50fcde44fabcbf600051c7cb950304b05e82f90f2ac4647016f39439608cd64ccc82fe6e996289bb2150e4e3ab08",
	}

	address = []string{
		"9e95a4e46da335a96845b4316251fc1bb197e1b8136d96ecc62bf6604eca9e49",
		"7a330039e77ca06bc127319fd707cc4911a80db489a39fcfb746283a05f61836",
		"131e2e717f2d33bdf7850c12b03dfe41ea8a5e76fdd6d4f23aebe558603e746f",
		"4c9e66b605882c1099088f26659692f084e41dc0dedfaedf6a6409af21c02aac",
		"90a66900634b206d20627fbaec432ebfbabeaf30b9e338af63191435e2e37022",
		"63f702e061385324a25dc4f1bcfc7e4f4692bcd80de71bd4dd7d6e2f67f92481",
	}
)

func createAndAssignNodes(ns *NodesSetup, numInitialNodes int) *NodesSetup {
	ns.InitialNodes = make([]*InitialNode, numInitialNodes)

	for i := 0; i < numInitialNodes; i++ {
		lookupIndex := i % len(pubKeys)
		ns.InitialNodes[i] = &InitialNode{}
		ns.InitialNodes[i].PubKey = pubKeys[lookupIndex]
		ns.InitialNodes[i].Address = address[lookupIndex]
	}

	err := ns.processConfig()
	if err != nil {
		return nil
	}

	ns.processMetaChainAssigment()
	ns.processShardAssignment()
	ns.createInitialNodesInfo()

	return ns
}

type argsTestNodesSetup struct {
	shardConsensusSize uint32
	shardMinNodes      uint32
	metaConsensusSize  uint32
	metaMinNodes       uint32
	numInitialNodes    uint32
	genesisMaxShards   uint32
}

func createTestNodesSetup(args argsTestNodesSetup) (*NodesSetup, error) {
	initialNodes := make([]*config.InitialNodeConfig, 0)
	for i := 0; uint32(i) < args.numInitialNodes; i++ {
		lookupIndex := i % len(pubKeys)
		initialNodes = append(initialNodes, &config.InitialNodeConfig{
			PubKey:  pubKeys[lookupIndex],
			Address: address[lookupIndex],
		})
	}
	ns, err := NewNodesSetup(
		config.NodesConfig{
			StartTime:    0,
			InitialNodes: initialNodes,
		},
		&chainParameters.ChainParametersHandlerStub{
			ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
				return config.ChainParametersByEpochConfig{
					EnableEpoch:                 0,
					ShardMinNumNodes:            args.shardMinNodes,
					ShardConsensusGroupSize:     args.shardConsensusSize,
					MetachainMinNumNodes:        args.metaMinNodes,
					MetachainConsensusGroupSize: args.metaConsensusSize,
				}, nil
			},
		},
		mock.NewPubkeyConverterMock(32),
		mock.NewPubkeyConverterMock(96),
		args.genesisMaxShards,
	)

	return ns, err
}

func createTestNodesSetupWithSpecificMockedComponents(args argsTestNodesSetup,
	initialNodes []*config.InitialNodeConfig,
	addressPubkeyConverter core.PubkeyConverter,
	validatorPubkeyConverter core.PubkeyConverter) (*NodesSetup, error) {

	ns, err := NewNodesSetup(
		config.NodesConfig{
			StartTime:    0,
			InitialNodes: initialNodes,
		},
		&chainParameters.ChainParametersHandlerStub{
			ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
				return config.ChainParametersByEpochConfig{
					EnableEpoch:                 0,
					ShardMinNumNodes:            args.shardMinNodes,
					ShardConsensusGroupSize:     args.shardConsensusSize,
					MetachainMinNumNodes:        args.metaMinNodes,
					MetachainConsensusGroupSize: args.metaConsensusSize,
				}, nil
			},
		},
		addressPubkeyConverter,
		validatorPubkeyConverter,
		args.genesisMaxShards,
	)

	return ns, err
}

func TestNodesSetup_ProcessConfigNodesWithIncompleteDataShouldErr(t *testing.T) {
	t.Parallel()

	noOfInitialNodes := 2
	ns := &NodesSetup{
		addressPubkeyConverter:   mock.NewPubkeyConverterMock(32),
		validatorPubkeyConverter: mock.NewPubkeyConverterMock(96),
	}

	ns.InitialNodes = make([]*InitialNode, noOfInitialNodes)

	ns.InitialNodes[0] = &InitialNode{}
	ns.InitialNodes[1] = &InitialNode{}

	ns.InitialNodes[0].PubKey = pubKeys[0]
	ns.InitialNodes[0].Address = address[0]

	err := ns.processConfig()

	require.NotNil(t, ns)
	require.Equal(t, ErrCouldNotParsePubKey, err)
}

func TestNodesSetup_ProcessConfigNodesShouldErrCouldNotParsePubKeyForString(t *testing.T) {
	t.Parallel()

	mockedNodes := make([]*config.InitialNodeConfig, 2)
	mockedNodes[0] = &config.InitialNodeConfig{
		PubKey:  pubKeys[0],
		Address: address[0],
	}

	mockedNodes[1] = &config.InitialNodeConfig{
		PubKey:  pubKeys[1],
		Address: address[1],
	}

	addressPubkeyConverterMocked := mock.NewPubkeyConverterMock(32)
	validatorPubkeyConverterMocked := &mock.PubkeyConverterMock{
		DecodeCalled: func() ([]byte, error) {
			return nil, ErrCouldNotParsePubKey
		},
	}

	_, err := createTestNodesSetupWithSpecificMockedComponents(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      1,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    2,
		genesisMaxShards:   3,
	},
		mockedNodes,
		addressPubkeyConverterMocked,
		validatorPubkeyConverterMocked,
	)

	require.ErrorIs(t, err, ErrCouldNotParsePubKey)
}

func TestNodesSetup_ProcessConfigNodesShouldErrCouldNotParseAddressForString(t *testing.T) {
	t.Parallel()

	mockedNodes := make([]*config.InitialNodeConfig, 2)
	mockedNodes[0] = &config.InitialNodeConfig{
		PubKey:  pubKeys[0],
		Address: address[0],
	}

	mockedNodes[1] = &config.InitialNodeConfig{
		PubKey:  pubKeys[1],
		Address: address[1],
	}

	addressPubkeyConverterMocked := &mock.PubkeyConverterMock{
		DecodeCalled: func() ([]byte, error) {
			return nil, ErrCouldNotParseAddress
		},
	}
	validatorPubkeyConverterMocked := mock.NewPubkeyConverterMock(96)

	_, err := createTestNodesSetupWithSpecificMockedComponents(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      1,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    2,
		genesisMaxShards:   3,
	},
		mockedNodes,
		addressPubkeyConverterMocked,
		validatorPubkeyConverterMocked,
	)

	require.ErrorIs(t, err, ErrCouldNotParseAddress)
}

func TestNodesSetup_ProcessConfigNodesWithEmptyDataShouldErrCouldNotParseAddress(t *testing.T) {
	t.Parallel()

	mockedNodes := make([]*config.InitialNodeConfig, 2)
	mockedNodes[0] = &config.InitialNodeConfig{
		PubKey:  pubKeys[0],
		Address: address[0],
	}

	mockedNodes[1] = &config.InitialNodeConfig{
		PubKey:  pubKeys[1],
		Address: "",
	}

	addressPubkeyConverterMocked := mock.NewPubkeyConverterMock(32)
	validatorPubkeyConverterMocked := mock.NewPubkeyConverterMock(96)

	_, err := createTestNodesSetupWithSpecificMockedComponents(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      1,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    2,
		genesisMaxShards:   3,
	},
		mockedNodes,
		addressPubkeyConverterMocked,
		validatorPubkeyConverterMocked,
	)

	require.ErrorIs(t, err, ErrCouldNotParseAddress)
}

func TestNodesSetup_ProcessConfigInvalidConsensusGroupSizeShouldErr(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 0,
		shardMinNodes:      0,
		metaConsensusSize:  0,
		metaMinNodes:       0,
		numInitialNodes:    0,
		genesisMaxShards:   3,
	})
	require.Equal(t, ErrNegativeOrZeroConsensusGroupSize, err)
	require.Nil(t, ns)
}

func TestNodesSetup_ProcessConfigInvalidMetaConsensusGroupSizeShouldErr(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      1,
		metaConsensusSize:  0,
		metaMinNodes:       0,
		numInitialNodes:    1,
		genesisMaxShards:   3,
	})
	require.Equal(t, ErrNegativeOrZeroConsensusGroupSize, err)
	require.Nil(t, ns)
}

func TestNodesSetup_ProcessConfigInvalidConsensusGroupSizeLargerThanNumOfNodesShouldErr(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 2,
		shardMinNodes:      0,
		metaConsensusSize:  0,
		metaMinNodes:       0,
		numInitialNodes:    2,
		genesisMaxShards:   3,
	})
	require.Equal(t, ErrMinNodesPerShardSmallerThanConsensusSize, err)
	require.Nil(t, ns)
}

func TestNodesSetup_ProcessConfigInvalidMetaConsensusGroupSizeLargerThanNumOfNodesShouldErr(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      1,
		metaConsensusSize:  2,
		metaMinNodes:       1,
		numInitialNodes:    2,
		genesisMaxShards:   3,
	})
	require.Equal(t, ErrMinNodesPerShardSmallerThanConsensusSize, err)
	require.Nil(t, ns)
}

func TestNodesSetup_ProcessConfigInvalidNumOfNodesSmallerThanMinNodesPerShardShouldErr(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 2,
		shardMinNodes:      3,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    2,
		genesisMaxShards:   3,
	})
	require.Nil(t, ns)
	require.Equal(t, ErrNodesSizeSmallerThanMinNoOfNodes, err)
}

func TestNodesSetup_ProcessConfigInvalidNumOfNodesSmallerThanTotalMinNodesShouldErr(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 2,
		shardMinNodes:      3,
		metaConsensusSize:  1,
		metaMinNodes:       3,
		numInitialNodes:    5,
		genesisMaxShards:   3,
	})
	require.Nil(t, ns)
	require.Equal(t, ErrNodesSizeSmallerThanMinNoOfNodes, err)
}

func TestNodesSetup_InitialNodesPubKeysWithHysteresis(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 63,
		shardMinNodes:      400,
		metaConsensusSize:  400,
		metaMinNodes:       400,
		numInitialNodes:    3000,
		genesisMaxShards:   100,
	})
	ns.Hysteresis = 0.2
	ns.Adaptivity = false
	require.NoError(t, err)

	ns = createAndAssignNodes(ns, 3000)
	require.Equal(t, 6, len(ns.eligible))
	for shard, shardNodes := range ns.eligible {
		require.Equal(t, 400, len(shardNodes))
		require.Equal(t, 100, len(ns.waiting[shard]))
	}

	ns = createAndAssignNodes(ns, 3570)
	require.Equal(t, 7, len(ns.eligible))
	for shard, shardNodes := range ns.eligible {
		require.Equal(t, 400, len(shardNodes))
		require.Equal(t, 110, len(ns.waiting[shard]))
	}

	ns = createAndAssignNodes(ns, 2400)
	require.Equal(t, 5, len(ns.eligible))
	for shard, shardNodes := range ns.eligible {
		require.Equal(t, 400, len(shardNodes))
		require.Equal(t, 80, len(ns.waiting[shard]))
	}
}

func TestNodesSetup_InitialNodesPubKeysForShardWrongShard(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      1,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    2,
		genesisMaxShards:   3,
	})
	require.NoError(t, err)
	eligible, waiting, err := ns.InitialNodesInfoForShard(1)

	require.NotNil(t, ns)
	require.Nil(t, eligible)
	require.Nil(t, waiting)
	require.NotNil(t, err)
}

func TestNodesSetup_InitialNodesPubKeysForShardGood(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      2,
		metaConsensusSize:  1,
		metaMinNodes:       2,
		numInitialNodes:    6,
		genesisMaxShards:   3,
	})
	require.NoError(t, err)

	eligible, waiting, err := ns.InitialNodesInfoForShard(1)

	require.NotNil(t, ns)
	require.Equal(t, 2, len(eligible))
	require.Equal(t, 0, len(waiting))
	require.Nil(t, err)
}

func TestNodesSetup_InitialNodesPubKeysForShardGoodMeta(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      2,
		metaConsensusSize:  2,
		metaMinNodes:       2,
		numInitialNodes:    6,
		genesisMaxShards:   3,
	})
	require.NoError(t, err)
	metaId := core.MetachainShardId
	eligible, waiting, err := ns.InitialNodesInfoForShard(metaId)

	require.NotNil(t, ns)
	require.Equal(t, 2, len(eligible))
	require.Equal(t, 0, len(waiting))
	require.Nil(t, err)
}

func TestNodesSetup_PublicKeyNotGood(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      5,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    6,
		genesisMaxShards:   3,
	})
	require.NoError(t, err)

	_, err = ns.GetShardIDForPubKey([]byte(pubKeys[0]))

	require.NotNil(t, ns)
	require.NotNil(t, err)
}

func TestNodesSetup_PublicKeyGood(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      5,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    6,
		genesisMaxShards:   3,
	})
	require.NoError(t, err)

	publicKey, _ := hex.DecodeString(pubKeys[2])

	selfId, err := ns.GetShardIDForPubKey(publicKey)

	require.NotNil(t, ns)
	require.Nil(t, err)
	require.Equal(t, uint32(0), selfId)
}

func TestNodesSetup_ShardPublicKeyGoodMeta(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      5,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    6,
		genesisMaxShards:   3,
	})
	require.NoError(t, err)
	publicKey, _ := hex.DecodeString(pubKeys[2])

	selfId, err := ns.GetShardIDForPubKey(publicKey)

	require.NotNil(t, ns)
	require.Nil(t, err)
	require.Equal(t, uint32(0), selfId)
}

func TestNodesSetup_MetaPublicKeyGoodMeta(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      5,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    6,
		genesisMaxShards:   3,
	})
	require.NoError(t, err)
	metaId := core.MetachainShardId
	publicKey, _ := hex.DecodeString(pubKeys[0])

	selfId, err := ns.GetShardIDForPubKey(publicKey)

	require.NotNil(t, ns)
	require.Nil(t, err)
	require.Equal(t, metaId, selfId)
}

func TestNodesSetup_MinNumberOfNodes(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 63,
		shardMinNodes:      400,
		metaConsensusSize:  400,
		metaMinNodes:       400,
		numInitialNodes:    2169,
		genesisMaxShards:   3,
	})
	ns.Hysteresis = 0.2
	ns.Adaptivity = false
	require.NoError(t, err)

	ns = createAndAssignNodes(ns, 2169)
	require.Equal(t, 4, len(ns.eligible))
	for shard, shardNodes := range ns.eligible {
		require.Equal(t, 400, len(shardNodes))
		require.LessOrEqual(t, len(ns.waiting[shard]), 143)
		require.GreaterOrEqual(t, len(ns.waiting[shard]), 142)
	}

	minNumNodes := ns.MinNumberOfNodes()
	require.Equal(t, uint32(1600), minNumNodes)

	minHysteresisNodesShard := ns.MinShardHysteresisNodes()
	require.Equal(t, uint32(80), minHysteresisNodesShard)

	minHysteresisNodesMeta := ns.MinMetaHysteresisNodes()
	require.Equal(t, uint32(80), minHysteresisNodesMeta)
}

func TestNewNodesSetup_InvalidMaxNumShardsShouldErr(t *testing.T) {
	t.Parallel()

	ns, err := NewNodesSetup(
		config.NodesConfig{},
		&chainParameters.ChainParametersHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		mock.NewPubkeyConverterMock(96),
		0,
	)

	require.Nil(t, ns)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), ErrInvalidMaximumNumberOfShards.Error())
}

func TestNewNodesSetup_ErrNilPubkeyConverterForAddressPubkeyConverter(t *testing.T) {
	t.Parallel()

	_, err := NewNodesSetup(
		config.NodesConfig{},
		&chainParameters.ChainParametersHandlerStub{},
		nil,
		mock.NewPubkeyConverterMock(96),
		3,
	)

	require.ErrorIs(t, err, ErrNilPubkeyConverter)
}

func TestNewNodesSetup_ErrNilPubkeyConverterForValidatorPubkeyConverter(t *testing.T) {
	t.Parallel()

	_, err := NewNodesSetup(
		config.NodesConfig{},
		&chainParameters.ChainParametersHandlerStub{},
		mock.NewPubkeyConverterMock(32),
		nil,
		3,
	)

	require.ErrorIs(t, err, ErrNilPubkeyConverter)
}

func TestNewNodesSetup_ErrNilChainParametersProvider(t *testing.T) {
	t.Parallel()

	_, err := NewNodesSetup(
		config.NodesConfig{},
		nil,
		mock.NewPubkeyConverterMock(32),
		mock.NewPubkeyConverterMock(96),
		3,
	)

	require.Equal(t, err, ErrNilChainParametersProvider)
}

func TestNewNodesSetup_ErrChainParametersForEpoch(t *testing.T) {
	t.Parallel()

	_, err := NewNodesSetup(
		config.NodesConfig{},
		&chainParameters.ChainParametersHandlerStub{
			ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
				return config.ChainParametersByEpochConfig{}, ErrInvalidChainParametersForEpoch
			},
		},
		mock.NewPubkeyConverterMock(32),
		mock.NewPubkeyConverterMock(96),
		3,
	)

	require.ErrorIs(t, err, ErrInvalidChainParametersForEpoch)
}

func TestNodesSetup_IfNodesWithinMaxShardLimitEquivalentDistribution(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 64,
		shardMinNodes:      400,
		metaConsensusSize:  400,
		metaMinNodes:       400,
		numInitialNodes:    2169,
		genesisMaxShards:   3,
	})
	ns.Hysteresis = 0.2
	ns.Adaptivity = false
	require.NoError(t, err)

	ns2 := &(*ns) //nolint
	ns2.genesisMaxNumShards = 3
	ns2 = createAndAssignNodes(ns2, 2169)

	require.Equal(t, 4, len(ns.eligible))
	require.Equal(t, 4, len(ns2.eligible))
	for shard, shardNodes := range ns.eligible {
		require.Equal(t, len(shardNodes), len(ns2.eligible[shard]))
		require.Equal(t, len(ns.waiting[shard]), len(ns2.waiting[shard]))
		require.GreaterOrEqual(t, len(ns.waiting[shard]), 142)
		require.Equal(t, len(ns.waiting[shard]), len(ns2.waiting[shard]))
		for i, node := range shardNodes {
			require.Equal(t, node, ns2.eligible[shard][i])
		}
		for i, node := range ns.waiting[shard] {
			require.Equal(t, node, ns2.waiting[shard][i])
		}
	}

	minNumNodes := ns.MinNumberOfNodes()
	require.Equal(t, minNumNodes, ns2.MinNumberOfNodes())

	minHysteresisNodesShard := ns.MinShardHysteresisNodes()
	require.Equal(t, minHysteresisNodesShard, ns2.MinShardHysteresisNodes())

	minHysteresisNodesMeta := ns.MinMetaHysteresisNodes()
	require.Equal(t, minHysteresisNodesMeta, ns2.MinMetaHysteresisNodes())
}

func TestNodesSetup_NodesAboveMaxShardLimit(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 63,
		shardMinNodes:      400,
		metaConsensusSize:  400,
		metaMinNodes:       400,
		numInitialNodes:    3200,
		genesisMaxShards:   3,
	})
	ns.Hysteresis = 0.2
	ns.Adaptivity = false
	require.NoError(t, err)

	require.Equal(t, 4, len(ns.eligible))
	for shard, shardNodes := range ns.eligible {
		require.Equal(t, 400, len(shardNodes))
		require.Equal(t, len(ns.waiting[shard]), 400)
	}

	minNumNodes := ns.MinNumberOfNodes()
	require.Equal(t, uint32(1600), minNumNodes)

	minHysteresisNodesShard := ns.MinShardHysteresisNodes()
	require.Equal(t, uint32(80), minHysteresisNodesShard)

	minHysteresisNodesMeta := ns.MinMetaHysteresisNodes()
	require.Equal(t, uint32(80), minHysteresisNodesMeta)

	ns = createAndAssignNodes(ns, 3600)
	for shard, shardNodes := range ns.eligible {
		require.Equal(t, 400, len(shardNodes))
		require.Equal(t, len(ns.waiting[shard]), 500)
	}

	minNumNodes = ns.MinNumberOfNodes()
	require.Equal(t, uint32(1600), minNumNodes)

	minHysteresisNodesShard = ns.MinShardHysteresisNodes()
	require.Equal(t, uint32(80), minHysteresisNodesShard)

	minHysteresisNodesMeta = ns.MinMetaHysteresisNodes()
	require.Equal(t, uint32(80), minHysteresisNodesMeta)
}

func TestNodesSetup_AllInitialNodesShouldWork(t *testing.T) {
	t.Parallel()

	noOfInitialNodes := 2

	var listOfInitialNodes = [2]InitialNode{
		{
			PubKey:  pubKeys[0],
			Address: address[0],
		},
		{
			PubKey:  pubKeys[1],
			Address: address[1],
		},
	}

	var expectedConvertedPubKeys = make([][]byte, 2)
	pubKeyConverter := mock.NewPubkeyConverterMock(96)

	for i, nod := range listOfInitialNodes {
		convertedValue, err := pubKeyConverter.Decode(nod.PubKey)
		require.Nil(t, err)
		require.NotNil(t, convertedValue)
		expectedConvertedPubKeys[i] = convertedValue
	}

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      1,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    2,
		genesisMaxShards:   1,
	})

	require.Nil(t, err)
	ns.Hysteresis = 0.2
	ns.Adaptivity = false

	ns = createAndAssignNodes(ns, noOfInitialNodes)

	allInitialNodes := ns.AllInitialNodes()

	for i, expectedConvertedKey := range expectedConvertedPubKeys {
		require.Equal(t, expectedConvertedKey, allInitialNodes[i].PubKeyBytes())
	}

}

func TestNodesSetup_InitialNodesInfoShouldWork(t *testing.T) {
	t.Parallel()

	noOfInitialNodes := 3

	var listOfInitialNodes = [3]InitialNode{
		{
			PubKey:  pubKeys[0],
			Address: address[0],
		},
		{
			PubKey:  pubKeys[1],
			Address: address[1],
		},
		{
			PubKey:  pubKeys[2],
			Address: address[2],
		},
	}

	var listOfExpectedConvertedPubKeysEligibleNodes = make([][]byte, 2)
	pubKeyConverter := mock.NewPubkeyConverterMock(96)

	for i := 0; i < 2; i++ {
		convertedValue, err := pubKeyConverter.Decode(listOfInitialNodes[i].PubKey)
		require.Nil(t, err)
		require.NotNil(t, convertedValue)
		listOfExpectedConvertedPubKeysEligibleNodes[i] = convertedValue
	}

	var listOfExpectedConvertedPubKeysWaitingNode = make([][]byte, 1)
	listOfExpectedConvertedPubKeysWaitingNode[0], _ = pubKeyConverter.Decode(listOfInitialNodes[2].PubKey)

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      1,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    3,
		genesisMaxShards:   1,
	})
	require.Nil(t, err)
	ns.Hysteresis = 0.2
	ns.Adaptivity = false

	ns = createAndAssignNodes(ns, noOfInitialNodes)

	allEligibleNodes, allWaitingNodes := ns.InitialNodesInfo()

	require.Equal(t, listOfExpectedConvertedPubKeysEligibleNodes[0], allEligibleNodes[(core.MetachainShardId)][0].PubKeyBytes())
	require.Equal(t, listOfExpectedConvertedPubKeysEligibleNodes[1], allEligibleNodes[0][0].PubKeyBytes())
	require.Equal(t, listOfExpectedConvertedPubKeysWaitingNode[0], allWaitingNodes[(core.MetachainShardId)][0].PubKeyBytes())

}

func TestNodesSetup_InitialNodesPubKeysShouldWork(t *testing.T) {
	t.Parallel()

	noOfInitialNodes := 3

	var listOfInitialNodes = [3]InitialNode{
		{
			PubKey:  pubKeys[0],
			Address: address[0],
		},
		{
			PubKey:  pubKeys[1],
			Address: address[1],
		},
		{
			PubKey:  pubKeys[2],
			Address: address[2],
		},
	}

	var listOfExpectedConvertedPubKeysEligibleNodes = make([]string, 2)
	pubKeyConverter := mock.NewPubkeyConverterMock(96)

	for i := 0; i < 2; i++ {
		convertedValue, err := pubKeyConverter.Decode(listOfInitialNodes[i].PubKey)
		require.Nil(t, err)
		require.NotNil(t, convertedValue)
		listOfExpectedConvertedPubKeysEligibleNodes[i] = string(convertedValue)
	}

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      1,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    3,
		genesisMaxShards:   1,
	})
	require.Nil(t, err)
	ns.Hysteresis = 0.2
	ns.Adaptivity = false

	ns = createAndAssignNodes(ns, noOfInitialNodes)

	allEligibleNodes := ns.InitialNodesPubKeys()

	require.Equal(t, listOfExpectedConvertedPubKeysEligibleNodes[0], allEligibleNodes[(core.MetachainShardId)][0])
	require.Equal(t, listOfExpectedConvertedPubKeysEligibleNodes[1], allEligibleNodes[0][0])

}

func TestNodesSetup_InitialEligibleNodesPubKeysForShardShouldErrShardIdOutOfRange(t *testing.T) {
	t.Parallel()

	noOfInitialNodes := 3

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      1,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    3,
		genesisMaxShards:   1,
	})
	require.Nil(t, err)
	ns.Hysteresis = 0.2
	ns.Adaptivity = false

	ns = createAndAssignNodes(ns, noOfInitialNodes)

	returnedPubKeys, err := ns.InitialEligibleNodesPubKeysForShard(1)
	require.Nil(t, returnedPubKeys)
	require.Equal(t, ErrShardIdOutOfRange, err)

}

func TestNodesSetup_InitialEligibleNodesPubKeysForShardShouldWork(t *testing.T) {
	t.Parallel()

	noOfInitialNodes := 3

	var listOfInitialNodes = [3]InitialNode{
		{
			PubKey:  pubKeys[0],
			Address: address[0],
		},
		{
			PubKey:  pubKeys[1],
			Address: address[1],
		},
		{
			PubKey:  pubKeys[2],
			Address: address[2],
		},
	}

	var listOfExpectedPubKeysEligibleNodes = make([]string, 2)
	pubKeyConverter := mock.NewPubkeyConverterMock(96)

	for i := 0; i < 2; i++ {
		convertedValue, err := pubKeyConverter.Decode(listOfInitialNodes[i].PubKey)
		require.Nil(t, err)
		require.NotNil(t, convertedValue)
		listOfExpectedPubKeysEligibleNodes[i] = string(convertedValue)
	}

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      1,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    3,
		genesisMaxShards:   1,
	})
	require.Nil(t, err)
	ns.Hysteresis = 0.2
	ns.Adaptivity = false

	ns = createAndAssignNodes(ns, noOfInitialNodes)

	allEligibleNodes, err := ns.InitialEligibleNodesPubKeysForShard(0)

	require.Nil(t, err)
	require.Equal(t, listOfExpectedPubKeysEligibleNodes[1], allEligibleNodes[0])
}

func TestNodesSetup_NumberOfShardsShouldWork(t *testing.T) {
	t.Parallel()

	noOfInitialNodes := 3

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      1,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    3,
		genesisMaxShards:   1,
	})
	require.Nil(t, err)
	require.NotNil(t, ns)

	ns.Hysteresis = 0.2
	ns.Adaptivity = false

	ns = createAndAssignNodes(ns, noOfInitialNodes)

	require.NotNil(t, ns)

	valReturned := ns.NumberOfShards()
	require.Equal(t, uint32(1), valReturned)

	valReturned = ns.MinNumberOfNodesWithHysteresis()
	require.Equal(t, uint32(2), valReturned)

	valReturned = ns.MinNumberOfShardNodes()
	require.Equal(t, uint32(1), valReturned)

	valReturned = ns.MinNumberOfShardNodes()
	require.Equal(t, uint32(1), valReturned)

	shardConsensusGroupSize := ns.GetShardConsensusGroupSize()
	require.Equal(t, uint32(1), shardConsensusGroupSize)

	metaConsensusGroupSize := ns.GetMetaConsensusGroupSize()
	require.Equal(t, uint32(1), metaConsensusGroupSize)

	ns.Hysteresis = 0.5
	hysteresis := ns.GetHysteresis()
	require.Equal(t, float32(0.5), hysteresis)

	ns.Adaptivity = true
	adaptivity := ns.GetAdaptivity()
	require.True(t, adaptivity)

	ns.StartTime = 2
	startTime := ns.GetStartTime()
	require.Equal(t, int64(2), startTime)

	ns.RoundDuration = 2
	roundDuration := ns.GetRoundDuration()
	require.Equal(t, uint64(2), roundDuration)

}

func TestNodesSetup_ExportNodesConfigShouldWork(t *testing.T) {
	t.Parallel()

	noOfInitialNodes := 3

	ns, err := createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      1,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    3,
		genesisMaxShards:   1,
	})
	require.Nil(t, err)

	ns.Hysteresis = 0.2
	ns.Adaptivity = false
	ns.StartTime = 10

	ns = createAndAssignNodes(ns, noOfInitialNodes)
	configNodes := ns.ExportNodesConfig()

	require.Equal(t, int64(10), configNodes.StartTime)

	var expectedNodesConfigs = make([]config.InitialNodeConfig, len(configNodes.InitialNodes))
	var actualNodesConfigs = make([]config.InitialNodeConfig, len(configNodes.InitialNodes))

	for i, nodeConfig := range configNodes.InitialNodes {
		expectedNodesConfigs[i] = config.InitialNodeConfig{PubKey: pubKeys[i], Address: address[i], InitialRating: 0}
		actualNodesConfigs[i] = config.InitialNodeConfig{PubKey: nodeConfig.PubKey, Address: nodeConfig.Address, InitialRating: nodeConfig.InitialRating}

	}

	for i := range configNodes.InitialNodes {
		require.Equal(t, expectedNodesConfigs[i], actualNodesConfigs[i])
	}

}

func TestNodesSetup_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	ns, _ := NewNodesSetup(config.NodesConfig{}, nil, nil, nil, 0)
	require.True(t, ns.IsInterfaceNil())

	ns, _ = createTestNodesSetup(argsTestNodesSetup{
		shardConsensusSize: 1,
		shardMinNodes:      1,
		metaConsensusSize:  1,
		metaMinNodes:       1,
		numInitialNodes:    3,
		genesisMaxShards:   1,
	})
	require.False(t, ns.IsInterfaceNil())
}
