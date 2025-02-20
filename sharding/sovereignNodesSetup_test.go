package sharding

import (
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/sharding/mock"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createSovereignMockNodesSetup() SovereignNodesSetup {
	return SovereignNodesSetup{
		NodesSetup: &NodesSetup{
			ConsensusGroupSize:          2,
			MinNodesPerShard:            2,
			MetaChainMinNodes:           0,
			MetaChainConsensusGroupSize: 0,
			addressPubkeyConverter:      mock.NewPubkeyConverterMock(32),
			validatorPubkeyConverter:    mock.NewPubkeyConverterMock(96),
			InitialNodes: []*InitialNode{
				{
					PubKey:  pubKeys[0],
					Address: address[0],
				},
				{
					PubKey:  pubKeys[1],
					Address: address[1],
				},
			},
		},
	}
}

func createSovereignNodesSetupArgs() *SovereignNodesSetupArgs {
	return &SovereignNodesSetupArgs{
		NodesFilePath:            "mock/testdata/sovereignNodesSetupMock.json",
		AddressPubKeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubKeyConverter: mock.NewPubkeyConverterMock(32),
	}
}

func TestNewSovereignNodesSetupErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("nil address converter", func(t *testing.T) {
		t.Parallel()

		args := createSovereignNodesSetupArgs()
		args.AddressPubKeyConverter = nil

		ns, err := NewSovereignNodesSetup(args)
		require.Nil(t, ns)
		require.ErrorIs(t, err, ErrNilPubkeyConverter)
		require.True(t, strings.Contains(err.Error(), "addressPubKeyConverter"))
	})

	t.Run("nil validator converter", func(t *testing.T) {
		t.Parallel()

		args := createSovereignNodesSetupArgs()
		args.ValidatorPubKeyConverter = nil

		ns, err := NewSovereignNodesSetup(args)
		require.Nil(t, ns)
		require.ErrorIs(t, err, ErrNilPubkeyConverter)
		require.True(t, strings.Contains(err.Error(), "validatorPubKeyConverter"))
	})

	t.Run("invalid nodes file path", func(t *testing.T) {
		t.Parallel()

		args := createSovereignNodesSetupArgs()
		args.NodesFilePath = ""

		ns, err := NewSovereignNodesSetup(args)
		require.Nil(t, ns)
		require.NotNil(t, err)
	})
}

