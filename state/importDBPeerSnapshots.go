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
}

func (adb *PeerAccountsDB) captureImportDBPeerRoot() bool {
	adb.mutOp.Lock()
	defer adb.mutOp.Unlock()
	r := adb.importDBSnapshots
	if r.latest == nil || !bytes.Equal(r.latest.root, adb.lastRootHash) {
		return false
	}
	// No newer account commit may write into the source between root selection
	// and the switch. A timed-out candidate reuses the existing replacement.
	r.store.Rotate()
	candidate := *r.latest
	r.candidate = &candidate
	r.waitCommits, r.attempted = 0, false
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
	}
	due := round >= r.lastRound && round-r.lastRound >= r.interval
	captured := false
	if r.candidate == nil {
		if (!due && !r.requested.Load()) || !adb.captureImportDBPeerRoot() {
			return
		}
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
				adb.captureImportDBPeerRoot()
			}
		}
		return
	}
	if r.attempted && !due && !r.requested.Load() {
		return
	}
	completed := r.snapshotter.SnapshotPeerState(candidate.root, candidate.epoch, adb.getMainTrie().GetStorageManager(), func() {
		// Reset only when the snapshot is actually accepted, not on capture,
		// timeout, skipped requests or completion.
		r.lastRound, r.attempted = round, true
		r.requested.Store(false)
	})
	if completed {
		r.store.SnapshotFinished()
		r.candidate = nil
	}
}

// SnapshotState coalesces ordinary epoch snapshots with the round trigger.
func (adb *PeerAccountsDB) SnapshotState(root []byte, epoch uint32) {
	if adb.importDBSnapshots == nil {
		adb.AccountsDB.SnapshotState(root, epoch)
		return
	}
	adb.importDBSnapshots.requested.Store(true)
}

// Memory generations have no snapshot state to recover after a restart.
func (adb *PeerAccountsDB) StartSnapshotIfNeeded() error {
	if adb.importDBSnapshots != nil {
		return nil
	}
	return adb.AccountsDB.StartSnapshotIfNeeded()
}
