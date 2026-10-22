package state

import (
	"bytes"
	"sync"
	"sync/atomic"

	"github.com/multiversx/mx-chain-core-go/data"
)

const importDBPeerSnapshotMaxWaitCommits = 5

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

// isDue is called with processMu held. A backwards round never underflows.
func (r *importDBPeerSnapshots) isDue(round uint64) bool {
	return round >= r.lastRound && round-r.lastRound >= r.interval
}

func (r *importDBPeerRoot) isFinal(header data.HeaderHandler, finalNonce uint64, finalHash []byte) bool {
	if !r.v3 {
		return header.GetNonce() == r.nonce+1 && bytes.Equal(header.GetPrevHash(), r.hash)
	}
	// Import replays the canonical chain. A later settled execution result
	// also settles earlier executions; equality must match the exact hash.
	return len(finalHash) > 0 && (finalNonce > r.nonce ||
		(finalNonce == r.nonce && bytes.Equal(finalHash, r.hash)))
}
