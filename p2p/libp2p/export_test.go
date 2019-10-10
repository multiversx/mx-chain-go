package libp2p

import (
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/whyrusleeping/timecache"
)

var MaxSendBuffSize = maxSendBuffSize
var BroadcastGoRoutines = broadcastGoRoutines

func (netMes *networkMessenger) ConnManager() connmgr.ConnManager {
	return netMes.ctxProvider.connHost.ConnManager()
}

func (netMes *networkMessenger) SetHost(newHost ConnectableHost) {
	netMes.ctxProvider.connHost = newHost
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

func (mh *MutexHolder) Mutexes() *lrucache.LRUCache {
	return mh.mutexes
}
