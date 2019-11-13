package networkSharding

import (
	"bytes"
	"errors"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewPeerShardMapper

func TestNewPeerShardMapper_NilNodesCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	psm, err := NewPeerShardMapper(nil)

	assert.True(t, check.IfNil(psm))
	assert.Equal(t, sharding.ErrNilNodesCoordinator, err)
}

func TestNewPeerShardMapper_ShouldWork(t *testing.T) {
	t.Parallel()

	psm, err := NewPeerShardMapper(&mock.NodesCoordinatorMock{})

	assert.False(t, check.IfNil(psm))
	assert.Nil(t, err)
}

//------- UpdatePeerIdPublicKey

func TestPeerShardMapper_UpdatePeerIdPublicKeyShouldWork(t *testing.T) {
	t.Parallel()

	psm, _ := NewPeerShardMapper(&mock.NodesCoordinatorMock{})
	pid := p2p.PeerID("dummy peer ID")
	pk := []byte("dummy pk")

	psm.UpdatePeerIdPublicKey(pid, pk)

	pkRecovered := psm.GetPkFromMap(pid)
	assert.Equal(t, pk, pkRecovered)
}

func TestPeerShardMapper_UpdatePeerIdPublicKeyShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	psm, _ := NewPeerShardMapper(&mock.NodesCoordinatorMock{})
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

	pkRecovered := psm.GetPkFromMap(pid)
	assert.Equal(t, pk, pkRecovered)
}

//------- UpdatePkShardId

func TestPeerShardMapper_UpdatePublicKeyShardIdShouldWork(t *testing.T) {
	t.Parallel()

	psm, _ := NewPeerShardMapper(&mock.NodesCoordinatorMock{})
	pk := []byte("dummy pk")
	shardId := uint32(67)

	psm.UpdatePublicKeyShardId(pk, shardId)

	shardidRecovered := psm.GetShardIdFromMap(pk)
	assert.Equal(t, shardId, shardidRecovered)
}

func TestPeerShardMapper_UpdatePublicKeyShardIdShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	psm, _ := NewPeerShardMapper(&mock.NodesCoordinatorMock{})
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

	shardidRecovered := psm.GetShardIdFromMap(pk)
	assert.Equal(t, shardId, shardidRecovered)
}

//------- ByID

func TestPeerShardMapper_ByIDPkNotFoundShouldReturnUnknown(t *testing.T) {
	t.Parallel()

	psm, _ := NewPeerShardMapper(&mock.NodesCoordinatorMock{})
	pid := p2p.PeerID("dummy peer ID")

	shardId := psm.ByID(pid)

	assert.Equal(t, sharding.UnknownShardId, shardId)
}

func TestPeerShardMapper_ByIDNodesCoordinatorHasTheShardId(t *testing.T) {
	t.Parallel()

	shardId := uint32(445)
	pk := []byte("dummy pk")
	psm, _ := NewPeerShardMapper(
		&mock.NodesCoordinatorMock{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator sharding.Validator, u uint32, e error) {
				if bytes.Equal(publicKey, pk) {
					return nil, shardId, nil
				}

				return nil, 0, nil
			},
		},
	)
	pid := p2p.PeerID("dummy peer ID")
	psm.UpdatePeerIdPublicKey(pid, pk)

	recoveredShardId := psm.ByID(pid)

	assert.Equal(t, shardId, recoveredShardId)
}

func TestPeerShardMapper_ByIDNodesCoordinatorDoesntHaveItShouldReturnIFromTheFallbackMap(t *testing.T) {
	t.Parallel()

	shardId := uint32(445)
	pk := []byte("dummy pk")
	psm, _ := NewPeerShardMapper(
		&mock.NodesCoordinatorMock{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator sharding.Validator, u uint32, e error) {
				return nil, 0, errors.New("not found")
			},
		},
	)
	pid := p2p.PeerID("dummy peer ID")
	psm.UpdatePeerIdPublicKey(pid, pk)
	psm.UpdatePublicKeyShardId(pk, shardId)

	recoveredShardId := psm.ByID(pid)

	assert.Equal(t, shardId, recoveredShardId)
}

func TestPeerShardMapper_ByIDShouldRetUnknownShardId(t *testing.T) {
	t.Parallel()

	pk := []byte("dummy pk")
	psm, _ := NewPeerShardMapper(
		&mock.NodesCoordinatorMock{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator sharding.Validator, u uint32, e error) {
				return nil, 0, errors.New("not found")
			},
		},
	)
	pid := p2p.PeerID("dummy peer ID")
	psm.UpdatePeerIdPublicKey(pid, pk)

	recoveredShardId := psm.ByID(pid)

	assert.Equal(t, sharding.UnknownShardId, recoveredShardId)
}

func TestPeerShardMapper_ByIDShouldWorkConcurrently(t *testing.T) {
	t.Parallel()

	shardId := uint32(445)
	pk := []byte("dummy pk")
	psm, _ := NewPeerShardMapper(
		&mock.NodesCoordinatorMock{
			GetValidatorWithPublicKeyCalled: func(publicKey []byte) (validator sharding.Validator, u uint32, e error) {
				return nil, 0, errors.New("not found")
			},
		},
	)
	pid := p2p.PeerID("dummy peer ID")
	psm.UpdatePeerIdPublicKey(pid, pk)
	psm.UpdatePublicKeyShardId(pk, shardId)

	numUpdates := 100
	wg := &sync.WaitGroup{}
	wg.Add(numUpdates)
	for i := 0; i < numUpdates; i++ {
		go func() {
			recoveredShardId := psm.ByID(pid)
			assert.Equal(t, shardId, recoveredShardId)

			wg.Done()
		}()
	}
	wg.Wait()
}
