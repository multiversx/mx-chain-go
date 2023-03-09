package storagerequesters

import (
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/epochproviders/disabled"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ dataRetriever.HeaderRequester = (*headerRequester)(nil)

var log = logger.GetOrCreate("dataretriever/storagerequesters")

// ArgHeaderRequester is the argument structure used to create new headerRequester instance
type ArgHeaderRequester struct {
	Messenger                dataRetriever.MessageHandler
	ResponseTopicName        string
	NonceConverter           typeConverters.Uint64ByteSliceConverter
	HdrStorage               storage.Storer
	HeadersNoncesStorage     storage.Storer
	ManualEpochStartNotifier dataRetriever.ManualEpochStartNotifier
	ChanGracefullyClose      chan endProcess.ArgEndProcess
	DelayBeforeGracefulClose time.Duration
}

type headerRequester struct {
	*storageRequester
	nonceConverter   typeConverters.Uint64ByteSliceConverter
	mutEpochHandler  sync.RWMutex
	epochHandler     dataRetriever.EpochHandler
	hdrStorage       storage.Storer
	hdrNoncesStorage storage.Storer
}

// NewHeaderRequester creates a new storage header resolver
func NewHeaderRequester(arg ArgHeaderRequester) (*headerRequester, error) {
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
	return &headerRequester{
		storageRequester: &storageRequester{
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
func (hdrReq *headerRequester) RequestDataFromHash(hash []byte, _ uint32) error {
	hdrReq.mutEpochHandler.RLock()
	metaEpoch := hdrReq.epochHandler.MetaEpoch()
	hdrReq.mutEpochHandler.RUnlock()

	hdrReq.manualEpochStartNotifier.NewEpoch(metaEpoch + 1)
	hdrReq.manualEpochStartNotifier.NewEpoch(metaEpoch + 2)

	buff, err := hdrReq.hdrStorage.SearchFirst(hash)
	if err != nil {
		hdrReq.signalGracefullyClose()
		return err
	}

	return hdrReq.sendToSelf(buff)
}

// RequestDataFromNonce requests a header by its nonce
func (hdrReq *headerRequester) RequestDataFromNonce(nonce uint64, epoch uint32) error {
	nonceKey := hdrReq.nonceConverter.ToByteSlice(nonce)
	hash, err := hdrReq.hdrNoncesStorage.SearchFirst(nonceKey)
	if err != nil {
		hdrReq.signalGracefullyClose()
		return err
	}

	return hdrReq.RequestDataFromHash(hash, epoch)
}

// RequestDataFromEpoch requests the epoch start block
func (hdrReq *headerRequester) RequestDataFromEpoch(identifier []byte) error {
	buff, err := hdrReq.resolveHeaderFromEpoch(identifier)
	if err != nil {
		hdrReq.signalGracefullyClose()
		return err
	}

	return hdrReq.sendToSelf(buff)
}

// resolveHeaderFromEpoch resolves a header using its key based on epoch
func (hdrReq *headerRequester) resolveHeaderFromEpoch(key []byte) ([]byte, error) {
	actualKey := key

	isUnknownEpoch, err := core.IsUnknownEpochIdentifier(key)
	if err != nil {
		return nil, err
	}
	if isUnknownEpoch {
		actualKey = []byte(core.EpochStartIdentifier(hdrReq.manualEpochStartNotifier.CurrentEpoch() - 1))
	}

	return hdrReq.hdrStorage.SearchFirst(actualKey)
}

// SetEpochHandler sets the epoch handler
func (hdrReq *headerRequester) SetEpochHandler(epochHandler dataRetriever.EpochHandler) error {
	if check.IfNil(epochHandler) {
		return dataRetriever.ErrNilEpochHandler
	}

	hdrReq.mutEpochHandler.Lock()
	hdrReq.epochHandler = epochHandler
	hdrReq.mutEpochHandler.Unlock()

	return nil
}

// Close will try to close the associated opened storers
func (hdrReq *headerRequester) Close() error {
	var lastErr error
	if !check.IfNil(hdrReq.hdrNoncesStorage) {
		lastErr = hdrReq.hdrNoncesStorage.Close()
		log.LogIfError(lastErr)
	}

	if !check.IfNil(hdrReq.hdrStorage) {
		lastErr = hdrReq.hdrStorage.Close()
		log.LogIfError(lastErr)
	}

	return lastErr
}

// IsInterfaceNil returns true if there is no value under the interface
func (hdrReq *headerRequester) IsInterfaceNil() bool {
	return hdrReq == nil
}
