package syncer

import (
	"context"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/trie/statistics"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process/factory"
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

	b := &baseAccountsSyncer{
		hasher:               args.Hasher,
		marshalizer:          args.Marshalizer,
		trieSyncers:          make(map[string]data.TrieSyncer),
		dataTries:            make(map[string]data.Trie),
		trieStorageManager:   args.TrieStorageManager,
		requestHandler:       args.RequestHandler,
		waitTime:             args.WaitTime,
		shardId:              core.MetachainShardId,
		cacher:               args.Cacher,
		rootHash:             nil,
		maxTrieLevelInMemory: args.MaxTrieLevelInMemory,
		name:                 "peer accounts",
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

	ctx, cancel := context.WithTimeout(context.Background(), v.waitTime)
	defer cancel()

	tss := statistics.NewTrieSyncStatistics()
	go v.printStatistics(tss, ctx)

	return v.syncMainTrie(rootHash, factory.ValidatorTrieNodesTopic, tss, ctx)
}
