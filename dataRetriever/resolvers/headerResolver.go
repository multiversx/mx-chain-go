package resolvers

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/epochproviders/disabled"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("dataRetriever/resolvers")

var _ dataRetriever.HeaderResolver = (*HeaderResolver)(nil)

// ArgHeaderResolver is the argument structure used to create new HeaderResolver instance
type ArgHeaderResolver struct {
	ArgBaseResolver
	Headers              dataRetriever.HeadersPool
	HdrStorage           storage.Storer
	HeadersNoncesStorage storage.Storer
	NonceConverter       typeConverters.Uint64ByteSliceConverter
	ShardCoordinator     sharding.Coordinator
	IsFullHistoryNode    bool
}

// HeaderResolver is a wrapper over Resolver that is specialized in resolving headers requests
type HeaderResolver struct {
	*baseResolver
	baseStorageResolver
	messageProcessor
	headers          dataRetriever.HeadersPool
	hdrNoncesStorage storage.Storer
	nonceConverter   typeConverters.Uint64ByteSliceConverter
	mutEpochHandler  sync.RWMutex
	epochHandler     dataRetriever.EpochHandler
	shardCoordinator sharding.Coordinator
}

// NewHeaderResolver creates a new header resolver
func NewHeaderResolver(arg ArgHeaderResolver) (*HeaderResolver, error) {
	err := checkArgHeaderResolver(arg)
	if err != nil {
		return nil, err
	}

	epochHandler := disabled.NewEpochHandler()
	hdrResolver := &HeaderResolver{
		baseResolver: &baseResolver{
			TopicResolverSender: arg.SenderResolver,
		},
		headers:             arg.Headers,
		baseStorageResolver: createBaseStorageResolver(arg.HdrStorage, arg.IsFullHistoryNode),
		hdrNoncesStorage:    arg.HeadersNoncesStorage,
		nonceConverter:      arg.NonceConverter,
		epochHandler:        epochHandler,
		shardCoordinator:    arg.ShardCoordinator,
		messageProcessor: messageProcessor{
			marshalizer:      arg.Marshaller,
			antifloodHandler: arg.AntifloodHandler,
			topic:            arg.SenderResolver.RequestTopic(),
			throttler:        arg.Throttler,
		},
	}

	return hdrResolver, nil
}

func checkArgHeaderResolver(arg ArgHeaderResolver) error {
	err := checkArgBase(arg.ArgBaseResolver)
	if err != nil {
		return err
	}
	if check.IfNil(arg.Headers) {
		return dataRetriever.ErrNilHeadersDataPool
	}
	if check.IfNil(arg.HdrStorage) {
		return dataRetriever.ErrNilHeadersStorage
	}
	if check.IfNil(arg.HeadersNoncesStorage) {
		return dataRetriever.ErrNilHeadersNoncesStorage
	}
	if check.IfNil(arg.NonceConverter) {
		return dataRetriever.ErrNilUint64ByteSliceConverter
	}
	if check.IfNil(arg.ShardCoordinator) {
		return dataRetriever.ErrNilShardCoordinator
	}
	return nil
}

// SetEpochHandler sets the epoch handler for this component
func (hdrRes *HeaderResolver) SetEpochHandler(epochHandler dataRetriever.EpochHandler) error {
	if check.IfNil(epochHandler) {
		return dataRetriever.ErrNilEpochHandler
	}

	hdrRes.mutEpochHandler.Lock()
	hdrRes.epochHandler = epochHandler
	hdrRes.mutEpochHandler.Unlock()

	return nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (hdrRes *HeaderResolver) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	err := hdrRes.canProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	hdrRes.throttler.StartProcessing()
	defer hdrRes.throttler.EndProcessing()

	rd, err := hdrRes.parseReceivedMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	var buff []byte

	switch rd.Type {
	case dataRetriever.HashType:
		buff, err = hdrRes.resolveHeaderFromHash(rd)
	case dataRetriever.NonceType:
		buff, err = hdrRes.resolveHeaderFromNonce(rd)
	case dataRetriever.EpochType:
		buff, err = hdrRes.resolveHeaderFromEpoch(rd.Value)
	default:
		return dataRetriever.ErrResolveTypeUnknown
	}
	if err != nil {
		hdrRes.DebugHandler().LogFailedToResolveData(
			hdrRes.topic,
			rd.Value,
			err,
		)
		return err
	}

	if buff == nil {
		hdrRes.DebugHandler().LogFailedToResolveData(
			hdrRes.topic,
			rd.Value,
			dataRetriever.ErrMissingData,
		)

		log.Trace("missing data",
			"data", rd)
		return nil
	}

	hdrRes.DebugHandler().LogSucceededToResolveData(hdrRes.topic, rd.Value)

	return hdrRes.Send(buff, message.Peer())
}

func (hdrRes *HeaderResolver) resolveHeaderFromNonce(rd *dataRetriever.RequestData) ([]byte, error) {
	// key is now an encoded nonce (uint64)
	nonce, err := hdrRes.nonceConverter.ToUint64(rd.Value)
	if err != nil {
		return nil, dataRetriever.ErrInvalidNonceByteSlice
	}

	epoch := rd.Epoch

	// header-nonces storer contains un-pruned data, so it is safe to search like this
	hash, err := hdrRes.hdrNoncesStorage.SearchFirst(rd.Value)
	if err != nil {
		log.Trace("hdrNoncesStorage.Get from calculated epoch", "error", err.Error())
		// Search the nonce-key pair in data pool
		var hdrBytes []byte
		hdrBytes, err = hdrRes.searchInCache(nonce)
		if err != nil {
			return nil, err
		}

		return hdrBytes, nil
	}

	newRd := &dataRetriever.RequestData{
		Type:  rd.Type,
		Value: hash,
		Epoch: epoch,
	}

	return hdrRes.resolveHeaderFromHash(newRd)
}

func (hdrRes *HeaderResolver) searchInCache(nonce uint64) ([]byte, error) {
	headers, _, err := hdrRes.headers.GetHeadersByNonceAndShardId(nonce, hdrRes.TargetShardID())
	if err != nil {
		return nil, err
	}

	hdr := headers[len(headers)-1]
	buff, err := hdrRes.marshalizer.Marshal(hdr)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

// resolveHeaderFromHash resolves a header using its key (header hash)
func (hdrRes *HeaderResolver) resolveHeaderFromHash(rd *dataRetriever.RequestData) ([]byte, error) {
	value, err := hdrRes.headers.GetHeaderByHash(rd.Value)
	if err != nil {
		return hdrRes.getFromStorage(rd.Value, rd.Epoch)
	}

	return hdrRes.marshalizer.Marshal(value)
}

// resolveHeaderFromEpoch resolves a header using its key based on epoch
func (hdrRes *HeaderResolver) resolveHeaderFromEpoch(key []byte) ([]byte, error) {
	actualKey := key

	isUnknownEpoch, err := core.IsUnknownEpochIdentifier(key)
	if err != nil {
		return nil, err
	}
	if isUnknownEpoch {
		hdrRes.mutEpochHandler.RLock()
		metaEpoch := hdrRes.epochHandler.MetaEpoch()
		hdrRes.mutEpochHandler.RUnlock()

		actualKey = []byte(core.EpochStartIdentifier(metaEpoch))
	}

	return hdrRes.searchFirst(actualKey)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hdrRes *HeaderResolver) IsInterfaceNil() bool {
	return hdrRes == nil
}
