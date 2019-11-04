package resolvers

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.DefaultLogger()

// HeaderResolver is a wrapper over Resolver that is specialized in resolving headers requests
type HeaderResolver struct {
	dataRetriever.TopicResolverSender
	headers          storage.Cacher
	hdrNonces        dataRetriever.Uint64SyncMapCacher
	hdrStorage       storage.Storer
	hdrNoncesStorage storage.Storer
	marshalizer      marshal.Marshalizer
	nonceConverter   typeConverters.Uint64ByteSliceConverter
}

// NewHeaderResolver creates a new header resolver
func NewHeaderResolver(
	senderResolver dataRetriever.TopicResolverSender,
	headers storage.Cacher,
	headersNonces dataRetriever.Uint64SyncMapCacher,
	hdrStorage storage.Storer,
	headersNoncesStorage storage.Storer,
	marshalizer marshal.Marshalizer,
	nonceConverter typeConverters.Uint64ByteSliceConverter,
) (*HeaderResolver, error) {

	if senderResolver == nil || senderResolver.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilResolverSender
	}
	if headers == nil || headers.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilHeadersDataPool
	}
	if headersNonces == nil || headersNonces.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilHeadersNoncesDataPool
	}
	if hdrStorage == nil || hdrStorage.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilHeadersStorage
	}
	if headersNoncesStorage == nil || headersNoncesStorage.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilHeadersNoncesStorage
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if nonceConverter == nil || nonceConverter.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilUint64ByteSliceConverter
	}

	hdrResolver := &HeaderResolver{
		TopicResolverSender: senderResolver,
		headers:             headers,
		hdrNonces:           headersNonces,
		hdrStorage:          hdrStorage,
		hdrNoncesStorage:    headersNoncesStorage,
		marshalizer:         marshalizer,
		nonceConverter:      nonceConverter,
	}

	return hdrResolver, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (hdrRes *HeaderResolver) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	rd, err := hdrRes.parseReceivedMessage(message)
	if err != nil {
		return err
	}
	var buff []byte

	switch rd.Type {
	case dataRetriever.HashType:
		buff, err = hdrRes.resolveHeaderFromHash(rd.Value)
	case dataRetriever.NonceType:
		buff, err = hdrRes.resolveHeaderFromNonce(rd.Value)
	default:
		return dataRetriever.ErrResolveTypeUnknown
	}
	if err != nil {
		return err
	}
	if buff == nil {
		log.Debug(fmt.Sprintf("missing data: %v", rd))
		return nil
	}

	return hdrRes.Send(buff, message.Peer())
}

func (hdrRes *HeaderResolver) resolveHeaderFromNonce(key []byte) ([]byte, error) {
	// key is now an encoded nonce (uint64)

	// Search the nonce-key pair in cache-storage
	hash, err := hdrRes.hdrNoncesStorage.Get(key)
	if err != nil {
		log.Debug(err.Error())
	}

	// Search the nonce-key pair in data pool
	if hash == nil {
		nonceBytes, err := hdrRes.nonceConverter.ToUint64(key)
		if err != nil {
			return nil, dataRetriever.ErrInvalidNonceByteSlice
		}

		value, ok := hdrRes.hdrNonces.Get(nonceBytes)
		if ok {
			value.Range(func(shardId uint32, existingHash []byte) bool {
				if shardId == hdrRes.TargetShardID() {
					hash = existingHash
					return false
				}

				return true
			})
		}

		if len(hash) == 0 {
			return nil, nil
		}
	}

	return hdrRes.resolveHeaderFromHash(hash)
}

// resolveHeaderFromHash resolves a header using its key (header hash)
func (hdrRes *HeaderResolver) resolveHeaderFromHash(key []byte) ([]byte, error) {
	value, ok := hdrRes.headers.Peek(key)
	if !ok {
		return hdrRes.hdrStorage.Get(key)
	}

	buff, err := hdrRes.marshalizer.Marshal(value)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

// parseReceivedMessage will transform the received p2p.Message in a RequestData object.
func (hdrRes *HeaderResolver) parseReceivedMessage(message p2p.MessageP2P) (*dataRetriever.RequestData, error) {
	rd := &dataRetriever.RequestData{}
	err := rd.Unmarshal(hdrRes.marshalizer, message)
	if err != nil {
		return nil, err
	}
	if rd.Value == nil {
		return nil, dataRetriever.ErrNilValue
	}

	return rd, nil
}

// RequestDataFromHash requests a header from other peers having input the hdr hash
func (hdrRes *HeaderResolver) RequestDataFromHash(hash []byte) error {
	return hdrRes.SendOnRequestTopic(&dataRetriever.RequestData{
		Type:  dataRetriever.HashType,
		Value: hash,
	})
}

// RequestDataFromNonce requests a header from other peers having input the hdr nonce
func (hdrRes *HeaderResolver) RequestDataFromNonce(nonce uint64) error {
	return hdrRes.SendOnRequestTopic(&dataRetriever.RequestData{
		Type:  dataRetriever.NonceType,
		Value: hdrRes.nonceConverter.ToByteSlice(nonce),
	})
}

// IsInterfaceNil returns true if there is no value under the interface
func (hdrRes *HeaderResolver) IsInterfaceNil() bool {
	if hdrRes == nil {
		return true
	}
	return false
}
