package interceptors

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// HeaderInterceptor represents an interceptor used for block headers
type HeaderInterceptor struct {
	hdrInterceptorBase *HeaderInterceptorBase
	headers            storage.Cacher
	headersNonces      dataRetriever.Uint64SyncMapCacher
	shardCoordinator   sharding.Coordinator
}

// NewHeaderInterceptor hooks a new interceptor for block headers
// Fetched block headers will be placed in a data pool
func NewHeaderInterceptor(
	marshalizer marshal.Marshalizer,
	headers storage.Cacher,
	headersNonces dataRetriever.Uint64SyncMapCacher,
	headerValidator process.HeaderHandlerProcessValidator,
	multiSigVerifier crypto.MultiSigVerifier,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
	chronologyValidator process.ChronologyValidator,
) (*HeaderInterceptor, error) {

	if headersNonces == nil {
		return nil, process.ErrNilHeadersNoncesDataPool
	}
	if headers == nil {
		return nil, process.ErrNilHeadersDataPool
	}
	hdrBaseInterceptor, err := NewHeaderInterceptorBase(
		marshalizer,
		headerValidator,
		multiSigVerifier,
		hasher,
		shardCoordinator,
		chronologyValidator,
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
	if err != nil {
		return err
	}

	go hi.processHeader(hdrIntercepted)

	return nil
}

func (hi *HeaderInterceptor) processHeader(hdrIntercepted *block.InterceptedHeader) {
	if !hi.hdrInterceptorBase.CheckHeaderForCurrentShard(hdrIntercepted) {
		return
	}

	isHeaderOkForProcessing := hi.hdrInterceptorBase.headerValidator.CheckHeaderHandlerValid(hdrIntercepted.Header)
	if !isHeaderOkForProcessing {
		log.Debug("intercepted block header can not be processed")
		return
	}

	hi.headers.HasOrAdd(hdrIntercepted.Hash(), hdrIntercepted.GetHeader())

	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(hdrIntercepted.ShardId, hdrIntercepted.Hash())
	hi.headersNonces.Merge(hdrIntercepted.Nonce, syncMap)
}
