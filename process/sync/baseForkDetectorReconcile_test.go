package sync

import (
	stdSync "sync"
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

func TestBaseForkDetector_AdvanceFinalAndSettledCheckpointConcurrentWithReconciliation(t *testing.T) {
	t.Parallel()

	for range 20 {
		lowerCheckpoint := &checkpointInfo{nonce: 4, round: 4, hash: []byte("lower")}
		targetCheckpoint := &checkpointInfo{nonce: 5, round: 5, hash: []byte("target")}
		bfd := &baseForkDetector{}
		bfd.fork.finalCheckpoint = targetCheckpoint
		bfd.fork.settledCheckpoint = lowerCheckpoint
		bfd.fork.checkpoint = []*checkpointInfo{lowerCheckpoint, targetCheckpoint}

		var waitGroup stdSync.WaitGroup
		waitGroup.Add(2)
		go func() {
			defer waitGroup.Done()
			bfd.advanceFinalAndSettledCheckpoint(targetCheckpoint)
		}()
		go func() {
			defer waitGroup.Done()
			bfd.ReconcileFinalCheckpoint(targetCheckpoint.nonce)
		}()
		waitGroup.Wait()

		require.Equal(t, targetCheckpoint, bfd.finalCheckpoint())
		require.Equal(t, targetCheckpoint, bfd.settledCheckpoint())
	}
}

func TestBaseForkDetector_ReconcileFinalCheckpointUsesRetainedProcessedParent(t *testing.T) {
	t.Parallel()

	settled := &checkpointInfo{nonce: 2, round: 2, hash: []byte("settled")}
	bfd := &baseForkDetector{
		headers: map[uint64][]*headerInfo{
			3: {{nonce: 3, round: 3, hash: []byte("parent"), state: process.BHProcessed}},
			4: {{nonce: 4, round: 4, hash: []byte("tip"), state: process.BHProcessed}},
		},
		fork: forkInfo{
			checkpoint:        []*checkpointInfo{{nonce: 4, round: 4, hash: []byte("tip")}},
			finalCheckpoint:   &checkpointInfo{nonce: 4, round: 4, hash: []byte("tip")},
			settledCheckpoint: settled,
		},
	}

	require.True(t, bfd.ReconcileFinalCheckpoint(4))
	require.Equal(t, &checkpointInfo{nonce: 3, round: 3, hash: []byte("parent")}, bfd.finalCheckpoint())
	require.Equal(t, settled, bfd.settledCheckpoint())
}

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
			testscommon.NewEnableRoundsHandlerStub(common.SupernovaRoundFlag),
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
			header := &block.HeaderV3{Epoch: 1, Nonce: nonce, Round: nonce, PrevHash: prevHash}
			err = sfd.AddHeader(header, hash, process.BHProcessed, nil, nil)
			require.Nil(t, err)
			prevHash = hash
		}
		require.Equal(t, uint64(4), sfd.finalCheckpoint().nonce)

		// a dead meta-notarized record above the regression target must not survive the purge
		added := sfd.append(&headerInfo{nonce: 4, round: 4, hash: []byte("deadNotarized"), state: process.BHNotarized, hasProof: true})
		require.True(t, added)

		sfd.mutFork.Lock()
		sfd.fork.settledCheckpoint = &checkpointInfo{nonce: 2, round: 2, hash: []byte{2}}
		sfd.mutFork.Unlock()

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

	t.Run("missing selected authority leaves the detector unchanged", func(t *testing.T) {
		t.Parallel()

		sfd := buildDetector()
		added := sfd.append(&headerInfo{
			nonce: 5, round: 5, hash: []byte("laterAuthority"), state: process.BHNotarized, hasProof: true,
		})
		require.True(t, added)
		finalBefore := sfd.finalCheckpoint()
		lastBefore := sfd.lastCheckpoint()
		probableBefore := sfd.ProbableHighestNonce()

		require.False(t, sfd.ReconcileFinalCheckpointFromAuthority(4, []byte("missingAuthority")))

		require.Equal(t, finalBefore, sfd.finalCheckpoint())
		require.Equal(t, lastBefore, sfd.lastCheckpoint())
		require.Equal(t, probableBefore, sfd.ProbableHighestNonce())
		sfd.mutHeaders.RLock()
		require.NotEmpty(t, sfd.headers[4])
		require.NotEmpty(t, sfd.headers[5])
		sfd.mutHeaders.RUnlock()
	})
}
