package networksharding_test

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/ElrondNetwork/elrond-go/sharding/networksharding"
	"github.com/stretchr/testify/assert"
)

//------- NewPeerShardMapper

func createPeerShardMapper() *networksharding.PeerShardMapper {
	psm, _ := networksharding.NewPeerShardMapper(
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		&nodesCoordinatorStub{},
		&mock.EpochHandlerMock{},
	)
	return psm
}

func TestNewPeerShardMapper_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	psm, err := networksharding.NewPeerShardMapper(
		&mock.CacherMock{},
		&mock.CacherMock{},
		&mock.CacherMock{},
		nil,
		&mock.EpochHandlerMock{},
	)

	assert.True(t, check.IfNil(psm))
	assert.Equal(t, sharding.ErrNilNodesCoordinator, err)
}

func TestNewPeerShardMapper_NilCacherForPeerIdPkShouldErr(t *testing.T) {
	t.Parallel()

	psm, err := networksharding.NewPeerShardMapper(
		nil,
		&mock.CacherMock{},
		&mock.CacherMock{},
		&nodesCoordinatorStub{},
		&mock.EpochHandlerMock{},
	)

	assert.True(t, check.IfNil(psm))
	assert.Equal(t, sharding.ErrNilCacher, err)
}

func TestNewPeerShardMapper_NilCacherForPkShardIdShouldErr(t *testing.T) {
	t.Parallel()

	psm, err := networksharding.NewPeerShardMapper(
		&mock.CacherMock{},
		nil,
		&mock.CacherMock{},
		&nodesCoordinatorStub{},
		&mock.EpochHandlerMock{},
	)

	assert.True(t, check.IfNil(psm))
	assert.Equal(t, sharding.ErrNilCacher, err)
}

func TestNewPeerShardMapper_NilCacherForPeerIdShardIdShouldErr(t *testing.T) {
	t.Parallel()

	psm, err := networksharding.NewPeerShardMapper(
		&mock.CacherMock{},
		&mock.CacherMock{},
		nil,
		&nodesCoordinatorStub{},
		&mock.EpochHandlerMock{},
	)

	assert.True(t, check.IfNil(psm))
	assert.Equal(t, sharding.ErrNilCacher, err)
}

func TestNewPeerShardMapper_NilEpochHandlerShouldErr(t *testing.T) {
	t.Parallel()

	psm, err := networksharding.NewPeerShardMapper(
		&mock.CacherMock{},
		&mock.CacherMock{},
		&mock.CacherMock{},
		&nodesCoordinatorStub{},
		nil,
	)

	assert.True(t, check.IfNil(psm))
	assert.Equal(t, sharding.ErrNilEpochHandler, err)
}

func TestNewPeerShardMapper_ShouldWork(t *testing.T) {
	t.Parallel()

	psm, err := networksharding.NewPeerShardMapper(
		&mock.CacherMock{},
		&mock.CacherMock{},
		&mock.CacherMock{},
		&nodesCoordinatorStub{},
		&mock.EpochHandlerMock{},
	)

	assert.False(t, check.IfNil(psm))
	assert.Nil(t, err)
}

//------- UpdatePeerIdPublicKey

func TestPeerShardMapper_UpdatePeerIdPublicKeyShouldWork(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pid := p2p.PeerID("dummy peer ID")
	pk := []byte("dummy pk")

	psm.UpdatePeerIdPublicKey(pid, pk)

	pkRecovered := psm.GetPkFromPidPk(pid)
	assert.Equal(t, pk, pkRecovered)
}

func TestPeerShardMapper_UpdatePeerIdPublicKeyMorePidsThanAllowedShouldTrim(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pk := []byte("dummy pk")
	pids := make([]p2p.PeerID, networksharding.MaxNumPidsPerPk+1)
	for i := 0; i < networksharding.MaxNumPidsPerPk+1; i++ {
		pids[i] = p2p.PeerID(fmt.Sprintf("pid %d", i))
		psm.UpdatePeerIdPublicKey(pids[i], pk)
	}

	for i := 0; i < networksharding.MaxNumPidsPerPk+1; i++ {
		shouldExists := i > 0 //the pid is evicted based on the first-in-first-out rule
		pkRecovered := psm.GetPkFromPidPk(pids[i])

		if shouldExists {
			assert.Equal(t, pk, pkRecovered)
		} else {
			assert.Nil(t, pkRecovered)
		}
	}
}

