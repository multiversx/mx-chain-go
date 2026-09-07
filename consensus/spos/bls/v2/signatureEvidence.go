package v2

import (
	"sync"
	"sync/atomic"

	"github.com/multiversx/mx-chain-core-go/core/check"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
)

// roundSignatureEvidence is a snapshot of the signature shares observed for one block,
// taken at the start of the following round, before the consensus state is reset.
type roundSignatureEvidence struct {
	roundIndex     int64
	nonce          uint64
	headerHash     []byte
	epoch          uint32
	headerRound    uint64
	shardID        uint32
	isStartOfEpoch bool
	threshold      int
	consensusGroup []string
	// assemblyRunning is true strictly while an assembly attempt executes;
	// nextAssemblyRound is the first round allowed to start the next attempt
	assemblyRunning   atomic.Bool
	nextAssemblyRound atomic.Int64

	mut    sync.RWMutex
	bitmap []byte
	shares [][]byte
	count  int
}

func (ev *roundSignatureEvidence) getCount() int {
	ev.mut.RLock()
	defer ev.mut.RUnlock()

	return ev.count
}

// getAggregationData returns copies of the bitmap and shares, safe to use outside the lock
func (ev *roundSignatureEvidence) getAggregationData() ([]byte, [][]byte) {
	ev.mut.RLock()
	defer ev.mut.RUnlock()

	bitmap := make([]byte, len(ev.bitmap))
	copy(bitmap, ev.bitmap)
	shares := make([][]byte, len(ev.shares))
	copy(shares, ev.shares)

	return bitmap, shares
}

// dropShares removes the shares at the given group indices and returns the remaining count
func (ev *roundSignatureEvidence) dropShares(indices []int) int {
	ev.mut.Lock()
	defer ev.mut.Unlock()

	for _, idx := range indices {
		if idx < 0 || idx >= len(ev.shares) || ev.shares[idx] == nil {
			continue
		}
		ev.shares[idx] = nil
		ev.bitmap[idx/8] &^= 1 << (uint16(idx) % 8)
		ev.count--
	}

	return ev.count
}

// signatureEvidenceHandler holds the signature evidence observed for the previous round's
// block, plus one retained quorum-level slot for a still unsettled nonce.
type signatureEvidenceHandler interface {
	Capture(ev *roundSignatureEvidence)
	GetPreviousRoundEvidence(nonce uint64, currentRound int64) (*roundSignatureEvidence, bool)
	GetRetainedQuorumEvidence(nonce uint64) (*roundSignatureEvidence, bool)
	GetAssemblyCandidate() (*roundSignatureEvidence, bool)
	DropRetained(nonce uint64)
	IsInterfaceNil() bool
}

type signatureEvidenceStore struct {
	proofsPool consensus.EquivalentProofsPool

	mut      sync.RWMutex
	previous *roundSignatureEvidence
	// retainedQuorum keeps the last quorum-level evidence for a still unsettled nonce beyond
	// the one-round lifetime of the previous slot, so self-assembly can be retried
	retainedQuorum *roundSignatureEvidence
}

func newSignatureEvidenceStore(proofsPool consensus.EquivalentProofsPool) (*signatureEvidenceStore, error) {
	if check.IfNil(proofsPool) {
		return nil, spos.ErrNilEquivalentProofPool
	}

	return &signatureEvidenceStore{
		proofsPool: proofsPool,
	}, nil
}

// Capture overwrites the previous-round slot (nil clears it); the replaced evidence is
// promoted to the retained slot when it holds quorum for a still unsettled nonce.
func (store *signatureEvidenceStore) Capture(ev *roundSignatureEvidence) {
	store.mut.Lock()
	defer store.mut.Unlock()

	replaced := store.previous
	store.previous = ev

	if replaced == nil {
		return
	}
	if replaced.getCount() < replaced.threshold {
		return
	}
	if store.isNonceSettled(replaced) {
		return
	}

	store.retainedQuorum = replaced
}

// GetPreviousRoundEvidence returns the previous-round slot only if it matches the given
// nonce and was captured exactly one round before currentRound
func (store *signatureEvidenceStore) GetPreviousRoundEvidence(nonce uint64, currentRound int64) (*roundSignatureEvidence, bool) {
	store.mut.RLock()
	defer store.mut.RUnlock()

	ev := store.previous
	if ev == nil || ev.nonce != nonce || ev.roundIndex != currentRound-1 {
		return nil, false
	}

	return ev, true
}

// GetRetainedQuorumEvidence returns the retained quorum slot for the nonce, if any
func (store *signatureEvidenceStore) GetRetainedQuorumEvidence(nonce uint64) (*roundSignatureEvidence, bool) {
	store.mut.RLock()
	defer store.mut.RUnlock()

	ev := store.retainedQuorum
	if ev == nil || ev.nonce != nonce {
		return nil, false
	}

	return ev, true
}

// GetAssemblyCandidate returns quorum-level evidence for a still unsettled nonce, if any
func (store *signatureEvidenceStore) GetAssemblyCandidate() (*roundSignatureEvidence, bool) {
	store.mut.RLock()
	defer store.mut.RUnlock()

	ev := store.previous
	if ev != nil && ev.getCount() >= ev.threshold && !store.isNonceSettled(ev) {
		return ev, true
	}

	ev = store.retainedQuorum
	if ev != nil && ev.getCount() >= ev.threshold && !store.isNonceSettled(ev) {
		return ev, true
	}

	return nil, false
}

// DropRetained clears the retained slot for the nonce, on settlement or demotion
func (store *signatureEvidenceStore) DropRetained(nonce uint64) {
	store.mut.Lock()
	defer store.mut.Unlock()

	if store.retainedQuorum != nil && store.retainedQuorum.nonce == nonce {
		store.retainedQuorum = nil
	}
}

func (store *signatureEvidenceStore) isNonceSettled(ev *roundSignatureEvidence) bool {
	proof, err := store.proofsPool.GetProofByNonce(ev.nonce, ev.shardID)

	return err == nil && !check.IfNil(proof)
}

// IsInterfaceNil returns true if there is no value under the interface
func (store *signatureEvidenceStore) IsInterfaceNil() bool {
	return store == nil
}
