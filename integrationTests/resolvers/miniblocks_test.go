package resolvers

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"

	"github.com/ElrondNetwork/elrond-go/process/factory"
)

func TestRequestResolveMiniblockByHashRequestingShardResolvingSameShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := newReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := createResolverRequester(shardId, shardId)
	miniblock, hash := createMiniblock(shardId, shardId)

	//add miniblock in pool
	_, _ = nResolver.ShardDataPool.MiniBlocks().HasOrAdd(hash, miniblock)

	//setup header received event
	nRequester.ShardDataPool.MiniBlocks().RegisterHandler(
		func(key []byte) {
			if bytes.Equal(key, hash) {
				log.Info("received miniblock", "hash", key)
				rm.done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.IntraShardResolver(factory.MiniBlocksTopic)
	log.LogIfError(err)
	err = resolver.RequestDataFromHash(hash)
	log.LogIfError(err)

	rm.waitWithTimeout()
}

func TestRequestResolveMiniblockByHashRequestingShardResolvingOtherShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := newReceiverMonitor(t)
	shardIdResolver := uint32(0)
	shardIdRequester := uint32(1)
	nResolver, nRequester := createResolverRequester(shardIdResolver, shardIdRequester)
	miniblock, hash := createMiniblock(shardIdResolver, shardIdRequester)

	//add miniblock in pool
	_, _ = nResolver.ShardDataPool.MiniBlocks().HasOrAdd(hash, miniblock)

	//setup header received event
	nRequester.ShardDataPool.MiniBlocks().RegisterHandler(
		func(key []byte) {
			if bytes.Equal(key, hash) {
				log.Info("received miniblock", "hash", key)
				rm.done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.MiniBlocksTopic, shardIdResolver)
	log.LogIfError(err)
	err = resolver.RequestDataFromHash(hash)
	log.LogIfError(err)

	rm.waitWithTimeout()
}

func TestRequestResolveMiniblockByHashRequestingShardResolvingMeta(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := newReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := createResolverRequester(core.MetachainShardId, shardId)
	miniblock, hash := createMiniblock(shardId, shardId)

	//add miniblock in pool
	_, _ = nResolver.MetaDataPool.MiniBlocks().HasOrAdd(hash, miniblock)

	//setup header received event
	nRequester.ShardDataPool.MiniBlocks().RegisterHandler(
		func(key []byte) {
			if bytes.Equal(key, hash) {
				log.Info("received miniblock", "hash", key)
				rm.done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.MiniBlocksTopic, core.MetachainShardId)
	log.LogIfError(err)
	err = resolver.RequestDataFromHash(hash)
	log.LogIfError(err)

	rm.waitWithTimeout()
}

func TestRequestResolveMiniblockByHashRequestingMetaResolvingShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := newReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := createResolverRequester(shardId, core.MetachainShardId)
	miniblock, hash := createMiniblock(shardId, core.MetachainShardId)

	//add miniblock in pool
	_, _ = nResolver.ShardDataPool.MiniBlocks().HasOrAdd(hash, miniblock)

	//setup header received event
	nRequester.MetaDataPool.MiniBlocks().RegisterHandler(
		func(key []byte) {
			if bytes.Equal(key, hash) {
				log.Info("received miniblock", "hash", key)
				rm.done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.MiniBlocksTopic, shardId)
	log.LogIfError(err)
	err = resolver.RequestDataFromHash(hash)
	log.LogIfError(err)

	rm.waitWithTimeout()
}
