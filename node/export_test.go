package node

import (
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

func (n *Node) HeartbeatMonitor() *heartbeat.Monitor {
	return n.heartbeatMonitor
}

func (n *Node) HeartbeatSender() *heartbeat.Sender {
	return n.heartbeatSender
}

func (n *Node) CreateConsensusTopic(messageProcessor p2p.MessageProcessor) error {
	return n.createConsensusTopic(messageProcessor)
}
