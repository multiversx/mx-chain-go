package interceptors

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// PeerBlockBodyInterceptor represents an interceptor used for transaction block bodies
type PeerBlockBodyInterceptor struct {
	*GenericBlockBodyInterceptor
}

// NewPeerBlockBodyInterceptor creates a new instance of a TxBlockBodyInterceptor
func NewPeerBlockBodyInterceptor(
	marshalizer marshal.Marshalizer,
	cache storage.Cacher,
	storer storage.Storer,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
) (*PeerBlockBodyInterceptor, error) {
	gbbi, err := NewGenericBlockBodyInterceptor(
		marshalizer,
		cache,
		storer,
		hasher,
		shardCoordinator)

	if err != nil {
		return nil, err
	}

	return &PeerBlockBodyInterceptor{GenericBlockBodyInterceptor: gbbi}, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (pbbi *PeerBlockBodyInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	err := pbbi.checkMessage(message)
	if err != nil {
		return err
	}

	peerBlockBody := block.NewInterceptedPeerBlockBody()
	err = pbbi.marshalizer.Unmarshal(peerBlockBody, message.Data())
	if err != nil {
		return err
	}

	return pbbi.processBlockBody(message.Data(), peerBlockBody)
}
