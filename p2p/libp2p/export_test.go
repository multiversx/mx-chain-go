package libp2p

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/whyrusleeping/timecache"
)

var MaxSendBuffSize = maxSendBuffSize
var BroadcastGoRoutines = broadcastGoRoutines

func (netMes *networkMessenger) SetHost(newHost ConnectableHost) {
	netMes.p2pHost = newHost
}

func (netMes *networkMessenger) SetLoadBalancer(outgoingPLB p2p.ChannelLoadBalancer) {
	netMes.outgoingPLB = outgoingPLB
}

func (netMes *networkMessenger) SetPeerDiscoverer(discoverer p2p.PeerDiscoverer) {
	netMes.peerDiscoverer = discoverer
}

func (ds *directSender) ProcessReceivedDirectMessage(message *pubsub_pb.Message, fromConnectedPeer peer.ID) error {
	return ds.processReceivedDirectMessage(message, fromConnectedPeer)
}

func (ds *directSender) SeenMessages() *timecache.TimeCache {
	return ds.seenMessages
}

func (ds *directSender) Counter() uint64 {
	return ds.counter
}

func (mh *MutexHolder) Mutexes() storage.Cacher {
	return mh.mutexes
}

func (ip *identityProvider) HandleStreams(s network.Stream) {
	ip.handleStreams(s)
}

func (ip *identityProvider) ProcessReceivedData(recvBuff []byte) error {
	return ip.processReceivedData(recvBuff)
}
