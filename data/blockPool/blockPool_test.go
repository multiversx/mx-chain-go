package blockPool_test

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockPool"
	"github.com/stretchr/testify/assert"
)

func TestNewBlockPool(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	assert.NotNil(t, bp)
}

func TestBlockPoolBodyStore(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bs := bp.BodyStore()
	assert.NotNil(t, bs)
}

func TestBlockPoolHeaderStore(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	hs := bp.HeaderStore()
	assert.NotNil(t, hs)
}

func TestBlockPoolAddHeader(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bp.RegisterHeaderHandler(receivedHeader)

	hdr := block.Header{Nonce: 0}

	key := make([]byte, 8)
	binary.PutUvarint(key, 0)

	bp.AddHeader(0, &hdr)
	assert.Equal(t, 1, bp.HeaderStore().Len())

	val, ok := bp.HeaderStore().Get(key)
	assert.Equal(t, true, ok)
	assert.Equal(t, &hdr, val.(*block.Header))

	hdr2 := block.Header{Nonce: 1}

	bp.AddHeader(0, &hdr2)
	assert.Equal(t, 1, bp.HeaderStore().Len())

	val, ok = bp.HeaderStore().Get(key)
	assert.Equal(t, true, ok)
	assert.Equal(t, &hdr, val.(*block.Header))
	assert.NotEqual(t, &hdr2, val.(*block.Header))

	bp.AddHeader(1, &hdr2)
	assert.Equal(t, 2, bp.HeaderStore().Len())

	key = make([]byte, 8)
	binary.PutUvarint(key, 1)

	val, ok = bp.HeaderStore().Get(key)
	assert.Equal(t, true, ok)
	assert.Equal(t, &hdr2, val.(*block.Header))

	key = make([]byte, 8)
	binary.PutUvarint(key, 2)

	val, ok = bp.HeaderStore().Get(key)
	assert.Equal(t, false, ok)
	assert.Nil(t, val)
}

func TestBlockPoolAddBody(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bp.RegisterBodyHandler(receivedBody)

	mblks := make([]block.MiniBlock, 0)
	mblk := block.MiniBlock{DestShardID: 0}
	mblks = append(mblks, mblk)

	blk := block.Block{MiniBlocks: mblks}

	key := make([]byte, 8)
	binary.PutUvarint(key, 0)

	bp.AddBody(0, &blk)
	assert.Equal(t, 1, bp.BodyStore().Len())

	val, ok := bp.BodyStore().Get(key)
	assert.Equal(t, true, ok)
	assert.Equal(t, &blk, val.(*block.Block))

	mblks2 := make([]block.MiniBlock, 0)
	mblk2 := block.MiniBlock{DestShardID: 1}
	mblks2 = append(mblks2, mblk2)

	blk2 := block.Block{MiniBlocks: mblks2}

	bp.AddBody(0, &blk2)
	assert.Equal(t, 1, bp.BodyStore().Len())

	val, ok = bp.BodyStore().Get(key)
	assert.Equal(t, true, ok)
	assert.Equal(t, &blk, val.(*block.Block))
	assert.NotEqual(t, &blk2, val.(*block.Block))

	bp.AddBody(1, &blk2)
	assert.Equal(t, 2, bp.BodyStore().Len())

	key = make([]byte, 8)
	binary.PutUvarint(key, 1)

	val, ok = bp.BodyStore().Get(key)
	assert.Equal(t, true, ok)
	assert.Equal(t, &blk2, val.(*block.Block))

	key = make([]byte, 8)
	binary.PutUvarint(key, 2)

	val, ok = bp.BodyStore().Get(key)
	assert.Equal(t, false, ok)
	assert.Nil(t, val)
}

func TestBlockPoolHeader(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bp.RegisterHeaderHandler(receivedHeader)

	hdr := block.Header{Nonce: 0}

	bp.AddHeader(0, &hdr)
	assert.Equal(t, 1, bp.HeaderStore().Len())

	val, ok := bp.Header(0)
	assert.Equal(t, true, ok)
	assert.Equal(t, &hdr, val.(*block.Header))

	val, ok = bp.Header(1)
	assert.Equal(t, false, ok)
	assert.Nil(t, val)
}

