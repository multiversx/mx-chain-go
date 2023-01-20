package blacklist_test

import (
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/consensus/blacklist"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/stretchr/testify/require"
)

func createMockPeerBlacklistArgs() blacklist.PeerBlackListArgs {
	return blacklist.PeerBlackListArgs{
		PeerCacher: &mock.PeerBlackListCacherStub{},
	}
}

func TestNewPeerBlacklist(t *testing.T) {
	t.Parallel()

	t.Run("nil peer cacher, should fail", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerBlacklistArgs()
		args.PeerCacher = nil

		pb, err := blacklist.NewPeerBlacklist(args)
		require.Nil(t, pb)
		require.Equal(t, spos.ErrNilPeerBlacklistCacher, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockPeerBlacklistArgs()

		wg := &sync.WaitGroup{}
		wg.Add(1)

		sweepWasCalled := false
		args.PeerCacher = &mock.PeerBlackListCacherStub{
			SweepCalled: func() {
				sweepWasCalled = true
				wg.Done()
			},
		}

		pb, err := blacklist.NewPeerBlacklist(args)
		require.Nil(t, err)
		require.False(t, pb.IsInterfaceNil())
		defer pb.Close()

		wg.Wait()

		require.True(t, sweepWasCalled)
	})
}

func TestBlacklistPeer(t *testing.T) {
	t.Parallel()

	args := createMockPeerBlacklistArgs()

	expPeer, _ := core.NewPeerID("peerID")

	hasWasCalled := false
	upsertWasCalled := false
	args.PeerCacher = &mock.PeerBlackListCacherStub{
		HasCalled: func(pid core.PeerID) bool {
			require.Equal(t, expPeer, pid)
			hasWasCalled = true
			return true
		},
		UpsertCalled: func(pid core.PeerID, span time.Duration) error {
			require.Equal(t, expPeer, pid)
			upsertWasCalled = true
			return nil
		},
	}

	pb, err := blacklist.NewPeerBlacklist(args)
	require.Nil(t, err)

	duration := 1 * time.Second

	pb.BlacklistPeer(expPeer, duration)
	require.True(t, hasWasCalled)
	require.True(t, upsertWasCalled)
}

func TestIsPeerBlacklisted(t *testing.T) {
	t.Parallel()

	args := createMockPeerBlacklistArgs()

	expPeer, _ := core.NewPeerID("peerID")

	hasWasCalled := false
	args.PeerCacher = &mock.PeerBlackListCacherStub{
		HasCalled: func(pid core.PeerID) bool {
			require.Equal(t, expPeer, pid)
			hasWasCalled = true
			return true
		},
	}

	pb, err := blacklist.NewPeerBlacklist(args)
	require.Nil(t, err)

	isBlacklisted := pb.IsPeerBlacklisted(expPeer)
	require.True(t, isBlacklisted)
	require.True(t, hasWasCalled)
}
