package headersCache_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool/headersCache"
	"github.com/stretchr/testify/assert"
)

func createASliceOfHeaders(numHeaders int, shardId uint32) ([]block.Header, [][]byte) {
	headers := make([]block.Header, 0)
	headersHashes := make([][]byte, 0)
	for i := 0; i < numHeaders; i++ {
		headers = append(headers, block.Header{Nonce: uint64(i), ShardId: shardId})
		headersHashes = append(headersHashes, []byte(fmt.Sprintf("%d_%d", shardId, i)))
	}

	return headers, headersHashes
}

func createASliceOfHeadersNonce0(numHeaders int, shardId uint32) ([]block.Header, [][]byte) {
	headers := make([]block.Header, 0)
	headersHashes := make([][]byte, 0)
	for i := 0; i < numHeaders; i++ {
		headers = append(headers, block.Header{Nonce: 0, ShardId: shardId})
		headersHashes = append(headersHashes, []byte(fmt.Sprintf("%d_%d", shardId, i)))
	}

	return headers, headersHashes
}

func TestNewHeadersCacher_AddHeadersInCache(t *testing.T) {
	t.Parallel()

	hdrsCacher, _ := headersCache.NewHeadersCacher(1000, 100)

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
	hdrs, _, err := hdrsCacher.GetHeaderByNonceAndShardId(nonce, shardId)
	assert.Nil(t, err)
	assert.Equal(t, expectedHeaders, hdrs)
}

func TestHeadersCacher_AddHeadersInCacheAndRemoveByHash(t *testing.T) {
	t.Parallel()

	hdrsCacher, _ := headersCache.NewHeadersCacher(1000, 100)

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
	assert.Equal(t, headersCache.ErrHeaderNotFound, err)

	hdrsCacher.RemoveHeaderByHash(hdrHash2)
	hdr, err = hdrsCacher.GetHeaderByHash(hdrHash2)
	assert.Nil(t, hdr)
	assert.Equal(t, headersCache.ErrHeaderNotFound, err)
}

func TestHeadersCacher_AddHeadersInCacheAndRemoveByNonceAndShadId(t *testing.T) {
	t.Parallel()

	hdrsCacher, _ := headersCache.NewHeadersCacher(1000, 100)

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
	assert.Equal(t, headersCache.ErrHeaderNotFound, err)

	hdr, err = hdrsCacher.GetHeaderByHash(hdrHash2)
	assert.Nil(t, hdr)
	assert.Equal(t, headersCache.ErrHeaderNotFound, err)
}

