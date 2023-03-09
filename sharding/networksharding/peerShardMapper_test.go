package networksharding_test

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/sharding/networksharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
)

// ------- NewPeerShardMapper

func createMockArgumentForPeerShardMapper() networksharding.ArgPeerShardMapper {
	return networksharding.ArgPeerShardMapper{
		PeerIdPkCache:         testscommon.NewCacherMock(),
		FallbackPkShardCache:  testscommon.NewCacherMock(),
		FallbackPidShardCache: testscommon.NewCacherMock(),
		NodesCoordinator:      &shardingMocks.NodesCoordinatorMock{},
		PreferredPeersHolder:  &p2pmocks.PeersHolderStub{},
	}
}

func createPeerShardMapper() *networksharding.PeerShardMapper {
	psm, _ := networksharding.NewPeerShardMapper(createMockArgumentForPeerShardMapper())
	return psm
}

func TestNewPeerShardMapper_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgumentForPeerShardMapper()
	arg.NodesCoordinator = nil
	psm, err := networksharding.NewPeerShardMapper(arg)

	assert.True(t, check.IfNil(psm))
	assert.Equal(t, nodesCoordinator.ErrNilNodesCoordinator, err)
}

func TestNewPeerShardMapper_NilCacherForPeerIdPkShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgumentForPeerShardMapper()
	arg.PeerIdPkCache = nil
	psm, err := networksharding.NewPeerShardMapper(arg)

	assert.True(t, check.IfNil(psm))
	assert.True(t, errors.Is(err, nodesCoordinator.ErrNilCacher))
}

func TestNewPeerShardMapper_NilCacherForPkShardIdShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgumentForPeerShardMapper()
	arg.FallbackPkShardCache = nil
	psm, err := networksharding.NewPeerShardMapper(arg)

	assert.True(t, check.IfNil(psm))
	assert.True(t, errors.Is(err, nodesCoordinator.ErrNilCacher))
}

func TestNewPeerShardMapper_NilCacherForPeerIdShardIdShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgumentForPeerShardMapper()
	arg.FallbackPidShardCache = nil
	psm, err := networksharding.NewPeerShardMapper(arg)

	assert.True(t, check.IfNil(psm))
	assert.True(t, errors.Is(err, nodesCoordinator.ErrNilCacher))
}

func TestNewPeerShardMapper_NilPreferredShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgumentForPeerShardMapper()
	arg.PreferredPeersHolder = nil
	psm, err := networksharding.NewPeerShardMapper(arg)

	assert.True(t, check.IfNil(psm))
	assert.Equal(t, p2p.ErrNilPreferredPeersHolder, err)
}

func TestNewPeerShardMapper_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgumentForPeerShardMapper()
	psm, err := networksharding.NewPeerShardMapper(arg)

	assert.False(t, check.IfNil(psm))
	assert.Nil(t, err)
}

// ------- UpdatePeerIdPublicKey

func TestPeerShardMapper_UpdatePeerIDInfoShouldWork(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pid := core.PeerID("dummy peer ID")
	pk := []byte("dummy pk")
	shardID := uint32(3737)

	psm.UpdatePeerIDInfo(pid, pk, shardID)

	pkRecovered := psm.GetPkFromPidPk(pid)
	assert.Equal(t, pk, pkRecovered)

	shIDFromPid := psm.GetShardIdFromPidShardId(pid)
	assert.Equal(t, shardID, shIDFromPid)

	shIDFromPk := psm.GetShardIdFromPkShardId(pk)
	assert.Equal(t, shardID, shIDFromPk)

	peerInfo := psm.GetPeerInfo(pid)
	assert.Equal(t,
		core.P2PPeerInfo{
			PeerType: core.ObserverPeer,
			ShardID:  shardID,
			PkBytes:  pk,
		},
		peerInfo)
}

func TestPeerShardMapper_UpdatePeerIDInfoMorePidsThanAllowedShouldTrim(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pk := []byte("dummy pk")
	pids := make([]core.PeerID, networksharding.MaxNumPidsPerPk+1)
	for i := 0; i < networksharding.MaxNumPidsPerPk+1; i++ {
		pids[i] = core.PeerID(fmt.Sprintf("pid %d", i))
		psm.UpdatePeerIDInfo(pids[i], pk, core.AllShardId)
	}

	for i := 0; i < networksharding.MaxNumPidsPerPk+1; i++ {
		shouldExists := i > 0 // the pid is evicted based on the first-in-first-out rule
		pkRecovered := psm.GetPkFromPidPk(pids[i])

		if shouldExists {
			assert.Equal(t, pk, pkRecovered)
		} else {
			assert.Nil(t, pkRecovered)
		}
	}
}

