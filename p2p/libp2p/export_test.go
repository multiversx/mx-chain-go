package libp2p

import (
	"github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-interface-connmgr"
	"github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/whyrusleeping/timecache"
)

func (netMes *networkMessenger) SetDiscoverer(discoverer discovery.Discoverer) {
	netMes.discoverer = discoverer
}

func (netMes *networkMessenger) ConnManager() ifconnmgr.ConnManager {
	return netMes.hostP2P.ConnManager()
}

func (netMes *networkMessenger) SetPeerDiscoveredHandler(handler PeerInfoHandler) {
	netMes.peerDiscoveredHandler = handler
}

func (netMes *networkMessenger) SetHost(host host.Host) {
	netMes.hostP2P = host
}

func (ds *directSender) ProcessReceivedDirectMessage(message *pubsub_pb.Message) error {
	return ds.processReceivedDirectMessage(message)
}

func (ds *directSender) SeenMessages() *timecache.TimeCache {
	return ds.seenMessages
}

func (ds *directSender) Counter() uint64 {
	return ds.counter
}
