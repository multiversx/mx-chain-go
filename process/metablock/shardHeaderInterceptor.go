package metablock

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var log = logger.DefaultLogger()

// ShardHeaderInterceptor represents an interceptor used for shard block headers by metachain nodes
type ShardHeaderInterceptor struct {
	hdrInterceptorBase *interceptors.HeaderInterceptorBase
	headers            storage.Cacher
	hdrsNonces         dataRetriever.Uint64Cacher
	storer             storage.Storer
}

// NewShardHeaderInterceptor hooks a new interceptor for shard block headers by metachain nodes
// Fetched block headers will be placed in a data pool
func NewShardHeaderInterceptor(
	marshalizer marshal.Marshalizer,
	headers storage.Cacher,
	hdrsNonces dataRetriever.Uint64Cacher,
	storer storage.Storer,
	multiSigVerifier crypto.MultiSigVerifier,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
	chronologyValidator process.ChronologyValidator,
) (*ShardHeaderInterceptor, error) {

	if headers == nil {
		return nil, process.ErrNilHeadersDataPool
	}
	if hdrsNonces == nil {
		return nil, process.ErrNilHeadersNoncesDataPool
	}

	hdrBaseInterceptor, err := interceptors.NewHeaderInterceptorBase(
		marshalizer,
		storer,
		multiSigVerifier,
		hasher,
		shardCoordinator,
		chronologyValidator,
	)
	if err != nil {
		return nil, err
	}

	return &ShardHeaderInterceptor{
		hdrInterceptorBase: hdrBaseInterceptor,
		headers:            headers,
		hdrsNonces:         hdrsNonces,
		storer:             storer,
	}, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (shi *ShardHeaderInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	hdrIntercepted, err := shi.hdrInterceptorBase.ParseReceivedMessage(message)
	if err != nil {
		return err
	}

	go shi.processHeader(hdrIntercepted)

	return nil
}

func (shi *ShardHeaderInterceptor) processHeader(hdrIntercepted *block.InterceptedHeader) {
	if !shi.hdrInterceptorBase.CheckHeaderForCurrentShard(hdrIntercepted) {
		return
	}

	err := shi.storer.Has(hdrIntercepted.Hash())
	isHeaderInStorage := err == nil
	if isHeaderInStorage {
		log.Debug("intercepted block header already processed")
		return
	}

	shi.headers.HasOrAdd(hdrIntercepted.Hash(), hdrIntercepted.GetHeader())

	nonce := hdrIntercepted.GetHeader().GetNonce()

	value, okPeek := shi.hdrsNonces.Peek(nonce)
	mapOfHashes, okTypeAssertion := value.(*sync.Map)

	if !okPeek || !okTypeAssertion {
		// repair saved data
		mapOfHashes = &sync.Map{}
	}

	mapOfHashes.Store(hdrIntercepted.GetHeader().GetShardID(), hdrIntercepted.Hash())
	shi.hdrsNonces.Put(nonce, mapOfHashes)
}
