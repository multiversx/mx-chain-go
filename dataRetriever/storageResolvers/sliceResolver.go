package storageResolvers

import (
	"fmt"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/data/endProcess"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// maxBuffToSend represents max buffer size to send in bytes
const maxBuffToSend = 1 << 18 //256KB

// ArgSliceResolver is the argument structure used to create a new sliceResolver instance
type ArgSliceResolver struct {
	Messenger                dataRetriever.MessageHandler
	ResponseTopicName        string
	Storage                  storage.Storer
	DataPacker               dataRetriever.DataPacker
	Marshalizer              marshal.Marshalizer
	ManualEpochStartNotifier dataRetriever.ManualEpochStartNotifier
	ChanGracefullyClose      chan endProcess.ArgEndProcess
	DelayBeforeGracefulClose time.Duration
}

type sliceResolver struct {
	*storageResolver
	storage     storage.Storer
	dataPacker  dataRetriever.DataPacker
	marshalizer marshal.Marshalizer
}

// NewSliceResolver is a wrapper over Resolver that is specialized in resolving single and multiple requests
func NewSliceResolver(arg ArgSliceResolver) (*sliceResolver, error) {
	if check.IfNil(arg.Messenger) {
		return nil, dataRetriever.ErrNilMessenger
	}
	if check.IfNil(arg.Storage) {
		return nil, dataRetriever.ErrNilStore
	}
	if check.IfNil(arg.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(arg.DataPacker) {
		return nil, dataRetriever.ErrNilDataPacker
	}
	if check.IfNil(arg.ManualEpochStartNotifier) {
		return nil, dataRetriever.ErrNilManualEpochStartNotifier
	}
	if arg.ChanGracefullyClose == nil {
		return nil, dataRetriever.ErrNilGracefullyCloseChannel
	}

	return &sliceResolver{
		storageResolver: &storageResolver{
			messenger:                arg.Messenger,
			responseTopicName:        arg.ResponseTopicName,
			manualEpochStartNotifier: arg.ManualEpochStartNotifier,
			chanGracefullyClose:      arg.ChanGracefullyClose,
			delayBeforeGracefulClose: arg.DelayBeforeGracefulClose,
		},
		storage:     arg.Storage,
		dataPacker:  arg.DataPacker,
		marshalizer: arg.Marshalizer,
	}, nil
}

// RequestDataFromHash searches the hash in provided storage and then will send to self the message
func (sliceRes *sliceResolver) RequestDataFromHash(hash []byte, _ uint32) error {
	mb, err := sliceRes.storage.Get(hash)
	if err != nil {
		sliceRes.signalGracefullyClose()

		return err
	}

	b := &batch.Batch{
		Data: [][]byte{mb},
	}
	buffToSend, err := sliceRes.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	return sliceRes.sendToSelf(buffToSend)
}

// RequestDataFromHashArray searches the hashes in provided storage and then will send to self the message(s)
func (sliceRes *sliceResolver) RequestDataFromHashArray(hashes [][]byte, _ uint32) error {
	var errFetch error
	errorsFound := 0
	mbsBuffSlice := make([][]byte, 0, len(hashes))
	for _, hash := range hashes {
		mb, errTemp := sliceRes.storage.Get(hash)
		if errTemp != nil {
			errFetch = fmt.Errorf("%w for hash %s", errTemp, logger.DisplayByteSlice(hash))
			log.Trace("fetchMbAsByteSlice missing",
				"error", errFetch.Error(),
				"hash", hash)
			errorsFound++

			continue
		}
		mbsBuffSlice = append(mbsBuffSlice, mb)
	}

	buffsToSend, errPack := sliceRes.dataPacker.PackDataInChunks(mbsBuffSlice, maxBuffToSend)
	if errPack != nil {
		return errPack
	}

	for _, buff := range buffsToSend {
		errSend := sliceRes.sendToSelf(buff)
		if errSend != nil {
			return errSend
		}
	}

	if errFetch != nil {
		errFetch = fmt.Errorf("resolveMbRequestByHashArray last error %w from %d encountered errors", errFetch, errorsFound)

		sliceRes.signalGracefullyClose()
	}

	return errFetch
}

// IsInterfaceNil returns true if there is no value under the interface
func (sliceRes *sliceResolver) IsInterfaceNil() bool {
	return sliceRes == nil
}
