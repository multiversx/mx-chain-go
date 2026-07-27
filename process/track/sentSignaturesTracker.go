package track

import (
	"bytes"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// externalPeerID is just a marker so the ResetRoundsWithoutReceivedMessages will know it is not an owned peer ID
// this is actually an invalid peer ID, it can not be obtained from a key
const externalPeerID = core.PeerID("external peer id")

type signedNonceInfo struct {
	headerHash []byte
	roundIndex int64
}

// roundsRetentionWindow bounds the per-key signed-rounds history; entries older than the window
// are removed on each reservation
const roundsRetentionWindow = 4

type sentSignaturesTracker struct {
	mut          sync.RWMutex
	sentFromSelf map[string]struct{}
	signedNonces map[string]map[uint64]*signedNonceInfo // pk -> nonce -> signedNonceInfo
	signedRounds map[string]map[int64][]byte            // pk -> round -> signed header hash
	keysHandler  KeysHandler
}

// NewSentSignaturesTracker will create a new instance of a tracker able to record if a signature was sent from self
func NewSentSignaturesTracker(keysHandler KeysHandler) (*sentSignaturesTracker, error) {
	if check.IfNil(keysHandler) {
		return nil, ErrNilKeysHandler
	}

	return &sentSignaturesTracker{
		sentFromSelf: make(map[string]struct{}),
		signedNonces: make(map[string]map[uint64]*signedNonceInfo),
		signedRounds: make(map[string]map[int64][]byte),
		keysHandler:  keysHandler,
	}, nil
}

// StartRound will initialize the tracker by removing any stored values
func (tracker *sentSignaturesTracker) StartRound() {
	tracker.mut.Lock()
	tracker.sentFromSelf = make(map[string]struct{})
	tracker.mut.Unlock()
}

// SignatureSent will record that the current host sent a signature for the provided public key
func (tracker *sentSignaturesTracker) SignatureSent(pkBytes []byte) {
	tracker.mut.Lock()
	tracker.sentFromSelf[string(pkBytes)] = struct{}{}
	tracker.mut.Unlock()
}

// ResetCountersForManagedBlockSigner is called at commit time and will call the reset rounds without received messages
// for the provided key that actually signed a block
func (tracker *sentSignaturesTracker) ResetCountersForManagedBlockSigner(signerPk []byte) {
	tracker.mut.RLock()
	defer tracker.mut.RUnlock()

	_, isSentFromSelf := tracker.sentFromSelf[string(signerPk)]
	if isSentFromSelf {
		return
	}

	tracker.keysHandler.ResetRoundsWithoutReceivedMessages(signerPk, externalPeerID)
}

// RecordSignedNonce records that a public key has signed a header hash for a given nonce.
// Most-recent-round-wins: overwrites an existing entry only if roundIndex is strictly greater
// than the previously recorded round. Same-round re-sends are idempotent (kept as-is).
// Automatically cleans up entries for nonces more than 1 behind the recorded nonce.
func (tracker *sentSignaturesTracker) RecordSignedNonce(pkBytes []byte, nonce uint64, headerHash []byte, roundIndex int64) {
	pk := string(pkBytes)

	tracker.mut.Lock()
	defer tracker.mut.Unlock()

	nonceMap, exists := tracker.signedNonces[pk]
	if !exists {
		nonceMap = make(map[uint64]*signedNonceInfo)
		tracker.signedNonces[pk] = nonceMap
	}

	if existing, alreadyRecorded := nonceMap[nonce]; alreadyRecorded {
		if roundIndex <= existing.roundIndex {
			return
		}
	}

	hashCopy := make([]byte, len(headerHash))
	copy(hashCopy, headerHash)
	nonceMap[nonce] = &signedNonceInfo{
		headerHash: hashCopy,
		roundIndex: roundIndex,
	}

	for oldNonce := range nonceMap {
		if nonce > oldNonce && nonce-oldNonce > 1 {
			delete(nonceMap, oldNonce)
		}
	}
}

// ReserveSignatureInRound atomically reserves the (key, round) signing slot, false when a different
// hash was already reserved in that round; in-memory only, crash-restart in-round out of scope
func (tracker *sentSignaturesTracker) ReserveSignatureInRound(pkBytes []byte, roundIndex int64, headerHash []byte) bool {
	pk := string(pkBytes)

	tracker.mut.Lock()
	defer tracker.mut.Unlock()

	roundMap, exists := tracker.signedRounds[pk]
	if !exists {
		roundMap = make(map[int64][]byte)
		tracker.signedRounds[pk] = roundMap
	}

	reservedHash, alreadyReserved := roundMap[roundIndex]
	if alreadyReserved {
		return bytes.Equal(reservedHash, headerHash)
	}

	hashCopy := make([]byte, len(headerHash))
	copy(hashCopy, headerHash)
	roundMap[roundIndex] = hashCopy

	for oldRound := range roundMap {
		if roundIndex-oldRound > roundsRetentionWindow {
			delete(roundMap, oldRound)
		}
	}

	return true
}

// GetSignedNonceInfo returns the header hash and round index previously signed by the given public key
// for the given nonce. Returns (hash, roundIndex, true) if found, (nil, 0, false) otherwise.
func (tracker *sentSignaturesTracker) GetSignedNonceInfo(pkBytes []byte, nonce uint64) ([]byte, int64, bool) {
	tracker.mut.RLock()
	defer tracker.mut.RUnlock()

	nonceMap, exists := tracker.signedNonces[string(pkBytes)]
	if !exists {
		return nil, 0, false
	}

	info, found := nonceMap[nonce]
	if !found {
		return nil, 0, false
	}

	return info.headerHash, info.roundIndex, true
}

// IsInterfaceNil returns true if there is no value under the interface
func (tracker *sentSignaturesTracker) IsInterfaceNil() bool {
	return tracker == nil
}
