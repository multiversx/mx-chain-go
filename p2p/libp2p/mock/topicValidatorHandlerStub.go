package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type TopicValidatorHandlerStub struct {
	ValidateCalled func(message p2p.MessageP2P) error
}

func (tvhs *TopicValidatorHandlerStub) Validate(message p2p.MessageP2P) error {
	return tvhs.ValidateCalled(message)
}