func TestPeerShardMapper_UpdatePeerIDInfoShouldUpdatePkForExistentPid(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pk1 := []byte("dummy pk1")
	pk2 := []byte("dummy pk2")
	pids := make([]core.PeerID, networksharding.MaxNumPidsPerPk+1)
	for i := 0; i < networksharding.MaxNumPidsPerPk; i++ {
		pids[i] = core.PeerID(fmt.Sprintf("pid %d", i))
	}

	newPid := core.PeerID("new pid")
	psm.UpdatePeerIDInfo(pids[0], pk1, 0)
	psm.UpdatePeerIDInfo(newPid, pk1, 0)

	for i := 0; i < networksharding.MaxNumPidsPerPk; i++ {
		psm.UpdatePeerIDInfo(pids[i], pk2, core.AllShardId)
	}

	for i := 0; i < networksharding.MaxNumPidsPerPk; i++ {
		pkRecovered := psm.GetPkFromPidPk(pids[i])

		assert.Equal(t, pk2, pkRecovered)
	}

	assert.Equal(t, []core.PeerID{newPid}, psm.GetFromPkPeerId(pk1))
}

func TestPeerShardMapper_UpdatePeerIDInfoWrongTypePkInPeerIdPkShouldRemove(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pk1 := []byte("dummy pk1")
	pid1 := core.PeerID("pid1")

	wrongTypePk := uint64(7)
	psm.PeerIdPk().Put([]byte(pid1), wrongTypePk, 8)

	psm.UpdatePeerIDInfo(pid1, pk1, core.AllShardId)

	pkRecovered := psm.GetPkFromPidPk(pid1)
	assert.Equal(t, pk1, pkRecovered)
}

func TestPeerShardMapper_UpdatePeerIDInfoShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pk := []byte("dummy pk")
	shardId := uint32(67)

	numUpdates := 100
	wg := &sync.WaitGroup{}
	wg.Add(numUpdates)
	for i := 0; i < numUpdates; i++ {
		go func() {
			psm.UpdatePeerIDInfo("", pk, shardId)
			wg.Done()
		}()
	}
	wg.Wait()

	shardidRecovered := psm.GetShardIdFromPkShardId(pk)
	assert.Equal(t, shardId, shardidRecovered)
}

// ------- UpdatePeerIDPublicKeyPair

func TestPeerShardMapper_UpdatePeerIDPublicKeyPairShouldWork(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pid := core.PeerID("dummy peer ID")
	pk := []byte("dummy pk")

	psm.UpdatePeerIDPublicKeyPair(pid, pk)

	pkRecovered := psm.GetPkFromPidPk(pid)
	assert.Equal(t, pk, pkRecovered)

	pidRecovered := psm.GetFromPkPeerId(pk)
	assert.Equal(t, []core.PeerID{pid}, pidRecovered)
}

// ------- GetPeerInfo

func TestPeerShardMapper_GetPeerInfoPkNotFoundShouldReturnUnknown(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pid := core.PeerID("dummy peer ID")

	peerInfo := psm.GetPeerInfo(pid)
	expectedPeerInfo := core.P2PPeerInfo{
		PeerType:    core.UnknownPeer,
		ShardID:     0,
		PeerSubType: core.RegularPeer,
	}

	assert.Equal(t, expectedPeerInfo, peerInfo)
}

func TestPeerShardMapper_GetPeerInfoNodesCoordinatorHasTheShardId(t *testing.T) {
	t.Parallel()

	shardId := uint32(445)
	pk := []byte("dummy pk")
	arg := createMockArgumentForPeerShardMapper()
	arg.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
		GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, u uint32, e error) {
			if bytes.Equal(publicKey, pk) {
				return nil, shardId, nil
			}

			return nil, 0, nil
		},
	}
	psm, _ := networksharding.NewPeerShardMapper(arg)
	pid := core.PeerID("dummy peer ID")
	psm.UpdatePeerIDInfo(pid, pk, core.AllShardId)

	peerInfo := psm.GetPeerInfo(pid)
	expectedPeerInfo := core.P2PPeerInfo{
		PeerType:    core.ValidatorPeer,
		PeerSubType: core.RegularPeer,
		ShardID:     shardId,
		PkBytes:     pk,
	}

	assert.Equal(t, expectedPeerInfo, peerInfo)
}

