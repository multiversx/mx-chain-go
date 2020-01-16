package sync

import (
	"bytes"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

type syncTries struct {
	tries       map[string]data.Trie
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
		tries:       make(map[string]data.Trie),
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
		errMeta := st.syncMeta(meta, &wg)
		if errMeta != nil {
			mutErr.Lock()
			errFound = errMeta
			mutErr.Unlock()
		}
		wg.Done()
	}()

	for _, shData := range meta.EpochStart.LastFinalizedHeaders {
		go func(shardData block.EpochStartShardData) {
			err := st.syncShard(shardData, &wg)
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

	if errFound == nil {
		st.mutSynced.Lock()
		st.synced = true
		st.mutSynced.Unlock()
	}

	return errFound
}

func (st *syncTries) syncMeta(meta *block.MetaBlock, wg *sync.WaitGroup) error {
	defer wg.Done()

	err := st.syncTrieOfType(factory.UserAccount, sharding.MetachainShardId, meta.RootHash)
	if err != nil {
		return nil
	}

	err = st.syncTrieOfType(factory.ValidatorAccount, sharding.MetachainShardId, meta.ValidatorStatsRootHash)
	if err != nil {
		return nil
	}

	return nil
}

func (st *syncTries) syncShard(shardData block.EpochStartShardData, wg *sync.WaitGroup) error {
	defer wg.Done()
	err := st.syncTrieOfType(factory.UserAccount, shardData.ShardId, shardData.RootHash)
	if err != nil {
		return err
	}
	return nil
}

func (st *syncTries) syncTrieOfType(accountType factory.Type, shardId uint32, rootHash []byte) error {
	accAdapterIdentifier := update.CreateTrieIdentifier(shardId, accountType)

	success := st.tryRecreateTrie(accAdapterIdentifier, rootHash)
	if success {
		return nil
	}

	trieSyncer, err := st.trieSyncers.Get(accAdapterIdentifier)
	if err != nil {
		// critical error - should not happen - maybe recreate trie syncer here
		return err
	}

	err = trieSyncer.StartSyncing(rootHash)
	if err != nil {
		// critical error - should not happen - maybe recreate trie syncer here
		return err
	}

	st.tries[accAdapterIdentifier] = trieSyncer.Trie()
	return nil
}

func (st *syncTries) tryRecreateTrie(id string, rootHash []byte) bool {
	savedTrie, ok := st.tries[id]
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

	st.tries[id] = trie
	return true
}

// GetTries returns the synced tries
func (st *syncTries) GetTries() (map[string]data.Trie, error) {
	st.mutSynced.Lock()
	defer st.mutSynced.Unlock()

	if !st.synced {
		return nil, update.ErrNotSynced
	}

	return st.tries, nil
}

// IsInterfaceNil returns nil if underlying object is nil
func (st *syncTries) IsInterfaceNil() bool {
	return st == nil
}
