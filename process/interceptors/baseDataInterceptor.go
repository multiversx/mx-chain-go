package interceptors

import (
	"bytes"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
)

type baseDataInterceptor struct {
	throttler            process.InterceptorThrottler
	antifloodHandler     process.P2PAntifloodHandler
	topic                string
	currentPeerId        core.PeerID
	processor            process.InterceptorProcessor
	mutDebugHandler      sync.RWMutex
	debugHandler         process.InterceptedDebugger
	preferredPeersHolder process.PreferredPeersHolderHandler
}

func (bdi *baseDataInterceptor) preProcessMesage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	if message == nil {
		return process.ErrNilMessage
	}
	if message.Data() == nil {
		return process.ErrNilDataToProcess
	}

	if !bdi.shouldSkipAntifloodChecks(fromConnectedPeer, message) {
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

func (bdi *baseDataInterceptor) shouldSkipAntifloodChecks(fromConnectedPeer core.PeerID, message p2p.MessageP2P) bool {
	return true

	if bdi.isMessageFromSelfToSelf(fromConnectedPeer, message) {
		return true
	}

	return bdi.preferredPeersHolder.Contains(fromConnectedPeer)
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
			"intercepted data", data.String(),
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
			"intercepted data", data.String(),
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
		"intercepted data", data.String(),
	)
	bdi.processDebugInterceptedData(data, err)
}

func (bdi *baseDataInterceptor) processDebugInterceptedData(interceptedData process.InterceptedData, err error) {
	identifiers := interceptedData.Identifiers()

	bdi.mutDebugHandler.RLock()
	bdi.debugHandler.LogProcessedHashes(bdi.topic, identifiers, err)
	bdi.mutDebugHandler.RUnlock()
}

func (bdi *baseDataInterceptor) receivedDebugInterceptedData(interceptedData process.InterceptedData) {
	identifiers := interceptedData.Identifiers()

	bdi.mutDebugHandler.RLock()
	bdi.debugHandler.LogReceivedHashes(bdi.topic, identifiers)
	bdi.mutDebugHandler.RUnlock()
}

// SetInterceptedDebugHandler will set a new intercepted debug handler
func (bdi *baseDataInterceptor) SetInterceptedDebugHandler(handler process.InterceptedDebugger) error {
	if check.IfNil(handler) {
		return process.ErrNilDebugger
	}

	bdi.mutDebugHandler.Lock()
	bdi.debugHandler = handler
	bdi.mutDebugHandler.Unlock()

	return nil
}