func TestHeadersCacher_EvictionShouldWork(t *testing.T) {
	t.Parallel()

	hdrs, hdrsHashes := createASliceOfHeaders(1000, 0)
	hdrsCacher, _ := headersCache.NewHeadersCacher(900, 100)

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

func TestHeadersCacher_ConcurrentRequestsShouldWorkNoEviction(t *testing.T) {
	t.Parallel()

	numHeadersToGenerate := 500

	hdrs, hdrsHashes := createASliceOfHeaders(numHeadersToGenerate, 0)
	hdrsCacher, _ := headersCache.NewHeadersCacher(numHeadersToGenerate+1, 10)

	for i := 0; i < numHeadersToGenerate; i++ {
		go func(index int) {
			hdrsCacher.Add(hdrsHashes[index], &hdrs[index])
			hdr, err := hdrsCacher.GetHeaderByHash(hdrsHashes[index])

			assert.Nil(t, err)
			assert.Equal(t, &hdrs[index], hdr)
		}(i)
	}
}

func TestHeadersCacher_ConcurrentRequestsShouldWorkWithEviction(t *testing.T) {
	shardId := uint32(0)
	cacheSize := 2
	numHeadersToGenerate := 500

	hdrs, hdrsHashes := createASliceOfHeaders(numHeadersToGenerate, shardId)
	hdrsCacher, _ := headersCache.NewHeadersCacher(cacheSize, 1)

	var waitgroup sync.WaitGroup
	for i := 0; i < numHeadersToGenerate; i++ {
		waitgroup.Add(1)
		go func(index int) {
			hdrsCacher.Add(hdrsHashes[index], &hdrs[index])
			waitgroup.Done()
		}(i)
	}
	waitgroup.Wait()
	// cache size after all eviction is finish should be 2
	assert.Equal(t, 2, hdrsCacher.GetNumHeadersFromCacheShard(shardId))

	numHeadersToGenerate = 3
	hdrs, hdrsHashes = createASliceOfHeaders(3, shardId)
	for i := 0; i < numHeadersToGenerate; i++ {
		hdrsCacher.Add(hdrsHashes[i], &hdrs[i])
	}

	assert.Equal(t, 2, hdrsCacher.GetNumHeadersFromCacheShard(shardId))
	hdr1, err := hdrsCacher.GetHeaderByHash(hdrsHashes[1])
	assert.Nil(t, err)
	assert.Equal(t, &hdrs[1], hdr1)

	hdr2, err := hdrsCacher.GetHeaderByHash(hdrsHashes[2])
	assert.Nil(t, err)
	assert.Equal(t, &hdrs[2], hdr2)
}

func TestHeadersCacher_AddHeadersWithSameNonceShouldBeRemovedAtEviction(t *testing.T) {
	t.Parallel()

	shardId := uint32(0)
	cacheSize := 2

	hash1, hash2, hash3 := []byte("hash1"), []byte("hash2"), []byte("hash3")
	hdr1, hdr2, hdr3 := &block.Header{Nonce: 0}, &block.Header{Nonce: 0}, &block.Header{Nonce: 1}

	hdrsCacher, _ := headersCache.NewHeadersCacher(cacheSize, 1)
	hdrsCacher.Add(hash1, hdr1)
	hdrsCacher.Add(hash2, hdr2)
	hdrsCacher.Add(hash3, hdr3)

	assert.Equal(t, 1, hdrsCacher.GetNumHeadersFromCacheShard(shardId))

	hdr, err := hdrsCacher.GetHeaderByHash(hash3)
	assert.Nil(t, err)
	assert.Equal(t, hdr3, hdr)
}

func TestHeadersCacher_AddALotOfHeadersAndCheckEviction(t *testing.T) {
	t.Parallel()

	cacheSize := 100
	numHeaders := 500
	shardId := uint32(0)
	hdrs, hdrsHash := createASliceOfHeaders(numHeaders, shardId)
	hdrsCacher, _ := headersCache.NewHeadersCacher(cacheSize, 50)

	var waitgroup sync.WaitGroup
	for i := 0; i < numHeaders; i++ {
		waitgroup.Add(1)
		go func(index int) {
			hdrsCacher.Add(hdrsHash[index], &hdrs[index])
			waitgroup.Done()
		}(i)
	}

	waitgroup.Wait()
	assert.Equal(t, 100, hdrsCacher.GetNumHeadersFromCacheShard(shardId))
}

func TestHeadersCacher_BigCacheALotOfHeadersShouldWork(t *testing.T) {
	t.Parallel()

	cacheSize := 100000
	numHeadersToGenerate := cacheSize
	shardId := uint32(0)

	hdrs, hdrsHash := createASliceOfHeaders(numHeadersToGenerate, shardId)
	hdrsCacher, _ := headersCache.NewHeadersCacher(cacheSize, 50)

	start := time.Now()
	for i := 0; i < numHeadersToGenerate; i++ {
		hdrsCacher.Add(hdrsHash[i], &hdrs[i])
	}
	elapsed := time.Since(start)
	fmt.Printf("insert %d took %s \n", numHeadersToGenerate, elapsed)

	start = time.Now()
	hdr, _ := hdrsCacher.GetHeaderByHash(hdrsHash[100])
	elapsed = time.Since(start)
	assert.Equal(t, &hdrs[100], hdr)
	fmt.Printf("get header by hash took %s \n", elapsed)

	start = time.Now()
	d, _, _ := hdrsCacher.GetHeaderByNonceAndShardId(uint64(100), shardId)
	elapsed = time.Since(start)
	fmt.Printf("get header by shard id and nonce took %s \n", elapsed)
	assert.Equal(t, &hdrs[100], d[0])

	start = time.Now()
	hdrsCacher.RemoveHeaderByNonceAndShardId(uint64(500), shardId)
	elapsed = time.Since(start)
	fmt.Printf("remove header by shard id and nonce took %s \n", elapsed)

	hdr, err := hdrsCacher.GetHeaderByHash(hdrsHash[500])
	assert.Error(t, headersCache.ErrHeaderNotFound, err)

	start = time.Now()
	hdrsCacher.RemoveHeaderByHash(hdrsHash[2012])
	elapsed = time.Since(start)
	fmt.Printf("remove header by hash took %s \n", elapsed)

	hdr, err = hdrsCacher.GetHeaderByHash(hdrsHash[2012])
	assert.Error(t, headersCache.ErrHeaderNotFound, err)
}

func TestHeadersCacher_AddHeadersWithDifferentShardIdOnMultipleGoroutines(t *testing.T) {
	t.Parallel()

	cacheSize := 1001
	numHdrsToGenerate := 1000

	hdrsShad0, hashesShad0 := createASliceOfHeadersNonce0(numHdrsToGenerate, 0)
	hdrsShad1, hashesShad1 := createASliceOfHeaders(numHdrsToGenerate, 1)
	hdrsShad2, hashesShad2 := createASliceOfHeaders(numHdrsToGenerate, 2)
	numElemsToRemove := 500

	hdrsCacher, _ := headersCache.NewHeadersCacher(cacheSize, numElemsToRemove)

	var waitgroup sync.WaitGroup
	start := time.Now()
	for i := 0; i < numHdrsToGenerate; i++ {
		waitgroup.Add(5)
		go func(index int) {
			hdrsCacher.Add(hashesShad0[index], &hdrsShad0[index])
			waitgroup.Done()
		}(i)

		go func(index int) {
			hdrsCacher.Add(hashesShad1[index], &hdrsShad1[index])
			go func(index int) {
				hdrsCacher.RemoveHeaderByHash(hashesShad1[index])
				waitgroup.Done()
			}(index)
			waitgroup.Done()
		}(i)

		go func(index int) {
			hdrsCacher.Add(hashesShad2[index], &hdrsShad2[index])
			go func(index int) {
				hdrsCacher.RemoveHeaderByHash(hashesShad2[index])
				waitgroup.Done()
			}(index)
			waitgroup.Done()
		}(i)
	}

	waitgroup.Wait()

	for i := 0; i < numHdrsToGenerate; i++ {
		waitgroup.Add(1)
		go func(index int) {
			hdrsCacher.RemoveHeaderByHash(hashesShad0[index])
			waitgroup.Done()
		}(i)
	}
	waitgroup.Wait()

	elapsed := time.Since(start)
	fmt.Printf("time need to add %d in cache %s \n", numHdrsToGenerate, elapsed)

	assert.Equal(t, 0, hdrsCacher.GetNumHeadersFromCacheShard(0))
	assert.Equal(t, 0, hdrsCacher.GetNumHeadersFromCacheShard(1))
	assert.Equal(t, 0, hdrsCacher.GetNumHeadersFromCacheShard(2))

}
