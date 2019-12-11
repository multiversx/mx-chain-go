package headersCash_test

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool/headersCash"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewHeadersCacher_AddHeadersInCache(t *testing.T) {
	t.Parallel()

	hdrsCacher := headersCash.NewHeadersCacher(1000, 100)

	nonce := uint64(1)
	shardId := uint32(0)

	hdrHash1 := []byte("hash1")
	hdrHash2 := []byte("hash2")
	testHdr1 := &block.Header{Nonce: nonce, ShardId: shardId}
	testHdr2 := &block.Header{Nonce: nonce, ShardId: shardId, Round: 100}

	hdrsCacher.Add(hdrHash1, testHdr1)
	hdrsCacher.Add(hdrHash2, testHdr2)

	hdr, err := hdrsCacher.GetHeaderByHash(hdrHash1)
	assert.Nil(t, err)
	assert.Equal(t, testHdr1, hdr)

	hdr, err = hdrsCacher.GetHeaderByHash(hdrHash2)
	assert.Nil(t, err)
	assert.Equal(t, testHdr2, hdr)

	expectedHeaders := []data.HeaderHandler{testHdr1, testHdr2}
	hdrs, err := hdrsCacher.GetHeaderByNonceAndShardId(nonce, shardId)
	assert.Nil(t, err)
	assert.Equal(t, expectedHeaders, hdrs)
}

func TestHeadersCacher_AddHeadersInCacheAndRemoveByHash(t *testing.T) {
	t.Parallel()

	hdrsCacher := headersCash.NewHeadersCacher(1000, 100)

	nonce := uint64(1)
	shardId := uint32(0)

	hdrHash1 := []byte("hash1")
	hdrHash2 := []byte("hash2")
	testHdr1 := &block.Header{Nonce: nonce, ShardId: shardId}
	testHdr2 := &block.Header{Nonce: nonce, ShardId: shardId, Round: 100}

	hdrsCacher.Add(hdrHash1, testHdr1)
	hdrsCacher.Add(hdrHash2, testHdr2)

	hdrsCacher.RemoveHeaderByHash(hdrHash1)
	hdr, err := hdrsCacher.GetHeaderByHash(hdrHash1)
	assert.Nil(t, hdr)
	assert.Equal(t, headersCash.ErrHeaderNotFound, err)

	hdrsCacher.RemoveHeaderByHash(hdrHash2)
	hdr, err = hdrsCacher.GetHeaderByHash(hdrHash2)
	assert.Nil(t, hdr)
	assert.Equal(t, headersCash.ErrHeaderNotFound, err)
}

func TestHeadersCacher_AddHeadersInCacheAndRemoveByNonceAndShadId(t *testing.T) {
	t.Parallel()

	hdrsCacher := headersCash.NewHeadersCacher(1000, 100)

	nonce := uint64(1)
	shardId := uint32(0)

	hdrHash1 := []byte("hash1")
	hdrHash2 := []byte("hash2")
	testHdr1 := &block.Header{Nonce: nonce, ShardId: shardId}
	testHdr2 := &block.Header{Nonce: nonce, ShardId: shardId, Round: 100}

	hdrsCacher.Add(hdrHash1, testHdr1)
	hdrsCacher.Add(hdrHash2, testHdr2)

	hdrsCacher.RemoveHeaderByNonceAndShardId(nonce, shardId)
	hdr, err := hdrsCacher.GetHeaderByHash(hdrHash1)
	assert.Nil(t, hdr)
	assert.Equal(t, headersCash.ErrHeaderNotFound, err)

	hdr, err = hdrsCacher.GetHeaderByHash(hdrHash2)
	assert.Nil(t, hdr)
	assert.Equal(t, headersCash.ErrHeaderNotFound, err)
}

func TestHeadersCacher_EvictionShouldWork(t *testing.T) {
	t.Parallel()

	hdrs, hdrsHashes := createASliceOfHeaders(1000, 0)
	hdrsCacher := headersCash.NewHeadersCacher(900, 100)

	for i := 0; i < 1000; i++ {
		hdrsCacher.Add(hdrsHashes[i], &hdrs[i])
	}

	// Cache will do eviction 2 times, in headers cache will be 800 headers
	for i := 200; i < 1000; i++ {
		hdr, err := hdrsCacher.GetHeaderByHash(hdrsHashes[i])
		assert.Nil(t, err)
		assert.Equal(t, &hdrs[i], hdr)
	}
}

func createASliceOfHeaders(numHeaders int, shardId uint32) ([]block.Header, [][]byte) {
	headers := make([]block.Header, 0)
	headersHashes := make([][]byte, 0)
	for i := 0; i < numHeaders; i++ {
		headers = append(headers, block.Header{Nonce: uint64(i), ShardId: shardId})
		headersHashes = append(headersHashes, []byte(fmt.Sprintf("%d", i)))
	}

	return headers, headersHashes
}
