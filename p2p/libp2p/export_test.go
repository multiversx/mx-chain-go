package libp2p

import (
	"context"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/go-libp2p-pubsub"
	pb "github.com/ElrondNetwork/go-libp2p-pubsub/pb"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/whyrusleeping/timecache"
)

var MaxSendBuffSize = maxSendBuffSize
var BroadcastGoRoutines = broadcastGoRoutines
var PubsubTimeCacheDuration = pubsubTimeCacheDuration
var AcceptMessagesInAdvanceDuration = acceptMessagesInAdvanceDuration

const CurrentTopicMessageVersion = currentTopicMessageVersion
const PollWaitForConnectionsInterval = pollWaitForConnectionsInterval

// SetHost -
func (netMes *networkMessenger) SetHost(newHost ConnectableHost) {
	netMes.p2pHost = newHost
}

// SetLoadBalancer -
func (netMes *networkMessenger) SetLoadBalancer(outgoingPLB p2p.ChannelLoadBalancer) {
	netMes.outgoingPLB = outgoingPLB
}

// SetPeerDiscoverer -
func (netMes *networkMessenger) SetPeerDiscoverer(discoverer p2p.PeerDiscoverer) {
	netMes.peerDiscoverer = discoverer
}

// PubsubCallback -
func (netMes *networkMessenger) PubsubCallback(handler p2p.MessageProcessor, topic string) func(ctx context.Context, pid peer.ID, message *pubsub.Message) bool {
	topicProcs := newTopicProcessors()
	_ = topicProcs.addTopicProcessor("identifier", handler)

	return netMes.pubsubCallback(topicProcs, topic)
}

// ValidMessageByTimestamp -
func (netMes *networkMessenger) ValidMessageByTimestamp(msg p2p.MessageP2P) error {
	return netMes.validMessageByTimestamp(msg)
}

// MapHistogram -
func (netMes *networkMessenger) MapHistogram(input map[uint32]int) string {
	return netMes.mapHistogram(input)
}

// PubsubHasTopic -
func (netMes *networkMessenger) PubsubHasTopic(expectedTopic string) bool {
	netMes.mutTopics.RLock()
	topics := netMes.pb.GetTopics()
	netMes.mutTopics.RUnlock()

	for _, topic := range topics {
		if topic == expectedTopic {
			return true
		}
	}
	return false
}

// HasProcessorForTopic -
func (netMes *networkMessenger) HasProcessorForTopic(expectedTopic string) bool {
	processor, found := netMes.processors[expectedTopic]

	return found && processor != nil
}

// Disconnect -
func (netMes *networkMessenger) Disconnect(pid core.PeerID) error {
	return netMes.p2pHost.Network().ClosePeer(peer.ID(pid))
}

// ProcessReceivedDirectMessage -
func (ds *directSender) ProcessReceivedDirectMessage(message *pb.Message, fromConnectedPeer peer.ID) error {
	return ds.processReceivedDirectMessage(message, fromConnectedPeer)
}

// SeenMessages -
func (ds *directSender) SeenMessages() *timecache.TimeCache {
	return ds.seenMessages
}

// Counter -
func (ds *directSender) Counter() uint64 {
	return ds.counter
}

// Mutexes -
func (mh *MutexHolder) Mutexes() storage.Cacher {
	return mh.mutexes
}
