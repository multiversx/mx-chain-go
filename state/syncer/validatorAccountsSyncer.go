package syncer

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/trie/statistics"
)

var _ epochStart.AccountsDBSyncer = (*validatorAccountsSyncer)(nil)

type validatorAccountsSyncer struct {
	*baseAccountsSyncer
}

// ArgsNewValidatorAccountsSyncer defines the arguments needed for the new account syncer
type ArgsNewValidatorAccountsSyncer struct {
	ArgsNewBaseAccountsSyncer
}

// NewValidatorAccountsSyncer creates a validator account syncer
func NewValidatorAccountsSyncer(args ArgsNewValidatorAccountsSyncer) (*validatorAccountsSyncer, error) {
	err := checkArgs(args.ArgsNewBaseAccountsSyncer)
	if err != nil {
		return nil, err
	}

	timeoutHandler, err := common.NewTimeoutHandler(args.Timeout)
	if err != nil {
		return nil, err
	}

	b := &baseAccountsSyncer{
		hasher:                            args.Hasher,
		marshalizer:                       args.Marshalizer,
		dataTries:                         make(map[string]struct{}),
		trieStorageManager:                args.TrieStorageManager,
		requestHandler:                    args.RequestHandler,
		timeoutHandler:                    timeoutHandler,
		shardId:                           core.MetachainShardId,
		cacher:                            args.Cacher,
		maxTrieLevelInMemory:              args.MaxTrieLevelInMemory,
		name:                              "peer accounts",
		maxHardCapForMissingNodes:         args.MaxHardCapForMissingNodes,
		trieSyncerVersion:                 args.TrieSyncerVersion,
		checkNodesOnDisk:                  args.CheckNodesOnDisk,
		storageMarker:                     args.StorageMarker,
		userAccountsSyncStatisticsHandler: statistics.NewTrieSyncStatistics(),
		appStatusHandler:                  args.AppStatusHandler,
	}

	u := &validatorAccountsSyncer{
		baseAccountsSyncer: b,
	}

	return u, nil
}

// SyncAccounts will launch the syncing method to gather all the data needed for validatorAccounts - it is a blocking method
func (v *validatorAccountsSyncer) SyncAccounts(rootHash []byte) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	v.timeoutHandler.ResetWatchdog()

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		v.cacher.Clear()
		cancel()
	}()

	go v.printStatisticsAndUpdateMetrics(ctx)

	err := v.syncMainTrie(
		rootHash,
		factory.ValidatorTrieNodesTopic,
		ctx,
		nil, // not used for validator accounts syncer
	)
	if err != nil {
		return err
	}

	v.storageMarker.MarkStorerAsSyncedAndActive(v.trieStorageManager)

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (v *validatorAccountsSyncer) IsInterfaceNil() bool {
	return v == nil
}
