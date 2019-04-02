package metablock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var log = logger.NewDefaultLogger()

// ShardHeaderInterceptor represents an interceptor used for shard block headers by metachain nodes
type ShardHeaderInterceptor struct {
	marshalizer      marshal.Marshalizer
	headers          storage.Cacher
	storer           storage.Storer
	multiSigVerifier crypto.MultiSigVerifier
	hasher           hashing.Hasher
	shardCoordinator sharding.Coordinator
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
	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
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

	hdrIntercept := &ShardHeaderInterceptor{
		marshalizer:      marshalizer,
		headers:          headers,
		storer:           storer,
		multiSigVerifier: multiSigVerifier,
		hasher:           hasher,
		shardCoordinator: shardCoordinator,
	}

	return hdrIntercept, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (shi *ShardHeaderInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	if message == nil {
		return process.ErrNilMessage
	}
	if message.Data() == nil {
		return process.ErrNilDataToProcess
	}

	hdrIntercepted := block.NewInterceptedHeader(shi.multiSigVerifier)
	err := shi.marshalizer.Unmarshal(hdrIntercepted, message.Data())
	if err != nil {
		return err
	}

	hashWithSig := shi.hasher.Compute(string(message.Data()))
	hdrIntercepted.SetHash(hashWithSig)

	err = hdrIntercepted.IntegrityAndValidity(shi.shardCoordinator)
	if err != nil {
		return err
	}

	err = hdrIntercepted.VerifySig()
	if err != nil {
		return err
	}

	isHeaderInStorage, _ := shi.storer.Has(hashWithSig)
	if isHeaderInStorage {
		log.Debug("intercepted block header already processed")
		return nil
	}

	_, _ = shi.headers.HasOrAdd(hashWithSig, hdrIntercepted.GetHeader())
	return nil
}
