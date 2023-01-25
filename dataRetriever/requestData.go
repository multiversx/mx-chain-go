//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. requestData.proto
package dataRetriever

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/p2p"
)

// UnmarshalWith sets the fields according to p2p.MessageP2P.Data() contents
// Errors if something went wrong
func (rd *RequestData) UnmarshalWith(marshalizer marshal.Marshalizer, message p2p.MessageP2P) error {
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return ErrNilMarshalizer
	}
	if check.IfNil(message) {
		return ErrNilMessage
	}
	if message.Data() == nil {
		return ErrNilDataToProcess
	}

	err := marshalizer.Unmarshal(rd, message.Data())
	if err != nil {
		return err
	}

	return nil
}
