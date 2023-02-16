package storagerequesters

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// maxBuffToSend represents max buffer size to send in bytes
const maxBuffToSend = 1 << 18 //256KB

// ArgSliceRequester is the argument structure used to create a new sliceRequester instance
type ArgSliceRequester struct {
	Messenger                dataRetriever.MessageHandler
	ResponseTopicName        string
	Storage                  storage.Storer
	DataPacker               dataRetriever.DataPacker
	Marshalizer              marshal.Marshalizer
	ManualEpochStartNotifier dataRetriever.ManualEpochStartNotifier
	ChanGracefullyClose      chan endProcess.ArgEndProcess
	DelayBeforeGracefulClose time.Duration
}

type sliceRequester struct {
	*storageRequester
	storage     storage.Storer
	dataPacker  dataRetriever.DataPacker
	marshalizer marshal.Marshalizer
}

// NewSliceRequester is a wrapper over Requester that is specialized in sending requests
func NewSliceRequester(arg ArgSliceRequester) (*sliceRequester, error) {
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

	return &sliceRequester{
		storageRequester: &storageRequester{
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
func (sliceReq *sliceRequester) RequestDataFromHash(hash []byte, _ uint32) error {
	mb, err := sliceReq.storage.Get(hash)
	if err != nil {
		sliceReq.signalGracefullyClose()
		return err
	}

	b := &batch.Batch{
		Data: [][]byte{mb},
	}
	buffToSend, err := sliceReq.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	return sliceReq.sendToSelf(buffToSend)
}

// RequestDataFromHashArray searches the hashes in provided storage and then will send to self the message(s)
func (sliceReq *sliceRequester) RequestDataFromHashArray(hashes [][]byte, _ uint32) error {
	var errFetch error
	errorsFound := 0
	mbsBuffSlice := make([][]byte, 0, len(hashes))
	for _, hash := range hashes {
		mb, errTemp := sliceReq.storage.Get(hash)
		if errTemp != nil {
			errFetch = fmt.Errorf("%w for hash %s", errTemp, logger.DisplayByteSlice(hash))
			log.Trace("fetchByteSlice missing",
				"error", errFetch.Error(),
				"hash", hash,
				"topic", sliceReq.responseTopicName)
			errorsFound++

			continue
		}
		mbsBuffSlice = append(mbsBuffSlice, mb)
	}

	buffsToSend, errPack := sliceReq.dataPacker.PackDataInChunks(mbsBuffSlice, maxBuffToSend)
	if errPack != nil {
		return errPack
	}

	for _, buff := range buffsToSend {
		errSend := sliceReq.sendToSelf(buff)
		if errSend != nil {
			return errSend
		}
	}

	if errFetch != nil {
		errFetch = fmt.Errorf("RequesterequestByHashArray on topic %s, last error %w from %d encountered errors",
			sliceReq.responseTopicName, errFetch, errorsFound)
		sliceReq.signalGracefullyClose()
	}

	return errFetch
}

// Close will try to close the associated opened storers
func (sliceReq *sliceRequester) Close() error {
	var err error
	if !check.IfNil(sliceReq.storage) {
		err = sliceReq.storage.Close()
	}
	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (sliceReq *sliceRequester) IsInterfaceNil() bool {
	return sliceReq == nil
}
