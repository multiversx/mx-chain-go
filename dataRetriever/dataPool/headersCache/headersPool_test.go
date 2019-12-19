package headersCache_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool/headersCache"
	"github.com/stretchr/testify/assert"
)

func TestNewHeadersCacher_AddHeadersInCache(t *testing.T) {
	t.Parallel()

	headersCacher, _ := headersCache.NewHeadersCacher(1000, 100)

	nonce := uint64(1)
	shardId := uint32(0)

	headerHash1 := []byte("hash1")
	headerHash2 := []byte("hash2")
	testHdr1 := &block.Header{Nonce: nonce, ShardId: shardId}
	testHdr2 := &block.Header{Nonce: nonce, ShardId: shardId, Round: 100}

	headersCacher.Add(headerHash1, testHdr1)
	headersCacher.Add(headerHash2, testHdr2)

	header, err := headersCacher.GetHeaderByHash(headerHash1)
	require.Nil(t, err)
	require.Equal(t, testHdr1, header)

	header, err = headersCacher.GetHeaderByHash(headerHash2)
	require.Nil(t, err)
	require.Equal(t, testHdr2, header)

	expectedHeaders := []data.HeaderHandler{testHdr1, testHdr2}
	headers, _, err := headersCacher.GetHeaderByNonceAndShardId(nonce, shardId)
	require.Nil(t, err)
	require.Equal(t, expectedHeaders, headers)
}

func TestHeadersCacher_AddHeadersInCacheAndRemoveByHash(t *testing.T) {
	t.Parallel()

	headersCacher, _ := headersCache.NewHeadersCacher(1000, 100)

	nonce := uint64(1)
	shardId := uint32(0)

	headerHash1 := []byte("hash1")
	headerHash2 := []byte("hash2")
	testHdr1 := &block.Header{Nonce: nonce, ShardId: shardId}
	testHdr2 := &block.Header{Nonce: nonce, ShardId: shardId, Round: 100}

	headersCacher.Add(headerHash1, testHdr1)
	headersCacher.Add(headerHash2, testHdr2)

	headersCacher.RemoveHeaderByHash(headerHash1)
	header, err := headersCacher.GetHeaderByHash(headerHash1)
	require.Nil(t, header)
	require.Equal(t, headersCache.ErrHeaderNotFound, err)

	headersCacher.RemoveHeaderByHash(headerHash2)
	header, err = headersCacher.GetHeaderByHash(headerHash2)
	require.Nil(t, header)
	require.Equal(t, headersCache.ErrHeaderNotFound, err)
}

func TestHeadersCacher_AddHeadersInCacheAndRemoveByNonceAndShadId(t *testing.T) {
	t.Parallel()

	headersCacher, _ := headersCache.NewHeadersCacher(1000, 100)

	nonce := uint64(1)
	shardId := uint32(0)

	headerHash1 := []byte("hash1")
	headerHash2 := []byte("hash2")
	testHdr1 := &block.Header{Nonce: nonce, ShardId: shardId}
	testHdr2 := &block.Header{Nonce: nonce, ShardId: shardId, Round: 100}

	headersCacher.Add(headerHash1, testHdr1)
	headersCacher.Add(headerHash2, testHdr2)

	headersCacher.RemoveHeaderByNonceAndShardId(nonce, shardId)
	header, err := headersCacher.GetHeaderByHash(headerHash1)
	require.Nil(t, header)
	require.Equal(t, headersCache.ErrHeaderNotFound, err)

	header, err = headersCacher.GetHeaderByHash(headerHash2)
	require.Nil(t, header)
	require.Equal(t, headersCache.ErrHeaderNotFound, err)
}

func TestHeadersCacher_EvictionShouldWork(t *testing.T) {
	t.Parallel()

	headers, headersHashes := createASliceOfHeaders(1000, 0)
	headersCacher, _ := headersCache.NewHeadersCacher(900, 100)

	for i := 0; i < 1000; i++ {
		headersCacher.Add(headersHashes[i], &headers[i])
	}

	// Cache will do eviction 2 times, in headers cache will be 800 headers
	for i := 200; i < 1000; i++ {
		header, err := headersCacher.GetHeaderByHash(headersHashes[i])
		assert.Nil(t, err)
		assert.Equal(t, &headers[i], header)
	}
}

