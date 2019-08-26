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
	headerValidator     process.HeaderValidator
	shardCoordinator    sharding.Coordinator
}

// NewHeaderInterceptor hooks a new interceptor for block headers
// Fetched block headers will be placed in a data pool
func NewHeaderInterceptor(
	marshalizer marshal.Marshalizer,
	headers storage.Cacher,
	headersNonces dataRetriever.Uint64SyncMapCacher,
	headerValidator process.HeaderValidator,
	multiSigVerifier crypto.MultiSigVerifier,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
	chronologyValidator process.ChronologyValidator,
) (*HeaderInterceptor, error) {

	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}
	if headersNonces == nil || headersNonces.IsInterfaceNil() {
		return nil, process.ErrNilHeadersNoncesDataPool
	}
	if headers == nil || headers.IsInterfaceNil() {
		return nil, process.ErrNilHeadersDataPool
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

	hdrInterceptor := &HeaderInterceptor{
		marshalizer:         marshalizer,
		multiSigVerifier:    multiSigVerifier,
		hasher:              hasher,
		shardCoordinator:    shardCoordinator,
		chronologyValidator: chronologyValidator,
		headers:             headers,
		headersNonces:       headersNonces,
		headerValidator:     headerValidator,
	}

	return hdrInterceptor, nil
}

// ParseReceivedMessage will transform the received p2p.Message in an InterceptedHeader.
// If the header hash is present in storage it will output an error
func (hi *HeaderInterceptor) ParseReceivedMessage(message p2p.MessageP2P) (*block.InterceptedHeader, error) {
	if message == nil || message.IsInterfaceNil() {
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
func (hi *HeaderInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	hdrIntercepted, err := hi.ParseReceivedMessage(message)
	if err != nil {
		return err
	}

	go hi.processHeader(hdrIntercepted)

	return nil
}

// checkHeaderForCurrentShard checks if the header is for current shard
func (hi *HeaderInterceptor) checkHeaderForCurrentShard(interceptedHdr *block.InterceptedHeader) bool {
	isHeaderForCurrentShard := hi.shardCoordinator.SelfId() == interceptedHdr.GetHeader().ShardId
	isMetachainShardCoordinator := hi.shardCoordinator.SelfId() == sharding.MetachainShardId

	return isHeaderForCurrentShard || isMetachainShardCoordinator
}

func (hi *HeaderInterceptor) processHeader(hdrIntercepted *block.InterceptedHeader) {
	if !hi.checkHeaderForCurrentShard(hdrIntercepted) {
		return
	}

	isHeaderOkForProcessing := hi.headerValidator.IsHeaderValidForProcessing(hdrIntercepted.Header)
	if !isHeaderOkForProcessing {
		log.Debug("intercepted block header can not be processed")
		return
	}

	hi.headers.HasOrAdd(hdrIntercepted.Hash(), hdrIntercepted.GetHeader())

	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(hdrIntercepted.ShardId, hdrIntercepted.Hash())
	hi.headersNonces.Merge(hdrIntercepted.Nonce, syncMap)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hi *HeaderInterceptor) IsInterfaceNil() bool {
	if hi == nil {
		return true
	}
	return false
}
