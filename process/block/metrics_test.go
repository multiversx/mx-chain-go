package block

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	outportStub "github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

func TestMetrics_CalculateRoundDuration(t *testing.T) {
	t.Parallel()

	lastBlockTimestamp := uint64(80)
	currentBlockTimestamp := uint64(100)
	lastBlockRound := uint64(5)
	currentBlockRound := uint64(10)
	expectedRoundDuration := uint64(4)

	roundDuration := calculateRoundDuration(lastBlockTimestamp, currentBlockTimestamp, lastBlockRound, currentBlockRound)
	assert.Equal(t, expectedRoundDuration, roundDuration)
}

func TestMetrics_IncrementMetricCountConsensusAcceptedBlocks(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("nodes coordinator returns error should early exit", func(t *testing.T) {
		t.Parallel()

		nodesCoord := &shardingMocks.NodesCoordinatorMock{
			GetValidatorsPublicKeysCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) (string, []string, error) {
				return "", nil, expectedErr
			},
		}
		statusHandler := &statusHandlerMock.AppStatusHandlerStub{
			IncrementHandler: func(_ string) {
				assert.Fail(t, "should have not been called")
			},
		}

		incrementMetricCountConsensusAcceptedBlocks(&block.HeaderV2{}, 0, nodesCoord, statusHandler, &testscommon.ManagedPeersHolderStub{})
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cntIncrement := 0
		mainKey := "main key"
		managedKeyInConsensus := "managed key in consensus"
		nodesCoord := &shardingMocks.NodesCoordinatorStub{
			GetOwnPublicKeyCalled: func() []byte {
				return []byte(mainKey)
			},
			GetValidatorsPublicKeysCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) (string, []string, error) {
				leader := "some leader"
				return leader, []string{
					leader,
					mainKey,
					managedKeyInConsensus,
					"some other key",
				}, nil
			},
		}
		statusHandler := &statusHandlerMock.AppStatusHandlerStub{
			IncrementHandler: func(_ string) {
				cntIncrement++
			},
		}
		managedPeersHolder := &testscommon.ManagedPeersHolderStub{
			IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
				return string(pkBytes) == managedKeyInConsensus
			},
		}

		incrementMetricCountConsensusAcceptedBlocks(&block.MetaBlock{}, 0, nodesCoord, statusHandler, managedPeersHolder)
		assert.Equal(t, 2, cntIncrement) // main key + managed key
	})
}

func TestMetrics_IndexRoundInfoShouldKeepSyntheticRoundTimestampsSplitByUnitsAfterSupernova(t *testing.T) {
	t.Parallel()

	var savedRoundsInfo *outportcore.RoundsInfo
	outportHandler := &outportStub.OutportStub{
		SaveRoundsInfoCalled: func(roundsInfo *outportcore.RoundsInfo) {
			savedRoundsInfo = roundsInfo
		},
	}
	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.SupernovaFlag && epoch == 7
		},
	}
	header := &testscommon.HeaderHandlerStub{
		EpochField:        7,
		RoundField:        10,
		TimestampField:    1774864086000,
		GetRandSeedCalled: func() []byte { return []byte("rand-seed") },
	}
	lastHeader := &testscommon.HeaderHandlerStub{
		EpochField:        7,
		RoundField:        8,
		TimestampField:    1774864080000,
		GetRandSeedCalled: func() []byte { return []byte("rand-seed") },
	}
	nodesCoordinator := &shardingMocks.NodesCoordinatorStub{
		GetValidatorsPublicKeysCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) (string, []string, error) {
			return "leader", []string{"pk1"}, nil
		},
		GetValidatorsIndexesCalled: func(_ []string, _ uint32) ([]uint64, error) {
			return []uint64{11}, nil
		},
	}
	roundHandler := &testscommon.RoundHandlerMock{
		GetTimeStampForRoundCalled: func(round uint64) uint64 {
			require.Equal(t, uint64(9), round)
			return 1774864082400
		},
	}

	indexRoundInfo(outportHandler, nodesCoordinator, 1, header, lastHeader, []uint64{22}, enableEpochsHandler, roundHandler)

	require.NotNil(t, savedRoundsInfo)
	require.Len(t, savedRoundsInfo.RoundsInfo, 2)

	currentRoundInfo := savedRoundsInfo.RoundsInfo[0]
	assert.Equal(t, uint64(1774864086), currentRoundInfo.Timestamp)
	assert.Equal(t, uint64(1774864086000), currentRoundInfo.TimestampMs)

	syntheticRoundInfo := savedRoundsInfo.RoundsInfo[1]
	assert.Equal(t, uint64(9), syntheticRoundInfo.Round)
	assert.Equal(t, uint64(1774864082), syntheticRoundInfo.Timestamp)
	assert.Equal(t, uint64(1774864082400), syntheticRoundInfo.TimestampMs)
	assert.Equal(t, []uint64{11}, syntheticRoundInfo.SignersIndexes)
}

