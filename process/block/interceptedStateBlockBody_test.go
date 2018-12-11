package block_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/stretchr/testify/assert"
)

//------- Check

func TestInterceptedStateBlockBodyCheckNilStateBlodyShouldRetTrue(t *testing.T) {
	inBB := block.NewInterceptedStateBlockBody()
	inBB.StateBlockBody = nil
	assert.False(t, inBB.Check())
}

func TestInterceptedStateBlockBodyCheckNilRootHashShouldRetFalse(t *testing.T) {
	inBB := block.NewInterceptedStateBlockBody()
	inBB.RootHash = nil
	assert.False(t, inBB.Check())
}

func TestInterceptedStateBlockBodyCheckEmptyRootHashShouldRetFalse(t *testing.T) {
	inBB := block.NewInterceptedStateBlockBody()
	inBB.RootHash = make([]byte, 0)
	assert.False(t, inBB.Check())
}

func TestInterceptedStateBlockBodyCheckOkValsShouldRetTrue(t *testing.T) {
	inBB := block.NewInterceptedStateBlockBody()

	inBB.RootHash = []byte("aaaa")
	assert.True(t, inBB.Check())
}

//------- Getters and Setters

func TestInterceptedStateBlockBodyAllGettersAndSettersShouldWork(t *testing.T) {
	t.Parallel()

	stateBlockBody := block.NewInterceptedStateBlockBody()

	stateBlockBody.ShardId = 45
	stateBlockBody.RootHash = []byte("aaa")

	assert.Equal(t, uint32(45), stateBlockBody.Shard())

	hash := []byte("aaaa")
	stateBlockBody.SetHash(hash)
	assert.Equal(t, hash, stateBlockBody.Hash())
	assert.Equal(t, string(hash), stateBlockBody.ID())

	newBB := stateBlockBody.New()
	assert.NotNil(t, newBB)
	assert.NotNil(t, newBB.(*block.InterceptedStateBlockBody).StateBlockBody)

	assert.Equal(t, uint32(45), stateBlockBody.Shard())
}
