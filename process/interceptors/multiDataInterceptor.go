package interceptors

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

var log = logger.GetOrCreate("process/interceptors")

// MultiDataInterceptor is used for intercepting packed multi data
type MultiDataInterceptor struct {
	marshalizer marshal.Marshalizer
	factory     process.InterceptedDataFactory
	processor   process.InterceptorProcessor
	throttler   process.InterceptorThrottler
}

// NewMultiDataInterceptor hooks a new interceptor for packed multi data
func NewMultiDataInterceptor(
	marshalizer marshal.Marshalizer,
	factory process.InterceptedDataFactory,
	processor process.InterceptorProcessor,
	throttler process.InterceptorThrottler,
) (*MultiDataInterceptor, error) {

	if check.IfNil(marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(factory) {
		return nil, process.ErrNilInterceptedDataFactory
	}
	if check.IfNil(processor) {
		return nil, process.ErrNilInterceptedDataProcessor
	}
	if check.IfNil(throttler) {
		return nil, process.ErrNilInterceptorThrottler
	}

	multiDataIntercept := &MultiDataInterceptor{
		marshalizer: marshalizer,
		factory:     factory,
		processor:   processor,
		throttler:   throttler,
	}

	return multiDataIntercept, nil
}

// ProcessReceivedMessage is the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (mdi *MultiDataInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error {
	err := preProcessMesage(mdi.throttler, message)
	if err != nil {
		return err
	}

	multiDataBuff := make([][]byte, 0)
	err = mdi.marshalizer.Unmarshal(&multiDataBuff, message.Data())
	if err != nil {
		mdi.throttler.EndProcessing()
		return err
	}
	if len(multiDataBuff) == 0 {
		mdi.throttler.EndProcessing()
		return process.ErrNoDataInMessage
	}

	filteredMultiDataBuff := make([][]byte, 0)
	interceptedMultiData := make([]process.InterceptedData, 0)
	lastErrEncountered := error(nil)
	wgProcess := &sync.WaitGroup{}
	wgProcess.Add(len(multiDataBuff))

	go func() {
		wgProcess.Wait()
		mdi.processor.SignalEndOfProcessing(interceptedMultiData)
		mdi.throttler.EndProcessing()
	}()

	for _, dataBuff := range multiDataBuff {
		interceptedData, err := mdi.interceptedData(dataBuff)
		if err != nil {
			lastErrEncountered = err
			wgProcess.Done()
			continue
		}

		interceptedMultiData = append(interceptedMultiData, interceptedData)

		//data is validated, add it to filtered out buff
		filteredMultiDataBuff = append(filteredMultiDataBuff, dataBuff)
		if !interceptedData.IsForCurrentShard() {
			log.Trace("intercepted data is for other shards")
			wgProcess.Done()
			continue
		}

		go processInterceptedData(mdi.processor, interceptedData, wgProcess)
	}

	var buffToSend []byte
	haveDataForBroadcast := len(filteredMultiDataBuff) > 0 && lastErrEncountered != nil
	if haveDataForBroadcast {
		buffToSend, err = mdi.marshalizer.Marshal(filteredMultiDataBuff)
		if err != nil {
			return err
		}

		if broadcastHandler != nil {
			broadcastHandler(buffToSend)
		}
	}

	return lastErrEncountered
}

func (mdi *MultiDataInterceptor) interceptedData(dataBuff []byte) (process.InterceptedData, error) {
	interceptedData, err := mdi.factory.Create(dataBuff)
	if err != nil {
		return nil, err
	}

	err = interceptedData.CheckValidity()
	if err != nil {
		return nil, err
	}

	return interceptedData, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mdi *MultiDataInterceptor) IsInterfaceNil() bool {
	if mdi == nil {
		return true
	}
	return false
}
