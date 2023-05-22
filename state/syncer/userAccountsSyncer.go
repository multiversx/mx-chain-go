package syncer

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ epochStart.AccountsDBSyncer = (*userAccountsSyncer)(nil)

var log = logger.GetOrCreate("syncer")

const timeBetweenRetries = 100 * time.Millisecond
const smallTrieThreshold = 1 * 1024 * 1024 // 1MB

type stats struct {
	address      []byte
	numBytes     uint64
	numTrieNodes uint64
	numLeaves    uint64
	duration     time.Duration
}

type userAccountsSyncer struct {
	*baseAccountsSyncer
	throttler      data.GoRoutineThrottler
	syncerMutex    sync.Mutex
	pubkeyCoverter core.PubkeyConverter

	mutStatistics sync.RWMutex
	largeTries    []*stats
	numSmallTries int
}

// ArgsNewUserAccountsSyncer defines the arguments needed for the new account syncer
type ArgsNewUserAccountsSyncer struct {
	ArgsNewBaseAccountsSyncer
	ShardId                uint32
	Throttler              data.GoRoutineThrottler
	AddressPubKeyConverter core.PubkeyConverter
}

// NewUserAccountsSyncer creates a user account syncer
func NewUserAccountsSyncer(args ArgsNewUserAccountsSyncer) (*userAccountsSyncer, error) {
	err := checkArgs(args.ArgsNewBaseAccountsSyncer)
	if err != nil {
		return nil, err
	}

	if check.IfNil(args.Throttler) {
		return nil, data.ErrNilThrottler
	}
	if check.IfNil(args.AddressPubKeyConverter) {
		return nil, ErrNilPubkeyConverter
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
		shardId:                           args.ShardId,
		cacher:                            args.Cacher,
		rootHash:                          nil,
		maxTrieLevelInMemory:              args.MaxTrieLevelInMemory,
		name:                              fmt.Sprintf("user accounts for shard %s", core.GetShardIDString(args.ShardId)),
		maxHardCapForMissingNodes:         args.MaxHardCapForMissingNodes,
		trieSyncerVersion:                 args.TrieSyncerVersion,
		checkNodesOnDisk:                  args.CheckNodesOnDisk,
		storageMarker:                     args.StorageMarker,
		userAccountsSyncStatisticsHandler: args.UserAccountsSyncStatisticsHandler,
		appStatusHandler:                  args.AppStatusHandler,
	}

	u := &userAccountsSyncer{
		baseAccountsSyncer: b,
		throttler:          args.Throttler,
		pubkeyCoverter:     args.AddressPubKeyConverter,
		largeTries:         make([]*stats, 0),
	}

	return u, nil
}

// SyncAccounts will launch the syncing method to gather all the data needed for userAccounts - it is a blocking method
func (u *userAccountsSyncer) SyncAccounts(rootHash []byte) error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.timeoutHandler.ResetWatchdog()

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		u.cacher.Clear()
		cancel()
	}()

	go u.printStatisticsAndUpdateMetrics(ctx)

	mainTrie, err := u.syncMainTrie(rootHash, factory.AccountTrieNodesTopic, ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = mainTrie.Close()
	}()

	log.Debug("main trie synced, starting to sync data tries", "num data tries", len(u.dataTries))

	err = u.syncAccountDataTries(mainTrie, ctx)
	if err != nil {
		return err
	}

	u.storageMarker.MarkStorerAsSyncedAndActive(mainTrie.GetStorageManager())

	return nil
}

func (u *userAccountsSyncer) syncDataTrie(rootHash []byte, address []byte, ctx context.Context) error {
	u.syncerMutex.Lock()
	_, ok := u.dataTries[string(rootHash)]
	if ok {
		u.syncerMutex.Unlock()
		return nil
	}

	u.dataTries[string(rootHash)] = struct{}{}
	u.syncerMutex.Unlock()

	arg := trie.ArgTrieSyncer{
		RequestHandler:            u.requestHandler,
		InterceptedNodes:          u.cacher,
		DB:                        u.trieStorageManager,
		Marshalizer:               u.marshalizer,
		Hasher:                    u.hasher,
		ShardId:                   u.shardId,
		Topic:                     factory.AccountTrieNodesTopic,
		TrieSyncStatistics:        u.userAccountsSyncStatisticsHandler,
		TimeoutHandler:            u.timeoutHandler,
		MaxHardCapForMissingNodes: u.maxHardCapForMissingNodes,
		CheckNodesOnDisk:          u.checkNodesOnDisk,
	}
	trieSyncer, err := trie.CreateTrieSyncer(arg, u.trieSyncerVersion)
	if err != nil {

		return err
	}

	err = trieSyncer.StartSyncing(rootHash, ctx)
	if err != nil {
		return err
	}

	u.updateDataTrieStatistics(trieSyncer, address)

	return nil
}

