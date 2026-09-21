package syncer

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/statistics"
)

var _ state.AccountsDBSyncer = (*validatorAccountsSyncer)(nil)

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
		userAccountsSyncStatisticsHandler: statistics.NewTrieSyncStatistics(),
		appStatusHandler:                  args.AppStatusHandler,
		enableEpochsHandler:               args.EnableEpochsHandler,
	}

	u := &validatorAccountsSyncer{
		baseAccountsSyncer: b,
	}

	return u, nil
}

// SyncAccounts will launch the syncing method to gather all the data needed for validatorAccounts - it is a blocking method
// TODO: handle trie storage statistics here
func (v *validatorAccountsSyncer) SyncAccounts(rootHash []byte, storageMarker common.StorageMarker) error {
	return v.syncAccounts(rootHash, storageMarker, v.checkNodesOnDisk, v.requestHandler)
}

// SyncAccountsWithDiskCheck completes local state and requests missing nodes for the given epoch.
func (v *validatorAccountsSyncer) SyncAccountsWithDiskCheck(rootHash []byte, storageMarker common.StorageMarker, epoch uint32) error {
	requestHandler, err := newRecoveryTrieRequestHandler(v.requestHandler, epoch)
	if err != nil {
		return err
	}
	return v.syncAccounts(rootHash, storageMarker, true, requestHandler)
}

func (v *validatorAccountsSyncer) syncAccounts(
	rootHash []byte,
	storageMarker common.StorageMarker,
	checkNodesOnDisk bool,
	requestHandler trie.RequestHandler,
) error {
	if check.IfNil(storageMarker) {
		return ErrNilStorageMarker
	}

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
		checkNodesOnDisk,
		requestHandler,
	)
	if err != nil {
		return err
	}

	storageMarker.MarkStorerAsSyncedAndActive(v.trieStorageManager)

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (v *validatorAccountsSyncer) IsInterfaceNil() bool {
	return v == nil
}
