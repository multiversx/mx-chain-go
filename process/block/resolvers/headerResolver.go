package resolvers

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var log = logger.NewDefaultLogger()

// HeaderResolver is a wrapper over Resolver that is specialized in resolving headers requests
type HeaderResolver struct {
	*HeaderResolverBase
	hdrNonces      data.Uint64Cacher
	headers        storage.Cacher
	nonceConverter typeConverters.Uint64ByteSliceConverter
}

// NewHeaderResolver creates a new header resolver
func NewHeaderResolver(
	senderResolver process.TopicResolverSender,
	pools data.PoolsHolder,
	hdrStorage storage.Storer,
	marshalizer marshal.Marshalizer,
	nonceConverter typeConverters.Uint64ByteSliceConverter,
) (*HeaderResolver, error) {

	if senderResolver == nil {
		return nil, process.ErrNilResolverSender
	}
	if pools == nil {
		return nil, process.ErrNilPoolsHolder
	}
	headers := pools.Headers()
	if headers == nil {
		return nil, process.ErrNilHeadersDataPool
	}
	headersNonces := pools.HeadersNonces()
	if headersNonces == nil {
		return nil, process.ErrNilHeadersNoncesDataPool
	}
	if nonceConverter == nil {
		return nil, process.ErrNilNonceConverter
	}
	hdrResolverBase, err := NewHeaderResolverBase(
		senderResolver,
		headers,
		hdrStorage,
		marshalizer,
	)
	if err != nil {
		return nil, err
	}

	hdrResolver := &HeaderResolver{
		hdrNonces:          pools.HeadersNonces(),
		headers:            headers,
		nonceConverter:     nonceConverter,
		HeaderResolverBase: hdrResolverBase,
	}

	return hdrResolver, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to, usually a request topic)
func (hdrRes *HeaderResolver) ProcessReceivedMessage(message p2p.MessageP2P) ([]byte, error) {
	rd, err := hdrRes.ParseReceivedMessage(message)
	if err != nil {
		return nil, err
	}
	var buff []byte

	switch rd.Type {
	case process.HashType:
		buff, err = hdrRes.ResolveHeaderFromHash(rd.Value)
	case process.NonceType:
		buff, err = hdrRes.resolveHeaderFromNonce(rd.Value)
	default:
		return nil, process.ErrResolveTypeUnknown
	}
	if err != nil {
		return nil, err
	}
	if buff == nil {
		log.Debug(fmt.Sprintf("missing data: %v", rd))
		return nil, nil
	}

	return nil, hdrRes.Send(buff, message.Peer())
}

func (hdrRes *HeaderResolver) resolveHeaderFromNonce(key []byte) ([]byte, error) {
	//key is now an encoded nonce (uint64)

	//Step 1. decode the nonce from the key
	nonce, err := hdrRes.nonceConverter.ToUint64(key)
	if err != nil {
		return nil, process.ErrInvalidNonceByteSlice
	}

	//Step 2. search the nonce-key pair
	hash, _ := hdrRes.hdrNonces.Get(nonce)
	if hash == nil {
		return nil, nil
	}

	//Step 3. search header by key (hash)
	value, ok := hdrRes.headers.Peek(hash)
	if !ok {
		return hdrRes.hdrStorage.Get(hash)
	}

	//since there might be multiple entries, it shall return the first one that it finds
	buff, err := hdrRes.marshalizer.Marshal(value)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

// RequestDataFromNonce requests a header from other peers having input the hdr nonce
func (hdrRes *HeaderResolver) RequestDataFromNonce(nonce uint64) error {
	return hdrRes.SendOnRequestTopic(&process.RequestData{
		Type:  process.NonceType,
		Value: hdrRes.nonceConverter.ToByteSlice(nonce),
	})
}
