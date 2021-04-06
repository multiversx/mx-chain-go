package syncer

import (
	"context"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type baseAccountsSyncer struct {
	hasher                    hashing.Hasher
	marshalizer               marshal.Marshalizer
	dataTries                 map[string]struct{}
	mutex                     sync.Mutex
	trieStorageManager        data.StorageManager
	requestHandler            trie.RequestHandler
	timeout                   time.Duration
	shardId                   uint32
	cacher                    storage.Cacher
	rootHash                  []byte
	maxTrieLevelInMemory      uint
	name                      string
	maxHardCapForMissingNodes int
}

const timeBetweenStatisticsPrints = time.Second * 2

// ArgsNewBaseAccountsSyncer defines the arguments needed for the new account syncer
type ArgsNewBaseAccountsSyncer struct {
	Hasher                    hashing.Hasher
	Marshalizer               marshal.Marshalizer
	TrieStorageManager        data.StorageManager
	RequestHandler            trie.RequestHandler
	Timeout                   time.Duration
	Cacher                    storage.Cacher
	MaxTrieLevelInMemory      uint
	MaxHardCapForMissingNodes int
}

func checkArgs(args ArgsNewBaseAccountsSyncer) error {
	if check.IfNil(args.Hasher) {
		return state.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return state.ErrNilMarshalizer
	}
	if check.IfNil(args.TrieStorageManager) {
		return state.ErrNilStorageManager
	}
	if check.IfNil(args.RequestHandler) {
		return state.ErrNilRequestHandler
	}
	if check.IfNil(args.Cacher) {
		return state.ErrNilCacher
	}
	if args.MaxHardCapForMissingNodes < 1 {
		return state.ErrInvalidMaxHardCapForMissingNodes
	}

	return nil
}

func (b *baseAccountsSyncer) syncMainTrie(
	rootHash []byte,
	trieTopic string,
	ssh data.SyncStatisticsHandler,
	ctx context.Context,
) (data.Trie, error) {
	b.rootHash = rootHash

	dataTrie, err := trie.NewTrie(b.trieStorageManager, b.marshalizer, b.hasher, b.maxTrieLevelInMemory)
	if err != nil {
		return nil, err
	}

	b.dataTries[string(rootHash)] = struct{}{}
	arg := trie.ArgTrieSyncer{
		RequestHandler:                 b.requestHandler,
		InterceptedNodes:               b.cacher,
		Trie:                           dataTrie,
		ShardId:                        b.shardId,
		Topic:                          trieTopic,
		TrieSyncStatistics:             ssh,
		TimeoutBetweenTrieNodesCommits: b.timeout,
		MaxHardCapForMissingNodes:      b.maxHardCapForMissingNodes,
	}
	trieSyncer, err := trie.NewTrieSyncer(arg)
	if err != nil {
		return nil, err
	}

	err = trieSyncer.StartSyncing(rootHash, ctx)
	if err != nil {
		return nil, err
	}

	return dataTrie, nil
}

func (b *baseAccountsSyncer) printStatistics(ssh data.SyncStatisticsHandler, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info("finished trie sync", "name", b.name, "num received", ssh.NumReceived(), "num missing", ssh.NumMissing())
			return
		case <-time.After(timeBetweenStatisticsPrints):
			log.Info("trie sync in progress", "name", b.name, "num received", ssh.NumReceived(), "num missing", ssh.NumMissing())
		}
	}
}

// Deprecated: GetSyncedTries returns the synced map of data trie. This is likely to case OOM exceptions
//TODO remove this function after fixing the hardfork sync state mechanism
func (b *baseAccountsSyncer) GetSyncedTries() map[string]data.Trie {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	dataTrie, err := trie.NewTrie(b.trieStorageManager, b.marshalizer, b.hasher, b.maxTrieLevelInMemory)
	if err != nil {
		log.Warn("error creating a new trie in baseAccountsSyncer.GetSyncedTries", "error", err)
		return make(map[string]data.Trie)
	}

	var recreatedTrie data.Trie
	clonedMap := make(map[string]data.Trie, len(b.dataTries))
	for key := range b.dataTries {
		recreatedTrie, err = dataTrie.Recreate([]byte(key))
		if err != nil {
			log.Warn("error recreating trie in baseAccountsSyncer.GetSyncedTries",
				"roothash", []byte(key), "error", err)
			continue
		}

		clonedMap[key] = recreatedTrie
	}

	return clonedMap
}

// IsInterfaceNil returns true if underlying object is nil
func (b *baseAccountsSyncer) IsInterfaceNil() bool {
	return b == nil
}
