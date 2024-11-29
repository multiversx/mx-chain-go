package metachain

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func createSovereignBaseEconomics() *sovereignBaseEconomics {
	args := createMockEpochEconomicsArguments()
	return &sovereignBaseEconomics{
		&baseEconomics{
			marshalizer:           args.Marshalizer,
			store:                 args.Store,
			shardCoordinator:      args.ShardCoordinator,
			economicsDataNotified: args.EconomicsDataNotified,
			genesisNonce:          1,
			genesisEpoch:          1,
		},
	}
}

func TestSovereignBaseEconomics_startNoncePerShardFromEpochStart(t *testing.T) {
	t.Parallel()

	args := createMockEpochEconomicsArguments()

	hdrPrevEpochStart := &block.SovereignChainHeader{
		EpochStart: block.EpochStartSovereign{
			Economics: block.Economics{
				TotalSupply:                      big.NewInt(2100),
				TotalToDistribute:                big.NewInt(0),
				TotalNewlyMinted:                 big.NewInt(0),
				RewardsPerBlock:                  big.NewInt(0),
				NodePrice:                        big.NewInt(10),
				RewardsForProtocolSustainability: big.NewInt(0),
			},
		},
		Header: &block.Header{
			SoftwareVersion: process.SovereignHeaderVersion,
			Epoch:           1,
			Nonce:           1,
		},
	}
	args.Store = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{GetCalled: func(key []byte) ([]byte, error) {
				hdrBytes, _ := args.Marshalizer.Marshal(hdrPrevEpochStart)
				return hdrBytes, nil
			}}, nil
		},
	}

	sbe := &sovereignBaseEconomics{
		&baseEconomics{
			marshalizer:           args.Marshalizer,
			store:                 args.Store,
			shardCoordinator:      args.ShardCoordinator,
			economicsDataNotified: args.EconomicsDataNotified,
			genesisNonce:          1,
			genesisEpoch:          1,
		},
	}

	noncesPerShardPrevEpoch, prevEpochStartHdr, err := sbe.startNoncePerShardFromEpochStart(2)
	require.Nil(t, err)
	require.Equal(t, map[uint32]uint64{
		core.SovereignChainShardId: 1,
	}, noncesPerShardPrevEpoch)
	require.Equal(t, hdrPrevEpochStart, prevEpochStartHdr)
}

func TestSovereignBaseEconomics_startNoncePerShardFromLastCrossNotarized(t *testing.T) {
	t.Parallel()

	sbe := createSovereignBaseEconomics()

	noncesPerShardCurrEpoch, err := sbe.startNoncePerShardFromLastCrossNotarized(2, nil)
	require.Nil(t, err)
	require.Equal(t, map[uint32]uint64{
		core.SovereignChainShardId: 2,
	}, noncesPerShardCurrEpoch)
}

func TestSovereignBaseEconomics_computeNumOfTotalCreatedBlocks(t *testing.T) {
	t.Parallel()

	sbe := createSovereignBaseEconomics()

	mapStartNonce := map[uint32]uint64{
		core.SovereignChainShardId: 3,
	}
	mapEndNonce := map[uint32]uint64{
		core.SovereignChainShardId: 7,
	}
	totalNumBlocksInEpoch := sbe.computeNumOfTotalCreatedBlocks(mapStartNonce, mapEndNonce)
	require.Equal(t, uint64(4), totalNumBlocksInEpoch)
}

func TestSovereignBaseEconomics_maxPossibleNotarizedBlocks(t *testing.T) {
	t.Parallel()

	sbe := createSovereignBaseEconomics()

	prevHeader := &block.SovereignChainHeader{
		Header: &block.Header{
			Round: 2,
		},
	}
	maxPossibleNotarizedBlocks := sbe.maxPossibleNotarizedBlocks(5, prevHeader)
	require.Equal(t, uint64(3), maxPossibleNotarizedBlocks)
}
