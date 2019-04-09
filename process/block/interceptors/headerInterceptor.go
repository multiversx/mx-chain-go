package interceptors

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// HeaderInterceptor represents an interceptor used for block headers
type HeaderInterceptor struct {
	hdrInterceptorBase *HeaderInterceptorBase
	headers            storage.Cacher
	headersNonces      data.Uint64Cacher
	shardCoordinator   sharding.Coordinator
}

// NewHeaderInterceptor hooks a new interceptor for block headers
// Fetched block headers will be placed in a data pool
func NewHeaderInterceptor(
	marshalizer marshal.Marshalizer,
	headers storage.Cacher,
	headersNonces data.Uint64Cacher,
	storer storage.Storer,
	multiSigVerifier crypto.MultiSigVerifier,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
) (*HeaderInterceptor, error) {

	if headersNonces == nil {
		return nil, process.ErrNilHeadersNoncesDataPool
	}
	if headers == nil {
		return nil, process.ErrNilHeadersDataPool
	}
	hdrBaseInterceptor, err := NewHeaderInterceptorBase(
		marshalizer,
		storer,
		multiSigVerifier,
		hasher,
		shardCoordinator,
	)
	if err != nil {
		return nil, err
	}

	hdrIntercept := &HeaderInterceptor{
		hdrInterceptorBase: hdrBaseInterceptor,
		headers:            headers,
		headersNonces:      headersNonces,
		shardCoordinator:   shardCoordinator,
	}

	return hdrIntercept, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (hi *HeaderInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	hdrIntercepted, err := hi.hdrInterceptorBase.ParseReceivedMessage(message)
	if err == process.ErrHeaderIsInStorage {
		log.Debug("intercepted block header already processed")
		return nil
	}
	if err != nil {
		return err
	}
	if !hi.checkHeaderForCurrentShard(hdrIntercepted) {
		return nil
	}

	_, _ = hi.headers.HasOrAdd(hdrIntercepted.Hash(), hdrIntercepted.GetHeader())
	_, _ = hi.headersNonces.HasOrAdd(hdrIntercepted.GetHeader().Nonce, hdrIntercepted.Hash())
	return nil
}

func (hi *HeaderInterceptor) checkHeaderForCurrentShard(header *block.InterceptedHeader) bool {
	return hi.shardCoordinator.SelfId() == header.GetHeader().ShardId
}
