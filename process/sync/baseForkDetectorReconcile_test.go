package sync

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	testscommonDataRetriever "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
)

func TestBaseForkDetector_ReconcileFinalCheckpointBelow(t *testing.T) {
	t.Parallel()

	buildDetector := func() *shardForkDetector {
		sfd, err := NewShardForkDetector(
			&mock.RoundHandlerMock{RoundIndex: 10},
			&testscommon.TimeCacheStub{},
			&mock.BlockTrackerMock{},
			0,
			0,
			&enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
					return flag == common.AndromedaFlag || flag == common.SupernovaFlag
				},
			},
			&testscommon.EnableRoundsHandlerStub{},
			&testscommonDataRetriever.ProofsPoolMock{
				HasProofCalled: func(shardID uint32, headerHash []byte) bool {
					return true
				},
			},
			&chainParameters.ChainParametersHandlerStub{},
			testscommon.GetDefaultProcessConfigsHandler(),
			0,
		)
		require.Nil(t, err)

		prevHash := []byte("hash0")
		for nonce := uint64(1); nonce <= 4; nonce++ {
			hash := []byte{byte(nonce)}
			header := &block.Header{Nonce: nonce, Round: nonce, PrevHash: prevHash, PubKeysBitmap: []byte("X")}
			err = sfd.AddHeader(header, hash, process.BHProcessed, nil, nil)
			require.Nil(t, err)
			prevHash = hash
		}
		require.Equal(t, uint64(4), sfd.finalCheckpoint().nonce)

		// a dead meta-notarized record above the regression target must not survive the purge
		added := sfd.append(&headerInfo{nonce: 4, round: 4, hash: []byte("deadNotarized"), state: process.BHNotarized, hasProof: true})
		require.True(t, added)

		sfd.setSettledCheckpoint(&checkpointInfo{nonce: 2, round: 2, hash: []byte{2}})

		return sfd
	}

	t.Run("nonce zero and the settled floor are refused", func(t *testing.T) {
		t.Parallel()

		sfd := buildDetector()
		require.False(t, sfd.ReconcileFinalCheckpointBelow(0))
		require.False(t, sfd.ReconcileFinalCheckpointBelow(1))
		require.False(t, sfd.ReconcileFinalCheckpointBelow(2))
		require.Equal(t, uint64(4), sfd.finalCheckpoint().nonce)
	})

	t.Run("purges records and checkpoints then lowers the final checkpoint", func(t *testing.T) {
		t.Parallel()

		sfd := buildDetector()
		require.True(t, sfd.ReconcileFinalCheckpointBelow(4))

		require.Equal(t, uint64(3), sfd.finalCheckpoint().nonce)
		require.Equal(t, uint64(3), sfd.ProbableHighestNonce())

		sfd.mutHeaders.RLock()
		_, hasPurgedNonce := sfd.headers[4]
		sfd.mutHeaders.RUnlock()
		require.False(t, hasPurgedNonce)

		require.LessOrEqual(t, sfd.lastCheckpoint().nonce, uint64(3))
	})

	t.Run("still purges when the final checkpoint is already below", func(t *testing.T) {
		t.Parallel()

		sfd := buildDetector()
		require.True(t, sfd.ReconcileFinalCheckpointBelow(4))
		require.Equal(t, uint64(3), sfd.finalCheckpoint().nonce)

		added := sfd.append(&headerInfo{nonce: 5, round: 5, hash: []byte("lateDeadNotarized"), state: process.BHNotarized, hasProof: true})
		require.True(t, added)

		require.True(t, sfd.ReconcileFinalCheckpointBelow(4))
		require.Equal(t, uint64(3), sfd.finalCheckpoint().nonce)

		sfd.mutHeaders.RLock()
		_, hasLateNonce := sfd.headers[5]
		sfd.mutHeaders.RUnlock()
		require.False(t, hasLateNonce)
	})
}
