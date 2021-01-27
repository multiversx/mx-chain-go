package interceptors

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/debug/resolver"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ArgSingleDataInterceptor is the argument for the single-data interceptor
type ArgSingleDataInterceptor struct {
	Topic            string
	DataFactory      process.InterceptedDataFactory
	Processor        process.InterceptorProcessor
	Throttler        process.InterceptorThrottler
	AntifloodHandler process.P2PAntifloodHandler
	WhiteListRequest process.WhiteListHandler
	CurrentPeerId    core.PeerID
}

// SingleDataInterceptor is used for intercepting packed multi data
type SingleDataInterceptor struct {
	*baseDataInterceptor
	factory          process.InterceptedDataFactory
	whiteListRequest process.WhiteListHandler
}

// NewSingleDataInterceptor hooks a new interceptor for single data
func NewSingleDataInterceptor(arg ArgSingleDataInterceptor) (*SingleDataInterceptor, error) {
	if len(arg.Topic) == 0 {
		return nil, process.ErrEmptyTopic
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

	singleDataIntercept := &SingleDataInterceptor{
		baseDataInterceptor: &baseDataInterceptor{
			throttler:        arg.Throttler,
			antifloodHandler: arg.AntifloodHandler,
			topic:            arg.Topic,
			currentPeerId:    arg.CurrentPeerId,
			processor:        arg.Processor,
			debugHandler:     resolver.NewDisabledInterceptorResolver(),
		},
		factory:          arg.DataFactory,
		whiteListRequest: arg.WhiteListRequest,
	}

	return singleDataIntercept, nil
}

// ProcessReceivedMessage is the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (sdi *SingleDataInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	sdi.mutDebugHandler.RLock()
	defer sdi.mutDebugHandler.RUnlock()

	err := sdi.preProcessMesage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	interceptedData, err := sdi.factory.Create(message.Data())
	if err != nil {
		sdi.throttler.EndProcessing()

		//this situation is so severe that we need to black list the peers
		reason := "can not create object from received bytes, topic " + sdi.topic + ", error " + err.Error()
		sdi.antifloodHandler.BlacklistPeer(message.Peer(), reason, core.InvalidMessageBlacklistDuration)
		sdi.antifloodHandler.BlacklistPeer(fromConnectedPeer, reason, core.InvalidMessageBlacklistDuration)

		return err
	}

	sdi.receivedDebugInterceptedData(interceptedData)

	err = interceptedData.CheckValidity()
	if err != nil {
		sdi.throttler.EndProcessing()
		sdi.processDebugInterceptedData(interceptedData, err)

		isWrongVersion := err == process.ErrInvalidTransactionVersion || err == process.ErrInvalidChainID
		if isWrongVersion {
			//this situation is so severe that we need to black list de peers
			reason := "wrong version of received intercepted data, topic " + sdi.topic + ", error " + err.Error()
			sdi.antifloodHandler.BlacklistPeer(message.Peer(), reason, core.InvalidMessageBlacklistDuration)
			sdi.antifloodHandler.BlacklistPeer(fromConnectedPeer, reason, core.InvalidMessageBlacklistDuration)
		}

		return err
	}

	errOriginator := sdi.antifloodHandler.IsOriginatorEligibleForTopic(message.Peer(), sdi.topic)
	isWhiteListed := sdi.whiteListRequest.IsWhiteListed(interceptedData)
	if !isWhiteListed && errOriginator != nil {
		log.Trace("got message from peer on topic only for validators",
			"originator", p2p.PeerIdToShortString(message.Peer()), "topic",
			sdi.topic, "err", errOriginator)
		sdi.throttler.EndProcessing()
		return errOriginator
	}

	isForCurrentShard := interceptedData.IsForCurrentShard()
	shouldProcess := isForCurrentShard || isWhiteListed
	if !shouldProcess {
		sdi.throttler.EndProcessing()
		log.Trace("intercepted data is for other shards",
			"pid", p2p.MessageOriginatorPid(message),
			"seq no", p2p.MessageOriginatorSeq(message),
			"topics", message.Topics(),
			"hash", interceptedData.Hash(),
			"is for current shard", isForCurrentShard,
			"is white listed", isWhiteListed,
		)

		return nil
	}

	go func() {
		sdi.processInterceptedData(interceptedData, message)
		sdi.throttler.EndProcessing()
	}()

	return nil
}

// RegisterHandler registers a callback function to be notified on received data
func (sdi *SingleDataInterceptor) RegisterHandler(handler func(topic string, hash []byte, data interface{})) {
	sdi.processor.RegisterHandler(handler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sdi *SingleDataInterceptor) IsInterfaceNil() bool {
	return sdi == nil
}
