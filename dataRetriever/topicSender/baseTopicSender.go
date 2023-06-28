package topicsender

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/debug/handler"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/p2p/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("dataretriever/topicsender")

const (
	minPeersToQuery    = 2
	preferredPeerIndex = -1
)

// ArgBaseTopicSender is the base DTO used to create a new topic sender instance
type ArgBaseTopicSender struct {
	Messenger            dataRetriever.MessageHandler
	TopicName            string
	OutputAntiflooder    dataRetriever.P2PAntifloodHandler
	PreferredPeersHolder dataRetriever.PreferredPeersHolderHandler
	TargetShardId        uint32
}

type baseTopicSender struct {
	messenger                   dataRetriever.MessageHandler
	topicName                   string
	outputAntiflooder           dataRetriever.P2PAntifloodHandler
	mutDebugHandler             sync.RWMutex
	debugHandler                dataRetriever.DebugHandler
	preferredPeersHolderHandler dataRetriever.PreferredPeersHolderHandler
	targetShardId               uint32
}

func createBaseTopicSender(args ArgBaseTopicSender) *baseTopicSender {
	return &baseTopicSender{
		messenger:                   args.Messenger,
		topicName:                   args.TopicName,
		outputAntiflooder:           args.OutputAntiflooder,
		debugHandler:                handler.NewDisabledInterceptorDebugHandler(),
		preferredPeersHolderHandler: args.PreferredPeersHolder,
		targetShardId:               args.TargetShardId,
	}
}

func checkBaseTopicSenderArgs(args ArgBaseTopicSender) error {
	if check.IfNil(args.Messenger) {
		return dataRetriever.ErrNilMessenger
	}
	if check.IfNil(args.OutputAntiflooder) {
		return dataRetriever.ErrNilAntifloodHandler
	}
	if check.IfNil(args.PreferredPeersHolder) {
		return dataRetriever.ErrNilPreferredPeersHolder
	}
	return nil
}

func (baseSender *baseTopicSender) sendToConnectedPeer(topic string, buff []byte, peer core.PeerID) error {
	msg := &factory.Message{
		DataField:  buff,
		PeerField:  peer,
		TopicField: topic,
	}

	shouldAvoidAntiFloodCheck := baseSender.preferredPeersHolderHandler.Contains(peer)
	if shouldAvoidAntiFloodCheck {
		return baseSender.messenger.SendToConnectedPeer(topic, buff, peer)
	}

	err := baseSender.outputAntiflooder.CanProcessMessage(msg, peer)
	if err != nil {
		return fmt.Errorf("%w while sending %d bytes to peer %s",
			err,
			len(buff),
			p2p.PeerIdToShortString(peer),
		)
	}

	return baseSender.messenger.SendToConnectedPeer(topic, buff, peer)
}

// DebugHandler returns the debug handler used in resolvers
func (baseSender *baseTopicSender) DebugHandler() dataRetriever.DebugHandler {
	baseSender.mutDebugHandler.RLock()
	defer baseSender.mutDebugHandler.RUnlock()

	return baseSender.debugHandler
}

// SetDebugHandler sets the debug handler used in resolvers
func (baseSender *baseTopicSender) SetDebugHandler(handler dataRetriever.DebugHandler) error {
	if check.IfNil(handler) {
		return dataRetriever.ErrNilDebugHandler
	}

	baseSender.mutDebugHandler.Lock()
	baseSender.debugHandler = handler
	baseSender.mutDebugHandler.Unlock()

	return nil
}

// RequestTopic returns the topic with the request suffix used for sending requests
func (baseSender *baseTopicSender) RequestTopic() string {
	return baseSender.topicName + core.TopicRequestSuffix
}

// TargetShardID returns the target shard ID for this resolver should serve data
func (baseSender *baseTopicSender) TargetShardID() uint32 {
	return baseSender.targetShardId
}