func TestPeerShardMapper_UpdatePeerIdPublicKeyShouldUpdatePkForExistentPid(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pk1 := []byte("dummy pk1")
	pk2 := []byte("dummy pk2")
	pids := make([]p2p.PeerID, networksharding.MaxNumPidsPerPk+1)
	for i := 0; i < networksharding.MaxNumPidsPerPk; i++ {
		pids[i] = p2p.PeerID(fmt.Sprintf("pid %d", i))
	}

	newPid := p2p.PeerID("new pid")
	psm.UpdatePeerIdPublicKey(pids[0], pk1)
	psm.UpdatePeerIdPublicKey(newPid, pk1)

	for i := 0; i < networksharding.MaxNumPidsPerPk; i++ {
		psm.UpdatePeerIdPublicKey(pids[i], pk2)
	}

	for i := 0; i < networksharding.MaxNumPidsPerPk; i++ {
		pkRecovered := psm.GetPkFromPidPk(pids[i])

		assert.Equal(t, pk2, pkRecovered)
	}

	assert.Equal(t, []p2p.PeerID{newPid}, psm.GetFromPkPeerId(pk1))
}

func TestPeerShardMapper_UpdatePeerIdPublicKeyWrongTypePkInPeerIdPkShouldRemove(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pk1 := []byte("dummy pk1")
	pid1 := p2p.PeerID("pid1")

	wrongTypePk := uint64(7)
	psm.PeerIdPk().Put([]byte(pid1), wrongTypePk)

	psm.UpdatePeerIdPublicKey(pid1, pk1)

	pkRecovered := psm.GetPkFromPidPk(pid1)
	assert.Equal(t, pk1, pkRecovered)
}

func TestPeerShardMapper_UpdatePeerIdPublicKeyShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pid := p2p.PeerID("dummy peer ID")
	pk := []byte("dummy pk")

	numUpdates := 100
	wg := &sync.WaitGroup{}
	wg.Add(numUpdates)
	for i := 0; i < numUpdates; i++ {
		go func() {
			psm.UpdatePeerIdPublicKey(pid, pk)
			wg.Done()
		}()
	}
	wg.Wait()

	pkRecovered := psm.GetPkFromPidPk(pid)
	assert.Equal(t, pk, pkRecovered)
}

//------- UpdatePublicKeyShardId

func TestPeerShardMapper_UpdatePublicKeyShardIdShouldWork(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pk := []byte("dummy pk")
	shardId := uint32(67)

	psm.UpdatePublicKeyShardId(pk, shardId)

	shardidRecovered := psm.GetShardIdFromPkShardId(pk)
	assert.Equal(t, shardId, shardidRecovered)
}

func TestPeerShardMapper_UpdatePublicKeyShardIdShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pk := []byte("dummy pk")
	shardId := uint32(67)

	numUpdates := 100
	wg := &sync.WaitGroup{}
	wg.Add(numUpdates)
	for i := 0; i < numUpdates; i++ {
		go func() {
			psm.UpdatePublicKeyShardId(pk, shardId)
			wg.Done()
		}()
	}
	wg.Wait()

	shardidRecovered := psm.GetShardIdFromPkShardId(pk)
	assert.Equal(t, shardId, shardidRecovered)
}

//------- UpdatePeerIdShardId

func TestPeerShardMapper_UpdatePeerIdShardIdShouldWork(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pid := p2p.PeerID("dummy peer ID")
	shardId := uint32(67)

	psm.UpdatePeerIdShardId(pid, shardId)

	shardidRecovered := psm.GetShardIdFromPidShardId(pid)
	assert.Equal(t, shardId, shardidRecovered)
}

