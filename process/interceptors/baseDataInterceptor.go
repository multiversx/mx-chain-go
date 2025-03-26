package interceptors

import (
	"bytes"
	"sync"
	"sync/atomic"
	"time"

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
	messagesMap          sync.Map
	bdiType              string
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
	value := int64(1)
	count, loaded := bdi.messagesMap.LoadOrStore(bdi.topic, &value)
	if loaded {
		atomic.AddInt64(count.(*int64), value)
	} else {
		bdi.messagesMap.Store(bdi.topic, &value)
	}

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

func (bdi *baseDataInterceptor) StartTimer() {
	go func() {
		now := time.Now()
		nextMinute := now.Truncate(time.Minute).Add(time.Minute)
		time.Sleep(time.Until(nextMinute))

		// Start ticker to run every minute
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()

		log.Info("Starting task execution at", "topic", bdi.topic)

		for _ = range ticker.C {
			value := int64(1)
			count, loaded := bdi.messagesMap.LoadOrStore(bdi.topic, &value)
			if loaded {
				currentCount := atomic.LoadInt64(count.(*int64))
				log.Info("messages statistics", "type", bdi.bdiType, "topic", bdi.topic, "count", currentCount)
			}
		}
	}()
}
