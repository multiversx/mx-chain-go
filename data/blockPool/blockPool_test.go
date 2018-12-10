package blockPool_test

import (
	"encoding/binary"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockPool"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewBlockPool(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	assert.NotNil(t, bp)
}

func TestBlockPool_MiniPoolBodyStore(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bs := bp.MiniPoolBodyStore()
	assert.NotNil(t, bs)
}

func TestBlockPool_MiniPoolHeaderStore(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	hs := bp.MiniPoolHeaderStore()
	assert.NotNil(t, hs)
}

func TestBlockPool_AddHeader(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bp.RegisterHeaderHandler(receivedHeader)

	hdr := block.Header{Nonce: 0}

	key := make([]byte, 8)
	binary.PutUvarint(key, 0)

	bp.AddHeader(0, &hdr)
	assert.Equal(t, 1, bp.MiniPoolHeaderStore().Len())

	val, ok := bp.MiniPoolHeaderStore().Get(key)
	assert.Equal(t, true, ok)
	assert.Equal(t, &hdr, val.(*block.Header))

	hdr2 := block.Header{Nonce: 1}

	bp.AddHeader(0, &hdr2)
	assert.Equal(t, 1, bp.MiniPoolHeaderStore().Len())

	val, ok = bp.MiniPoolHeaderStore().Get(key)
	assert.Equal(t, true, ok)
	assert.Equal(t, &hdr, val.(*block.Header))
	assert.NotEqual(t, &hdr2, val.(*block.Header))

	bp.AddHeader(1, &hdr2)
	assert.Equal(t, 2, bp.MiniPoolHeaderStore().Len())

	key = make([]byte, 8)
	binary.PutUvarint(key, 1)

	val, ok = bp.MiniPoolHeaderStore().Get(key)
	assert.Equal(t, true, ok)
	assert.Equal(t, &hdr2, val.(*block.Header))

	key = make([]byte, 8)
	binary.PutUvarint(key, 2)

	val, ok = bp.MiniPoolHeaderStore().Get(key)
	assert.Equal(t, false, ok)
	assert.Nil(t, val)
}

func TestBlockPool_AddBody(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bp.RegisterBodyHandler(receivedBody)

	mblks := make([]block.MiniBlock, 0)
	mblk := block.MiniBlock{DestShardID: 0}
	mblks = append(mblks, mblk)

	blk := block.Block{MiniBlocks: mblks}

	key := make([]byte, 8)
	binary.PutUvarint(key, 0)

	bp.AddBody(0, &blk)
	assert.Equal(t, 1, bp.MiniPoolBodyStore().Len())

	val, ok := bp.MiniPoolBodyStore().Get(key)
	assert.Equal(t, true, ok)
	assert.Equal(t, &blk, val.(*block.Block))

	mblks2 := make([]block.MiniBlock, 0)
	mblk2 := block.MiniBlock{DestShardID: 1}
	mblks2 = append(mblks2, mblk2)

	blk2 := block.Block{MiniBlocks: mblks2}

	bp.AddBody(0, &blk2)
	assert.Equal(t, 1, bp.MiniPoolBodyStore().Len())

	val, ok = bp.MiniPoolBodyStore().Get(key)
	assert.Equal(t, true, ok)
	assert.Equal(t, &blk, val.(*block.Block))
	assert.NotEqual(t, &blk2, val.(*block.Block))

	bp.AddBody(1, &blk2)
	assert.Equal(t, 2, bp.MiniPoolBodyStore().Len())

	key = make([]byte, 8)
	binary.PutUvarint(key, 1)

	val, ok = bp.MiniPoolBodyStore().Get(key)
	assert.Equal(t, true, ok)
	assert.Equal(t, &blk2, val.(*block.Block))

	key = make([]byte, 8)
	binary.PutUvarint(key, 2)

	val, ok = bp.MiniPoolBodyStore().Get(key)
	assert.Equal(t, false, ok)
	assert.Nil(t, val)
}

func TestBlockPool_Header(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bp.RegisterHeaderHandler(receivedHeader)

	hdr := block.Header{Nonce: 0}

	bp.AddHeader(0, &hdr)
	assert.Equal(t, 1, bp.MiniPoolHeaderStore().Len())

	val, ok := bp.Header(0)
	assert.Equal(t, true, ok)
	assert.Equal(t, &hdr, val.(*block.Header))

	val, ok = bp.Header(1)
	assert.Equal(t, false, ok)
	assert.Nil(t, val)
}

