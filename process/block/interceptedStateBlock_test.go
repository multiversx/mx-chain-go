package block_test

import (
	"testing"

	block2 "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestInterceptedStateBlockBody_NewShouldNotCreateNilBlock(t *testing.T) {
	t.Parallel()

	stateBlockBody := block.NewInterceptedStateBlockBody()

	assert.NotNil(t, stateBlockBody.StateBlockBody)
}

func TestInterceptedStateBlockBody_GetUnderlingObjectShouldReturnBlock(t *testing.T) {
	t.Parallel()

	stateBlockBody := block.NewInterceptedStateBlockBody()

	assert.True(t, stateBlockBody.GetUnderlyingObject() == stateBlockBody.StateBlockBody)
}

func TestInterceptedStateBlockBody_GetterSetterHashID(t *testing.T) {
	t.Parallel()

	hash := []byte("hash")

	stateBlockBody := block.NewInterceptedStateBlockBody()
	stateBlockBody.SetHash(hash)

	assert.Equal(t, hash, stateBlockBody.Hash())
	assert.Equal(t, string(hash), stateBlockBody.ID())
}

func TestInterceptedStateBlockBody_ShardShouldWork(t *testing.T) {
	t.Parallel()

	shard := uint32(78)

	stateBlockBody := block.NewInterceptedStateBlockBody()
	stateBlockBody.ShardID = shard

	assert.Equal(t, shard, stateBlockBody.Shard())
}

func TestInterceptedStateBlockBody_CreateShouldNotProduceNils(t *testing.T) {
	t.Parallel()

	stateBlockBody := block.NewInterceptedStateBlockBody()
	stateBlockCreated := stateBlockBody.Create()

	assert.NotNil(t, stateBlockCreated)
	assert.NotNil(t, stateBlockCreated.(*block.InterceptedStateBlockBody).StateBlockBody)
}

func TestInterceptedStateBlockBody_CreateShouldNotProduceSameObject(t *testing.T) {
	t.Parallel()

	stateBlockBody := block.NewInterceptedStateBlockBody()
	stateBlockCreated := stateBlockBody.Create()

	assert.False(t, stateBlockBody == stateBlockCreated)
	assert.False(t, stateBlockCreated.(*block.InterceptedStateBlockBody).StateBlockBody ==
		stateBlockBody.StateBlockBody)
}

func TestInterceptedStateBlockBody_IntegrityNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	stateBlk := &block.InterceptedStateBlockBody{StateBlockBody: &block2.StateBlockBody{}}

	stateBlk.RootHash = make([]byte, 0)
	stateBlk.ShardID = 0

	assert.Equal(t, process.ErrNilShardCoordinator, stateBlk.Integrity(nil))
}

func TestInterceptedStateBlockBody_IntegrityInvalidShardShouldErr(t *testing.T) {
	t.Parallel()

	stateBlk := &block.InterceptedStateBlockBody{StateBlockBody: &block2.StateBlockBody{}}

	stateBlk.RootHash = make([]byte, 0)
	stateBlk.ShardID = 6

	assert.Equal(t, process.ErrInvalidShardId, stateBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedStateBlockBody_IntegrityNilRootHashShouldErr(t *testing.T) {
	t.Parallel()

	stateBlk := &block.InterceptedStateBlockBody{StateBlockBody: &block2.StateBlockBody{}}

	stateBlk.RootHash = nil
	stateBlk.ShardID = 0

	assert.Equal(t, process.ErrNilRootHash, stateBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedStateBlockBody_IntegrityNilStateBlockBodyShouldErr(t *testing.T) {
	t.Parallel()

	stateBlk := &block.InterceptedStateBlockBody{StateBlockBody: &block2.StateBlockBody{}}
	stateBlk.StateBlockBody = nil

	assert.Equal(t, process.ErrNilStateBlockBody, stateBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedStateBlockBody_IntegrityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	stateBlk := &block.InterceptedStateBlockBody{StateBlockBody: &block2.StateBlockBody{}}

	stateBlk.RootHash = make([]byte, 0)
	stateBlk.ShardID = 0

	assert.Nil(t, stateBlk.Integrity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedStateBlockBody_IntegrityAndValidityIntegrityDoesNotPassShouldErr(t *testing.T) {
	t.Parallel()

	stateBlk := &block.InterceptedStateBlockBody{StateBlockBody: &block2.StateBlockBody{}}

	stateBlk.RootHash = make([]byte, 0)
	stateBlk.ShardID = 6

	assert.Equal(t, process.ErrInvalidShardId, stateBlk.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}

func TestInterceptedStateBlockBody_IntegrityAndValidityOkValsShouldWork(t *testing.T) {
	t.Parallel()

	stateBlk := &block.InterceptedStateBlockBody{StateBlockBody: &block2.StateBlockBody{}}

	stateBlk.RootHash = make([]byte, 0)
	stateBlk.ShardID = 0

	assert.Nil(t, stateBlk.IntegrityAndValidity(mock.NewOneShardCoordinatorMock()))
}