func TestPeerShardMapper_GetPeerInfoNodesCoordinatorWrongTypeInCacheShouldReturnUnknown(t *testing.T) {
	t.Parallel()

	wrongTypePk := uint64(6)
	psm, _ := networksharding.NewPeerShardMapper(createMockArgumentForPeerShardMapper())
	pid := core.PeerID("dummy peer ID")
	psm.PeerIdPk().Put([]byte(pid), wrongTypePk, 8)

	peerInfo := psm.GetPeerInfo(pid)
	expectedPeerInfo := core.P2PPeerInfo{
		PeerType:    core.UnknownPeer,
		ShardID:     0,
		PeerSubType: core.RegularPeer,
	}

	assert.Equal(t, expectedPeerInfo, peerInfo)
}

func TestPeerShardMapper_GetPeerInfoNodesCoordinatorDoesntHaveItShouldReturnFromTheFallbackMap(t *testing.T) {
	t.Parallel()

	shardId := uint32(445)
	pk := []byte("dummy pk")
	arg := createMockArgumentForPeerShardMapper()
	arg.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
		GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, u uint32, e error) {
			return nil, 0, errors.New("not found")
		},
	}
	psm, _ := networksharding.NewPeerShardMapper(arg)
	pid := core.PeerID("dummy peer ID")
	psm.UpdatePeerIDInfo(pid, pk, shardId)

	peerInfo := psm.GetPeerInfo(pid)
	expectedPeerInfo := core.P2PPeerInfo{
		PeerType:    core.ObserverPeer,
		PeerSubType: core.RegularPeer,
		ShardID:     shardId,
		PkBytes:     pk,
	}

	assert.Equal(t, expectedPeerInfo, peerInfo)
}

func TestPeerShardMapper_GetPeerInfoNodesCoordinatorDoesntHaveItWrongTypeInCacheShouldReturnUnknown(t *testing.T) {
	t.Parallel()

	pk := []byte("dummy pk")
	arg := createMockArgumentForPeerShardMapper()
	arg.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
		GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, u uint32, e error) {
			return nil, 0, errors.New("not found")
		},
	}
	psm, _ := networksharding.NewPeerShardMapper(arg)
	pid := core.PeerID("dummy peer ID")
	psm.UpdatePeerIDInfo(pid, pk, core.AllShardId)
	wrongTypeShardId := "shard 4"
	psm.FallbackPkShard().Put(pk, wrongTypeShardId, len(wrongTypeShardId))

	peerInfo := psm.GetPeerInfo(pid)
	expectedPeerInfo := core.P2PPeerInfo{
		PeerType:    core.UnknownPeer,
		ShardID:     0,
		PeerSubType: core.RegularPeer,
	}

	assert.Equal(t, expectedPeerInfo, peerInfo)
}

func TestPeerShardMapper_GetPeerInfoNodesCoordinatorDoesntHaveItShouldReturnFromTheSecondFallbackMap(t *testing.T) {
	t.Parallel()

	shardId := uint32(445)
	pk := []byte("dummy pk")
	arg := createMockArgumentForPeerShardMapper()
	arg.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
		GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, u uint32, e error) {
			return nil, 0, errors.New("not found")
		},
	}
	psm, _ := networksharding.NewPeerShardMapper(arg)
	pid := core.PeerID("dummy peer ID")
	psm.UpdatePeerIDInfo(pid, pk, shardId)

	peerInfo := psm.GetPeerInfo(pid)
	expectedPeerInfo := core.P2PPeerInfo{
		PeerType:    core.ObserverPeer,
		ShardID:     shardId,
		PeerSubType: core.RegularPeer,
		PkBytes:     pk,
	}

	assert.Equal(t, expectedPeerInfo, peerInfo)
}

func TestPeerShardMapper_GetPeerInfoShouldRetUnknownShardId(t *testing.T) {
	t.Parallel()

	pk := []byte("dummy pk")
	arg := createMockArgumentForPeerShardMapper()
	arg.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
		GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, u uint32, e error) {
			return nil, 0, errors.New("not found")
		},
	}
	psm, _ := networksharding.NewPeerShardMapper(arg)
	pid := core.PeerID("dummy peer ID")
	psm.UpdatePeerIDInfo(pid, pk, core.AllShardId)

	peerInfo := psm.GetPeerInfo(pid)
	expectedPeerInfo := core.P2PPeerInfo{
		PeerType:    core.UnknownPeer,
		ShardID:     0,
		PeerSubType: core.RegularPeer,
	}

	assert.Equal(t, expectedPeerInfo, peerInfo)
}

