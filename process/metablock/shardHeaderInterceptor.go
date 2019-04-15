package metablock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var log = logger.DefaultLogger()

// ShardHeaderInterceptor represents an interceptor used for shard block headers by metachain nodes
type ShardHeaderInterceptor struct {
	hdrInterceptorBase *interceptors.HeaderInterceptorBase
	headers            storage.Cacher
}

// NewShardHeaderInterceptor hooks a new interceptor for shard block headers by metachain nodes
// Fetched block headers will be placed in a data pool
func NewShardHeaderInterceptor(
	marshalizer marshal.Marshalizer,
	headers storage.Cacher,
	storer storage.Storer,
	multiSigVerifier crypto.MultiSigVerifier,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
) (*ShardHeaderInterceptor, error) {

	if headers == nil {
		return nil, process.ErrNilHeadersDataPool
	}

	hdrBaseInterceptor, err := interceptors.NewHeaderInterceptorBase(
		marshalizer,
		storer,
		multiSigVerifier,
		hasher,
		shardCoordinator,
	)
	if err != nil {
		return nil, err
	}

	return &ShardHeaderInterceptor{
		hdrInterceptorBase: hdrBaseInterceptor,
		headers:            headers,
	}, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (shi *ShardHeaderInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	hdrIntercepted, err := shi.hdrInterceptorBase.ParseReceivedMessage(message)
	if err == process.ErrHeaderIsInStorage {
		log.Debug("intercepted block header already processed")
		return nil
	}
	if err != nil {
		return err
	}

	_, _ = shi.headers.HasOrAdd(hdrIntercepted.Hash(), hdrIntercepted.GetHeader())
	return nil
}
