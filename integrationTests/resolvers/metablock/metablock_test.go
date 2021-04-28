package metablock

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/resolvers"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

//------- Request resolve by hash

func TestRequestResolveMetaHeadersByHashRequestingShardResolvingShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := resolvers.NewReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := resolvers.CreateResolverRequester(shardId, shardId)
	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()
	headerNonce := uint64(0)
	header, hash := resolvers.CreateMetaHeader(headerNonce, integrationTests.ChainID)

	//add header with nonce 0 in pool
	nResolver.DataPool.Headers().AddHeader(hash, header)

	//setup header received event
	nRequester.DataPool.Headers().RegisterHandler(
		func(header data.HeaderHandler, key []byte) {
			if bytes.Equal(key, hash) {
				resolvers.Log.Info("received meta header", "hash", key)
				rm.Done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.MetaChainResolver(factory.MetachainBlocksTopic)
	resolvers.Log.LogIfError(err)
	err = resolver.RequestDataFromHash(hash, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

func TestRequestResolveMetaHeadersByHashRequestingMetaResolvingShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := resolvers.NewReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := resolvers.CreateResolverRequester(shardId, core.MetachainShardId)
	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()
	headerNonce := uint64(0)
	header, hash := resolvers.CreateMetaHeader(headerNonce, integrationTests.ChainID)

	//add header with nonce 0 in pool
	nResolver.DataPool.Headers().AddHeader(hash, header)

	//setup header received event
	nRequester.DataPool.Headers().RegisterHandler(
		func(header data.HeaderHandler, key []byte) {
			if bytes.Equal(key, hash) {
				resolvers.Log.Info("received meta header", "hash", key)
				rm.Done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.MetaChainResolver(factory.MetachainBlocksTopic)
	resolvers.Log.LogIfError(err)
	err = resolver.RequestDataFromHash(hash, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

func TestRequestResolveMetaHeadersByHashRequestingShardResolvingMeta(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := resolvers.NewReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := resolvers.CreateResolverRequester(core.MetachainShardId, shardId)
	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()
	headerNonce := uint64(0)
	header, hash := resolvers.CreateMetaHeader(headerNonce, integrationTests.ChainID)

	//add header with nonce 0 in pool
	nResolver.DataPool.Headers().AddHeader(hash, header)

	//setup header received event
	nRequester.DataPool.Headers().RegisterHandler(
		func(header data.HeaderHandler, key []byte) {
			if bytes.Equal(key, hash) {
				resolvers.Log.Info("received meta header", "hash", key)
				rm.Done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.MetaChainResolver(factory.MetachainBlocksTopic)
	resolvers.Log.LogIfError(err)
	err = resolver.RequestDataFromHash(hash, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

//------- Request resolve by nonce

func TestRequestResolveMetaHeadersByNonceRequestingShardResolvingShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := resolvers.NewReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := resolvers.CreateResolverRequester(shardId, shardId)
	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()
	headerNonce := uint64(0)
	header, hash := resolvers.CreateMetaHeader(headerNonce, integrationTests.ChainID)

	//add header with nonce 0 in pool
	nResolver.DataPool.Headers().AddHeader(hash, header)

	//setup header received event
	nRequester.DataPool.Headers().RegisterHandler(
		func(header data.HeaderHandler, key []byte) {
			if bytes.Equal(key, hash) {
				resolvers.Log.Info("received header", "hash", key)
				rm.Done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.MetaChainResolver(factory.MetachainBlocksTopic)
	resolvers.Log.LogIfError(err)
	headerResolver, ok := resolver.(dataRetriever.HeaderResolver)
	assert.True(t, ok)
	err = headerResolver.RequestDataFromNonce(headerNonce, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

func TestRequestResolveMetaHeadersByNonceRequestingMetaResolvingShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := resolvers.NewReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := resolvers.CreateResolverRequester(shardId, core.MetachainShardId)
	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()
	headerNonce := uint64(0)
	header, hash := resolvers.CreateMetaHeader(headerNonce, integrationTests.ChainID)

	//add header with nonce 0 in pool
	nResolver.DataPool.Headers().AddHeader(hash, header)

	//setup header received event
	nRequester.DataPool.Headers().RegisterHandler(
		func(header data.HeaderHandler, key []byte) {
			if bytes.Equal(key, hash) {
				resolvers.Log.Info("received header", "hash", key)
				rm.Done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.MetaChainResolver(factory.MetachainBlocksTopic)
	resolvers.Log.LogIfError(err)
	headerResolver, ok := resolver.(dataRetriever.HeaderResolver)
	assert.True(t, ok)
	err = headerResolver.RequestDataFromNonce(headerNonce, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

func TestRequestResolveMetaHeadersByNonceRequestingShardResolvingMeta(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := resolvers.NewReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := resolvers.CreateResolverRequester(core.MetachainShardId, shardId)
	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()
	headerNonce := uint64(0)
	header, hash := resolvers.CreateMetaHeader(headerNonce, integrationTests.ChainID)

	//add header with nonce 0 in pool
	nResolver.DataPool.Headers().AddHeader(hash, header)

	//setup header received event
	nRequester.DataPool.Headers().RegisterHandler(
		func(header data.HeaderHandler, key []byte) {
			if bytes.Equal(key, hash) {
				resolvers.Log.Info("received header", "hash", key)
				rm.Done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.MetaChainResolver(factory.MetachainBlocksTopic)
	resolvers.Log.LogIfError(err)
	headerResolver, ok := resolver.(dataRetriever.HeaderResolver)
	assert.True(t, ok)
	err = headerResolver.RequestDataFromNonce(headerNonce, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}
