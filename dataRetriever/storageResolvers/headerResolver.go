package storageResolvers

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/epochproviders/disabled"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("dataretriever/storageresolvers")

// ArgHeaderResolver is the argument structure used to create new HeaderResolver instance
type ArgHeaderResolver struct {
	Messenger                dataRetriever.MessageHandler
	ResponseTopicName        string
	NonceConverter           typeConverters.Uint64ByteSliceConverter
	HdrStorage               storage.Storer
	HeadersNoncesStorage     storage.Storer
	ManualEpochStartNotifier dataRetriever.ManualEpochStartNotifier
	ChanGracefullyClose      chan endProcess.ArgEndProcess
	DelayBeforeGracefulClose time.Duration
}

type headerResolver struct {
	*storageResolver
	nonceConverter   typeConverters.Uint64ByteSliceConverter
	mutEpochHandler  sync.RWMutex
	epochHandler     dataRetriever.EpochHandler
	hdrStorage       storage.Storer
	hdrNoncesStorage storage.Storer
}

// NewHeaderResolver creates a new storage header resolver
func NewHeaderResolver(arg ArgHeaderResolver) (*headerResolver, error) {
	if check.IfNil(arg.Messenger) {
		return nil, dataRetriever.ErrNilMessenger
	}
	if check.IfNil(arg.HdrStorage) {
		return nil, dataRetriever.ErrNilHeadersStorage
	}
	if check.IfNil(arg.HeadersNoncesStorage) {
		return nil, dataRetriever.ErrNilHeadersNoncesStorage
	}
	if check.IfNil(arg.NonceConverter) {
		return nil, dataRetriever.ErrNilUint64ByteSliceConverter
	}
	if check.IfNil(arg.ManualEpochStartNotifier) {
		return nil, dataRetriever.ErrNilManualEpochStartNotifier
	}
	if arg.ChanGracefullyClose == nil {
		return nil, dataRetriever.ErrNilGracefullyCloseChannel
	}

	epochHandler := disabled.NewEpochHandler()
	return &headerResolver{
		storageResolver: &storageResolver{
			messenger:                arg.Messenger,
			responseTopicName:        arg.ResponseTopicName,
			manualEpochStartNotifier: arg.ManualEpochStartNotifier,
			chanGracefullyClose:      arg.ChanGracefullyClose,
			delayBeforeGracefulClose: arg.DelayBeforeGracefulClose,
		},
		hdrStorage:       arg.HdrStorage,
		hdrNoncesStorage: arg.HeadersNoncesStorage,
		nonceConverter:   arg.NonceConverter,
		epochHandler:     epochHandler,
	}, nil
}

// RequestDataFromHash searches the hash in provided storage and then will send to self the message
func (hdrRes *headerResolver) RequestDataFromHash(hash []byte, _ uint32) error {
	hdrRes.mutEpochHandler.RLock()
	metaEpoch := hdrRes.epochHandler.MetaEpoch()
	hdrRes.mutEpochHandler.RUnlock()

	hdrRes.manualEpochStartNotifier.NewEpoch(metaEpoch + 1)
	hdrRes.manualEpochStartNotifier.NewEpoch(metaEpoch + 2)

	buff, err := hdrRes.hdrStorage.SearchFirst(hash)
	if err != nil {
		hdrRes.signalGracefullyClose()
		return err
	}

	return hdrRes.sendToSelf(buff)
}

// RequestDataFromNonce requests a header by its nonce
func (hdrRes *headerResolver) RequestDataFromNonce(nonce uint64, epoch uint32) error {
	nonceKey := hdrRes.nonceConverter.ToByteSlice(nonce)
	hash, err := hdrRes.hdrNoncesStorage.SearchFirst(nonceKey)
	if err != nil {
		hdrRes.signalGracefullyClose()
		return err
	}

	return hdrRes.RequestDataFromHash(hash, epoch)
}

// RequestDataFromEpoch requests the epoch start block
func (hdrRes *headerResolver) RequestDataFromEpoch(identifier []byte) error {
	buff, err := hdrRes.resolveHeaderFromEpoch(identifier)
	if err != nil {
		hdrRes.signalGracefullyClose()
		return err
	}

	return hdrRes.sendToSelf(buff)
}

// resolveHeaderFromEpoch resolves a header using its key based on epoch
func (hdrRes *headerResolver) resolveHeaderFromEpoch(key []byte) ([]byte, error) {
	actualKey := key

	isUnknownEpoch, err := core.IsUnknownEpochIdentifier(key)
	if err != nil {
		return nil, err
	}
	if isUnknownEpoch {
		actualKey = []byte(core.EpochStartIdentifier(hdrRes.manualEpochStartNotifier.CurrentEpoch() - 1))
	}

	return hdrRes.hdrStorage.SearchFirst(actualKey)
}

// SetEpochHandler sets the epoch handler
func (hdrRes *headerResolver) SetEpochHandler(epochHandler dataRetriever.EpochHandler) error {
	if check.IfNil(epochHandler) {
		return dataRetriever.ErrNilEpochHandler
	}

	hdrRes.mutEpochHandler.Lock()
	hdrRes.epochHandler = epochHandler
	hdrRes.mutEpochHandler.Unlock()

	return nil
}

// Close will try to close the associated opened storers
func (hdrRes *headerResolver) Close() error {
	errNonces := hdrRes.hdrNoncesStorage.Close()
	log.LogIfError(errNonces)

	errHeaders := hdrRes.hdrStorage.Close()
	log.LogIfError(errHeaders)

	if errNonces != nil {
		return errNonces
	}

	return errHeaders
}

// IsInterfaceNil returns true if there is no value under the interface
func (hdrRes *headerResolver) IsInterfaceNil() bool {
	return hdrRes == nil
}
