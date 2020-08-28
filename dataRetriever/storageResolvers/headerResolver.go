package storageResolvers

import (
	"fmt"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/epochproviders"
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
}

type headerResolver struct {
	*storageResolver
	nonceConverter           typeConverters.Uint64ByteSliceConverter
	mutEpochHandler          sync.RWMutex
	epochHandler             dataRetriever.EpochHandler
	hdrStorage               storage.Storer
	hdrNoncesStorage         storage.Storer
	manualEpochStartNotifier dataRetriever.ManualEpochStartNotifier
	chanGracefullyClose      chan endProcess.ArgEndProcess
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

	epochHandler := epochproviders.NewNilEpochHandler()
	return &headerResolver{
		storageResolver: &storageResolver{
			messenger:         arg.Messenger,
			responseTopicName: arg.ResponseTopicName,
		},
		hdrStorage:               arg.HdrStorage,
		hdrNoncesStorage:         arg.HeadersNoncesStorage,
		nonceConverter:           arg.NonceConverter,
		epochHandler:             epochHandler,
		manualEpochStartNotifier: arg.ManualEpochStartNotifier,
		chanGracefullyClose:      arg.ChanGracefullyClose,
	}, nil
}

// RequestDataFromHash searches the hash in provided storage and then will send to self the message
func (hdrRes *headerResolver) RequestDataFromHash(hash []byte, _ uint32) error {
	hdrRes.mutEpochHandler.RLock()
	metaEpoch := hdrRes.epochHandler.MetaEpoch()
	hdrRes.mutEpochHandler.RUnlock()

	hdrRes.manualEpochStartNotifier.NewEpoch(metaEpoch + 1)

	buff, err := hdrRes.hdrStorage.SearchFirst(hash)
	if err != nil {
		crtEpoch := hdrRes.manualEpochStartNotifier.CurrentEpoch()

		argEndProcess := endProcess.ArgEndProcess{
			Reason: core.ImportComplete,
			Description: fmt.Sprintf("import ended because data from epochs %d or %d does not exist",
				crtEpoch-1, crtEpoch),
		}

		select {
		case hdrRes.chanGracefullyClose <- argEndProcess:
		default:
			log.Debug("headerResolver.RequestDataFromHash: could not wrote on the end chan")
		}

		return err
	}

	return hdrRes.sendToSelf(buff)
}

// RequestDataFromNonce requests a header by its nonce
func (hdrRes *headerResolver) RequestDataFromNonce(nonce uint64, epoch uint32) error {
	nonceKey := hdrRes.nonceConverter.ToByteSlice(nonce)
	hash, err := hdrRes.hdrNoncesStorage.SearchFirst(nonceKey)
	if err != nil {
		return err
	}

	return hdrRes.RequestDataFromHash(hash, epoch)
}

// RequestDataFromEpoch requests the epoch start block
func (hdrRes *headerResolver) RequestDataFromEpoch(identifier []byte) error {
	return hdrRes.RequestDataFromHash(identifier, 0)
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

// IsInterfaceNil returns true if there is no value under the interface
func (hdrRes *headerResolver) IsInterfaceNil() bool {
	return hdrRes == nil
}
