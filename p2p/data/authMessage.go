//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. proto/authMessage.proto
//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. proto/topicMessage.proto
package data

// AuthMessage represents the authentication message used in the handshake process of 2 peers
type AuthMessage struct {
	AuthMessagePb
}

// IsInterfaceNil returns true if there is no value under the interface
func (am *AuthMessage) IsInterfaceNil() bool {
	return am == nil
}
