package libp2p

import (
	"github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p-interface-connmgr"
	"github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/whyrusleeping/timecache"
)

func (p2pMes *libp2pMessenger) SetDiscoverer(discoverer discovery.Discoverer) {
	p2pMes.discoverer = discoverer
}

func (p2pMes *libp2pMessenger) ConnManager() ifconnmgr.ConnManager {
	return p2pMes.connManager
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
