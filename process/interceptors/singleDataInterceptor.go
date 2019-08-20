package interceptors

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// SingleDataInterceptor is used for intercepting packed multi data
type SingleDataInterceptor struct {
	factory                  process.InterceptedDataFactory
	processor                process.InterceptorProcessor
	throttler                process.InterceptorThrottler
	broadcastCallbackHandler func(buffToSend []byte)
}

// NewSingleDataInterceptor hooks a new interceptor for single data
func NewSingleDataInterceptor(
	factory process.InterceptedDataFactory,
	processor process.InterceptorProcessor,
	throttler process.InterceptorThrottler,
) (*SingleDataInterceptor, error) {

	if factory == nil {
		return nil, process.ErrNilInterceptedDataFactory
	}
	if processor == nil {
		return nil, process.ErrNilInterceptedDataProcessor
	}
	if throttler == nil {
		return nil, process.ErrNilInterceptorThrottler
	}

	singleDataIntercept := &SingleDataInterceptor{
		factory:   factory,
		processor: processor,
		throttler: throttler,
	}

	return singleDataIntercept, nil
}

// ProcessReceivedMessage is the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (sdi *SingleDataInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	err := preProcessMesage(sdi.throttler, message)
	if err != nil {
		return err
	}

	interceptedData, err := sdi.factory.Create(message.Data())
	if err != nil {
		sdi.throttler.EndProcess()
		return err
	}

	err = interceptedData.CheckValidity()
	if err != nil {
		sdi.throttler.EndProcess()
		return err
	}

	if !interceptedData.IsForMyShard() {
		sdi.throttler.EndProcess()
		log.Debug("intercepted data is for other shards")
		return nil
	}

	wgProcess := &sync.WaitGroup{}
	wgProcess.Add(1)
	go func() {
		wgProcess.Wait()
		sdi.throttler.EndProcess()
	}()

	go processInterceptedData(sdi.processor, interceptedData, wgProcess)

	return nil
}
