package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// HeaderInterceptor represents an interceptor used for block headers
type HeaderInterceptor struct {
	process.Interceptor
	headers          data.ShardedDataCacherNotifier
	storer           storage.Storer
	headersNonces    data.Uint64Cacher
	hasher           hashing.Hasher
	shardCoordinator sharding.ShardCoordinator
}

// GenericBlockBodyInterceptor represents an interceptor used for all types of block bodies
type GenericBlockBodyInterceptor struct {
	process.Interceptor
	cache            storage.Cacher
	hasher           hashing.Hasher
	storer           storage.Storer
	shardCoordinator sharding.ShardCoordinator
}

//------- HeaderInterceptor

// NewHeaderInterceptor hooks a new interceptor for block headers
// Fetched block headers will be placed in a data pool
func NewHeaderInterceptor(
	interceptor process.Interceptor,
	headers data.ShardedDataCacherNotifier,
	headersNonces data.Uint64Cacher,
	storer storage.Storer,
	hasher hashing.Hasher,
	shardCoordinator sharding.ShardCoordinator,
) (*HeaderInterceptor, error) {

	if interceptor == nil {
		return nil, process.ErrNilInterceptor
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

	if hasher == nil {
		return nil, process.ErrNilHasher
	}

	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}

	hdrIntercept := &HeaderInterceptor{
		Interceptor:      interceptor,
		headers:          headers,
		headersNonces:    headersNonces,
		storer:           storer,
		hasher:           hasher,
		shardCoordinator: shardCoordinator,
	}

	interceptor.SetReceivedMessageHandler(hdrIntercept.processHdr)

	return hdrIntercept, nil
}

func (hi *HeaderInterceptor) processHdr(message p2p.MessageP2P) error {
	err := checkMessage(message)
	if err != nil {
		return err
	}

	marshalizer := hi.Marshalizer()
	if marshalizer == nil {
		return process.ErrNilMarshalizer
	}

	hdrIntercepted := NewInterceptedHeader()
	err = marshalizer.Unmarshal(hdrIntercepted, message.Data())
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

func checkMessage(message p2p.MessageP2P) error {
	if message == nil {
		return process.ErrNilMessage
	}

	if message.Data() == nil {
		return process.ErrNilDataToProcess
	}

	return nil
}

func (hi *HeaderInterceptor) checkHeaderForCurrentShard(header *InterceptedHeader) bool {
	return hi.shardCoordinator.ShardForCurrentNode() == header.GetHeader().ShardId
}

//------- GenericBlockBodyInterceptor

// NewGenericBlockBodyInterceptor hooks a new interceptor for block bodies
// Fetched data blocks will be placed inside the cache
func NewGenericBlockBodyInterceptor(
	interceptor process.Interceptor,
	cache storage.Cacher,
	storer storage.Storer,
	hasher hashing.Hasher,
	shardCoordinator sharding.ShardCoordinator,
) (*GenericBlockBodyInterceptor, error) {

	if interceptor == nil {
		return nil, process.ErrNilInterceptor
	}

	if cache == nil {
		return nil, process.ErrNilCacher
	}

	if storer == nil {
		return nil, process.ErrNilBlockBodyStorage
	}

	if hasher == nil {
		return nil, process.ErrNilHasher
	}

	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}

	bbIntercept := &GenericBlockBodyInterceptor{
		Interceptor:      interceptor,
		cache:            cache,
		storer:           storer,
		hasher:           hasher,
		shardCoordinator: shardCoordinator,
	}

	return bbIntercept, nil
}

// ProcessTxBodyBlock process incoming message as tx block body
func (gbbi *GenericBlockBodyInterceptor) ProcessTxBodyBlock(message p2p.MessageP2P) error {
	err := checkMessage(message)
	if err != nil {
		return err
	}

	marshalizer := gbbi.Marshalizer()
	if marshalizer == nil {
		return process.ErrNilMarshalizer
	}

	txBlockBody := NewInterceptedTxBlockBody()
	err = marshalizer.Unmarshal(txBlockBody, message.Data())
	if err != nil {
		return err
	}

	return gbbi.processBlockBody(message.Data(), txBlockBody)
}

// ProcessStateBodyBlock process incoming message as state block body
func (gbbi *GenericBlockBodyInterceptor) ProcessStateBodyBlock(message p2p.MessageP2P) error {
	err := checkMessage(message)
	if err != nil {
		return err
	}

	marshalizer := gbbi.Marshalizer()
	if marshalizer == nil {
		return process.ErrNilMarshalizer
	}

	stateBlockBody := NewInterceptedStateBlockBody()
	err = marshalizer.Unmarshal(stateBlockBody, message.Data())
	if err != nil {
		return err
	}

	return gbbi.processBlockBody(message.Data(), stateBlockBody)
}

// ProcessPeerChangeBodyBlock process the message as peer change block body
func (gbbi *GenericBlockBodyInterceptor) ProcessPeerChangeBodyBlock(message p2p.MessageP2P) error {
	err := checkMessage(message)
	if err != nil {
		return err
	}

	marshalizer := gbbi.Marshalizer()
	if marshalizer == nil {
		return process.ErrNilMarshalizer
	}

	peerChangeBlockBody := NewInterceptedPeerBlockBody()
	err = marshalizer.Unmarshal(peerChangeBlockBody, message.Data())
	if err != nil {
		return err
	}

	return gbbi.processBlockBody(message.Data(), peerChangeBlockBody)
}

func (gbbi *GenericBlockBodyInterceptor) processBlockBody(messageData []byte, body process.InterceptedBlockBody) error {
	hash := gbbi.hasher.Compute(string(messageData))
	body.SetHash(hash)

	err := body.IntegrityAndValidity(gbbi.shardCoordinator)
	if err != nil {
		return err
	}

	isBlockInStorage, _ := gbbi.storer.Has(hash)

	if isBlockInStorage {
		log.Debug("intercepted block body already processed")
		return nil
	}

	_ = gbbi.cache.Put(hash, body.GetUnderlyingObject())
	return nil
}
