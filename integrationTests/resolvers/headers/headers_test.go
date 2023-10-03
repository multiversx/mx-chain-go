package headers

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/resolvers"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/stretchr/testify/assert"
)

//------- Request resolve by hash

func TestRequestResolveShardHeadersByHashRequestingShardResolvingShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := resolvers.NewReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := resolvers.CreateResolverRequester(shardId, shardId)
	defer func() {
		nRequester.Close()
		nResolver.Close()
	}()
	headerNonce := uint64(0)
	header, hash := resolvers.CreateShardHeader(headerNonce, integrationTests.ChainID)

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
	requester, err := nRequester.RequestersFinder.CrossShardRequester(factory.ShardBlocksTopic, core.MetachainShardId)
	resolvers.Log.LogIfError(err)
	err = requester.RequestDataFromHash(hash, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

func TestRequestResolveShardHeadersByHashRequestingMetaResolvingShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := resolvers.NewReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := resolvers.CreateResolverRequester(shardId, core.MetachainShardId)
	defer func() {
		nRequester.Close()
		nResolver.Close()
	}()
	headerNonce := uint64(0)
	header, hash := resolvers.CreateShardHeader(headerNonce, integrationTests.ChainID)

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
	requester, err := nRequester.RequestersFinder.CrossShardRequester(factory.ShardBlocksTopic, shardId)
	resolvers.Log.LogIfError(err)
	err = requester.RequestDataFromHash(hash, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

func TestRequestResolveShardHeadersByHashRequestingShardResolvingMeta(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := resolvers.NewReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := resolvers.CreateResolverRequester(core.MetachainShardId, shardId)
	defer func() {
		nRequester.Close()
		nResolver.Close()
	}()
	headerNonce := uint64(0)
	header, hash := resolvers.CreateShardHeader(headerNonce, integrationTests.ChainID)

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
	requester, err := nRequester.RequestersFinder.CrossShardRequester(factory.ShardBlocksTopic, core.MetachainShardId)
	resolvers.Log.LogIfError(err)
	err = requester.RequestDataFromHash(hash, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

//------- Request resolve by nonce

func TestRequestResolveShardHeadersByNonceRequestingShardResolvingShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := resolvers.NewReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := resolvers.CreateResolverRequester(shardId, shardId)
	defer func() {
		nRequester.Close()
		nResolver.Close()
	}()
	headerNonce := uint64(0)
	header, hash := resolvers.CreateShardHeader(headerNonce, integrationTests.ChainID)

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
	requester, err := nRequester.RequestersFinder.CrossShardRequester(factory.ShardBlocksTopic, core.MetachainShardId)
	resolvers.Log.LogIfError(err)
	headerRequester, ok := requester.(requestHandlers.HeaderRequester)
	assert.True(t, ok)
	err = headerRequester.RequestDataFromNonce(headerNonce, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

func TestRequestResolveShardHeadersByNonceRequestingMetaResolvingShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := resolvers.NewReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := resolvers.CreateResolverRequester(shardId, core.MetachainShardId)
	defer func() {
		nRequester.Close()
		nResolver.Close()
	}()
	headerNonce := uint64(0)
	header, hash := resolvers.CreateShardHeader(headerNonce, integrationTests.ChainID)

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
	requester, err := nRequester.RequestersFinder.CrossShardRequester(factory.ShardBlocksTopic, shardId)
	resolvers.Log.LogIfError(err)
	headerRequester, ok := requester.(requestHandlers.HeaderRequester)
	assert.True(t, ok)
	err = headerRequester.RequestDataFromNonce(headerNonce, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

func TestRequestResolveShardHeadersByNonceRequestingShardResolvingMeta(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := resolvers.NewReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := resolvers.CreateResolverRequester(core.MetachainShardId, shardId)
	defer func() {
		nRequester.Close()
		nResolver.Close()
	}()
	headerNonce := uint64(0)
	header, hash := resolvers.CreateShardHeader(headerNonce, integrationTests.ChainID)

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
	requester, err := nRequester.RequestersFinder.CrossShardRequester(factory.ShardBlocksTopic, core.MetachainShardId)
	resolvers.Log.LogIfError(err)
	headerRequester, ok := requester.(requestHandlers.HeaderRequester)
	assert.True(t, ok)
	err = headerRequester.RequestDataFromNonce(headerNonce, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}
