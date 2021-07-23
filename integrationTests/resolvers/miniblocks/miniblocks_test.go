package miniblocks

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests/resolvers"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

func TestRequestResolveMiniblockByHashRequestingShardResolvingSameShard(t *testing.T) {
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
	miniblock, hash := resolvers.CreateMiniblock(shardId, shardId)

	//add miniblock in pool
	_, _ = nResolver.DataPool.MiniBlocks().HasOrAdd(hash, miniblock, miniblock.Size())

	//setup header received event
	nRequester.DataPool.MiniBlocks().RegisterHandler(func(key []byte, value interface{}) {
		if bytes.Equal(key, hash) {
			resolvers.Log.Info("received miniblock", "hash", key)
			rm.Done()
		}
	}, core.UniqueIdentifier())

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.IntraShardResolver(factory.MiniBlocksTopic)
	resolvers.Log.LogIfError(err)
	nRequester.WhiteListHandler.Add([][]byte{hash})
	err = resolver.RequestDataFromHash(hash, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

func TestRequestResolveMiniblockByHashRequestingShardResolvingOtherShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := resolvers.NewReceiverMonitor(t)
	shardIdResolver := uint32(0)
	shardIdRequester := uint32(1)
	nResolver, nRequester := resolvers.CreateResolverRequester(shardIdResolver, shardIdRequester)
	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()
	miniblock, hash := resolvers.CreateMiniblock(shardIdResolver, shardIdRequester)

	//add miniblock in pool
	_, _ = nResolver.DataPool.MiniBlocks().HasOrAdd(hash, miniblock, miniblock.Size())

	//setup header received event
	nRequester.DataPool.MiniBlocks().RegisterHandler(func(key []byte, value interface{}) {
		if bytes.Equal(key, hash) {
			resolvers.Log.Info("received miniblock", "hash", key)
			rm.Done()
		}
	}, core.UniqueIdentifier())

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.MiniBlocksTopic, shardIdResolver)
	resolvers.Log.LogIfError(err)
	nRequester.WhiteListHandler.Add([][]byte{hash})
	err = resolver.RequestDataFromHash(hash, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

func TestRequestResolveMiniblockByHashRequestingShardResolvingMeta(t *testing.T) {
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
	miniblock, hash := resolvers.CreateMiniblock(shardId, shardId)

	//add miniblock in pool
	_, _ = nResolver.DataPool.MiniBlocks().HasOrAdd(hash, miniblock, miniblock.Size())

	//setup header received event
	nRequester.DataPool.MiniBlocks().RegisterHandler(func(key []byte, value interface{}) {
		if bytes.Equal(key, hash) {
			resolvers.Log.Info("received miniblock", "hash", key)
			rm.Done()
		}
	}, core.UniqueIdentifier())

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.MiniBlocksTopic, core.MetachainShardId)
	resolvers.Log.LogIfError(err)
	nRequester.WhiteListHandler.Add([][]byte{hash})
	err = resolver.RequestDataFromHash(hash, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

func TestRequestResolveMiniblockByHashRequestingMetaResolvingShard(t *testing.T) {
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
	miniblock, hash := resolvers.CreateMiniblock(shardId, core.MetachainShardId)

	//add miniblock in pool
	_, _ = nResolver.DataPool.MiniBlocks().HasOrAdd(hash, miniblock, miniblock.Size())

	//setup header received event
	nRequester.DataPool.MiniBlocks().RegisterHandler(func(key []byte, value interface{}) {
		if bytes.Equal(key, hash) {
			resolvers.Log.Info("received miniblock", "hash", key)
			rm.Done()
		}
	}, core.UniqueIdentifier())

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.MiniBlocksTopic, shardId)
	resolvers.Log.LogIfError(err)
	nRequester.WhiteListHandler.Add([][]byte{hash})
	err = resolver.RequestDataFromHash(hash, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

func TestRequestResolvePeerMiniblockByHashRequestingShardResolvingSameShard(t *testing.T) {
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
	miniblock, hash := resolvers.CreateMiniblock(core.MetachainShardId, core.AllShardId)

	//add miniblock in pool
	_, _ = nResolver.DataPool.MiniBlocks().HasOrAdd(hash, miniblock, miniblock.Size())

	//setup header received event
	nRequester.DataPool.MiniBlocks().RegisterHandler(func(key []byte, value interface{}) {
		if bytes.Equal(key, hash) {
			resolvers.Log.Info("received miniblock", "hash", key)
			rm.Done()
		}
	}, core.UniqueIdentifier())

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.MiniBlocksTopic, core.AllShardId)
	resolvers.Log.LogIfError(err)
	nRequester.WhiteListHandler.Add([][]byte{hash})
	err = resolver.RequestDataFromHash(hash, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

func TestRequestResolvePeerMiniblockByHashRequestingShardResolvingOtherShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := resolvers.NewReceiverMonitor(t)
	shardIdResolver := uint32(0)
	shardIdRequester := uint32(1)
	nResolver, nRequester := resolvers.CreateResolverRequester(shardIdResolver, shardIdRequester)
	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()
	miniblock, hash := resolvers.CreateMiniblock(shardIdResolver, core.AllShardId)

	//add miniblock in pool
	_, _ = nResolver.DataPool.MiniBlocks().HasOrAdd(hash, miniblock, miniblock.Size())

	//setup header received event
	nRequester.DataPool.MiniBlocks().RegisterHandler(func(key []byte, value interface{}) {
		if bytes.Equal(key, hash) {
			resolvers.Log.Info("received miniblock", "hash", key)
			rm.Done()
		}
	}, core.UniqueIdentifier())

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.MiniBlocksTopic, core.AllShardId)
	resolvers.Log.LogIfError(err)
	nRequester.WhiteListHandler.Add([][]byte{hash})
	err = resolver.RequestDataFromHash(hash, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

func TestRequestResolvePeerMiniblockByHashRequestingShardResolvingMeta(t *testing.T) {
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
	miniblock, hash := resolvers.CreateMiniblock(shardId, core.AllShardId)

	//add miniblock in pool
	_, _ = nResolver.DataPool.MiniBlocks().HasOrAdd(hash, miniblock, miniblock.Size())

	//setup header received event
	nRequester.DataPool.MiniBlocks().RegisterHandler(func(key []byte, value interface{}) {
		if bytes.Equal(key, hash) {
			resolvers.Log.Info("received miniblock", "hash", key)
			rm.Done()
		}
	}, core.UniqueIdentifier())

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.MiniBlocksTopic, core.AllShardId)
	resolvers.Log.LogIfError(err)
	nRequester.WhiteListHandler.Add([][]byte{hash})
	err = resolver.RequestDataFromHash(hash, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}

func TestRequestResolvePeerMiniblockByHashRequestingMetaResolvingShard(t *testing.T) {
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
	miniblock, hash := resolvers.CreateMiniblock(shardId, core.AllShardId)

	//add miniblock in pool
	_, _ = nResolver.DataPool.MiniBlocks().HasOrAdd(hash, miniblock, miniblock.Size())

	//setup header received event
	nRequester.DataPool.MiniBlocks().RegisterHandler(func(key []byte, value interface{}) {
		if bytes.Equal(key, hash) {
			resolvers.Log.Info("received miniblock", "hash", key)
			rm.Done()
		}
	}, core.UniqueIdentifier())

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.MiniBlocksTopic, core.AllShardId)
	resolvers.Log.LogIfError(err)
	nRequester.WhiteListHandler.Add([][]byte{hash})
	err = resolver.RequestDataFromHash(hash, 0)
	resolvers.Log.LogIfError(err)

	rm.WaitWithTimeout()
}
