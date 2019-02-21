package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type TopicValidatorStub struct {
	ValidateCalled func(message p2p.MessageP2P) error
}

func (tvs *TopicValidatorStub) Validate(message p2p.MessageP2P) error {
	return tvs.ValidateCalled(message)
}
