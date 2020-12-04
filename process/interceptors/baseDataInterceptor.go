package interceptors

import (
	"bytes"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

type baseDataInterceptor struct {
	throttler        process.InterceptorThrottler
	antifloodHandler process.P2PAntifloodHandler
	topic            string
	currentPeerId    core.PeerID
	processor        process.InterceptorProcessor
	mutDebugHandler  sync.RWMutex
	debugHandler     process.InterceptedDebugger
}

func (bdi *baseDataInterceptor) preProcessMesage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	if message == nil {
		return process.ErrNilMessage
	}
	if message.Data() == nil {
		return process.ErrNilDataToProcess
	}

	if !bdi.isMessageFromSelfToSelf(fromConnectedPeer, message) {
		err := bdi.antifloodHandler.CanProcessMessage(message, fromConnectedPeer)
		if err != nil {
			return err
		}
		err = bdi.antifloodHandler.CanProcessMessagesOnTopic(fromConnectedPeer, bdi.topic, 1, uint64(len(message.Data())), message.SeqNo())
		if err != nil {
			return err
		}

		if !bdi.throttler.CanProcess() {
			return process.ErrSystemBusy
		}
	}

	bdi.throttler.StartProcessing()
	return nil
}

func (bdi *baseDataInterceptor) isMessageFromSelfToSelf(fromConnectedPeer core.PeerID, message p2p.MessageP2P) bool {
	return bytes.Equal(message.Signature(), message.From()) &&
		bytes.Equal(message.From(), bdi.currentPeerId.Bytes()) &&
		fromConnectedPeer == bdi.currentPeerId
}

func (bdi *baseDataInterceptor) processInterceptedData(data process.InterceptedData, msg p2p.MessageP2P) {
	err := bdi.processor.Validate(data, msg.Peer())
	if err != nil {
		log.Trace("intercepted data is not valid",
			"hash", data.Hash(),
			"type", data.Type(),
			"pid", p2p.MessageOriginatorPid(msg),
			"seq no", p2p.MessageOriginatorSeq(msg),
			"data", data.String(),
			"error", err.Error(),
		)
		bdi.processDebugInterceptedData(data, err)

		return
	}

	err = bdi.processor.Save(data, msg.Peer(), bdi.topic)
	if err != nil {
		log.Trace("intercepted data can not be processed",
			"hash", data.Hash(),
			"type", data.Type(),
			"pid", p2p.MessageOriginatorPid(msg),
			"seq no", p2p.MessageOriginatorSeq(msg),
			"data", data.String(),
			"error", err.Error(),
		)
		bdi.processDebugInterceptedData(data, err)

		return
	}

	log.Trace("intercepted data is processed",
		"hash", data.Hash(),
		"type", data.Type(),
		"pid", p2p.MessageOriginatorPid(msg),
		"seq no", p2p.MessageOriginatorSeq(msg),
		"data", data.String(),
	)
	bdi.processDebugInterceptedData(data, err)
}

func (bdi *baseDataInterceptor) processDebugInterceptedData(interceptedData process.InterceptedData, err error) {
	identifiers := interceptedData.Identifiers()
	bdi.debugHandler.LogProcessedHashes(bdi.topic, identifiers, err)
}

func (bdi *baseDataInterceptor) receivedDebugInterceptedData(interceptedData process.InterceptedData) {
	identifiers := interceptedData.Identifiers()
	bdi.debugHandler.LogReceivedHashes(bdi.topic, identifiers)
}
