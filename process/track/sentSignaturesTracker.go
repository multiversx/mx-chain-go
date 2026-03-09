package track

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// externalPeerID is just a marker so the ResetRoundsWithoutReceivedMessages will know it is not an owned peer ID
// this is actually an invalid peer ID, it can not be obtained from a key
const externalPeerID = core.PeerID("external peer id")

type sentSignaturesTracker struct {
	mut          sync.RWMutex
	sentFromSelf map[string]struct{}
	signedNonces map[string]map[uint64][]byte // pk -> nonce -> headerHash
	keysHandler  KeysHandler
}

// NewSentSignaturesTracker will create a new instance of a tracker able to record if a signature was sent from self
func NewSentSignaturesTracker(keysHandler KeysHandler) (*sentSignaturesTracker, error) {
	if check.IfNil(keysHandler) {
		return nil, ErrNilKeysHandler
	}

	return &sentSignaturesTracker{
		sentFromSelf: make(map[string]struct{}),
		signedNonces: make(map[string]map[uint64][]byte),
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
// First-write-wins: if an entry already exists for this (pk, nonce), it is not overwritten.
// Automatically cleans up entries for nonces more than 1 behind the recorded nonce.
func (tracker *sentSignaturesTracker) RecordSignedNonce(pkBytes []byte, nonce uint64, headerHash []byte) {
	pk := string(pkBytes)

	tracker.mut.Lock()
	defer tracker.mut.Unlock()

	nonceMap, exists := tracker.signedNonces[pk]
	if !exists {
		nonceMap = make(map[uint64][]byte)
		tracker.signedNonces[pk] = nonceMap
	}

	if _, alreadyRecorded := nonceMap[nonce]; alreadyRecorded {
		return
	}

	hashCopy := make([]byte, len(headerHash))
	copy(hashCopy, headerHash)
	nonceMap[nonce] = hashCopy

	for oldNonce := range nonceMap {
		if nonce > oldNonce && nonce-oldNonce > 1 {
			delete(nonceMap, oldNonce)
		}
	}
}

// GetSignedHash returns the header hash previously signed by the given public key for the given nonce.
// Returns (hash, true) if found, (nil, false) otherwise.
func (tracker *sentSignaturesTracker) GetSignedHash(pkBytes []byte, nonce uint64) ([]byte, bool) {
	tracker.mut.RLock()
	defer tracker.mut.RUnlock()

	nonceMap, exists := tracker.signedNonces[string(pkBytes)]
	if !exists {
		return nil, false
	}

	hash, found := nonceMap[nonce]
	return hash, found
}

// IsInterfaceNil returns true if there is no value under the interface
func (tracker *sentSignaturesTracker) IsInterfaceNil() bool {
	return tracker == nil
}
