package sharding

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/stretchr/testify/require"
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

func createAndAssignNodes(ns *NodesSetup, noOfInitialNodes int) *NodesSetup {
	ns.InitialNodes = make([]*InitialNode, noOfInitialNodes)

	for i := 0; i < noOfInitialNodes; i++ {
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

func createTestNodesSetup(shardConsensusSize uint32, minShardNodes uint32, metaConsensusSize uint32, minMetaNodes uint32, numInitialNodes uint32, genesisMaxShards uint32) (*NodesSetup, error) {
	initialNodes := make([]*config.InitialNodeConfig, 0)
	for i := 0; uint32(i) < numInitialNodes; i++ {
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
		[]config.ChainParametersByEpochConfig{
			{
				EnableEpoch:                 0,
				ShardMinNodes:               minShardNodes,
				ShardConsensusGroupSize:     shardConsensusSize,
				MetachainMinNumNodes:        minMetaNodes,
				MetachainConsensusGroupSize: metaConsensusSize,
			},
		},
		mock.NewPubkeyConverterMock(32),
		mock.NewPubkeyConverterMock(96),
		genesisMaxShards,
		&epochNotifier.EpochNotifierStub{},
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

func TestNodesSetup_ProcessConfigInvalidConsensusGroupSizeShouldErr(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(0, 0, 0, 0, 0, 3)
	require.Equal(t, ErrNegativeOrZeroConsensusGroupSize, err)
	require.Nil(t, ns)
}

func TestNodesSetup_ProcessConfigInvalidMetaConsensusGroupSizeShouldErr(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(1, 1, 0, 0, 1, 3)
	require.Equal(t, ErrNegativeOrZeroConsensusGroupSize, err)
	require.Nil(t, ns)
}

func TestNodesSetup_ProcessConfigInvalidConsensusGroupSizeLargerThanNumOfNodesShouldErr(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(2, 0, 0, 0, 2, 3)
	require.Equal(t, ErrMinNodesPerShardSmallerThanConsensusSize, err)
	require.Nil(t, ns)
}

func TestNodesSetup_ProcessConfigInvalidMetaConsensusGroupSizeLargerThanNumOfNodesShouldErr(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(1, 1, 2, 1, 2, 3)
	require.Equal(t, ErrMinNodesPerShardSmallerThanConsensusSize, err)
	require.Nil(t, ns)
}

func TestNodesSetup_ProcessConfigInvalidNumOfNodesSmallerThanMinNodesPerShardShouldErr(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(2, 3, 1, 1, 2, 3)
	require.Nil(t, ns)
	require.Equal(t, ErrNodesSizeSmallerThanMinNoOfNodes, err)
}

func TestNodesSetup_InitialNodesPubKeysWithHysteresis(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(63, 400, 400, 400, 3000, 100)
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

	ns, _ := createTestNodesSetup(1, 1, 1, 1, 1, 3)
	eligible, waiting, err := ns.InitialNodesInfoForShard(1)

	require.NotNil(t, ns)
	require.Nil(t, eligible)
	require.Nil(t, waiting)
	require.NotNil(t, err)
}

func TestNodesSetup_InitialNodesPubKeysForShardGood(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(1, 2, 1, 2, 7, 3)
	require.NoError(t, err)

	eligible, waiting, err := ns.InitialNodesInfoForShard(1)

	require.NotNil(t, ns)
	require.Equal(t, 2, len(eligible))
	require.Equal(t, 0, len(waiting))
	require.Nil(t, err)
}

func TestNodesSetup_InitialNodesPubKeysForShardGoodMeta(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(1, 2, 2, 2, 7, 3)
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

	ns, err := createTestNodesSetup(1, 5, 1, 1, 6, 3)
	require.NoError(t, err)

	_, err = ns.GetShardIDForPubKey([]byte(pubKeys[0]))

	require.NotNil(t, ns)
	require.NotNil(t, err)
}

func TestNodesSetup_PublicKeyGood(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(1, 5, 1, 1, 6, 3)
	require.NoError(t, err)

	publicKey, _ := hex.DecodeString(pubKeys[2])

	selfId, err := ns.GetShardIDForPubKey(publicKey)

	require.NotNil(t, ns)
	require.Nil(t, err)
	require.Equal(t, uint32(0), selfId)
}

func TestNodesSetup_ShardPublicKeyGoodMeta(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(1, 5, 1, 1, 6, 3)
	require.NoError(t, err)
	publicKey, _ := hex.DecodeString(pubKeys[2])

	selfId, err := ns.GetShardIDForPubKey(publicKey)

	require.NotNil(t, ns)
	require.Nil(t, err)
	require.Equal(t, uint32(0), selfId)
}

func TestNodesSetup_MetaPublicKeyGoodMeta(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(1, 5, 1, 1, 6, 3)
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

	ns, err := createTestNodesSetup(63, 400, 400, 400, 2169, 3)
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
		[]config.ChainParametersByEpochConfig{},
		mock.NewPubkeyConverterMock(32),
		mock.NewPubkeyConverterMock(96),
		0,
		&epochNotifier.EpochNotifierStub{},
	)

	require.Nil(t, ns)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), ErrInvalidMaximumNumberOfShards.Error())
}

func TestNodesSetup_IfNodesWithinMaxShardLimitEquivalentDistribution(t *testing.T) {
	t.Parallel()

	ns, err := createTestNodesSetup(63, 400, 400, 400, 2169, 3)
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

	ns, err := createTestNodesSetup(63, 400, 400, 400, 3200, 3)
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
