package state

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
)

const importDBPeerSnapshotMaxWaitCommits = 5

type importDBPeerMemory interface {
	Rotate()
	SnapshotFinished()
}

type importDBPeerSnapshotter interface {
	SnapshotPeerState([]byte, uint32, common.StorageManager, func()) bool
}

type importDBPeerRoot struct {
	root, hash []byte
	nonce      uint64
	epoch      uint32
	v3         bool
}

type importDBPeerSnapshots struct {
	store       importDBPeerMemory
	snapshotter importDBPeerSnapshotter
	interval    uint64
	requested   atomic.Bool
	// latest is protected by AccountsDB.mutOp, together with lastRootHash.
	latest *importDBPeerRoot
	// Only imported headers drive the lifecycle. Execution can keep committing.
	processMu   sync.Mutex
	candidate   *importDBPeerRoot
	lastHeader  []byte
	lastRound   uint64
	initialized bool
	waitCommits int
	attempted   bool
}

// EnableImportDBPeerSnapshots is called before processing starts. Optional
// interfaces keep memory rotation out of the regular accounts and storer APIs.
func (adb *PeerAccountsDB) EnableImportDBPeerSnapshots(store storage.Storer, interval uint64) error {
	memory, ok := store.(importDBPeerMemory)
	snapshotter, supported := adb.snapshotsManger.(importDBPeerSnapshotter)
	if !ok || !supported || interval == 0 || adb.pruningEnabled || !adb.getMainTrie().GetStorageManager().IsSnapshotSupported() {
		return fmt.Errorf("import peer snapshots require memory rotation, snapshots and disabled peer pruning")
	}
	adb.importDBSnapshots = &importDBPeerSnapshots{store: memory, snapshotter: snapshotter, interval: interval}
	log.Debug("import-db peer snapshots enabled", "rounds interval", interval, "max wait commits", importDBPeerSnapshotMaxWaitCommits)
	return nil
}

// InitializeImportDBPeerState copies the selected startup checkpoint from disk
// before replay starts. Bootstrap may try older roots until it selects one, so
// this must run after storage restoration, rather than during RecreateTrie.
func (adb *PeerAccountsDB) InitializeImportDBPeerState() error {
	r := adb.importDBSnapshots
	if r == nil {
		return nil
	}
	r.processMu.Lock()
	defer r.processMu.Unlock()
	source, ok := r.store.(interface{ HasBootstrapSource() bool })
	if !ok || !source.HasBootstrapSource() {
		log.Debug("import-db peer startup snapshot skipped, no disk source")
		return nil
	}
	root, err := adb.RootHash()
	if err != nil {
		log.Debug("import-db peer startup root unavailable", "error", err)
		return err
	}
	manager, epoch, err := adb.getTrieStorageManagerAndLatestEpoch(adb.getMainTrie())
	if err != nil {
		log.Debug("import-db peer startup storage epoch unavailable", "rootHash", root, "error", err)
		return err
	}
	log.Info("starting import-db peer startup snapshot", "rootHash", root, "epoch", epoch)
	if r.snapshotter.SnapshotPeerState(root, epoch, manager, func() {
		log.Debug("import-db peer startup snapshot started", "rootHash", root, "epoch", epoch)
	}) {
		r.store.SnapshotFinished()
		log.Debug("import-db peer startup snapshot completed", "rootHash", root, "epoch", epoch)
	} else {
		log.Debug("import-db peer startup snapshot incomplete or skipped, retaining disk source", "rootHash", root, "epoch", epoch)
	}
	return nil
}

// RecordImportDBPeerCommit associates stored state with the block actually
// executed, not the independently advancing header or its last included result.
func (adb *PeerAccountsDB) RecordImportDBPeerCommit(header data.HeaderHandler, hash []byte) {
	if adb.importDBSnapshots == nil {
		return
	}
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()
	adb.importDBSnapshots.latest = &importDBPeerRoot{
		root: bytes.Clone(adb.lastRootHash), hash: bytes.Clone(hash),
		nonce: header.GetNonce(), epoch: header.GetEpoch(), v3: header.IsHeaderV3(),
	}
	log.Debug("import-db peer committed root recorded", "nonce", header.GetNonce(), "hash", hash,
		"rootHash", adb.lastRootHash, "epoch", header.GetEpoch(), "v3", header.IsHeaderV3())
}

func (adb *PeerAccountsDB) captureImportDBPeerRoot() bool {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()
	r := adb.importDBSnapshots
	if r.latest == nil || !bytes.Equal(r.latest.root, adb.lastRootHash) {
		log.Debug("import-db peer root capture deferred", "has committed root", r.latest != nil, "current rootHash", adb.lastRootHash)
		return false
	}
	// No newer account commit may write into the source between root selection
	// and the switch. A timed-out candidate reuses the existing replacement.
	r.store.Rotate()
	candidate := *r.latest
	r.candidate = &candidate
	r.waitCommits, r.attempted = 0, false
	log.Debug("import-db peer snapshot root captured", "nonce", candidate.nonce, "hash", candidate.hash,
		"rootHash", candidate.root, "epoch", candidate.epoch, "v3", candidate.v3)
	return true
}