func TestNewSovereignNodesSetupShouldWork(t *testing.T) {
	t.Parallel()

	addrPubKeyConverter := mock.NewPubkeyConverterMock(32)
	validatorPubKeyConverter := mock.NewPubkeyConverterMock(96)
	ns, err := NewSovereignNodesSetup(
		&SovereignNodesSetupArgs{
			NodesFilePath:            "mock/testdata/sovereignNodesSetupMock.json",
			AddressPubKeyConverter:   addrPubKeyConverter,
			ValidatorPubKeyConverter: validatorPubKeyConverter,
		},
	)
	require.Nil(t, err)
	require.NotNil(t, ns)

	encodedPubKeys := []string{
		"cbba7cf4ad9d443b8535b7dd94ee79e0ec7d17a2f8367983e7a1ab8e9ee66736be85e6ae0576194b2dafc2265ce2530d43421239a41fae8d75a1c6689d78089efce6a5f2883d7089da73448f3e9413f4e90023015474d681c15cdfe221733306",
		"c274c4aa1eb936f8e707f69eb865104fb6facb7e320229113fed709c5f6b533f6946aaa3943780b655382838b52fc20181120ef2893ab8aa0183f206a36b3812b92bac8636ed95c2bc40d9978f4b3ba06f77458823ece42e686fc4355e573396",
	}
	encodedAddresses := []string{
		"9e95a4e46da335a96845b4316251fc1bb197e1b8136d96ecc62bf6604eca9e49",
		"7a330039e77ca06bc127319fd707cc4911a80db489a39fcfb746283a05f61836",
	}

	decodedPubKey1, _ := addrPubKeyConverter.Decode(encodedPubKeys[0])
	decodedPubKey2, _ := addrPubKeyConverter.Decode(encodedPubKeys[1])

	decodedAddr1, _ := addrPubKeyConverter.Decode(encodedAddresses[0])
	decodedAddr2, _ := addrPubKeyConverter.Decode(encodedAddresses[1])

	nodesInfo := []nodeInfo{
		{
			assignedShard: core.SovereignChainShardId,
			eligible:      true,
			pubKey:        decodedPubKey1,
			address:       decodedAddr1,
			initialRating: defaultInitialRating,
		},
		{
			assignedShard: core.SovereignChainShardId,
			eligible:      true,
			pubKey:        decodedPubKey2,
			address:       decodedAddr2,
			initialRating: defaultInitialRating,
		},
	}

	expectedInitialNodes := []nodesCoordinator.GenesisNodeInfoHandler{
		&InitialNode{
			PubKey:        encodedPubKeys[0],
			Address:       encodedAddresses[0],
			InitialRating: 0,
			nodeInfo:      nodesInfo[0],
		},
		&InitialNode{
			PubKey:        encodedPubKeys[1],
			Address:       encodedAddresses[1],
			InitialRating: 0,
			nodeInfo:      nodesInfo[1],
		},
	}

	require.Equal(t, expectedInitialNodes, ns.AllInitialNodes())

	eligible, waiting := ns.InitialNodesInfo()
	require.Empty(t, waiting)
	require.Equal(t, []nodesCoordinator.GenesisNodeInfoHandler{&nodesInfo[0], &nodesInfo[1]}, eligible[core.SovereignChainShardId])

	shardID, err := ns.GetShardIDForPubKey(decodedPubKey1)
	require.Nil(t, err)
	require.Equal(t, core.SovereignChainShardId, shardID)

	shardID, err = ns.GetShardIDForPubKey(decodedPubKey2)
	require.Nil(t, err)
	require.Equal(t, core.SovereignChainShardId, shardID)

	require.Equal(t, int64(1689935785), ns.GetStartTime())
	require.Equal(t, uint64(5000), ns.GetRoundDuration())
	require.Equal(t, uint32(1), ns.GetShardConsensusGroupSize())
	require.Equal(t, uint32(0), ns.GetMetaConsensusGroupSize())
	require.Equal(t, uint32(1), ns.NumberOfShards())
	require.Equal(t, uint32(1), ns.MinNumberOfNodes())
	require.Equal(t, uint32(1), ns.MinNumberOfShardNodes())
	require.Equal(t, uint32(0), ns.MinNumberOfMetaNodes())
	require.Equal(t, float32(0), ns.GetHysteresis())
	require.Equal(t, uint32(1), ns.MinNumberOfNodesWithHysteresis())
	require.Equal(t, false, ns.GetAdaptivity())
}

func TestProcessSovereignConfigErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("invalid consensus size", func(t *testing.T) {
		t.Parallel()

		ns := createSovereignMockNodesSetup()
		ns.ConsensusGroupSize = 0

		err := ns.processSovereignConfig()
		assert.Equal(t, ErrNegativeOrZeroConsensusGroupSize, err)
	})

	t.Run("invalid min nodes vs consensus size", func(t *testing.T) {
		t.Parallel()

		ns := createSovereignMockNodesSetup()
		ns.ConsensusGroupSize = 4
		ns.MinNodesPerShard = 3

		err := ns.processSovereignConfig()
		assert.Equal(t, ErrMinNodesPerShardSmallerThanConsensusSize, err)
	})

	t.Run("invalid min nodes vs num of actual nodes", func(t *testing.T) {
		t.Parallel()

		ns := createSovereignMockNodesSetup()
		ns.MinNodesPerShard = 9999

		err := ns.processSovereignConfig()
		assert.Equal(t, ErrNodesSizeSmallerThanMinNoOfNodes, err)
	})

	t.Run("invalid meta chain num nodes", func(t *testing.T) {
		t.Parallel()

		ns := createSovereignMockNodesSetup()

		ns.MetaChainMinNodes = 1
		ns.MetaChainConsensusGroupSize = 0
		err := ns.processSovereignConfig()
		require.ErrorIs(t, err, errSovereignInvalidMetaConsensusSize)

		ns.MetaChainMinNodes = 0
		ns.MetaChainConsensusGroupSize = 1
		err = ns.processSovereignConfig()
		require.ErrorIs(t, err, errSovereignInvalidMetaConsensusSize)

		ns.MetaChainMinNodes = 1
		ns.MetaChainConsensusGroupSize = 1
		err = ns.processSovereignConfig()
		require.ErrorIs(t, err, errSovereignInvalidMetaConsensusSize)
	})

}
