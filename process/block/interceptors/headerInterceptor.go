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
	*messageChecker
	marshalizer      marshal.Marshalizer
	headers          data.ShardedDataCacherNotifier
	storer           storage.Storer
	headersNonces    data.Uint64Cacher
	multiSigVerifier crypto.MultiSigVerifier
	hasher           hashing.Hasher
	shardCoordinator sharding.ShardCoordinator
}

// NewHeaderInterceptor hooks a new interceptor for block headers
// Fetched block headers will be placed in a data pool
func NewHeaderInterceptor(
	marshalizer marshal.Marshalizer,
	headers data.ShardedDataCacherNotifier,
	headersNonces data.Uint64Cacher,
	storer storage.Storer,
	multiSigVerifier crypto.MultiSigVerifier,
	hasher hashing.Hasher,
	shardCoordinator sharding.ShardCoordinator,
) (*HeaderInterceptor, error) {
	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}

	if headers == nil {
		return nil, process.ErrNilHeadersDataPool
	}

	if headersNonces == nil {
		return nil, process.ErrNilHeadersNoncesDataPool
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

	hdrIntercept := &HeaderInterceptor{
		messageChecker:   &messageChecker{},
		marshalizer:      marshalizer,
		headers:          headers,
		headersNonces:    headersNonces,
		storer:           storer,
		multiSigVerifier: multiSigVerifier,
		hasher:           hasher,
		shardCoordinator: shardCoordinator,
	}

	return hdrIntercept, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (hi *HeaderInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	err := hi.checkMessage(message)
	if err != nil {
		return err
	}

	hdrIntercepted := block.NewInterceptedHeader(hi.multiSigVerifier)
	err = hi.marshalizer.Unmarshal(hdrIntercepted, message.Data())
	if err != nil {
		return err
	}

	hashWithSig := hi.hasher.Compute(string(message.Data()))
	hdrIntercepted.SetHash(hashWithSig)

	err = hdrIntercepted.IntegrityAndValidity(hi.shardCoordinator)
	if err != nil {
		return err
	}

	err = hdrIntercepted.VerifySig()
	if err != nil {
		return err
	}

	isHeaderInStorage, _ := hi.storer.Has(hashWithSig)

	if isHeaderInStorage {
		log.Debug("intercepted block header already processed")
		return nil
	}

	hi.headers.AddData(hashWithSig, hdrIntercepted.GetHeader(), hdrIntercepted.Shard())
	if hi.checkHeaderForCurrentShard(hdrIntercepted) {
		_, _ = hi.headersNonces.HasOrAdd(hdrIntercepted.GetHeader().Nonce, hashWithSig)
	}
	return nil
}

func (hi *HeaderInterceptor) checkHeaderForCurrentShard(header *block.InterceptedHeader) bool {
	return hi.shardCoordinator.ShardForCurrentNode() == header.GetHeader().ShardId
}