// ProcessImportDBPeerSnapshot is driven by canonical import commits. finalNonce
// and finalHash identify settled execution, as exposed by GetFinalBlockInfo.
func (adb *PeerAccountsDB) ProcessImportDBPeerSnapshot(header data.HeaderHandler, hash []byte, finalNonce uint64, finalHash []byte) {
	r := adb.importDBSnapshots
	if r == nil || check.IfNil(header) {
		return
	}
	r.processMu.Lock()
	defer r.processMu.Unlock()
	if bytes.Equal(hash, r.lastHeader) {
		return
	}
	r.lastHeader = bytes.Clone(hash)
	round := header.GetRound()
	if !r.initialized {
		r.lastRound, r.initialized = round, true
		log.Debug("import-db peer snapshot round counter initialized", "round", round, "rounds interval", r.interval)
	}
	due := round >= r.lastRound && round-r.lastRound >= r.interval
	captured := false
	if r.candidate == nil {
		if (!due && !r.requested.Load()) || !adb.captureImportDBPeerRoot() {
			return
		}
		log.Debug("import-db peer snapshot trigger reached", "round", round, "last snapshot round", r.lastRound,
			"round interval reached", due, "epoch snapshot requested", r.requested.Load())
		captured = true
	}
	candidate := r.candidate
	ready := header.GetNonce() == candidate.nonce+1 && bytes.Equal(header.GetPrevHash(), candidate.hash)
	if candidate.v3 {
		// Import replays the canonical chain. A later settled execution result
		// also settles earlier executions; equality must match the exact hash.
		ready = len(finalHash) > 0 && (finalNonce > candidate.nonce ||
			(finalNonce == candidate.nonce && bytes.Equal(finalHash, candidate.hash)))
	}
	if !ready {
		if !captured {
			r.waitCommits++
			if r.waitCommits >= importDBPeerSnapshotMaxWaitCommits {
				log.Debug("import-db peer snapshot finality wait expired, recapturing root", "nonce", candidate.nonce,
					"rootHash", candidate.root, "wait commits", r.waitCommits, "final nonce", finalNonce, "final hash", finalHash)
				adb.captureImportDBPeerRoot()
			}
		}
		log.Debug("import-db peer snapshot waiting for finality", "nonce", r.candidate.nonce, "hash", r.candidate.hash,
			"round", round, "committed nonce", header.GetNonce(), "final nonce", finalNonce, "final hash", finalHash, "wait commits", r.waitCommits)
		return
	}
	if r.attempted && !due && !r.requested.Load() {
		log.Debug("import-db peer snapshot retry waiting for trigger", "nonce", candidate.nonce, "round", round, "last snapshot round", r.lastRound)
		return
	}
	log.Debug("import-db peer finalized snapshot requested", "nonce", candidate.nonce, "hash", candidate.hash,
		"rootHash", candidate.root, "epoch", candidate.epoch, "round", round)
	completed := r.snapshotter.SnapshotPeerState(candidate.root, candidate.epoch, adb.getMainTrie().GetStorageManager(), func() {
		// Reset only when the snapshot is actually accepted, not on capture,
		// timeout, skipped requests or completion.
		r.lastRound, r.attempted = round, true
		r.requested.Store(false)
		log.Debug("import-db peer finalized snapshot started, round counter reset", "nonce", candidate.nonce,
			"rootHash", candidate.root, "epoch", candidate.epoch, "round", round)
	})
	if completed {
		r.store.SnapshotFinished()
		r.candidate = nil
		log.Debug("import-db peer finalized snapshot completed", "nonce", candidate.nonce, "rootHash", candidate.root, "epoch", candidate.epoch)
	} else {
		log.Debug("import-db peer finalized snapshot incomplete or skipped, retaining source", "nonce", candidate.nonce, "rootHash", candidate.root, "epoch", candidate.epoch)
	}
}

// SnapshotState coalesces ordinary epoch snapshots with the round trigger.
func (adb *PeerAccountsDB) SnapshotState(root []byte, epoch uint32) {
	if adb.importDBSnapshots == nil {
		adb.AccountsDB.SnapshotState(root, epoch)
		return
	}
	adb.importDBSnapshots.requested.Store(true)
	log.Debug("import-db peer epoch snapshot queued for finalized root", "rootHash", root, "epoch", epoch)
}

// Disk seeding runs after checkpoint restoration in InitializeImportDBPeerState.
// Memory generations have no snapshot state to recover before that point.
func (adb *PeerAccountsDB) StartSnapshotIfNeeded() error {
	if adb.importDBSnapshots != nil {
		log.Debug("import-db peer restart snapshot deferred until checkpoint restoration")
		return nil
	}
	return adb.AccountsDB.StartSnapshotIfNeeded()
}