func TestBlockPool_Body(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bp.RegisterBodyHandler(receivedBody)

	blk := block.Block{}

	bp.AddBody(0, &blk)
	assert.Equal(t, 1, bp.MiniPoolBodyStore().Len())

	val, ok := bp.Body(0)
	assert.Equal(t, true, ok)
	assert.Equal(t, &blk, val.(*block.Block))

	val, ok = bp.Body(1)
	assert.Equal(t, false, ok)
	assert.Nil(t, val)
}

func TestBlockPool_RemoveHeader(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bp.RegisterHeaderHandler(receivedHeader)

	hdr := block.Header{Nonce: 0}

	bp.AddHeader(0, &hdr)
	assert.Equal(t, 1, bp.MiniPoolHeaderStore().Len())

	hdr2 := block.Header{Nonce: 1}

	bp.AddHeader(1, &hdr2)
	assert.Equal(t, 2, bp.MiniPoolHeaderStore().Len())

	bp.RemoveHeader(2)
	assert.Equal(t, 2, bp.MiniPoolHeaderStore().Len())

	bp.RemoveHeader(0)
	assert.Equal(t, 1, bp.MiniPoolHeaderStore().Len())

	key := make([]byte, 8)
	binary.PutUvarint(key, 1)

	val, ok := bp.MiniPoolHeaderStore().Get(key)
	assert.Equal(t, true, ok)
	assert.Equal(t, &hdr2, val.(*block.Header))
	assert.NotEqual(t, &hdr, val.(*block.Header))
}

func TestBlockPool_RemoveBlock(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bp.RegisterBodyHandler(receivedBody)

	mblks := make([]block.MiniBlock, 0)
	mblk := block.MiniBlock{DestShardID: 0}
	mblks = append(mblks, mblk)

	blk := block.Block{MiniBlocks: mblks}

	bp.AddBody(0, &blk)
	assert.Equal(t, 1, bp.MiniPoolBodyStore().Len())

	mblks2 := make([]block.MiniBlock, 0)
	mblk2 := block.MiniBlock{DestShardID: 1}
	mblks2 = append(mblks2, mblk2)

	blk2 := block.Block{MiniBlocks: mblks2}

	bp.AddBody(1, &blk2)
	assert.Equal(t, 2, bp.MiniPoolBodyStore().Len())

	bp.RemoveBody(2)
	assert.Equal(t, 2, bp.MiniPoolBodyStore().Len())

	bp.RemoveBody(0)
	assert.Equal(t, 1, bp.MiniPoolBodyStore().Len())

	key := make([]byte, 8)
	binary.PutUvarint(key, 1)

	val, ok := bp.MiniPoolBodyStore().Get(key)
	assert.Equal(t, true, ok)
	assert.Equal(t, &blk2, val.(*block.Block))
	assert.NotEqual(t, &blk, val.(*block.Block))
}

func TestBlockPool_ClearHeaderMiniPool(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bp.RegisterHeaderHandler(receivedHeader)

	hdr := block.Header{Nonce: 0}

	bp.AddHeader(0, &hdr)
	bp.AddHeader(1, &hdr)
	bp.AddHeader(2, &hdr)
	assert.Equal(t, 3, bp.MiniPoolHeaderStore().Len())

	bp.ClearHeaderMiniPool()
	assert.Equal(t, 0, bp.MiniPoolHeaderStore().Len())
}

func TestBlockPool_ClearBodyMiniPool(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bp.RegisterBodyHandler(receivedBody)

	blk := block.Block{}

	bp.AddBody(0, &blk)
	bp.AddBody(1, &blk)
	bp.AddBody(2, &blk)
	assert.Equal(t, 3, bp.MiniPoolBodyStore().Len())

	bp.ClearBodyMiniPool()
	assert.Equal(t, 0, bp.MiniPoolBodyStore().Len())
}

func receivedHeader(nounce uint64) {
	fmt.Println(nounce)
}

func receivedBody(nounce uint64) {
	fmt.Println(nounce)
}
