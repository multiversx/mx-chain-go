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
	headerValidator        process.HeaderValidator
	multiSigVerifier       crypto.MultiSigVerifier
	hasher                 hashing.Hasher
	shardCoordinator       sharding.Coordinator
	chronologyValidator    process.ChronologyValidator
}

// NewMetachainHeaderInterceptor hooks a new interceptor for metachain block headers
// Fetched metachain block headers will be placed in a data pool
func NewMetachainHeaderInterceptor(
	marshalizer marshal.Marshalizer,
	metachainHeaders storage.Cacher,
	metachainHeadersNonces dataRetriever.Uint64SyncMapCacher,
	headerValidator process.HeaderValidator,
	multiSigVerifier crypto.MultiSigVerifier,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
	chronologyValidator process.ChronologyValidator,
) (*MetachainHeaderInterceptor, error) {

	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}
	if metachainHeaders == nil || metachainHeaders.IsInterfaceNil() {
		return nil, process.ErrNilMetaHeadersDataPool
	}
	if metachainHeadersNonces == nil || metachainHeadersNonces.IsInterfaceNil() {
		return nil, process.ErrNilMetaHeadersNoncesDataPool
	}
	if headerValidator == nil || headerValidator.IsInterfaceNil() {
		return nil, process.ErrNilHeaderHandlerValidator
	}
	if multiSigVerifier == nil || multiSigVerifier.IsInterfaceNil() {
		return nil, process.ErrNilMultiSigVerifier
	}
	if hasher == nil || hasher.IsInterfaceNil() {
		return nil, process.ErrNilHasher
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if chronologyValidator == nil || chronologyValidator.IsInterfaceNil() {
		return nil, process.ErrNilChronologyValidator
	}

	return &MetachainHeaderInterceptor{
		messageChecker:         &messageChecker{},
		marshalizer:            marshalizer,
		metachainHeaders:       metachainHeaders,
		headerValidator:        headerValidator,
		multiSigVerifier:       multiSigVerifier,
		hasher:                 hasher,
		shardCoordinator:       shardCoordinator,
		chronologyValidator:    chronologyValidator,
		metachainHeadersNonces: metachainHeadersNonces,
	}, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (mhi *MetachainHeaderInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	err := mhi.checkMessage(message)
	if err != nil {
		return err
	}

	metaHdrIntercepted := block.NewInterceptedMetaHeader(mhi.multiSigVerifier, mhi.chronologyValidator)
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
	err := mhi.headerValidator.HeaderValidForProcessing(metaHdrIntercepted)
	if err != nil {
		log.Debug("intercepted meta block header already processed: " + err.Error())
		return
	}

	mhi.metachainHeaders.HasOrAdd(metaHdrIntercepted.Hash(), metaHdrIntercepted.HeaderHandler())

	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(sharding.MetachainShardId, metaHdrIntercepted.Hash())
	mhi.metachainHeadersNonces.Merge(metaHdrIntercepted.Nonce, syncMap)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mhi *MetachainHeaderInterceptor) IsInterfaceNil() bool {
	if mhi == nil {
		return true
	}
	return false
}
