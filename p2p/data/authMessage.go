package data

import (
	protobuf "github.com/ElrondNetwork/elrond-go/p2p/data/proto"
)

// AuthMessage represents the authentication message used in the handshake process of 2 peers
type AuthMessage struct {
	protobuf.AuthMessagePb
}

// IsInterfaceNil returns true if there is no value under the interface
func (am *AuthMessage) IsInterfaceNil() bool {
	return am == nil
}
