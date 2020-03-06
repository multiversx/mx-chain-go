package syncer

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type validatorAccountsSyncer struct {
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

// ArgsNewValidatorAccountsSyncer defines the arguments needed for the new account syncer
type ArgsNewValidatorAccountsSyncer struct {
	Hasher             hashing.Hasher
	Marshalizer        marshal.Marshalizer
	TrieStorageManager data.StorageManager
	RequestHandler     trie.RequestHandler
	WaitTime           time.Duration
	ShardId            uint32
	Cacher             storage.Cacher
}

// NewValidatorAccountsSyncer creates a validator account syncer
func NewValidatorAccountsSyncer(args ArgsNewValidatorAccountsSyncer) (*validatorAccountsSyncer, error) {
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

	u := &validatorAccountsSyncer{
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

// SyncAccounts will launch the syncing method to gather all the data needed for validatorAccounts
func (v *validatorAccountsSyncer) SyncAccounts(rootHash []byte) error {
	v.rootHash = rootHash

	dataTrie, err := trie.NewTrie(v.trieStorageManager, v.marshalizer, v.hasher)
	if err != nil {
		return err
	}

	v.dataTries[string(rootHash)] = dataTrie
	trieSyncer, err := trie.NewTrieSyncer(v.requestHandler, v.cacher, dataTrie, v.waitTime, v.shardId, factory.ValidatorTrieNodesTopic)
	if err != nil {
		return err
	}
	v.trieSyncers[string(rootHash)] = trieSyncer

	err = trieSyncer.StartSyncing(rootHash)
	if err != nil {
		return err
	}

	return nil
}

// GetSyncedTries returns the synced map of data trie
func (v *validatorAccountsSyncer) GetSyncedTries() map[string]data.Trie {
	return v.dataTries
}

// IsInterfaceNil returns true if underlying object is nil
func (v *validatorAccountsSyncer) IsInterfaceNil() bool {
	return v == nil
}
