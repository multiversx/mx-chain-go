package factory

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/throttler"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state/syncer"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/statistics"
	"github.com/multiversx/mx-chain-go/update"
	containers "github.com/multiversx/mx-chain-go/update/container"
	"github.com/multiversx/mx-chain-go/update/genesis"
)

// ArgsNewAccountsDBSyncersContainerFactory defines the arguments needed to create accounts DB syncers container
type ArgsNewAccountsDBSyncersContainerFactory struct {
	TrieCacher                storage.Cacher
	RequestHandler            update.RequestHandler
	ShardCoordinator          sharding.Coordinator
	Hasher                    hashing.Hasher
	Marshalizer               marshal.Marshalizer
	TrieStorageManager        common.StorageManager
	TimoutGettingTrieNode     time.Duration
	MaxTrieLevelInMemory      uint
	NumConcurrentTrieSyncers  int
	MaxHardCapForMissingNodes int
	TrieSyncerVersion         int
	CheckNodesOnDisk          bool
	AddressPubKeyConverter    core.PubkeyConverter
	EnableEpochsHandler       common.EnableEpochsHandler
}

type accountDBSyncersContainerFactory struct {
	trieCacher                storage.Cacher
	requestHandler            update.RequestHandler
	container                 update.AccountsDBSyncContainer
	shardCoordinator          sharding.Coordinator
	hasher                    hashing.Hasher
	marshalizer               marshal.Marshalizer
	timeoutGettingTrieNode    time.Duration
	trieStorageManager        common.StorageManager
	maxTrieLevelinMemory      uint
	numConcurrentTrieSyncers  int
	maxHardCapForMissingNodes int
	trieSyncerVersion         int
	checkNodesOnDisk          bool
	addressPubKeyConverter    core.PubkeyConverter
	enableEpochsHandler       common.EnableEpochsHandler
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
	if args.NumConcurrentTrieSyncers < 1 {
		return nil, update.ErrInvalidNumConcurrentTrieSyncers
	}
	if args.MaxHardCapForMissingNodes < 1 {
		return nil, update.ErrInvalidMaxHardCapForMissingNodes
	}
	err := trie.CheckTrieSyncerVersion(args.TrieSyncerVersion)
	if err != nil {
		return nil, err
	}
	if check.IfNil(args.AddressPubKeyConverter) {
		return nil, update.ErrNilPubKeyConverter
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, update.ErrNilEnableEpochsHandler
	}

	t := &accountDBSyncersContainerFactory{
		shardCoordinator:          args.ShardCoordinator,
		trieCacher:                args.TrieCacher,
		requestHandler:            args.RequestHandler,
		hasher:                    args.Hasher,
		marshalizer:               args.Marshalizer,
		trieStorageManager:        args.TrieStorageManager,
		timeoutGettingTrieNode:    args.TimoutGettingTrieNode,
		maxTrieLevelinMemory:      args.MaxTrieLevelInMemory,
		numConcurrentTrieSyncers:  args.NumConcurrentTrieSyncers,
		maxHardCapForMissingNodes: args.MaxHardCapForMissingNodes,
		trieSyncerVersion:         args.TrieSyncerVersion,
		checkNodesOnDisk:          args.CheckNodesOnDisk,
		addressPubKeyConverter:    args.AddressPubKeyConverter,
		enableEpochsHandler:       args.EnableEpochsHandler,
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
	thr, err := throttler.NewNumGoRoutinesThrottler(int32(a.numConcurrentTrieSyncers))
	if err != nil {
		return err
	}

	args := syncer.ArgsNewUserAccountsSyncer{
		ArgsNewBaseAccountsSyncer: syncer.ArgsNewBaseAccountsSyncer{
			Hasher:                            a.hasher,
			Marshalizer:                       a.marshalizer,
			TrieStorageManager:                a.trieStorageManager,
			RequestHandler:                    a.requestHandler,
			Timeout:                           a.timeoutGettingTrieNode,
			Cacher:                            a.trieCacher,
			MaxTrieLevelInMemory:              a.maxTrieLevelinMemory,
			MaxHardCapForMissingNodes:         a.maxHardCapForMissingNodes,
			TrieSyncerVersion:                 a.trieSyncerVersion,
			CheckNodesOnDisk:                  a.checkNodesOnDisk,
			UserAccountsSyncStatisticsHandler: statistics.NewTrieSyncStatistics(),
			AppStatusHandler:                  disabled.NewAppStatusHandler(),
			EnableEpochsHandler:               a.enableEpochsHandler,
		},
		ShardId:                shardId,
		Throttler:              thr,
		AddressPubKeyConverter: a.addressPubKeyConverter,
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
			Hasher:                            a.hasher,
			Marshalizer:                       a.marshalizer,
			TrieStorageManager:                a.trieStorageManager,
			RequestHandler:                    a.requestHandler,
			Timeout:                           a.timeoutGettingTrieNode,
			Cacher:                            a.trieCacher,
			MaxTrieLevelInMemory:              a.maxTrieLevelinMemory,
			MaxHardCapForMissingNodes:         a.maxHardCapForMissingNodes,
			TrieSyncerVersion:                 a.trieSyncerVersion,
			CheckNodesOnDisk:                  a.checkNodesOnDisk,
			UserAccountsSyncStatisticsHandler: statistics.NewTrieSyncStatistics(),
			AppStatusHandler:                  disabled.NewAppStatusHandler(),
			EnableEpochsHandler:               a.enableEpochsHandler,
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