func (u *userAccountsSyncer) updateDataTrieStatistics(trieSyncer trie.TrieSyncer, address []byte) {
	isSmallTrie := trieSyncer.NumBytes() < smallTrieThreshold

	u.mutStatistics.Lock()
	defer u.mutStatistics.Unlock()

	if isSmallTrie {
		u.numSmallTries++
		return
	}

	trieStats := &stats{
		address:      address,
		numBytes:     trieSyncer.NumBytes(),
		numTrieNodes: trieSyncer.NumTrieNodes(),
		numLeaves:    trieSyncer.NumLeaves(),
		duration:     trieSyncer.Duration(),
	}
	u.largeTries = append(u.largeTries, trieStats)
}

func (u *userAccountsSyncer) syncAccountDataTries(
	mainTrie common.Trie,
	ctx context.Context,
) error {
	defer u.printDataTrieStatistics()

	mainRootHash, err := mainTrie.RootHash()
	if err != nil {
		return err
	}

	leavesChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelDefaultCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	err = mainTrie.GetAllLeavesOnChannel(leavesChannels, context.Background(), mainRootHash, keyBuilder.NewDisabledKeyBuilder())
	if err != nil {
		return err
	}

	var errFound error
	errMutex := sync.Mutex{}
	wg := sync.WaitGroup{}

	for leaf := range leavesChannels.LeavesChan {
		u.resetTimeoutHandlerWatchdog()

		account := state.NewEmptyUserAccount()
		err = u.marshalizer.Unmarshal(account, leaf.Value())
		if err != nil {
			log.Trace("this must be a leaf with code", "err", err)
			continue
		}

		if len(account.RootHash) == 0 {
			continue
		}

		err = u.checkGoRoutinesThrottler(ctx)
		if err != nil {
			return err
		}

		u.throttler.StartProcessing()
		wg.Add(1)
		atomic.AddInt32(&u.numMaxTries, 1)

		go func(trieRootHash []byte, address []byte) {
			defer u.throttler.EndProcessing()

			log.Trace("sync data trie", "roothash", trieRootHash)
			newErr := u.syncDataTrie(trieRootHash, address, ctx)
			if newErr != nil {
				errMutex.Lock()
				errFound = newErr
				errMutex.Unlock()
			}
			atomic.AddInt32(&u.numTriesSynced, 1)
			log.Trace("finished sync data trie", "roothash", trieRootHash)
			wg.Done()
		}(account.RootHash, account.Address)
	}

	wg.Wait()

	err = leavesChannels.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return err
	}

	return errFound
}

func (u *userAccountsSyncer) printDataTrieStatistics() {
	u.mutStatistics.Lock()
	defer u.mutStatistics.Unlock()

	log.Debug("user accounts tries sync has finished",
		"num small data tries", u.numSmallTries, "threshold", core.ConvertBytes(uint64(smallTrieThreshold)))

	sort.Slice(u.largeTries, func(i, j int) bool {
		trieI := u.largeTries[i]
		trieJ := u.largeTries[j]

		return trieI.numBytes >= trieJ.numBytes
	})

	for _, trieStat := range u.largeTries {
		address := u.pubkeyCoverter.SilentEncode(trieStat.address, log)

		log.Debug("datatrie for "+address,
			"num trie nodes", trieStat.numTrieNodes,
			"num leaves", trieStat.numLeaves,
			"size", core.ConvertBytes(trieStat.numBytes),
			"time used to sync the trie", trieStat.duration.Truncate(time.Second),
		)
	}

}

func (u *userAccountsSyncer) checkGoRoutinesThrottler(ctx context.Context) error {
	for {
		if u.throttler.CanProcess() {
			break
		}

		select {
		case <-time.After(timeBetweenRetries):
			continue
		case <-ctx.Done():
			return data.ErrTimeIsOut
		}
	}

	return nil
}

// resetTimeoutHandlerWatchdog this method should be called whenever the syncer is doing something other than
// requesting trie nodes as to prevent the sync process being terminated prematurely.
func (u *userAccountsSyncer) resetTimeoutHandlerWatchdog() {
	u.timeoutHandler.ResetWatchdog()
}

// IsInterfaceNil returns true if there is no value under the interface
func (u *userAccountsSyncer) IsInterfaceNil() bool {
	return u == nil
}
