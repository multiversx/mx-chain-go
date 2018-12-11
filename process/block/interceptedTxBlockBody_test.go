package block_test

import (
	block2 "github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/stretchr/testify/assert"
	"testing"
)

//------- Check

func TestInterceptedTxBlockBodyCheckNilTxBlodyShouldRetTrue(t *testing.T) {
	inBB := block.NewInterceptedTxBlockBody()
	inBB.TxBlockBody = nil
	assert.False(t, inBB.Check())
}

func TestInterceptedTxBlockBodyCheckNilMiniBlocksShouldRetFalse(t *testing.T) {
	inBB := block.NewInterceptedTxBlockBody()
	inBB.MiniBlocks = nil
	assert.False(t, inBB.Check())
}

func TestInterceptedTxBlockBodyCheckEmptyMiniBlocksShouldRetFalse(t *testing.T) {
	inBB := block.NewInterceptedTxBlockBody()
	inBB.MiniBlocks = make([]block2.MiniBlock, 0)
	assert.False(t, inBB.Check())
}

func TestInterceptedTxBlockBodyCheckEmptyMiniBlockShouldRetFalse(t *testing.T) {
	inBB := block.NewInterceptedTxBlockBody()

	miniBlock := block2.MiniBlock{}

	inBB.MiniBlocks = make([]block2.MiniBlock, 0)
	inBB.MiniBlocks = append(inBB.MiniBlocks, miniBlock)
	assert.False(t, inBB.Check())
}

func TestInterceptedTxBlockBodyCheckOkValsShouldRetTrue(t *testing.T) {
	inBB := block.NewInterceptedTxBlockBody()

	miniBlock := block2.MiniBlock{}
	miniBlock.TxHashes = append(miniBlock.TxHashes, []byte{65})

	inBB.MiniBlocks = make([]block2.MiniBlock, 0)
	inBB.MiniBlocks = append(inBB.MiniBlocks, miniBlock)
	assert.True(t, inBB.Check())
}

//------- Getters and Setters

func TestInterceptedTxBlockBodyAllGettersAndSettersShouldWork(t *testing.T) {
	t.Parallel()

	txBlockBody := block.NewInterceptedTxBlockBody()

	txBlockBody.TxBlockBody.ShardId = 45
	txBlockBody.TxBlockBody.MiniBlocks = make([]block2.MiniBlock, 0)

	assert.Equal(t, uint32(45), txBlockBody.Shard())
	assert.Equal(t, 0, len(txBlockBody.MiniBlocks))

	hash := []byte("aaaa")
	txBlockBody.SetHash(hash)
	assert.Equal(t, hash, txBlockBody.Hash())
	assert.Equal(t, string(hash), txBlockBody.ID())

	newTxBB := txBlockBody.New()
	assert.NotNil(t, newTxBB)
	assert.NotNil(t, newTxBB.(*block.InterceptedTxBlockBody).TxBlockBody)

	assert.Equal(t, uint32(45), txBlockBody.Shard())
}
