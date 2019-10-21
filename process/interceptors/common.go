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

	//TODO(jls) check this. Maybe should treat the error and return nil on ProcessReceivedMessage
	// as to not disturb the message propagation
	//if !throttler.CanProcess() {
	//	return process.ErrSystemBusy
	//}

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
		log.Debug(fmt.Sprintf("intercepted data is not valid: %s", err.Error()))
		wgProcess.Done()
		return
	}

	err = processor.Save(data)
	if err != nil {
		log.Debug(fmt.Sprintf("intercepted data can not be processed: %s", err.Error()))
	}

	wgProcess.Done()
}
