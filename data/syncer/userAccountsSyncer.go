package syncer

import (
	"context"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

var log = logger.GetOrCreate("syncer")

type userAccountsSyncer struct {
	*baseAccountsSyncer
}

// ArgsNewUserAccountsSyncer defines the arguments needed for the new account syncer
type ArgsNewUserAccountsSyncer struct {
	ArgsNewBaseAccountsSyncer
	ShardId uint32
}

// NewUserAccountsSyncer creates a user account syncer
func NewUserAccountsSyncer(args ArgsNewUserAccountsSyncer) (*userAccountsSyncer, error) {
	err := checkArgs(args.ArgsNewBaseAccountsSyncer)
	if err != nil {
		return nil, err
	}

	b := &baseAccountsSyncer{
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

	u := &userAccountsSyncer{
		baseAccountsSyncer: b,
	}

	return u, nil
}

// SyncAccounts will launch the syncing method to gather all the data needed for userAccounts
func (u *userAccountsSyncer) SyncAccounts(rootHash []byte) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), u.waitTime)
	defer cancel()
	u.ctx = ctx

	err := u.syncMainTrie(rootHash, factory.AccountTrieNodesTopic)
	if err != nil {
		return nil
	}

	mainTrie := u.dataTries[string(rootHash)]
	rootHashes, err := u.findAllAccountRootHashes(mainTrie)
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
		trieSyncer, err := trie.NewTrieSyncer(u.requestHandler, u.cacher, dataTrie, u.shardId, factory.AccountTrieNodesTopic)
		if err != nil {
			return err
		}
		u.trieSyncers[string(rootHash)] = trieSyncer

		err = trieSyncer.StartSyncing(rootHash, u.ctx)
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
		account := state.NewEmptyUserAccount()
		err = u.marshalizer.Unmarshal(account, leaf)
		if err != nil {
			log.Trace("this must be a leaf with code", "err", err)
			continue
		}

		if len(account.RootHash) > 0 {
			rootHashes = append(rootHashes, account.RootHash)
		}
	}

	return rootHashes, nil
}
