package spos

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/consensus"
)

// externalPeerID is just a marker so the ResetRoundsWithoutReceivedMessages will know it is not an owned peer ID
// this is actually an invalid peer ID, it can not be obtained from a key
const externalPeerID = core.PeerID("external peer id")

type sentSignaturesTracker struct {
	mut          sync.RWMutex
	sentFromSelf map[string]struct{}
	keysHandler  consensus.KeysHandler
}

// NewSentSignaturesTracker will create a new instance of a tracker able to record if a signature was sent from self
func NewSentSignaturesTracker(keysHandler consensus.KeysHandler) (*sentSignaturesTracker, error) {
	if check.IfNil(keysHandler) {
		return nil, ErrNilKeysHandler
	}

	return &sentSignaturesTracker{
		sentFromSelf: make(map[string]struct{}),
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

// ReceivedActualSigners is called whenever a final info is received. If a signer public key did not send a signature
// from the current host, it will call the reset rounds without received message. This is the case when another instance of a
// multikey node (possibly running as main) broadcast only the final info as it contained the leader + a few signers
func (tracker *sentSignaturesTracker) ReceivedActualSigners(signersPks []string) {
	tracker.mut.RLock()
	defer tracker.mut.RUnlock()

	for _, signerPk := range signersPks {
		_, isSentFromSelf := tracker.sentFromSelf[signerPk]
		if isSentFromSelf {
			continue
		}

		tracker.keysHandler.ResetRoundsWithoutReceivedMessages([]byte(signerPk), externalPeerID)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (tracker *sentSignaturesTracker) IsInterfaceNil() bool {
	return tracker == nil
}
