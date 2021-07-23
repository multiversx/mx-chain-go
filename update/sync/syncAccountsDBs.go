package sync

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/ElrondNetwork/elrond-go/update/genesis"
)

var _ update.EpochStartTriesSyncHandler = (*syncAccountsDBs)(nil)

type syncAccountsDBs struct {
	tries              *concurrentTriesMap
	accountsBDsSyncers update.AccountsDBSyncContainer
	activeAccountsDBs  map[state.AccountsDbIdentifier]state.AccountsAdapter
	mutSynced          sync.Mutex
	synced             bool
}

// ArgsNewSyncAccountsDBsHandler is the argument structured to create a sync tries handler
type ArgsNewSyncAccountsDBsHandler struct {
	AccountsDBsSyncers update.AccountsDBSyncContainer
	ActiveAccountsDBs  map[state.AccountsDbIdentifier]state.AccountsAdapter
}

// NewSyncAccountsDBsHandler creates a new syncAccountsDBs
func NewSyncAccountsDBsHandler(args ArgsNewSyncAccountsDBsHandler) (*syncAccountsDBs, error) {
	if check.IfNil(args.AccountsDBsSyncers) {
		return nil, update.ErrNilAccountsDBSyncContainer
	}

	st := &syncAccountsDBs{
		tries:              newConcurrentTriesMap(),
		accountsBDsSyncers: args.AccountsDBsSyncers,
		activeAccountsDBs:  make(map[state.AccountsDbIdentifier]state.AccountsAdapter),
		synced:             false,
		mutSynced:          sync.Mutex{},
	}
	for key, value := range args.ActiveAccountsDBs {
		st.activeAccountsDBs[key] = value
	}

	return st, nil
}

// SyncTriesFrom syncs all the state tries from an epoch start metachain
func (st *syncAccountsDBs) SyncTriesFrom(meta *block.MetaBlock) error {
	if !meta.IsStartOfEpochBlock() && meta.Nonce > 0 {
		return update.ErrNotEpochStartBlock
	}

	var errFound error
	mutErr := sync.Mutex{}

	st.mutSynced.Lock()
	st.synced = false
	st.mutSynced.Unlock()

	wg := sync.WaitGroup{}
	wg.Add(1 + len(meta.EpochStart.LastFinalizedHeaders))

	// TODO: might think of a way to stop waiting at a signal
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

	<-chDone

	if errFound != nil {
		return errFound
	}

	st.mutSynced.Lock()
	st.synced = true
	st.mutSynced.Unlock()

	return nil
}

func (st *syncAccountsDBs) syncMeta(meta *block.MetaBlock) error {
	err := st.syncAccountsOfType(genesis.UserAccount, state.UserAccountsState, core.MetachainShardId, meta.RootHash)
	if err != nil {
		return fmt.Errorf("%w UserAccount, shard: meta", err)
	}

	err = st.syncAccountsOfType(genesis.ValidatorAccount, state.PeerAccountsState, core.MetachainShardId, meta.ValidatorStatsRootHash)
	if err != nil {
		return fmt.Errorf("%w ValidatorAccount, shard: meta", err)
	}

	return nil
}

func (st *syncAccountsDBs) syncShard(shardData block.EpochStartShardData) error {
	err := st.syncAccountsOfType(genesis.UserAccount, state.UserAccountsState, shardData.ShardID, shardData.RootHash)
	if err != nil {
		return fmt.Errorf("%w UserAccount, shard: %d", err, shardData.ShardID)
	}
	return nil
}

func (st *syncAccountsDBs) syncAccountsOfType(accountType genesis.Type, trieID state.AccountsDbIdentifier, shardId uint32, rootHash []byte) error {
	accAdapterIdentifier := genesis.CreateTrieIdentifier(shardId, accountType)

	log.Debug("syncing accounts",
		"type", accAdapterIdentifier,
		"shard ID", shardId,
		"root hash", rootHash,
	)

	success := st.tryRecreateTrie(shardId, accAdapterIdentifier, trieID, rootHash)
	if success {
		return nil
	}

	accountsDBSyncer, err := st.accountsBDsSyncers.Get(accAdapterIdentifier)
	if err != nil {
		return err
	}

	err = accountsDBSyncer.SyncAccounts(rootHash)
	if err != nil {
		// TODO: critical error - should not happen - maybe recreate trie syncer here
		return err
	}

	st.setTries(shardId, accAdapterIdentifier, rootHash, accountsDBSyncer.GetSyncedTries())

	return nil
}

func (st *syncAccountsDBs) setTries(shId uint32, initialID string, rootHash []byte, tries map[string]temporary.Trie) {
	for hash, currentTrie := range tries {
		if bytes.Equal(rootHash, []byte(hash)) {
			st.tries.setTrie(initialID, currentTrie)
			continue
		}

		dataTrieIdentifier := genesis.CreateTrieIdentifier(shId, genesis.DataTrie)
		identifier := genesis.AddRootHashToIdentifier(dataTrieIdentifier, hash)
		st.tries.setTrie(identifier, currentTrie)
	}
}

func (st *syncAccountsDBs) tryRecreateTrie(shardId uint32, id string, trieID state.AccountsDbIdentifier, rootHash []byte) bool {
	savedTrie, ok := st.tries.getTrie(id)
	if ok {
		currHash, err := savedTrie.RootHash()
		if err == nil && bytes.Equal(currHash, rootHash) {
			return true
		}
	}

	activeTrie := st.activeAccountsDBs[trieID]
	if check.IfNil(activeTrie) {
		return false
	}

	tries, err := activeTrie.RecreateAllTries(rootHash)
	if err != nil {
		return false
	}

	for _, recreatedTrie := range tries {
		err = recreatedTrie.Commit()
		if err != nil {
			return false
		}
	}

	st.setTries(shardId, id, rootHash, tries)

	return true
}

// GetTries returns the synced tries
func (st *syncAccountsDBs) GetTries() (map[string]temporary.Trie, error) {
	st.mutSynced.Lock()
	defer st.mutSynced.Unlock()

	if !st.synced {
		return nil, update.ErrNotSynced
	}

	return st.tries.getTries(), nil
}

// IsInterfaceNil returns nil if underlying object is nil
func (st *syncAccountsDBs) IsInterfaceNil() bool {
	return st == nil
}
