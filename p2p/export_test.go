package p2p

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
)

func (cn *ConnNotifier) ComputeInboundOutboundConns(conns []net.Conn) (inConns, outConns int) {
	return cn.computeInboundOutboundConns(conns)
}

func (t *Topic) EventBusData() []DataReceivedHandler {
	return t.eventBusDataRcvHandlers
}

func (t *Topic) Marsh() marshal.Marshalizer {
	return t.marsh
}

var DurTimeCache = durTimeCache
var MutGloballyRegPeers = &mutGloballyRegPeers
var Log = log

func RecreateGlobbalyRegisteredMemPeersMap() {
	globallyRegisteredPeers = make(map[peer.ID]*MemMessenger)
}