func TestPeerShardMapper_UpdatePeerIdShardIdShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pid := p2p.PeerID("dummy peer ID")
	shardId := uint32(67)

	numUpdates := 100
	wg := &sync.WaitGroup{}
	wg.Add(numUpdates)
	for i := 0; i < numUpdates; i++ {
		go func() {
			psm.UpdatePeerIdShardId(pid, shardId)
			wg.Done()
		}()
	}
	wg.Wait()

	shardidRecovered := psm.GetShardIdFromPidShardId(pid)
	assert.Equal(t, shardId, shardidRecovered)
}

//------- GetShardID

func TestPeerShardMapper_GetShardIDPkNotFoundShouldReturnUnknown(t *testing.T) {
	t.Parallel()

	psm := createPeerShardMapper()
	pid := p2p.PeerID("dummy peer ID")

	shardId := psm.GetShardID(pid)

	assert.Equal(t, core.UnknownShardId, shardId)
}

func TestPeerShardMapper_GetShardIDNodesCoordinatorHasTheShardId(t *testing.T) {
	t.Parallel()

	shardId := uint32(445)
	pk := []byte("dummy pk")
	psm, _ := networksharding.NewPeerShardMapper(
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		&nodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte, epoch uint32) (validator sharding.Validator, u uint32, e error) {
				if bytes.Equal(publicKey, pk) {
					return nil, shardId, nil
				}

				return nil, 0, nil
			},
		},
		&mock.EpochHandlerMock{},
	)
	pid := p2p.PeerID("dummy peer ID")
	psm.UpdatePeerIdPublicKey(pid, pk)

	recoveredShardId := psm.GetShardID(pid)

	assert.Equal(t, shardId, recoveredShardId)
}

func TestPeerShardMapper_GetShardIDNodesCoordinatorWrongTypeInCacheShouldReturnUnknown(t *testing.T) {
	t.Parallel()

	wrongTypePk := uint64(6)
	psm, _ := networksharding.NewPeerShardMapper(
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		&nodesCoordinatorStub{},
		&mock.EpochHandlerMock{},
	)
	pid := p2p.PeerID("dummy peer ID")
	psm.PeerIdPk().Put([]byte(pid), wrongTypePk)

	recoveredShardId := psm.GetShardID(pid)

	assert.Equal(t, core.UnknownShardId, recoveredShardId)
}

func TestPeerShardMapper_GetShardIDNodesCoordinatorDoesntHaveItShouldReturnFromTheFallbackMap(t *testing.T) {
	t.Parallel()

	shardId := uint32(445)
	pk := []byte("dummy pk")
	psm, _ := networksharding.NewPeerShardMapper(
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		&nodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte, epoch uint32) (validator sharding.Validator, u uint32, e error) {
				return nil, 0, errors.New("not found")
			},
		},
		&mock.EpochHandlerMock{},
	)
	pid := p2p.PeerID("dummy peer ID")
	psm.UpdatePeerIdPublicKey(pid, pk)
	psm.UpdatePublicKeyShardId(pk, shardId)

	recoveredShardId := psm.GetShardID(pid)

	assert.Equal(t, shardId, recoveredShardId)
}

func TestPeerShardMapper_GetShardIDNodesCoordinatorDoesntHaveItWrongTypeInCacheShouldReturnUnknown(t *testing.T) {
	t.Parallel()

	pk := []byte("dummy pk")
	psm, _ := networksharding.NewPeerShardMapper(
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		&nodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte, epoch uint32) (validator sharding.Validator, u uint32, e error) {
				return nil, 0, errors.New("not found")
			},
		},
		&mock.EpochHandlerMock{},
	)
	pid := p2p.PeerID("dummy peer ID")
	psm.UpdatePeerIdPublicKey(pid, pk)
	wrongTypeShardId := "shard 4"
	psm.FallbackPkShard().Put(pk, wrongTypeShardId)

	recoveredShardId := psm.GetShardID(pid)

	assert.Equal(t, core.UnknownShardId, recoveredShardId)
}

