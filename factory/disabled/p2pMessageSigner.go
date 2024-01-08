package disabled

import "github.com/multiversx/mx-chain-go/p2p"

// MessageSigner implements P2PSigningHandler interface but it does nothing as it is disabled
type MessageSigner struct {
}

// Verify does nothing
func (ms *MessageSigner) Verify(message p2p.MessageP2P) error {
	return nil
}

// Serialize does nothing
func (ms *MessageSigner) Serialize(messages []p2p.MessageP2P) ([]byte, error) {
	return nil, nil
}

// Deserialize does nothing
func (ms *MessageSigner) Deserialize(messagesBytes []byte) ([]p2p.MessageP2P, error) {
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ms *MessageSigner) IsInterfaceNil() bool {
	return ms == nil
}
