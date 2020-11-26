//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. esdt.proto
package esdt

import "math/big"

// New returns a new batch from given buffers
func New() *ESDigitalToken {
	return &ESDigitalToken{
		Value: big.NewInt(0),
	}
}
