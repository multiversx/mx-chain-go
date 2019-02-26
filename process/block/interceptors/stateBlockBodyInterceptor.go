package interceptors

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// StateBlockBodyInterceptor represents an interceptor used for state block bodies
type StateBlockBodyInterceptor struct {
	*GenericBlockBodyInterceptor
}

// NewStateBlockBodyInterceptor creates a new instance of a TxBlockBodyInterceptor
func NewStateBlockBodyInterceptor(
	marshalizer marshal.Marshalizer,
	cache storage.Cacher,
	storer storage.Storer,
	hasher hashing.Hasher,
	shardCoordinator sharding.ShardCoordinator,
) (*StateBlockBodyInterceptor, error) {
	gbbi, err := NewGenericBlockBodyInterceptor(
		marshalizer,
		cache,
		storer,
		hasher,
		shardCoordinator)

	if err != nil {
		return nil, err
	}

	return &StateBlockBodyInterceptor{GenericBlockBodyInterceptor: gbbi}, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (sbbi *StateBlockBodyInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	err := sbbi.checkMessage(message)
	if err != nil {
		return err
	}

	stateBlockBody := block.NewInterceptedStateBlockBody()
	err = sbbi.marshalizer.Unmarshal(stateBlockBody, message.Data())
	if err != nil {
		return err
	}

	return sbbi.processBlockBody(message.Data(), stateBlockBody)
}
