package interceptors

import (
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

	if !throttler.CanProcess() {
		return process.ErrSystemBusy
	}

	throttler.StartProcessing()
	return nil
}

func processInterceptedData(
	processor process.InterceptorProcessor,
	data process.InterceptedData,
	wgProcess *sync.WaitGroup,
) {
	err := processor.Validate(data)
	if err != nil {
		log.Trace("intercepted data is not valid",
			"error", err.Error(),
		)
		wgProcess.Done()
		return
	}

	err = processor.Save(data)
	if err != nil {
		log.Trace("intercepted data can not be processed",
			"error", err.Error(),
		)
	}

	wgProcess.Done()
}