func TestPeerShardMapper_GetPeerInfoWithWrongTypeInCacheShouldReturnUnknown(t *testing.T) {
	t.Parallel()

	arg := createMockArgumentForPeerShardMapper()
	arg.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
		GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, u uint32, e error) {
			return nil, 0, errors.New("not found")
		},
	}
	psm, _ := networksharding.NewPeerShardMapper(arg)
	pid := core.PeerID("dummy peer ID")
	wrongTypeShardId := "shard 4"
	psm.FallbackPidShard().Put([]byte(pid), wrongTypeShardId, len(wrongTypeShardId))

	peerInfo := psm.GetPeerInfo(pid)
	expectedPeerInfo := core.P2PPeerInfo{
		PeerType:    core.UnknownPeer,
		ShardID:     0,
		PeerSubType: core.RegularPeer,
	}

	assert.Equal(t, expectedPeerInfo, peerInfo)
}

func TestPeerShardMapper_GetPeerInfoShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	shardId := uint32(445)
	pk := []byte("dummy pk")
	arg := createMockArgumentForPeerShardMapper()
	arg.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
		GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator nodesCoordinator.Validator, u uint32, e error) {
			return nil, 0, errors.New("not found")
		},
	}
	psm, _ := networksharding.NewPeerShardMapper(arg)
	pid := core.PeerID("dummy peer ID")
	psm.UpdatePeerIDInfo(pid, pk, shardId)

	numUpdates := 100
	wg := &sync.WaitGroup{}
	wg.Add(numUpdates)
	for i := 0; i < numUpdates; i++ {
		go func() {
			peerInfo := psm.GetPeerInfo(pid)
			expectedPeerInfo := core.P2PPeerInfo{
				PeerType:    core.ObserverPeer,
				PeerSubType: core.RegularPeer,
				ShardID:     shardId,
				PkBytes:     pk,
			}

			assert.Equal(t, expectedPeerInfo, peerInfo)

			wg.Done()
		}()
	}
	wg.Wait()
}

func TestPeerShardMapper_NotifyOrder(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()

	assert.Equal(t, uint32(common.NetworkShardingOrder), psm.NotifyOrder())
}

func TestPeerShardMapper_UpdatePeerIDPublicKey(t *testing.T) {
	t.Parallel()

	pid1 := core.PeerID("pid1")
	pid2 := core.PeerID("pid2")
	pk1 := []byte("pk1")
	pk2 := []byte("pk2")

	t.Run("peer shard mapper is new, adding a pair should return true", func(t *testing.T) {
		t.Parallel()

		psm := createPeerShardMapper()

		assert.True(t, psm.UpdatePeerIDPublicKey(pid1, pk1))
	})
	t.Run("pair exists, adding the pairs multiple times should return false", func(t *testing.T) {
		t.Parallel()

		psm := createPeerShardMapper()

		_ = psm.UpdatePeerIDPublicKey(pid1, pk1)
		numTries := 1000
		for i := 0; i < numTries; i++ {
			assert.False(t, psm.UpdatePeerIDPublicKey(pid1, pk1))
		}
	})
	t.Run("pair exists but on a different public key returns true", func(t *testing.T) {
		t.Parallel()

		psm := createPeerShardMapper()

		_ = psm.UpdatePeerIDPublicKey(pid1, pk1)
		assert.True(t, psm.UpdatePeerIDPublicKey(pid1, pk2))
		assert.True(t, psm.UpdatePeerIDPublicKey(pid1, pk1))
		assert.True(t, psm.UpdatePeerIDPublicKey(pid1, pk2))
	})
	t.Run("pair exists but on a different pid returns true", func(t *testing.T) {
		t.Parallel()

		psm := createPeerShardMapper()

		_ = psm.UpdatePeerIDPublicKey(pid1, pk1)
		assert.True(t, psm.UpdatePeerIDPublicKey(pid2, pk1))
		assert.False(t, psm.UpdatePeerIDPublicKey(pid1, pk1))
		assert.False(t, psm.UpdatePeerIDPublicKey(pid2, pk1))
	})
}

func TestPeerShardMapper_PutPeerIdShardId(t *testing.T) {
	t.Parallel()

	providedPid := core.PeerID("provided pid")
	providedShardID := uint32(123)
	wasCalled := false
	args := createMockArgumentForPeerShardMapper()
	args.PreferredPeersHolder = &p2pmocks.PeersHolderStub{
		PutShardIDCalled: func(peerID core.PeerID, shardID uint32) {
			wasCalled = true
			assert.Equal(t, providedPid, peerID)
			assert.Equal(t, providedShardID, shardID)
		},
	}
	psm, _ := networksharding.NewPeerShardMapper(args)
	assert.False(t, check.IfNil(psm))

	psm.PutPeerIdShardId(providedPid, providedShardID)
	assert.True(t, wasCalled)
}
