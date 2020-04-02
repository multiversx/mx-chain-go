package interceptors

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

func preProcessMesage(
	throttler process.InterceptorThrottler,
	antifloodHandler process.P2PAntifloodHandler,
	message p2p.MessageP2P,
	fromConnectedPeer p2p.PeerID,
	topic string,
) error {

	if message == nil {
		return process.ErrNilMessage
	}
	if message.Data() == nil {
		return process.ErrNilDataToProcess
	}
	err := antifloodHandler.CanProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}
	err = antifloodHandler.CanProcessMessageOnTopic(fromConnectedPeer, topic)
	if err != nil {
		return err
	}

	if !throttler.CanProcess() {
		return process.ErrSystemBusy
	}

	throttler.StartProcessing()
	return nil
}

func processInterceptedData(
	processor process.InterceptorProcessor,
	handler process.InterceptedDebugHandler,
	data process.InterceptedData,
	topic string,
	wgProcess *sync.WaitGroup,
	msg p2p.MessageP2P,
) {
	err := processor.Validate(data)
	if err != nil {
		log.Trace("intercepted data is not valid",
			"hash", data.Hash(),
			"type", data.Type(),
			"pid", p2p.MessageOriginatorPid(msg),
			"seq no", p2p.MessageOriginatorSeq(msg),
			"data", data.String(),
			"error", err.Error(),
		)
		wgProcess.Done()
		processDebugInterceptedData(handler, data, topic, err)

		return
	}

	err = processor.Save(data)
	if err != nil {
		log.Trace("intercepted data can not be processed",
			"hash", data.Hash(),
			"type", data.Type(),
			"pid", p2p.MessageOriginatorPid(msg),
			"seq no", p2p.MessageOriginatorSeq(msg),
			"data", data.String(),
			"error", err.Error(),
		)
		wgProcess.Done()
		processDebugInterceptedData(handler, data, topic, err)

		return
	}

	log.Trace("intercepted data is processed",
		"hash", data.Hash(),
		"type", data.Type(),
		"pid", p2p.MessageOriginatorPid(msg),
		"seq no", p2p.MessageOriginatorSeq(msg),
		"data", data.String(),
	)
	processDebugInterceptedData(handler, data, topic, err)

	wgProcess.Done()
}

func processDebugInterceptedData(
	debugHandler process.InterceptedDebugHandler,
	interceptedData process.InterceptedData,
	topic string,
	err error,
) {
	if !debugHandler.Enabled() {
		return
	}

	identifiers := interceptedData.Identifiers()
	for _, identifier := range identifiers {
		debugHandler.ProcessedHash(topic, identifier, err)
	}
}

func receivedDebugInterceptedData(
	debugHandler process.InterceptedDebugHandler,
	interceptedData process.InterceptedData,
	topic string,
) {
	if !debugHandler.Enabled() {
		return
	}

	identifiers := interceptedData.Identifiers()
	for _, identifier := range identifiers {
		debugHandler.ReceivedHash(topic, identifier)
	}
}
