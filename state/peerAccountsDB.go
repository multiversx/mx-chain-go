package state

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/common"
)

// PeerAccountsDB will save and synchronize data from peer processor, plus will synchronize with nodesCoordinator
type PeerAccountsDB struct {
	*AccountsDB
}

// NewPeerAccountsDB creates a new account manager
func NewPeerAccountsDB(args ArgsAccountsDB) (*PeerAccountsDB, error) {
	err := checkArgsAccountsDB(args)
	if err != nil {
		return nil, err
	}

	adb := &PeerAccountsDB{
		&AccountsDB{
			mainTrie:       args.Trie,
			hasher:         args.Hasher,
			marshaller:     args.Marshaller,
			accountFactory: args.AccountFactory,
			entries:        make([]JournalEntry, 0),
			dataTries:      NewDataTriesHolder(),
			mutOp:          sync.RWMutex{},
			loadCodeMeasurements: &loadingMeasurements{
				identifier: "load code",
			},
			storagePruningManager: args.StoragePruningManager,
			processingMode:        args.ProcessingMode,
			lastSnapshot:          &snapshotInfo{},
			processStatusHandler:  args.ProcessStatusHandler,
		},
	}

	return adb, nil
}

// StartSnapshotIfNeeded starts the snapshot if the previous snapshot process was not fully completed
func (adb *PeerAccountsDB) StartSnapshotIfNeeded() error {
	return startSnapshotIfNeeded(adb, adb.trieSyncer, adb.mainTrie.GetStorageManager(), adb.processingMode)
}

// MarkSnapshotDone will mark that the snapshot process has been completed
func (adb *PeerAccountsDB) MarkSnapshotDone() {
	trieStorageManager, epoch, err := adb.getTrieStorageManagerAndLatestEpoch()
	if err != nil {
		log.Error("MarkSnapshotDone error", "err", err.Error())
		return
	}

	err = trieStorageManager.PutInEpochWithoutCache([]byte(common.ActiveDBKey), []byte(common.ActiveDBVal), epoch)
	handleLoggingWhenError("error while putting active DB value into main storer", err)
}

func (adb *PeerAccountsDB) getTrieStorageManagerAndLatestEpoch() (common.StorageManager, uint32, error) {
	trieStorageManager := adb.mainTrie.GetStorageManager()
	epoch, err := trieStorageManager.GetLatestStorageEpoch()
	if err != nil {
		return nil, 0, fmt.Errorf("%w while getting the latest storage epoch", err)
	}

	return trieStorageManager, epoch, nil
}

// SnapshotState triggers the snapshotting process of the state trie
func (adb *PeerAccountsDB) SnapshotState(rootHash []byte) {
	log.Trace("peerAccountsDB.SnapshotState", "root hash", rootHash)
	trieStorageManager, epoch, err := adb.getTrieStorageManagerAndLatestEpoch()
	if err != nil {
		log.Error("SnapshotState error", "err", err.Error())
		return
	}

	if !trieStorageManager.ShouldTakeSnapshot() {
		log.Debug("skipping snapshot for rootHash", "hash", rootHash)
		return
	}

	log.Info("starting snapshot peer trie", "rootHash", rootHash, "epoch", epoch)

	adb.lastSnapshot.rootHash = rootHash
	adb.lastSnapshot.epoch = epoch
	err = trieStorageManager.Put([]byte(lastSnapshotStarted), rootHash)
	if err != nil {
		log.Warn("could not set lastSnapshotStarted", "err", err, "rootHash", rootHash)
	}

	missingNodesChannel := make(chan []byte, missingNodesChannelSize)
	stats := newSnapshotStatistics(0, 1)

	trieStorageManager.EnterPruningBufferingMode()
	stats.NewSnapshotStarted()
	errChan := make(chan error, 1)
	trieStorageManager.TakeSnapshot(rootHash, rootHash, nil, missingNodesChannel, errChan, stats, epoch)

	go adb.syncMissingNodes(missingNodesChannel, stats)

	go adb.markActiveDBAfterSnapshot(stats, missingNodesChannel, errChan, rootHash, "snapshotState peer trie", epoch)

	adb.waitForCompletionIfAppropriate(stats)
}

// SetStateCheckpoint triggers the checkpointing process of the state trie
func (adb *PeerAccountsDB) SetStateCheckpoint(rootHash []byte) {
	log.Trace("peerAccountsDB.SetStateCheckpoint", "root hash", rootHash)
	trieStorageManager := adb.mainTrie.GetStorageManager()

	missingNodesChannel := make(chan []byte, missingNodesChannelSize)
	stats := newSnapshotStatistics(0, 1)

	trieStorageManager.EnterPruningBufferingMode()
	stats.NewSnapshotStarted()
	errChan := make(chan error, 1)
	trieStorageManager.SetCheckpoint(rootHash, rootHash, nil, missingNodesChannel, errChan, stats)

	go adb.syncMissingNodes(missingNodesChannel, stats)

	// TODO decide if we need to take some actions whenever we hit an error that occurred in the checkpoint process
	//  that will be present in the errChan var
	go adb.finishSnapshotOperation(rootHash, stats, missingNodesChannel, "setStateCheckpoint peer trie")

	adb.waitForCompletionIfAppropriate(stats)
}

// RecreateAllTries recreates all the tries from the accounts DB
func (adb *PeerAccountsDB) RecreateAllTries(rootHash []byte) (map[string]common.Trie, error) {
	recreatedTrie, err := adb.mainTrie.Recreate(rootHash)
	if err != nil {
		return nil, err
	}

	allTries := make(map[string]common.Trie)
	allTries[string(rootHash)] = recreatedTrie

	return allTries, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (adb *PeerAccountsDB) IsInterfaceNil() bool {
	return adb == nil
}
