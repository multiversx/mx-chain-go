package interceptors

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

func preProcessMesage(throttler process.InterceptorThrottler, message p2p.MessageP2P) error {
	if message == nil {
		return process.ErrNilMessage
	}
	if message.Data() == nil {
		return process.ErrNilDataToProcess
	}

	if !throttler.CanProcessMessage() {
		//can not process current message, send it to other peers
		return nil
	}

	throttler.StartMessageProcessing()
	return nil
}

func processInterceptedData(
	processor process.InterceptorProcessor,
	data process.InterceptedData,
	wgProcess *sync.WaitGroup,
) {
	err := processor.CheckValidForProcessing(data)
	if err != nil {
		log.Debug(fmt.Sprintf("intercepted data is not valid: %s", err.Error()))
		wgProcess.Done()
		return
	}

	err = processor.ProcessInteceptedData(data)
	if err != nil {
		log.Debug(fmt.Sprintf("intercepted data can not be processed: %s", err.Error()))
	}

	wgProcess.Done()
}