func TestHeadersCacher_ConcurrentRequestsShouldWorkNoEviction(t *testing.T) {
	t.Parallel()

	numHeadersToGenerate := 500

	headers, headersHashes := createASliceOfHeaders(numHeadersToGenerate, 0)
	headersCacher, _ := headersCache.NewHeadersCacher(numHeadersToGenerate+1, 10)

	for i := 0; i < numHeadersToGenerate; i++ {
		go func(index int) {
			headersCacher.Add(headersHashes[index], &headers[index])
			header, err := headersCacher.GetHeaderByHash(headersHashes[index])

			assert.Nil(t, err)
			assert.Equal(t, &headers[index], header)
		}(i)
	}
}

func TestHeadersCacher_ConcurrentRequestsShouldWorkWithEviction(t *testing.T) {
	shardId := uint32(0)
	cacheSize := 2
	numHeadersToGenerate := 500

	headers, headersHashes := createASliceOfHeaders(numHeadersToGenerate, shardId)
	headersCacher, _ := headersCache.NewHeadersCacher(cacheSize, 1)

	var waitgroup sync.WaitGroup
	for i := 0; i < numHeadersToGenerate; i++ {
		waitgroup.Add(1)
		go func(index int) {
			headersCacher.Add(headersHashes[index], &headers[index])
			waitgroup.Done()
		}(i)
	}
	waitgroup.Wait()
	// cache size after all eviction is finish should be 2
	assert.Equal(t, 2, headersCacher.GetNumHeadersFromCacheShard(shardId))

	numHeadersToGenerate = 3
	headers, headersHashes = createASliceOfHeaders(3, shardId)
	for i := 0; i < numHeadersToGenerate; i++ {
		headersCacher.Add(headersHashes[i], &headers[i])
		time.Sleep(time.Microsecond)
	}

	assert.Equal(t, 2, headersCacher.GetNumHeadersFromCacheShard(shardId))
	header1, err := headersCacher.GetHeaderByHash(headersHashes[1])
	assert.Nil(t, err)
	assert.Equal(t, &headers[1], header1)

	header2, err := headersCacher.GetHeaderByHash(headersHashes[2])
	assert.Nil(t, err)
	assert.Equal(t, &headers[2], header2)
}

func TestHeadersCacher_AddHeadersWithSameNonceShouldBeRemovedAtEviction(t *testing.T) {
	t.Parallel()

	shardId := uint32(0)
	cacheSize := 2

	hash1, hash2, hash3 := []byte("hash1"), []byte("hash2"), []byte("hash3")
	header1, header2, header3 := &block.Header{Nonce: 0}, &block.Header{Nonce: 0}, &block.Header{Nonce: 1}

	headersCacher, _ := headersCache.NewHeadersCacher(cacheSize, 1)
	headersCacher.Add(hash1, header1)
	headersCacher.Add(hash2, header2)
	headersCacher.Add(hash3, header3)

	assert.Equal(t, 1, headersCacher.GetNumHeadersFromCacheShard(shardId))

	header, err := headersCacher.GetHeaderByHash(hash3)
	assert.Nil(t, err)
	assert.Equal(t, header3, header)
}

func TestHeadersCacher_AddALotOfHeadersAndCheckEviction(t *testing.T) {
	t.Parallel()

	cacheSize := 100
	numHeaders := 500
	shardId := uint32(0)
	headers, headersHash := createASliceOfHeaders(numHeaders, shardId)
	headersCacher, _ := headersCache.NewHeadersCacher(cacheSize, 50)

	var waitgroup sync.WaitGroup
	for i := 0; i < numHeaders; i++ {
		waitgroup.Add(1)
		go func(index int) {
			headersCacher.Add(headersHash[index], &headers[index])
			waitgroup.Done()
		}(i)
	}

	waitgroup.Wait()
	assert.Equal(t, 100, headersCacher.GetNumHeadersFromCacheShard(shardId))
}

