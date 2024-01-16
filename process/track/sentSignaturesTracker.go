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
	keysHandler  KeysHandler
}

// NewSentSignaturesTracker will create a new instance of a tracker able to record if a signature was sent from self
func NewSentSignaturesTracker(keysHandler KeysHandler) (*sentSignaturesTracker, error) {
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

// ResetCountersManagedBlockSigners is called at commit time and will call the reset rounds without received messages
// for each managed key that actually signed a block
func (tracker *sentSignaturesTracker) ResetCountersManagedBlockSigners(signersPKs [][]byte) {
	tracker.mut.RLock()
	defer tracker.mut.RUnlock()

	for _, signerPk := range signersPKs {
		_, isSentFromSelf := tracker.sentFromSelf[string(signerPk)]
		if isSentFromSelf {
			continue
		}

		tracker.keysHandler.ResetRoundsWithoutReceivedMessages(signerPk, externalPeerID)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (tracker *sentSignaturesTracker) IsInterfaceNil() bool {
	return tracker == nil
}
