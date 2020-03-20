package mock

import "github.com/ElrondNetwork/elrond-go/p2p"

// NilAntifloodHandler is an empty implementation of P2PAntifloodHandler
// it does nothing
type NilAntifloodHandler struct {
}

// ResetForTopic won't do anything
func (nah *NilAntifloodHandler) ResetForTopic(_ string) {
}

// SetMaxMessagesForTopic won't do anything
func (nah *NilAntifloodHandler) SetMaxMessagesForTopic(_ string, _ uint32) {
}

// CanProcessMessage will always return nil, allowing messages to go to interceptors
func (nah *NilAntifloodHandler) CanProcessMessage(_ p2p.MessageP2P, _ p2p.PeerID) error {
	return nil
}

// CanProcessMessageOnTopic will always return nil, allowing messages to go to interceptors
func (nah *NilAntifloodHandler) CanProcessMessageOnTopic(_ p2p.PeerID, _ string) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (nah *NilAntifloodHandler) IsInterfaceNil() bool {
	return nah == nil
}
