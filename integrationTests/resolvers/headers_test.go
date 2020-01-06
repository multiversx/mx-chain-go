package resolvers

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

//------- Request resolve by hash

func TestRequestResolveShardHeadersByHashRequestingShardResolvingShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := newReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := createResolverRequester(shardId, shardId)
	headerNonce := uint64(0)
	header, hash := createShardHeader(headerNonce, integrationTests.IntegrationTestsChainID)

	//add header with nonce 0 in pool
	_, _ = nResolver.ShardDataPool.Headers().HasOrAdd(hash, header)

	//setup header received event
	nRequester.ShardDataPool.Headers().RegisterHandler(
		func(key []byte) {
			if bytes.Equal(key, hash) {
				log.Info("received header", "hash", key)
				rm.done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.ShardBlocksTopic, sharding.MetachainShardId)
	log.LogIfError(err)
	err = resolver.RequestDataFromHash(hash)
	log.LogIfError(err)

	rm.waitWithTimeout()
}

func TestRequestResolveShardHeadersByHashRequestingMetaResolvingShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := newReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := createResolverRequester(shardId, sharding.MetachainShardId)
	headerNonce := uint64(0)
	header, hash := createShardHeader(headerNonce, integrationTests.IntegrationTestsChainID)

	//add header with nonce 0 in pool
	_, _ = nResolver.ShardDataPool.Headers().HasOrAdd(hash, header)

	//setup header received event
	nRequester.MetaDataPool.ShardHeaders().RegisterHandler(
		func(key []byte) {
			if bytes.Equal(key, hash) {
				log.Info("received header", "hash", key)
				rm.done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.ShardBlocksTopic, shardId)
	log.LogIfError(err)
	err = resolver.RequestDataFromHash(hash)
	log.LogIfError(err)

	rm.waitWithTimeout()
}

func TestRequestResolveShardHeadersByHashRequestingShardResolvingMeta(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := newReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := createResolverRequester(sharding.MetachainShardId, shardId)
	headerNonce := uint64(0)
	header, hash := createShardHeader(headerNonce, integrationTests.IntegrationTestsChainID)

	//add header with nonce 0 in pool
	_, _ = nResolver.MetaDataPool.ShardHeaders().HasOrAdd(hash, header)

	//setup header received event
	nRequester.ShardDataPool.Headers().RegisterHandler(
		func(key []byte) {
			if bytes.Equal(key, hash) {
				log.Info("received header", "hash", key)
				rm.done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.ShardBlocksTopic, sharding.MetachainShardId)
	log.LogIfError(err)
	err = resolver.RequestDataFromHash(hash)
	log.LogIfError(err)

	rm.waitWithTimeout()
}

//------- Request resolve by nonce

func TestRequestResolveShardHeadersByNonceRequestingShardResolvingShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := newReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := createResolverRequester(shardId, shardId)
	headerNonce := uint64(0)
	header, hash := createShardHeader(headerNonce, integrationTests.IntegrationTestsChainID)

	//add header with nonce 0 in pool
	_, _ = nResolver.ShardDataPool.Headers().HasOrAdd(hash, header)
	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(shardId, hash)
	nResolver.ShardDataPool.HeadersNonces().Merge(headerNonce, syncMap)

	//setup header received event
	nRequester.ShardDataPool.Headers().RegisterHandler(
		func(key []byte) {
			if bytes.Equal(key, hash) {
				log.Info("received header", "hash", key)
				rm.done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.ShardBlocksTopic, sharding.MetachainShardId)
	log.LogIfError(err)
	headerResolver, ok := resolver.(dataRetriever.HeaderResolver)
	assert.True(t, ok)
	err = headerResolver.RequestDataFromNonce(headerNonce)
	log.LogIfError(err)

	rm.waitWithTimeout()
}

func TestRequestResolveShardHeadersByNonceRequestingMetaResolvingShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := newReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := createResolverRequester(shardId, sharding.MetachainShardId)
	headerNonce := uint64(0)
	header, hash := createShardHeader(headerNonce, integrationTests.IntegrationTestsChainID)

	//add header with nonce 0 in pool
	_, _ = nResolver.ShardDataPool.Headers().HasOrAdd(hash, header)
	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(shardId, hash)
	nResolver.ShardDataPool.HeadersNonces().Merge(headerNonce, syncMap)

	//setup header received event
	nRequester.MetaDataPool.ShardHeaders().RegisterHandler(
		func(key []byte) {
			if bytes.Equal(key, hash) {
				log.Info("received header", "hash", key)
				rm.done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.ShardBlocksTopic, shardId)
	log.LogIfError(err)
	headerResolver, ok := resolver.(dataRetriever.HeaderResolver)
	assert.True(t, ok)
	err = headerResolver.RequestDataFromNonce(headerNonce)
	log.LogIfError(err)

	rm.waitWithTimeout()
}

func TestRequestResolveShardHeadersByNonceRequestingShardResolvingMeta(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := newReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := createResolverRequester(sharding.MetachainShardId, shardId)
	headerNonce := uint64(0)
	header, hash := createShardHeader(headerNonce, integrationTests.IntegrationTestsChainID)

	//add header with nonce 0 in pool
	_, _ = nResolver.MetaDataPool.ShardHeaders().HasOrAdd(hash, header)
	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(shardId, hash)
	nResolver.MetaDataPool.HeadersNonces().Merge(headerNonce, syncMap)

	//setup header received event
	nRequester.ShardDataPool.Headers().RegisterHandler(
		func(key []byte) {
			if bytes.Equal(key, hash) {
				log.Info("received header", "hash", key)
				rm.done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.ShardBlocksTopic, sharding.MetachainShardId)
	log.LogIfError(err)
	headerResolver, ok := resolver.(dataRetriever.HeaderResolver)
	assert.True(t, ok)
	err = headerResolver.RequestDataFromNonce(headerNonce)
	log.LogIfError(err)

	rm.waitWithTimeout()
}
