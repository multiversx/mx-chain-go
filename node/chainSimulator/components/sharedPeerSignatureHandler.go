package components

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	crypto "github.com/multiversx/mx-chain-crypto-go"
)

// peerSignatureVerificationKey identifies the static validator-key to peer-ID binding carried by
// consensus messages. String fields keep arbitrary binary values distinct without delimiter
// ambiguities.
type peerSignatureVerificationKey struct {
	publicKey string
	peerID    string
	signature string
}

type peerSignatureVerification struct {
	done chan struct{}
	err  error
}

// sharedPeerSignatureVerifier deduplicates the same real BLS verification across the physical nodes
// of one in-process simulator. Production nodes verify independently because they do not share
// trust; the simulator's network already routes a message from one registered sender to every
// receiver in the same process, so repeating the immutable key-to-peer binding on every receiver
// adds CPU cost without exercising a different consensus decision.
type sharedPeerSignatureVerifier struct {
	mut           sync.Mutex
	verifications map[peerSignatureVerificationKey]*peerSignatureVerification
}

func newSharedPeerSignatureVerifier() *sharedPeerSignatureVerifier {
	return &sharedPeerSignatureVerifier{
		verifications: make(map[peerSignatureVerificationKey]*peerSignatureVerification),
	}
}

func (verifier *sharedPeerSignatureVerifier) wrap(handler crypto.PeerSignatureHandler) crypto.PeerSignatureHandler {
	return &cachedPeerSignatureHandler{
		handler:  handler,
		verifier: verifier,
	}
}

type cachedPeerSignatureHandler struct {
	handler  crypto.PeerSignatureHandler
	verifier *sharedPeerSignatureVerifier
}

func (handler *cachedPeerSignatureHandler) VerifyPeerSignature(
	publicKey []byte,
	peerID core.PeerID,
	signature []byte,
) error {
	key := peerSignatureVerificationKey{
		publicKey: string(publicKey),
		peerID:    string(peerID),
		signature: string(signature),
	}

	handler.verifier.mut.Lock()
	verification, found := handler.verifier.verifications[key]
	if found {
		handler.verifier.mut.Unlock()
		<-verification.done
		return verification.err
	}

	verification = &peerSignatureVerification{done: make(chan struct{})}
	handler.verifier.verifications[key] = verification
	handler.verifier.mut.Unlock()

	verification.err = handler.handler.VerifyPeerSignature(publicKey, peerID, signature)
	close(verification.done)

	return verification.err
}

func (handler *cachedPeerSignatureHandler) GetPeerSignature(
	privateKey crypto.PrivateKey,
	peerID []byte,
) ([]byte, error) {
	return handler.handler.GetPeerSignature(privateKey, peerID)
}

func (handler *cachedPeerSignatureHandler) IsInterfaceNil() bool {
	return handler == nil
}
