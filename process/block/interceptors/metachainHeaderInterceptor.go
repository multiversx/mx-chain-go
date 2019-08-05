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

// MetachainHeaderInterceptor represents an interceptor used for metachain block headers
type MetachainHeaderInterceptor struct {
	*messageChecker
	marshalizer            marshal.Marshalizer
	metachainHeaders       storage.Cacher
	metachainHeadersNonces dataRetriever.Uint64SyncMapCacher
	storer                 storage.Storer
	multiSigVerifier       crypto.MultiSigVerifier
	hasher                 hashing.Hasher
	shardCoordinator       sharding.Coordinator
	nodesCoordinator       sharding.NodesCoordinator
}

// NewMetachainHeaderInterceptor hooks a new interceptor for metachain block headers
// Fetched metachain block headers will be placed in a data pool
func NewMetachainHeaderInterceptor(
	marshalizer marshal.Marshalizer,
	metachainHeaders storage.Cacher,
	metachainHeadersNonces dataRetriever.Uint64SyncMapCacher,
	storer storage.Storer,
	multiSigVerifier crypto.MultiSigVerifier,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
) (*MetachainHeaderInterceptor, error) {

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if metachainHeaders == nil {
		return nil, process.ErrNilMetaHeadersDataPool
	}
	if metachainHeadersNonces == nil {
		return nil, process.ErrNilMetaHeadersNoncesDataPool
	}
	if storer == nil {
		return nil, process.ErrNilMetaHeadersStorage
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
	if nodesCoordinator == nil {
		return nil, process.ErrNilNodesCoordinator
	}

	return &MetachainHeaderInterceptor{
		messageChecker:         &messageChecker{},
		marshalizer:            marshalizer,
		metachainHeaders:       metachainHeaders,
		storer:                 storer,
		multiSigVerifier:       multiSigVerifier,
		hasher:                 hasher,
		shardCoordinator:       shardCoordinator,
		nodesCoordinator:       nodesCoordinator,
		metachainHeadersNonces: metachainHeadersNonces,
	}, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (mhi *MetachainHeaderInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	err := mhi.checkMessage(message)
	if err != nil {
		return err
	}

	metaHdrIntercepted := block.NewInterceptedMetaHeader(
		mhi.multiSigVerifier,
		mhi.nodesCoordinator,
		mhi.marshalizer,
		mhi.hasher,
	)
	err = mhi.marshalizer.Unmarshal(metaHdrIntercepted, message.Data())
	if err != nil {
		return err
	}

	hashWithSig := mhi.hasher.Compute(string(message.Data()))
	metaHdrIntercepted.SetHash(hashWithSig)

	err = metaHdrIntercepted.IntegrityAndValidity(mhi.shardCoordinator)
	if err != nil {
		return err
	}

	err = metaHdrIntercepted.VerifySig()
	if err != nil {
		return err
	}

	go mhi.processMetaHeader(metaHdrIntercepted)

	return nil
}

func (mhi *MetachainHeaderInterceptor) processMetaHeader(metaHdrIntercepted *block.InterceptedMetaHeader) {
	err := mhi.storer.Has(metaHdrIntercepted.Hash())
	isHeaderInStorage := err == nil
	if isHeaderInStorage {
		log.Debug("intercepted meta block header already processed")
		return
	}

	mhi.metachainHeaders.HasOrAdd(metaHdrIntercepted.Hash(), metaHdrIntercepted.GetMetaHeader())

	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(sharding.MetachainShardId, metaHdrIntercepted.Hash())
	mhi.metachainHeadersNonces.Merge(metaHdrIntercepted.Nonce, syncMap)
}
