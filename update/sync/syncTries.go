package sync

import (
	"bytes"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/genesis"
)

type syncTries struct {
	tries       *concurrentTriesMap
	trieSyncers update.TrieSyncContainer
	activeTries state.TriesHolder
	mutSynced   sync.Mutex
	synced      bool
}

// ArgsNewSyncTriesHandler is the argument structed to create a sync tries handler
type ArgsNewSyncTriesHandler struct {
	TrieSyncers update.TrieSyncContainer
	ActiveTries state.TriesHolder
}

// NewSyncTriesHandler creates a new syncTries
func NewSyncTriesHandler(args ArgsNewSyncTriesHandler) (*syncTries, error) {
	if check.IfNil(args.TrieSyncers) {
		return nil, update.ErrNilTrieSyncers
	}
	if check.IfNil(args.ActiveTries) {
		return nil, update.ErrNilActiveTries
	}

	st := &syncTries{
		tries:       newConcurrentTriesMap(),
		trieSyncers: args.TrieSyncers,
		activeTries: args.ActiveTries,
		synced:      false,
		mutSynced:   sync.Mutex{},
	}

	return st, nil
}

// SyncTriesFrom syncs all the state tries from an epoch start metachain
func (st *syncTries) SyncTriesFrom(meta *block.MetaBlock, waitTime time.Duration) error {
	if !meta.IsStartOfEpochBlock() {
		return update.ErrNotEpochStartBlock
	}

	var errFound error
	mutErr := sync.Mutex{}

	st.synced = false
	wg := sync.WaitGroup{}
	wg.Add(1 + len(meta.EpochStart.LastFinalizedHeaders))

	chDone := make(chan bool)
	go func() {
		wg.Wait()
		chDone <- true
	}()

	go func() {
		errMeta := st.syncMeta(meta)
		if errMeta != nil {
			mutErr.Lock()
			errFound = errMeta
			mutErr.Unlock()
		}
		wg.Done()
	}()

	for _, shData := range meta.EpochStart.LastFinalizedHeaders {
		go func(shardData block.EpochStartShardData) {
			err := st.syncShard(shardData)
			if err != nil {
				mutErr.Lock()
				errFound = err
				mutErr.Unlock()
			}
			wg.Done()
		}(shData)
	}

	err := WaitFor(chDone, waitTime)
	if err != nil {
		return err
	}

	if errFound != nil {
		return errFound
	}

	st.mutSynced.Lock()
	st.synced = true
	st.mutSynced.Unlock()

	return nil
}

func (st *syncTries) syncMeta(meta *block.MetaBlock) error {
	err := st.syncTrieOfType(factory.UserAccount, core.MetachainShardId, meta.RootHash)
	if err != nil {
		return err
	}

	err = st.syncTrieOfType(factory.ValidatorAccount, core.MetachainShardId, meta.ValidatorStatsRootHash)
	if err != nil {
		return err
	}

	return nil
}

func (st *syncTries) syncShard(shardData block.EpochStartShardData) error {
	err := st.syncTrieOfType(factory.UserAccount, shardData.ShardId, shardData.RootHash)
	if err != nil {
		return err
	}
	return nil
}

func (st *syncTries) syncTrieOfType(accountType factory.Type, shardId uint32, rootHash []byte) error {
	accAdapterIdentifier := genesis.CreateTrieIdentifier(shardId, accountType)

	success := st.tryRecreateTrie(accAdapterIdentifier, rootHash)
	if success {
		return nil
	}

	trieSyncer, err := st.trieSyncers.Get(accAdapterIdentifier)
	if err != nil {
		return err
	}

	err = trieSyncer.StartSyncing(rootHash)
	if err != nil {
		// critical error - should not happen - maybe recreate trie syncer here
		return err
	}

	st.tries.setTrie(accAdapterIdentifier, trieSyncer.Trie())
	return nil
}

func (st *syncTries) tryRecreateTrie(id string, rootHash []byte) bool {
	savedTrie, ok := st.tries.getTrie(id)
	if ok {
		currHash, err := savedTrie.Root()
		if err == nil && bytes.Equal(currHash, rootHash) {
			return true
		}
	}

	accounts := st.activeTries.Get([]byte(id))
	if check.IfNil(accounts) {
		return false
	}

	trie, err := accounts.Recreate(rootHash)
	if err != nil {
		return false
	}

	err = trie.Commit()
	if err != nil {
		return false
	}

	st.tries.setTrie(id, trie)
	return true
}

// GetTries returns the synced tries
func (st *syncTries) GetTries() (map[string]data.Trie, error) {
	if !st.synced {
		return nil, update.ErrNotSynced
	}

	return st.tries.getTries(), nil
}

// IsInterfaceNil returns nil if underlying object is nil
func (st *syncTries) IsInterfaceNil() bool {
	return st == nil
}
