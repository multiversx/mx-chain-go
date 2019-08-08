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
	marshalizer         marshal.Marshalizer
	storer              storage.Storer
	multiSigVerifier    crypto.MultiSigVerifier
	hasher              hashing.Hasher
	chronologyValidator process.ChronologyValidator
	headers             storage.Cacher
	headersNonces       dataRetriever.Uint64SyncMapCacher
	shardCoordinator    sharding.Coordinator
}

// NewHeaderInterceptor hooks a new interceptor for block headers
// Fetched block headers will be placed in a data pool
func NewHeaderInterceptor(
	marshalizer marshal.Marshalizer,
	headers storage.Cacher,
	headersNonces dataRetriever.Uint64SyncMapCacher,
	storer storage.Storer,
	multiSigVerifier crypto.MultiSigVerifier,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
	chronologyValidator process.ChronologyValidator,
) (*HeaderInterceptor, error) {

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if headersNonces == nil {
		return nil, process.ErrNilHeadersNoncesDataPool
	}
	if headers == nil {
		return nil, process.ErrNilHeadersDataPool
	}
	if storer == nil {
		return nil, process.ErrNilHeadersStorage
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
	if chronologyValidator == nil {
		return nil, process.ErrNilChronologyValidator
	}

	hdrInterceptor := &HeaderInterceptor{
		marshalizer:         marshalizer,
		storer:              storer,
		multiSigVerifier:    multiSigVerifier,
		hasher:              hasher,
		shardCoordinator:    shardCoordinator,
		chronologyValidator: chronologyValidator,
		headers:             headers,
		headersNonces:       headersNonces,
	}

	return hdrInterceptor, nil
}

// ParseReceivedMessage will transform the received p2p.Message in an InterceptedHeader.
// If the header hash is present in storage it will output an error
func (hi *HeaderInterceptor) ParseReceivedMessage(message p2p.MessageP2P) (*block.InterceptedHeader, error) {
	if message == nil {
		return nil, process.ErrNilMessage
	}
	if message.Data() == nil {
		return nil, process.ErrNilDataToProcess
	}

	hdrIntercepted := block.NewInterceptedHeader(hi.multiSigVerifier, hi.chronologyValidator)
	err := hi.marshalizer.Unmarshal(hdrIntercepted, message.Data())
	if err != nil {
		return nil, err
	}

	hashWithSig := hi.hasher.Compute(string(message.Data()))
	hdrIntercepted.SetHash(hashWithSig)

	err = hdrIntercepted.IntegrityAndValidity(hi.shardCoordinator)
	if err != nil {
		return nil, err
	}

	err = hdrIntercepted.VerifySig()
	if err != nil {
		return nil, err
	}

	return hdrIntercepted, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (hi *HeaderInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	hdrIntercepted, err := hi.ParseReceivedMessage(message)
	if err != nil {
		return err
	}

	go hi.processHeader(hdrIntercepted)

	return nil
}

// CheckHeaderForCurrentShard checks if the header is for current shard
func (hi *HeaderInterceptor) checkHeaderForCurrentShard(interceptedHdr *block.InterceptedHeader) bool {
	isHeaderForCurrentShard := hi.shardCoordinator.SelfId() == interceptedHdr.GetHeader().ShardId
	isMetachainShardCoordinator := hi.shardCoordinator.SelfId() == sharding.MetachainShardId

	return isHeaderForCurrentShard || isMetachainShardCoordinator
}

func (hi *HeaderInterceptor) processHeader(hdrIntercepted *block.InterceptedHeader) {
	if !hi.checkHeaderForCurrentShard(hdrIntercepted) {
		return
	}

	err := hi.storer.Has(hdrIntercepted.Hash())
	isHeaderInStorage := err == nil
	if isHeaderInStorage {
		log.Debug("intercepted block header already processed")
		return
	}

	hi.headers.HasOrAdd(hdrIntercepted.Hash(), hdrIntercepted.GetHeader())

	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(hdrIntercepted.ShardId, hdrIntercepted.Hash())
	hi.headersNonces.Merge(hdrIntercepted.Nonce, syncMap)
}
