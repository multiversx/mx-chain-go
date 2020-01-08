package genesis

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
)

type syncState struct {
	shardCoordinator sharding.Coordinator
	trieSyncers      update.TrieSyncContainer
	metaBlockCache   storage.Cacher
	metaBlockStorage storage.Storer
	epochHandler     update.EpochHandler
	pruningStorer    update.HistoryStorer
	accountsAdapters update.AccountsHandlerContainer

	tries           map[uint32]data.Trie
	lastSyncedEpoch uint32
}

// Arguments for the NewSync
type ArgsNewSyncState struct {
	ShardCoordinator sharding.Coordinator
	TrieSyncers      update.TrieSyncContainer
	EpochHandler     update.EpochHandler
}

// NewSyncState creates a complete syncer which saves the state of the blockchain with pending values as well
func NewSyncState(args ArgsNewSyncState) (*syncState, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, data.ErrNilShardCoordinator
	}
	if check.IfNil(args.TrieSyncers) {
		return nil, dataRetriever.ErrNilResolverContainer
	}

	ss := &syncState{
		shardCoordinator: args.ShardCoordinator,
		trieSyncers:      args.TrieSyncers,
		lastSyncedEpoch:  0,
	}

	return ss, nil
}

// SyncAllState gets an epoch number and will sync the complete data for that epoch start metablock
func (ss *syncState) SyncAllState(epoch uint32) error {
	if epoch == ss.lastSyncedEpoch {
		return nil
	}

	meta, err := ss.getEpochStartMetaHeader(epoch)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(len(meta.EpochStart.LastFinalizedHeaders) + 1)

	go ss.syncMeta(meta, &wg)

	for _, shardData := range meta.EpochStart.LastFinalizedHeaders {
		go ss.syncShard(shardData, &wg)
	}

	wg.Wait()

	ss.lastSyncedEpoch = epoch

	return nil
}

func CreateAccountAdapterIdentifier(shID uint32, accountType factory.Type) string {
	return fmt.Sprint("%d %d", shID, accountType)
}

func (ss *syncState) syncShard(shardData block.EpochStartShardData, wg *sync.WaitGroup) {
	accAdapterIdentifier := CreateAccountAdapterIdentifier(shardData.ShardId, factory.UserAccount)

	accounts, err := ss.accountsAdapters.Get(accAdapterIdentifier)
	if err != nil {
		accounts, err := state.NewAccountsDB()
		if err != nil {
			// critical error - maybe restart node
			return
		}
	}

	accounts.RecreateTrie()

	rootHash, err := accounts.RootHash()
	if err != nil {
		accounts, err := state.NewAccountsDB()
		if err != nil {
			return
		}
	}

	accounts.RecreateTrie()

	wg.Done()
}

func (ss *syncState) syncMeta(meta *block.MetaBlock, wg *sync.WaitGroup) {
	wg.Done()
}

func (ss *syncState) getEpochStartMetaHeader(epoch uint32) (*block.MetaBlock, error) {
	return nil, nil
}

// IsInterfaceNil returns if underlying objects in nil
func (ss *syncState) IsInterfaceNil() bool {
	return ss == nil
}