func TestPeerShardMapper_GetShardIDNodesCoordinatorDoesntHaveItShouldReturnFromTheSecondFallbackMap(t *testing.T) {
	t.Parallel()

	shardId := uint32(445)
	pk := []byte("dummy pk")
	psm, _ := networksharding.NewPeerShardMapper(
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		&nodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte, epoch uint32) (validator sharding.Validator, u uint32, e error) {
				return nil, 0, errors.New("not found")
			},
		},
		&mock.EpochHandlerMock{},
	)
	pid := p2p.PeerID("dummy peer ID")
	psm.UpdatePeerIdPublicKey(pid, pk)
	psm.UpdatePeerIdShardId(pid, shardId)

	recoveredShardId := psm.GetShardID(pid)

	assert.Equal(t, shardId, recoveredShardId)
}

func TestPeerShardMapper_GetShardIDWrongDataInPeerIdMapShouldErr(t *testing.T) {
	t.Parallel()

	shardId := uint32(445)
	pk := []byte("dummy pk")

	psm, _ := networksharding.NewPeerShardMapper(
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		&nodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte, epoch uint32) (validator sharding.Validator, u uint32, e error) {
				return nil, 0, errors.New("not found")
			},
		},
		&mock.EpochHandlerMock{},
	)
	pid := p2p.PeerID("dummy peer ID")
	psm.UpdatePeerIdPublicKey(pid, pk)
	psm.UpdatePublicKeyShardId(pk, shardId)

	recoveredShardId := psm.GetShardID(pid)

	assert.Equal(t, shardId, recoveredShardId)
}

func TestPeerShardMapper_GetShardIDShouldRetUnknownShardId(t *testing.T) {
	t.Parallel()

	pk := []byte("dummy pk")
	psm, _ := networksharding.NewPeerShardMapper(
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		&nodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte, epoch uint32) (validator sharding.Validator, u uint32, e error) {
				return nil, 0, errors.New("not found")
			},
		},
		&mock.EpochHandlerMock{},
	)
	pid := p2p.PeerID("dummy peer ID")
	psm.UpdatePeerIdPublicKey(pid, pk)

	recoveredShardId := psm.GetShardID(pid)

	assert.Equal(t, core.UnknownShardId, recoveredShardId)
}

func TestPeerShardMapper_GetShardIDShouldWrongTypeInCacheShouldReturnUnknown(t *testing.T) {
	t.Parallel()

	psm, _ := networksharding.NewPeerShardMapper(
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		&nodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte, epoch uint32) (validator sharding.Validator, u uint32, e error) {
				return nil, 0, errors.New("not found")
			},
		},
		&mock.EpochHandlerMock{},
	)
	pid := p2p.PeerID("dummy peer ID")
	wrongTypeShardId := "shard 4"
	psm.FallbackPidShard().Put([]byte(pid), wrongTypeShardId)

	recoveredShardId := psm.GetShardID(pid)

	assert.Equal(t, core.UnknownShardId, recoveredShardId)
}

func TestPeerShardMapper_GetShardIDShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	shardId := uint32(445)
	pk := []byte("dummy pk")
	psm, _ := networksharding.NewPeerShardMapper(
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		mock.NewCacherMock(),
		&nodesCoordinatorStub{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte, epoch uint32) (validator sharding.Validator, u uint32, e error) {
				return nil, 0, errors.New("not found")
			},
		},
		&mock.EpochHandlerMock{},
	)
	pid := p2p.PeerID("dummy peer ID")
	psm.UpdatePeerIdPublicKey(pid, pk)
	psm.UpdatePublicKeyShardId(pk, shardId)

	numUpdates := 100
	wg := &sync.WaitGroup{}
	wg.Add(numUpdates)
	for i := 0; i < numUpdates; i++ {
		go func() {
			recoveredShardId := psm.GetShardID(pid)
			assert.Equal(t, shardId, recoveredShardId)

			wg.Done()
		}()
	}
	wg.Wait()
}