func TestMetrics_IndexRoundInfoSplitsLargeGapIntoMultipleBatches(t *testing.T) {
	t.Parallel()

	lastRound := uint64(0)
	currentRound := uint64(2*maxRoundsInfoPerBatch + 505)
	totalEntries := int(currentRound - lastRound) // 1 current + (current-last-1) missed
	require.Equal(t, 2*maxRoundsInfoPerBatch+505, totalEntries)

	var savedBatches []*outportcore.RoundsInfo
	outportHandler := &outportStub.OutportStub{
		SaveRoundsInfoCalled: func(roundsInfo *outportcore.RoundsInfo) {
			savedBatches = append(savedBatches, roundsInfo)
		},
	}
	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.SupernovaFlag && epoch == 7
		},
	}
	header := &testscommon.HeaderHandlerStub{
		EpochField:        7,
		RoundField:        currentRound,
		TimestampField:    1774864086000,
		GetRandSeedCalled: func() []byte { return []byte("rand-seed") },
	}
	lastHeader := &testscommon.HeaderHandlerStub{
		EpochField:        7,
		RoundField:        lastRound,
		TimestampField:    1774864080000,
		GetRandSeedCalled: func() []byte { return []byte("rand-seed") },
	}
	nodesCoordinator := &shardingMocks.NodesCoordinatorStub{
		GetValidatorsPublicKeysCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) (string, []string, error) {
			return "leader", []string{"pk1"}, nil
		},
		GetValidatorsIndexesCalled: func(_ []string, _ uint32) ([]uint64, error) {
			return []uint64{11}, nil
		},
	}
	roundHandler := &testscommon.RoundHandlerMock{
		GetTimeStampForRoundCalled: func(round uint64) uint64 {
			return 1774864080000 + round*600
		},
	}

	indexRoundInfo(outportHandler, nodesCoordinator, 1, header, lastHeader, []uint64{22}, enableEpochsHandler, roundHandler)

	// 2505 entries with batch size 1000 must result in 3 SaveRoundsInfo calls: 1000, 1000, 505
	require.Len(t, savedBatches, 3)
	require.Len(t, savedBatches[0].RoundsInfo, maxRoundsInfoPerBatch)
	require.Len(t, savedBatches[1].RoundsInfo, maxRoundsInfoPerBatch)
	require.Len(t, savedBatches[2].RoundsInfo, 505)

	totalSaved := 0
	for _, batch := range savedBatches {
		require.LessOrEqual(t, len(batch.RoundsInfo), maxRoundsInfoPerBatch)
		assert.Equal(t, uint32(1), batch.ShardID)
		totalSaved += len(batch.RoundsInfo)
	}
	assert.Equal(t, totalEntries, totalSaved)

	// first entry is the current block, the rest are missed rounds in order
	currentRoundInfo := savedBatches[0].RoundsInfo[0]
	assert.Equal(t, currentRound, currentRoundInfo.Round)
	assert.True(t, currentRoundInfo.BlockWasProposed)

	flattened := make([]*outportcore.RoundInfo, 0, totalSaved)
	for _, batch := range savedBatches {
		flattened = append(flattened, batch.RoundsInfo...)
	}
	for idx := uint64(1); idx < currentRound; idx++ {
		missed := flattened[idx]
		assert.Equal(t, idx, missed.Round)
		assert.False(t, missed.BlockWasProposed)
		assert.Equal(t, 1774864080000+idx*600, missed.TimestampMs)
		assert.Equal(t, (1774864080000+idx*600)/1000, missed.Timestamp)
	}
}
