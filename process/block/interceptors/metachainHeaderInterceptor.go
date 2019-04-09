package interceptors

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// MetachainHeaderInterceptor represents an interceptor used for metachain block headers
type MetachainHeaderInterceptor struct {
	*messageChecker
	marshalizer      marshal.Marshalizer
	metachainHeaders storage.Cacher
	storer           storage.Storer
	multiSigVerifier crypto.MultiSigVerifier
	hasher           hashing.Hasher
	shardCoordinator sharding.Coordinator
}

// NewMetachainHeaderInterceptor hooks a new interceptor for metachain block headers
// Fetched metachain block headers will be placed in a data pool
func NewMetachainHeaderInterceptor(
	marshalizer marshal.Marshalizer,
	metachainHeaders storage.Cacher,
	storer storage.Storer,
	multiSigVerifier crypto.MultiSigVerifier,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
) (*MetachainHeaderInterceptor, error) {
	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if metachainHeaders == nil {
		return nil, process.ErrNilMetachainHeadersDataPool
	}
	if storer == nil {
		return nil, process.ErrNilMetachainHeadersStorage
	}
	if multiSigVerifier == nil {
		return nil, process.ErrNilMultiSigVerifier
	}
	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}

	return &MetachainHeaderInterceptor{
		messageChecker:   &messageChecker{},
		marshalizer:      marshalizer,
		metachainHeaders: metachainHeaders,
		storer:           storer,
		multiSigVerifier: multiSigVerifier,
		hasher:           hasher,
		shardCoordinator: shardCoordinator,
	}, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (mhi *MetachainHeaderInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	err := mhi.checkMessage(message)
	if err != nil {
		return err
	}

	//TODO: modify this to actually fetch metachain blocks. For now, we place shard headers on this topic
	hdrIntercepted := block.NewInterceptedHeader(mhi.multiSigVerifier)
	err = mhi.marshalizer.Unmarshal(hdrIntercepted, message.Data())
	if err != nil {
		return err
	}

	hashWithSig := mhi.hasher.Compute(string(message.Data()))
	hdrIntercepted.SetHash(hashWithSig)

	err = hdrIntercepted.IntegrityAndValidity(mhi.shardCoordinator)
	if err != nil {
		return err
	}

	err = hdrIntercepted.VerifySig()
	if err != nil {
		return err
	}

	isHeaderInStorage, _ := mhi.storer.Has(hashWithSig)
	if isHeaderInStorage {
		log.Debug("intercepted block header already processed")
		return nil
	}

	mhi.metachainHeaders.HasOrAdd(hashWithSig, hdrIntercepted.GetHeader())
	return nil
}
