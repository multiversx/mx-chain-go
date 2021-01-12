package factory

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/throttler"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/syncer"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
	containers "github.com/ElrondNetwork/elrond-go/update/container"
	"github.com/ElrondNetwork/elrond-go/update/genesis"
)

const numConcurrentTrieSyncers = 50

// ArgsNewAccountsDBSyncersContainerFactory defines the arguments needed to create accounts DB syncers container
type ArgsNewAccountsDBSyncersContainerFactory struct {
	TrieCacher            storage.Cacher
	RequestHandler        update.RequestHandler
	ShardCoordinator      sharding.Coordinator
	Hasher                hashing.Hasher
	Marshalizer           marshal.Marshalizer
	TrieStorageManager    data.StorageManager
	TimoutGettingTrieNode time.Duration
	MaxTrieLevelInMemory  uint
}

type accountDBSyncersContainerFactory struct {
	trieCacher             storage.Cacher
	requestHandler         update.RequestHandler
	container              update.AccountsDBSyncContainer
	shardCoordinator       sharding.Coordinator
	hasher                 hashing.Hasher
	marshalizer            marshal.Marshalizer
	timeoutGettingTrieNode time.Duration
	trieStorageManager     data.StorageManager
	maxTrieLevelinMemory   uint
}

// NewAccountsDBSContainerFactory creates a factory for trie syncers container
func NewAccountsDBSContainerFactory(args ArgsNewAccountsDBSyncersContainerFactory) (*accountDBSyncersContainerFactory, error) {
	if check.IfNil(args.RequestHandler) {
		return nil, update.ErrNilRequestHandler
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}
	if check.IfNil(args.TrieCacher) {
		return nil, update.ErrNilCacher
	}
	if check.IfNil(args.Hasher) {
		return nil, update.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, update.ErrNilMarshalizer
	}
	if check.IfNil(args.TrieStorageManager) {
		return nil, update.ErrNilStorageManager
	}

	t := &accountDBSyncersContainerFactory{
		shardCoordinator:       args.ShardCoordinator,
		trieCacher:             args.TrieCacher,
		requestHandler:         args.RequestHandler,
		hasher:                 args.Hasher,
		marshalizer:            args.Marshalizer,
		trieStorageManager:     args.TrieStorageManager,
		timeoutGettingTrieNode: args.TimoutGettingTrieNode,
		maxTrieLevelinMemory:   args.MaxTrieLevelInMemory,
	}

	return t, nil
}

// Create creates all the needed syncers and returns the container
func (a *accountDBSyncersContainerFactory) Create() (update.AccountsDBSyncContainer, error) {
	a.container = containers.NewAccountsDBSyncersContainer()

	for i := uint32(0); i < a.shardCoordinator.NumberOfShards(); i++ {
		err := a.createUserAccountsSyncer(i)
		if err != nil {
			return nil, err
		}
	}

	err := a.createUserAccountsSyncer(core.MetachainShardId)
	if err != nil {
		return nil, err
	}

	err = a.createValidatorAccountsSyncer(core.MetachainShardId)
	if err != nil {
		return nil, err
	}

	return a.container, nil
}

func (a *accountDBSyncersContainerFactory) createUserAccountsSyncer(shardId uint32) error {
	thr, err := throttler.NewNumGoRoutinesThrottler(numConcurrentTrieSyncers)
	if err != nil {
		return err
	}

	args := syncer.ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: syncer.ArgsNewBaseAccountsSyncer{
			Hasher:               a.hasher,
			Marshalizer:          a.marshalizer,
			TrieStorageManager:   a.trieStorageManager,
			RequestHandler:       a.requestHandler,
			Timeout:              a.timeoutGettingTrieNode,
			Cacher:               a.trieCacher,
			MaxTrieLevelInMemory: a.maxTrieLevelinMemory,
		},
		ShardId:   shardId,
		Throttler: thr,
	}
	accountSyncer, err := syncer.NewUserAccountsSyncer(args)
	if err != nil {
		return err
	}
	trieId := genesis.CreateTrieIdentifier(shardId, genesis.UserAccount)

	return a.container.Add(trieId, accountSyncer)
}

func (a *accountDBSyncersContainerFactory) createValidatorAccountsSyncer(shardId uint32) error {
	args := syncer.ArgsNewValidatorAccountsSyncer{
		ArgsNewBaseAccountsSyncer: syncer.ArgsNewBaseAccountsSyncer{
			Hasher:               a.hasher,
			Marshalizer:          a.marshalizer,
			TrieStorageManager:   a.trieStorageManager,
			RequestHandler:       a.requestHandler,
			Timeout:              a.timeoutGettingTrieNode,
			Cacher:               a.trieCacher,
			MaxTrieLevelInMemory: a.maxTrieLevelinMemory,
		},
	}
	accountSyncer, err := syncer.NewValidatorAccountsSyncer(args)
	if err != nil {
		return err
	}
	trieId := genesis.CreateTrieIdentifier(shardId, genesis.ValidatorAccount)

	return a.container.Add(trieId, accountSyncer)
}

// IsInterfaceNil returns true if the underlying object is nil
func (a *accountDBSyncersContainerFactory) IsInterfaceNil() bool {
	return a == nil
}