func TestHeadersCacher_BigCacheALotOfHeadersShouldWork(t *testing.T) {
	t.Parallel()

	cacheSize := 100000
	numHeadersToGenerate := cacheSize
	shardId := uint32(0)

	headers, headersHash := createASliceOfHeaders(numHeadersToGenerate, shardId)
	headersCacher, _ := headersCache.NewHeadersCacher(cacheSize, 50)

	start := time.Now()
	for i := 0; i < numHeadersToGenerate; i++ {
		headersCacher.Add(headersHash[i], &headers[i])
	}
	elapsed := time.Since(start)
	fmt.Printf("insert %d took %s \n", numHeadersToGenerate, elapsed)

	start = time.Now()
	header, _ := headersCacher.GetHeaderByHash(headersHash[100])
	elapsed = time.Since(start)
	assert.Equal(t, &headers[100], header)
	fmt.Printf("get header by hash took %s \n", elapsed)

	start = time.Now()
	d, _, _ := headersCacher.GetHeaderByNonceAndShardId(uint64(100), shardId)
	elapsed = time.Since(start)
	fmt.Printf("get header by shard id and nonce took %s \n", elapsed)
	assert.Equal(t, &headers[100], d[0])

	start = time.Now()
	headersCacher.RemoveHeaderByNonceAndShardId(uint64(500), shardId)
	elapsed = time.Since(start)
	fmt.Printf("remove header by shard id and nonce took %s \n", elapsed)

	header, err := headersCacher.GetHeaderByHash(headersHash[500])
	assert.Error(t, headersCache.ErrHeaderNotFound, err)

	start = time.Now()
	headersCacher.RemoveHeaderByHash(headersHash[2012])
	elapsed = time.Since(start)
	fmt.Printf("remove header by hash took %s \n", elapsed)

	header, err = headersCacher.GetHeaderByHash(headersHash[2012])
	assert.Error(t, headersCache.ErrHeaderNotFound, err)
}

func TestHeadersCacher_AddHeadersWithDifferentShardIdOnMultipleGoroutines(t *testing.T) {
	t.Parallel()

	cacheSize := 1001
	numHdrsToGenerate := 1000

	headersShard0, hashesShad0 := createASliceOfHeadersNonce0(numHdrsToGenerate, 0)
	headersShard1, hashesShad1 := createASliceOfHeaders(numHdrsToGenerate, 1)
	headersShard2, hashesShad2 := createASliceOfHeaders(numHdrsToGenerate, 2)
	numElemsToRemove := 500

	headersCacher, _ := headersCache.NewHeadersCacher(cacheSize, numElemsToRemove)

	var waitgroup sync.WaitGroup
	start := time.Now()
	for i := 0; i < numHdrsToGenerate; i++ {
		waitgroup.Add(5)
		go func(index int) {
			headersCacher.Add(hashesShad0[index], &headersShard0[index])
			waitgroup.Done()
		}(i)

		go func(index int) {
			headersCacher.Add(hashesShad1[index], &headersShard1[index])
			go func(index int) {
				headersCacher.RemoveHeaderByHash(hashesShad1[index])
				waitgroup.Done()
			}(index)
			waitgroup.Done()
		}(i)

		go func(index int) {
			headersCacher.Add(hashesShad2[index], &headersShard2[index])
			go func(index int) {
				headersCacher.RemoveHeaderByHash(hashesShad2[index])
				waitgroup.Done()
			}(index)
			waitgroup.Done()
		}(i)
	}

	waitgroup.Wait()

	for i := 0; i < numHdrsToGenerate; i++ {
		waitgroup.Add(1)
		go func(index int) {
			headersCacher.RemoveHeaderByHash(hashesShad0[index])
			waitgroup.Done()
		}(i)
	}
	waitgroup.Wait()

	elapsed := time.Since(start)
	fmt.Printf("time need to add %d in cache %s \n", numHdrsToGenerate, elapsed)

	assert.Equal(t, 0, headersCacher.GetNumHeadersFromCacheShard(0))
	assert.Equal(t, 0, headersCacher.GetNumHeadersFromCacheShard(1))
	assert.Equal(t, 0, headersCacher.GetNumHeadersFromCacheShard(2))
}

func TestHeadersCacher_TestEvictionRemoveCorrectHeader(t *testing.T) {
	t.Parallel()

	shardId := uint32(0)
	cacheSize := 2
	numHeadersToGenerate := 3

	headers, headersHashes := createASliceOfHeaders(numHeadersToGenerate, shardId)
	headersCacher, _ := headersCache.NewHeadersCacher(cacheSize, 1)

	for i := 0; i < numHeadersToGenerate-1; i++ {
		headersCacher.Add(headersHashes[i], &headers[i])
		time.Sleep(time.Microsecond)
	}

	header, err := headersCacher.GetHeaderByHash(headersHashes[0])
	assert.Nil(t, err)
	assert.Equal(t, &headers[0], header)

	headersCacher.Add(headersHashes[2], &headers[2])

	header, err = headersCacher.GetHeaderByHash(headersHashes[0])
	assert.Nil(t, err)
	assert.Equal(t, &headers[0], header)

	header, err = headersCacher.GetHeaderByHash(headersHashes[2])
	assert.Nil(t, err)
	assert.Equal(t, &headers[2], header)

	header, err = headersCacher.GetHeaderByHash(headersHashes[1])
	assert.Equal(t, headersCache.ErrHeaderNotFound, err)
}

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
