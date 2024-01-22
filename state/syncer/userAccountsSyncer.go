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
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/trie"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ state.AccountsDBSyncer = (*userAccountsSyncer)(nil)

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
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, ErrNilEnableEpochsHandler
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
		maxTrieLevelInMemory:              args.MaxTrieLevelInMemory,
		name:                              fmt.Sprintf("user accounts for shard %s", core.GetShardIDString(args.ShardId)),
		maxHardCapForMissingNodes:         args.MaxHardCapForMissingNodes,
		trieSyncerVersion:                 args.TrieSyncerVersion,
		checkNodesOnDisk:                  args.CheckNodesOnDisk,
		userAccountsSyncStatisticsHandler: args.UserAccountsSyncStatisticsHandler,
		appStatusHandler:                  args.AppStatusHandler,
		enableEpochsHandler:               args.EnableEpochsHandler,
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
func (u *userAccountsSyncer) SyncAccounts(rootHash []byte, storageMarker common.StorageMarker) error {
	if check.IfNil(storageMarker) {
		return ErrNilStorageMarker
	}

	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.timeoutHandler.ResetWatchdog()

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		u.cacher.Clear()
		cancel()
	}()

	go u.printStatisticsAndUpdateMetrics(ctx)

	leavesChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder, common.TrieLeavesChannelSyncCapacity),
		ErrChan:    errChan.NewErrChanWrapper(),
	}

	wgSyncMainTrie := &sync.WaitGroup{}
	wgSyncMainTrie.Add(1)

	go func() {
		err := u.syncMainTrie(rootHash, factory.AccountTrieNodesTopic, ctx, leavesChannels.LeavesChan)
		if err != nil {
			leavesChannels.ErrChan.WriteInChanNonBlocking(err)
		}

		common.CloseKeyValueHolderChan(leavesChannels.LeavesChan)

		wgSyncMainTrie.Done()
	}()

	err := u.syncAccountDataTries(leavesChannels, ctx)
	if err != nil {
		return err
	}

	wgSyncMainTrie.Wait()

	err = leavesChannels.ErrChan.ReadFromChanNonBlocking()
	if err != nil {
		return err
	}

	storageMarker.MarkStorerAsSyncedAndActive(u.trieStorageManager)

	log.Debug("main trie and data tries synced", "main trie root hash", rootHash, "num data tries", len(u.dataTries))

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

	trieSyncer, err := u.createAndStartSyncer(ctx, rootHash, u.checkNodesOnDisk)
	if err != nil {
		return err
	}

	u.updateDataTrieStatistics(trieSyncer, address)

	return nil
}

func (u *userAccountsSyncer) createAndStartSyncer(
	ctx context.Context,
	hash []byte,
	checkNodesOnDisk bool,
) (trie.TrieSyncer, error) {
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
		CheckNodesOnDisk:          checkNodesOnDisk,
		LeavesChan:                nil, // not used for data tries
	}
	trieSyncer, err := trie.CreateTrieSyncer(arg, u.trieSyncerVersion)
	if err != nil {
		return nil, err
	}

	err = trieSyncer.StartSyncing(hash, ctx)
	if err != nil {
		return nil, err
	}

	return trieSyncer, nil
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
	leavesChannels *common.TrieIteratorChannels,
	ctx context.Context,
) error {
	if leavesChannels == nil {
		return trie.ErrNilTrieIteratorChannels
	}

	defer u.printDataTrieStatistics()

	wg := sync.WaitGroup{}
	for leaf := range leavesChannels.LeavesChan {
		u.resetTimeoutHandlerWatchdog()

		accountData := &accounts.UserAccountData{}
		err := u.marshalizer.Unmarshal(accountData, leaf.Value())
		if err != nil {
			log.Trace("this must be a leaf with code", "leaf key", leaf.Key(), "err", err)
			continue
		}

		if len(accountData.CodeHash) != 0 && leaf.Version() == core.WithoutCodeLeaf {
			u.syncAccountCode(accountData.CodeHash, &wg, ctx, leavesChannels.ErrChan)
		}

		if common.IsEmptyTrie(accountData.RootHash) {
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
			err := u.syncDataTrie(trieRootHash, address, ctx)
			if err != nil {
				leavesChannels.ErrChan.WriteInChanNonBlocking(err)
			}
			atomic.AddInt32(&u.numTriesSynced, 1)
			log.Trace("finished sync data trie", "roothash", trieRootHash)
			wg.Done()
		}(accountData.RootHash, accountData.Address)
	}

	wg.Wait()

	return nil
}

func (u *userAccountsSyncer) syncAccountCode(codeHash []byte, wg *sync.WaitGroup, ctx context.Context, errChan common.BufferedErrChan) {
	u.throttler.StartProcessing()
	wg.Add(1)

	waitTimeBetweenChecks := time.Millisecond * 100

	go func(codeHash []byte) {
		defer u.throttler.EndProcessing()

		for {
			codeData, ok := u.cacher.Get(codeHash)
			if ok {
				code, ok := codeData.([]byte)
				if !ok {
					errChan.WriteInChanNonBlocking(errors.ErrWrongTypeAssertion)
					break
				}

				err := u.trieStorageManager.Put(codeHash, code)
				if err == nil {
					break
				}

				log.Trace("failed to put code into storage", "codeHash", codeHash, "error", err.Error())
			}

			u.requestHandler.RequestTrieNodes(u.shardId, [][]byte{codeHash}, factory.AccountTrieNodesTopic)

			log.Trace("requested trie node", "codeHash", codeHash)

			select {
			case <-time.After(waitTimeBetweenChecks):
				continue
			case <-ctx.Done():
				break
			}
		}

		wg.Done()
	}(codeHash)
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

// MissingDataTrieNodeFound is called whenever a missing data trie node is found.
// This will trigger the sync process for the whole sub trie, starting from the given hash.
func (u *userAccountsSyncer) MissingDataTrieNodeFound(hash []byte) {
	defer u.printDataTrieStatistics()

	u.timeoutHandler.ResetWatchdog()

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		u.cacher.Clear()
		cancel()
	}()

	trieSyncer, err := u.createAndStartSyncer(ctx, hash, true)
	if err != nil {
		log.Error("cannot sync trie", "err", err, "hash", hash)
		return
	}

	u.updateDataTrieStatistics(trieSyncer, hash)

	log.Debug("finished sync data trie", "hash", hash)
}

// IsInterfaceNil returns true if there is no value under the interface
func (u *userAccountsSyncer) IsInterfaceNil() bool {
	return u == nil
}
