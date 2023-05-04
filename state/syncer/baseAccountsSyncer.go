package syncer

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/trie"
)

type baseAccountsSyncer struct {
	hasher                            hashing.Hasher
	marshalizer                       marshal.Marshalizer
	dataTries                         map[string]struct{}
	mutex                             sync.Mutex
	trieStorageManager                common.StorageManager
	requestHandler                    trie.RequestHandler
	timeoutHandler                    trie.TimeoutHandler
	shardId                           uint32
	cacher                            storage.Cacher
	rootHash                          []byte
	maxTrieLevelInMemory              uint
	name                              string
	maxHardCapForMissingNodes         int
	checkNodesOnDisk                  bool
	userAccountsSyncStatisticsHandler common.SizeSyncStatisticsHandler
	appStatusHandler                  core.AppStatusHandler

	trieSyncerVersion int
	numTriesSynced    int32
	numMaxTries       int32
}

const timeBetweenStatisticsPrints = time.Second * 2

// ArgsNewBaseAccountsSyncer defines the arguments needed for the new account syncer
type ArgsNewBaseAccountsSyncer struct {
	Hasher                            hashing.Hasher
	Marshalizer                       marshal.Marshalizer
	TrieStorageManager                common.StorageManager
	RequestHandler                    trie.RequestHandler
	Timeout                           time.Duration
	Cacher                            storage.Cacher
	UserAccountsSyncStatisticsHandler common.SizeSyncStatisticsHandler
	AppStatusHandler                  core.AppStatusHandler
	MaxTrieLevelInMemory              uint
	MaxHardCapForMissingNodes         int
	TrieSyncerVersion                 int
	CheckNodesOnDisk                  bool
}

func checkArgs(args ArgsNewBaseAccountsSyncer) error {
	if check.IfNil(args.Hasher) {
		return state.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return state.ErrNilMarshalizer
	}
	if check.IfNil(args.TrieStorageManager) {
		return state.ErrNilStorageManager
	}
	if check.IfNil(args.RequestHandler) {
		return state.ErrNilRequestHandler
	}
	if check.IfNil(args.Cacher) {
		return state.ErrNilCacher
	}
	if check.IfNil(args.UserAccountsSyncStatisticsHandler) {
		return state.ErrNilSyncStatisticsHandler
	}
	if check.IfNil(args.AppStatusHandler) {
		return state.ErrNilAppStatusHandler
	}
	if args.MaxHardCapForMissingNodes < 1 {
		return state.ErrInvalidMaxHardCapForMissingNodes
	}

	return trie.CheckTrieSyncerVersion(args.TrieSyncerVersion)
}

func (b *baseAccountsSyncer) syncMainTrie(
	rootHash []byte,
	trieTopic string,
	ctx context.Context,
) (common.Trie, error) {
	b.rootHash = rootHash
	atomic.AddInt32(&b.numMaxTries, 1)

	log.Trace("syncing main trie", "roothash", rootHash)
	dataTrie, err := trie.NewTrie(b.trieStorageManager, b.marshalizer, b.hasher, b.maxTrieLevelInMemory)
	if err != nil {
		return nil, err
	}

	b.dataTries[string(rootHash)] = struct{}{}
	arg := trie.ArgTrieSyncer{
		RequestHandler:            b.requestHandler,
		InterceptedNodes:          b.cacher,
		DB:                        b.trieStorageManager,
		Marshalizer:               b.marshalizer,
		Hasher:                    b.hasher,
		ShardId:                   b.shardId,
		Topic:                     trieTopic,
		TrieSyncStatistics:        b.userAccountsSyncStatisticsHandler,
		TimeoutHandler:            b.timeoutHandler,
		MaxHardCapForMissingNodes: b.maxHardCapForMissingNodes,
		CheckNodesOnDisk:          b.checkNodesOnDisk,
	}
	trieSyncer, err := trie.CreateTrieSyncer(arg, b.trieSyncerVersion)
	if err != nil {
		return nil, err
	}

	err = trieSyncer.StartSyncing(rootHash, ctx)
	if err != nil {
		return nil, err
	}

	atomic.AddInt32(&b.numTriesSynced, 1)

	log.Trace("finished syncing main trie", "roothash", rootHash)

	return dataTrie.Recreate(rootHash)
}

