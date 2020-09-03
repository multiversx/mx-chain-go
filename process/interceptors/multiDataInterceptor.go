package interceptors

import (
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/debug/resolver"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

var log = logger.GetOrCreate("process/interceptors")

// ArgMultiDataInterceptor is the argument for the multi-data interceptor
type ArgMultiDataInterceptor struct {
	Topic            string
	Marshalizer      marshal.Marshalizer
	DataFactory      process.InterceptedDataFactory
	Processor        process.InterceptorProcessor
	Throttler        process.InterceptorThrottler
	AntifloodHandler process.P2PAntifloodHandler
	WhiteListRequest process.WhiteListHandler
	CurrentPeerId    core.PeerID
}

// MultiDataInterceptor is used for intercepting packed multi data
type MultiDataInterceptor struct {
	topic                      string
	marshalizer                marshal.Marshalizer
	factory                    process.InterceptedDataFactory
	processor                  process.InterceptorProcessor
	throttler                  process.InterceptorThrottler
	whiteListRequest           process.WhiteListHandler
	antifloodHandler           process.P2PAntifloodHandler
	mutInterceptedDebugHandler sync.RWMutex
	interceptedDebugHandler    process.InterceptedDebugger
	currentPeerId              core.PeerID
}

// NewMultiDataInterceptor hooks a new interceptor for packed multi data
func NewMultiDataInterceptor(arg ArgMultiDataInterceptor) (*MultiDataInterceptor, error) {
	if len(arg.Topic) == 0 {
		return nil, process.ErrEmptyTopic
	}
	if check.IfNil(arg.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(arg.DataFactory) {
		return nil, process.ErrNilInterceptedDataFactory
	}
	if check.IfNil(arg.Processor) {
		return nil, process.ErrNilInterceptedDataProcessor
	}
	if check.IfNil(arg.Throttler) {
		return nil, process.ErrNilInterceptorThrottler
	}
	if check.IfNil(arg.AntifloodHandler) {
		return nil, process.ErrNilAntifloodHandler
	}
	if check.IfNil(arg.WhiteListRequest) {
		return nil, process.ErrNilWhiteListHandler
	}
	if len(arg.CurrentPeerId) == 0 {
		return nil, process.ErrEmptyPeerID
	}

	multiDataIntercept := &MultiDataInterceptor{
		topic:            arg.Topic,
		marshalizer:      arg.Marshalizer,
		factory:          arg.DataFactory,
		processor:        arg.Processor,
		throttler:        arg.Throttler,
		whiteListRequest: arg.WhiteListRequest,
		antifloodHandler: arg.AntifloodHandler,
		currentPeerId:    arg.CurrentPeerId,
	}
	multiDataIntercept.interceptedDebugHandler = resolver.NewDisabledInterceptorResolver()

	return multiDataIntercept, nil
}

// ProcessReceivedMessage is the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (mdi *MultiDataInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	err := preProcessMesage(mdi.throttler, mdi.antifloodHandler, message, fromConnectedPeer, mdi.topic, mdi.currentPeerId)
	if err != nil {
		return err
	}

	b := batch.Batch{}
	err = mdi.marshalizer.Unmarshal(&b, message.Data())
	if err != nil {
		mdi.throttler.EndProcessing()

		//this situation is so severe that we need to black list de peers
		reason := "unmarshalable data got on topic " + mdi.topic
		mdi.antifloodHandler.BlacklistPeer(message.Peer(), reason, core.InvalidMessageBlacklistDuration)
		mdi.antifloodHandler.BlacklistPeer(fromConnectedPeer, reason, core.InvalidMessageBlacklistDuration)

		return err
	}
	multiDataBuff := b.Data
	lenMultiData := len(multiDataBuff)
	if lenMultiData == 0 {
		mdi.throttler.EndProcessing()
		return process.ErrNoDataInMessage
	}

	err = mdi.antifloodHandler.CanProcessMessagesOnTopic(
		fromConnectedPeer,
		mdi.topic,
		uint32(lenMultiData),
		uint64(len(message.Data())),
		message.SeqNo(),
	)
	if err != nil {
		mdi.throttler.EndProcessing()
		return err
	}

	listInterceptedData := make([]process.InterceptedData, len(multiDataBuff))
	errOriginator := mdi.antifloodHandler.IsOriginatorEligibleForTopic(message.Peer(), mdi.topic)

	for index, dataBuff := range multiDataBuff {
		var interceptedData process.InterceptedData
		interceptedData, err = mdi.interceptedData(dataBuff, message.Peer(), fromConnectedPeer)
		listInterceptedData[index] = interceptedData
		if err != nil {
			mdi.throttler.EndProcessing()
			return err
		}

		isWhiteListed := mdi.whiteListRequest.IsWhiteListed(interceptedData)
		if !isWhiteListed && errOriginator != nil {
			mdi.throttler.EndProcessing()
			log.Trace("got message from peer on topic only for validators", "originator",
				p2p.PeerIdToShortString(message.Peer()),
				"topic", mdi.topic,
				"err", errOriginator)
			return errOriginator
		}

		isForCurrentShard := interceptedData.IsForCurrentShard()
		shouldProcess := isForCurrentShard || isWhiteListed
		if !shouldProcess {
			log.Trace("intercepted data should not be processed",
				"pid", p2p.MessageOriginatorPid(message),
				"seq no", p2p.MessageOriginatorSeq(message),
				"topics", message.Topics(),
				"hash", interceptedData.Hash(),
				"is for this shard", isForCurrentShard,
				"is white listed", isWhiteListed,
			)
			mdi.throttler.EndProcessing()
			return process.ErrInterceptedDataNotForCurrentShard
		}
	}

	go func() {
		for _, interceptedData := range listInterceptedData {
			processInterceptedData(
				mdi.processor,
				mdi.interceptedDebugHandler,
				interceptedData,
				mdi.topic,
				message,
			)
		}
		mdi.throttler.EndProcessing()
	}()

	return nil
}

func (mdi *MultiDataInterceptor) interceptedData(dataBuff []byte, originator core.PeerID, fromConnectedPeer core.PeerID) (process.InterceptedData, error) {
	interceptedData, err := mdi.factory.Create(dataBuff)
	if err != nil {
		//this situation is so severe that we need to black list de peers
		reason := "can not create object from received bytes, topic " + mdi.topic
		mdi.antifloodHandler.BlacklistPeer(originator, reason, core.InvalidMessageBlacklistDuration)
		mdi.antifloodHandler.BlacklistPeer(fromConnectedPeer, reason, core.InvalidMessageBlacklistDuration)

		return nil, err
	}

	receivedDebugInterceptedData(mdi.interceptedDebugHandler, interceptedData, mdi.topic)

	err = interceptedData.CheckValidity()
	if err != nil {
		processDebugInterceptedData(mdi.interceptedDebugHandler, interceptedData, mdi.topic, err)

		isWrongVersion := err == process.ErrInvalidTransactionVersion || err == process.ErrInvalidChainID
		if isWrongVersion {
			//this situation is so severe that we need to black list de peers
			reason := "wrong version of received intercepted data, topic " + mdi.topic + ", error " + err.Error()
			mdi.antifloodHandler.BlacklistPeer(originator, reason, core.InvalidMessageBlacklistDuration)
			mdi.antifloodHandler.BlacklistPeer(fromConnectedPeer, reason, core.InvalidMessageBlacklistDuration)
		}

		return nil, err
	}

	return interceptedData, nil
}

// SetInterceptedDebugHandler will set a new intercepted debug handler
func (mdi *MultiDataInterceptor) SetInterceptedDebugHandler(handler process.InterceptedDebugger) error {
	if check.IfNil(handler) {
		return process.ErrNilDebugger
	}

	mdi.mutInterceptedDebugHandler.Lock()
	mdi.interceptedDebugHandler = handler
	mdi.mutInterceptedDebugHandler.Unlock()

	return nil
}

// RegisterHandler registers a callback function to be notified on received data
func (mdi *MultiDataInterceptor) RegisterHandler(handler func(topic string, hash []byte, data interface{})) {
	mdi.processor.RegisterHandler(handler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mdi *MultiDataInterceptor) IsInterfaceNil() bool {
	return mdi == nil
}
