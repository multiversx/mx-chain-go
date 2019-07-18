package metablock

import (
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptors"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.DefaultLogger()

// ShardHeaderInterceptor represents an interceptor used for shard block headers by metachain nodes
type ShardHeaderInterceptor struct {
	hdrInterceptorBase *interceptors.HeaderInterceptorBase
	headers            storage.Cacher
	hdrsNonces         dataRetriever.Uint64SyncMapCacher
	storer             storage.Storer
}

// NewShardHeaderInterceptor hooks a new interceptor for shard block headers by metachain nodes
// Fetched block headers will be placed in a data pool
func NewShardHeaderInterceptor(
	marshalizer marshal.Marshalizer,
	headers storage.Cacher,
	hdrsNonces dataRetriever.Uint64SyncMapCacher,
	storer storage.Storer,
	multiSigVerifier crypto.MultiSigVerifier,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
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
		nodesCoordinator,
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

	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(hdrIntercepted.GetHeader().GetShardID(), hdrIntercepted.Hash())
	shi.hdrsNonces.Merge(nonce, syncMap)
}
