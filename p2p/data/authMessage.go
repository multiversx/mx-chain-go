//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. authMessage.proto
//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. topicMessage.proto
package data

// AuthMessage represents the authentication message used in the handshake process of 2 peers
type AuthMessage struct {
	AuthMessagePb
}

// IsInterfaceNil returns true if there is no value under the interface
func (am *AuthMessage) IsInterfaceNil() bool {
	return am == nil
}
