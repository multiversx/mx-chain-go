package syncer

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("syncer")

type userAccountsSyncer struct {
	hasher             hashing.Hasher
	marshalizer        marshal.Marshalizer
	trieSyncers        map[string]data.TrieSyncer
	dataTries          map[string]data.Trie
	trieStorageManager data.StorageManager
	requestHandler     trie.RequestHandler
	waitTime           time.Duration
	shardId            uint32
	cacher             storage.Cacher
	rootHash           []byte
}

// ArgsNewUserAccountsSyncer defines the arguments needed for the new account syncer
type ArgsNewUserAccountsSyncer struct {
	Hasher             hashing.Hasher
	Marshalizer        marshal.Marshalizer
	TrieStorageManager data.StorageManager
	RequestHandler     trie.RequestHandler
	WaitTime           time.Duration
	ShardId            uint32
	Cacher             storage.Cacher
}

// NewUserAccountsSyncer creates a user account syncer
func NewUserAccountsSyncer(args ArgsNewUserAccountsSyncer) (*userAccountsSyncer, error) {
	if check.IfNil(args.Hasher) {
		return nil, state.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, state.ErrNilMarshalizer
	}
	if check.IfNil(args.TrieStorageManager) {
		return nil, state.ErrNilStorageManager
	}
	if check.IfNil(args.RequestHandler) {
		return nil, state.ErrNilRequestHandler
	}
	if args.WaitTime < time.Second {
		return nil, state.ErrInvalidWaitTime
	}

	u := &userAccountsSyncer{
		hasher:             args.Hasher,
		marshalizer:        args.Marshalizer,
		trieSyncers:        make(map[string]data.TrieSyncer),
		dataTries:          make(map[string]data.Trie),
		trieStorageManager: args.TrieStorageManager,
		requestHandler:     args.RequestHandler,
		waitTime:           args.WaitTime,
		shardId:            args.ShardId,
		cacher:             args.Cacher,
		rootHash:           nil,
	}

	return u, nil
}

// SyncAccounts will launch the syncing method to gather all the data needed for userAccounts
func (u *userAccountsSyncer) SyncAccounts(rootHash []byte) error {
	u.rootHash = rootHash

	dataTrie, err := trie.NewTrie(u.trieStorageManager, u.marshalizer, u.hasher)
	if err != nil {
		return err
	}

	u.dataTries[string(rootHash)] = dataTrie
	trieSyncer, err := trie.NewTrieSyncer(u.requestHandler, u.cacher, dataTrie, u.waitTime, u.shardId, factory.AccountTrieNodesTopic)
	if err != nil {
		return err
	}
	u.trieSyncers[string(rootHash)] = trieSyncer

	err = trieSyncer.StartSyncing(rootHash)
	if err != nil {
		return err
	}

	rootHashes, err := u.findAllAccountRootHashes(dataTrie)
	if err != nil {
		return err
	}

	err = u.syncAccountDataTries(rootHashes)
	if err != nil {
		return err
	}

	return nil
}

func (u *userAccountsSyncer) syncAccountDataTries(rootHashes [][]byte) error {
	for _, rootHash := range rootHashes {
		dataTrie, err := trie.NewTrie(u.trieStorageManager, u.marshalizer, u.hasher)
		if err != nil {
			return err
		}

		u.dataTries[string(rootHash)] = dataTrie
		trieSyncer, err := trie.NewTrieSyncer(u.requestHandler, u.cacher, dataTrie, u.waitTime, u.shardId, factory.AccountTrieNodesTopic)
		if err != nil {
			return err
		}
		u.trieSyncers[string(rootHash)] = trieSyncer

		err = trieSyncer.StartSyncing(rootHash)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *userAccountsSyncer) findAllAccountRootHashes(mainTrie data.Trie) ([][]byte, error) {
	leafs, err := mainTrie.GetAllLeaves()
	if err != nil {
		return nil, err
	}

	rootHashes := make([][]byte, 0)
	for _, leaf := range leafs {
		account := &state.Account{}
		err := u.marshalizer.Unmarshal(account, leaf)
		if err != nil {
			log.Trace("this must be a leaf with code", "err", err)
		}

		if len(account.RootHash) > 0 {
			rootHashes = append(rootHashes, account.RootHash)
		}
	}

	return rootHashes, nil
}

// GetSyncedTries returns the synced map of data trie
func (u *userAccountsSyncer) GetSyncedTries() map[string]data.Trie {
	return u.dataTries
}

// IsInterfaceNil returns true if underlying object is nil
func (u *userAccountsSyncer) IsInterfaceNil() bool {
	return u == nil
}
