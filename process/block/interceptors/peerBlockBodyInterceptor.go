package interceptors

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// PeerBlockBodyInterceptor represents an interceptor used for peer changes block bodies
type PeerBlockBodyInterceptor struct {
	*messageChecker
	marshalizer      marshal.Marshalizer
	cache            storage.Cacher
	hasher           hashing.Hasher
	storer           storage.Storer
	shardCoordinator sharding.Coordinator
}

// NewPeerBlockBodyInterceptor creates a new instance of a TxBlockBodyInterceptor
func NewPeerBlockBodyInterceptor(
	marshalizer marshal.Marshalizer,
	cache storage.Cacher,
	storer storage.Storer,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
) (*PeerBlockBodyInterceptor, error) {

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if cache == nil {
		return nil, process.ErrNilCacher
	}
	if storer == nil {
		return nil, process.ErrNilBlockBodyStorage
	}
	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}

	return &PeerBlockBodyInterceptor{
		messageChecker:   &messageChecker{},
		marshalizer:      marshalizer,
		cache:            cache,
		storer:           storer,
		hasher:           hasher,
		shardCoordinator: shardCoordinator,
	}, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (pbbi *PeerBlockBodyInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	err := pbbi.checkMessage(message)
	if err != nil {
		return err
	}

	peerChBlockBody := block.NewInterceptedPeerBlockBody()
	err = pbbi.marshalizer.Unmarshal(peerChBlockBody, message.Data())
	if err != nil {
		return err
	}

	err = peerChBlockBody.IntegrityAndValidity(pbbi.shardCoordinator)
	if err != nil {
		return err
	}

	go pbbi.processPeerChangeBlockBody(peerChBlockBody)

	return nil
}

func (pbbi *PeerBlockBodyInterceptor) processPeerChangeBlockBody(peerChBlockBody *block.InterceptedPeerBlockBody) {
	err := pbbi.storer.Has(peerChBlockBody.Hash())
	isBlockInStorage := err == nil
	if isBlockInStorage {
		log.Debug("intercepted peer change block body already processed")
		return
	}

	pbbi.cache.HasOrAdd(peerChBlockBody.Hash(), peerChBlockBody)
}