func (b *baseAccountsSyncer) printStatisticsAndUpdateMetrics(ctx context.Context) {
	lastDataReceived := uint64(0)
	peakDataReceived := uint64(0)
	startedSync := time.Now()
	for {
		select {
		case <-ctx.Done():
			peakSpeed := convertBytesPerIntervalToSpeed(peakDataReceived, timeBetweenStatisticsPrints)
			finishedSync := time.Now()
			totalSyncDuration := finishedSync.Sub(startedSync)
			averageSpeed := convertBytesPerIntervalToSpeed(b.userAccountsSyncStatisticsHandler.NumBytesReceived(), totalSyncDuration)

			log.Info("finished trie sync",
				"name", b.name,
				"time elapsed", totalSyncDuration.Truncate(time.Second),
				"num processed", b.userAccountsSyncStatisticsHandler.NumProcessed(),
				"num large nodes", b.userAccountsSyncStatisticsHandler.NumLarge(),
				"num missing", b.userAccountsSyncStatisticsHandler.NumMissing(),
				"state data size", core.ConvertBytes(b.userAccountsSyncStatisticsHandler.NumBytesReceived()),
				"total iterations", b.userAccountsSyncStatisticsHandler.NumIterations(),
				"total CPU time", b.userAccountsSyncStatisticsHandler.ProcessingTime(),
				"peak processing speed", peakSpeed,
				"average processing speed", averageSpeed,
			)
			b.userAccountsSyncStatisticsHandler.Reset()
			b.updateMetrics()
			return
		case <-time.After(timeBetweenStatisticsPrints):
			bytesReceivedDelta := b.userAccountsSyncStatisticsHandler.NumBytesReceived() - lastDataReceived
			if b.userAccountsSyncStatisticsHandler.NumBytesReceived() < lastDataReceived {
				bytesReceivedDelta = 0
			}
			lastDataReceived = b.userAccountsSyncStatisticsHandler.NumBytesReceived()

			speed := convertBytesPerIntervalToSpeed(bytesReceivedDelta, timeBetweenStatisticsPrints)
			if peakDataReceived < bytesReceivedDelta {
				peakDataReceived = bytesReceivedDelta
			}

			log.Info("trie sync in progress",
				"name", b.name,
				"time elapsed", time.Since(startedSync).Truncate(time.Second),
				"num tries currently syncing", b.userAccountsSyncStatisticsHandler.NumTries(),
				"num processed", b.userAccountsSyncStatisticsHandler.NumProcessed(),
				"num large nodes", b.userAccountsSyncStatisticsHandler.NumLarge(),
				"num missing", b.userAccountsSyncStatisticsHandler.NumMissing(),
				"num tries synced", fmt.Sprintf("%d/%d", atomic.LoadInt32(&b.numTriesSynced), atomic.LoadInt32(&b.numMaxTries)),
				"intercepted trie nodes cache size", core.ConvertBytes(b.cacher.SizeInBytesContained()),
				"num of intercepted trie nodes", b.cacher.Len(),
				"state data size", core.ConvertBytes(b.userAccountsSyncStatisticsHandler.NumBytesReceived()),
				"iterations", b.userAccountsSyncStatisticsHandler.NumIterations(),
				"CPU time", b.userAccountsSyncStatisticsHandler.ProcessingTime(),
				"processing speed", speed)

			b.updateMetrics()
		}
	}
}

func (b *baseAccountsSyncer) updateMetrics() {
	b.appStatusHandler.SetUInt64Value(common.MetricTrieSyncNumProcessedNodes, uint64(b.userAccountsSyncStatisticsHandler.NumProcessed()))
	b.appStatusHandler.SetUInt64Value(common.MetricTrieSyncNumReceivedBytes, b.userAccountsSyncStatisticsHandler.NumBytesReceived())
}

func convertBytesPerIntervalToSpeed(bytes uint64, interval time.Duration) string {
	if interval < time.Millisecond {
		// con not compute precisely, highly likely to get an overflow
		return "N/A"
	}

	bytesReceivedPerSec := float64(bytes) / interval.Seconds()
	uint64Val := uint64(bytesReceivedPerSec)

	return fmt.Sprintf("%s/s", core.ConvertBytes(uint64Val))
}

// Deprecated: GetSyncedTries returns the synced map of data trie. This is likely to case OOM exceptions
// TODO remove this function after fixing the hardfork sync state mechanism
func (b *baseAccountsSyncer) GetSyncedTries() map[string]common.Trie {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	dataTrie, err := trie.NewTrie(b.trieStorageManager, b.marshalizer, b.hasher, b.maxTrieLevelInMemory)
	if err != nil {
		log.Warn("error creating a new trie in baseAccountsSyncer.GetSyncedTries", "error", err)
		return make(map[string]common.Trie)
	}

	var recreatedTrie common.Trie
	clonedMap := make(map[string]common.Trie, len(b.dataTries))
	for key := range b.dataTries {
		recreatedTrie, err = dataTrie.Recreate([]byte(key))
		if err != nil {
			log.Warn("error recreating trie in baseAccountsSyncer.GetSyncedTries",
				"roothash", []byte(key), "error", err)
			continue
		}

		clonedMap[key] = recreatedTrie
	}

	return clonedMap
}

// IsInterfaceNil returns true if underlying object is nil
func (b *baseAccountsSyncer) IsInterfaceNil() bool {
	return b == nil
}