func TestBlockPoolBody(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bp.RegisterBodyHandler(receivedBody)

	blk := block.Block{}

	bp.AddBody(0, &blk)
	assert.Equal(t, 1, bp.BodyStore().Len())

	val, ok := bp.Body(0)
	assert.Equal(t, true, ok)
	assert.Equal(t, &blk, val.(*block.Block))

	val, ok = bp.Body(1)
	assert.Equal(t, false, ok)
	assert.Nil(t, val)
}

func TestBlockPoolRemoveHeader(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bp.RegisterHeaderHandler(receivedHeader)

	hdr := block.Header{Nonce: 0}

	bp.AddHeader(0, &hdr)
	assert.Equal(t, 1, bp.HeaderStore().Len())

	hdr2 := block.Header{Nonce: 1}

	bp.AddHeader(1, &hdr2)
	assert.Equal(t, 2, bp.HeaderStore().Len())

	bp.RemoveHeader(2)
	assert.Equal(t, 2, bp.HeaderStore().Len())

	bp.RemoveHeader(0)
	assert.Equal(t, 1, bp.HeaderStore().Len())

	key := make([]byte, 8)
	binary.PutUvarint(key, 1)

	val, ok := bp.HeaderStore().Get(key)
	assert.Equal(t, true, ok)
	assert.Equal(t, &hdr2, val.(*block.Header))
	assert.NotEqual(t, &hdr, val.(*block.Header))
}

func TestBlockPoolRemoveBlock(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bp.RegisterBodyHandler(receivedBody)

	mblks := make([]block.MiniBlock, 0)
	mblk := block.MiniBlock{DestShardID: 0}
	mblks = append(mblks, mblk)

	blk := block.Block{MiniBlocks: mblks}

	bp.AddBody(0, &blk)
	assert.Equal(t, 1, bp.BodyStore().Len())

	mblks2 := make([]block.MiniBlock, 0)
	mblk2 := block.MiniBlock{DestShardID: 1}
	mblks2 = append(mblks2, mblk2)

	blk2 := block.Block{MiniBlocks: mblks2}

	bp.AddBody(1, &blk2)
	assert.Equal(t, 2, bp.BodyStore().Len())

	bp.RemoveBody(2)
	assert.Equal(t, 2, bp.BodyStore().Len())

	bp.RemoveBody(0)
	assert.Equal(t, 1, bp.BodyStore().Len())

	key := make([]byte, 8)
	binary.PutUvarint(key, 1)

	val, ok := bp.BodyStore().Get(key)
	assert.Equal(t, true, ok)
	assert.Equal(t, &blk2, val.(*block.Block))
	assert.NotEqual(t, &blk, val.(*block.Block))
}

func TestBlockPoolClearHeaderStore(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bp.RegisterHeaderHandler(receivedHeader)

	hdr := block.Header{Nonce: 0}

	bp.AddHeader(0, &hdr)
	bp.AddHeader(1, &hdr)
	bp.AddHeader(2, &hdr)
	assert.Equal(t, 3, bp.HeaderStore().Len())

	bp.ClearHeaderStore()
	assert.Equal(t, 0, bp.HeaderStore().Len())
}

func TestBlockPoolClearBodyStore(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bp.RegisterBodyHandler(receivedBody)

	blk := block.Block{}

	bp.AddBody(0, &blk)
	bp.AddBody(1, &blk)
	bp.AddBody(2, &blk)
	assert.Equal(t, 3, bp.BodyStore().Len())

	bp.ClearBodyStore()
	assert.Equal(t, 0, bp.BodyStore().Len())
}

func receivedHeader(nonce uint64) {
	fmt.Println(nonce)
}

func receivedBody(nonce uint64) {
	fmt.Println(nonce)
}
