//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. requestData.proto
package dataRetriever

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// UnmarshalWith sets the fields according to p2p.MessageP2P.Data() contents
// Errors if something went wrong
func (rd *RequestData) UnmarshalWith(marshaller marshal.Marshalizer, message p2p.MessageP2P) error {
	if marshaller == nil || check.IfNil(marshaller) {
		return ErrNilMarshaller
	}
	if check.IfNil(message) {
		return ErrNilMessage
	}
	if message.Data() == nil {
		return ErrNilDataToProcess
	}

	err := marshaller.Unmarshal(rd, message.Data())
	if err != nil {
		return err
	}

	return nil
}
