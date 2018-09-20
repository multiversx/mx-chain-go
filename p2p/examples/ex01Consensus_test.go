package examples

//
//import (
//	"testing"
//	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
//	"context"
//	"github.com/ElrondNetwork/elrond-go-sandbox/service"
//	"github.com/stretchr/testify/assert"
//	"time"
//	"fmt"
//)
//
//func recv(sender *p2p.Node, peerID string, m *p2p.Message){
//	fmt.Printf("%s got a message: %v and traversed %v\n", sender.P2pNode.ID().Pretty(), m.Payload, m.Peers)
//
//	m.AddHop(sender.P2pNode.ID().Pretty())
//
//	sender.BroadcastMessage(m, []string{peerID})
//}
//
//func TestCons(t *testing.T) {
//
//	nodes := []p2p.Node{}
//
//	for i := 0; i < 10; i++{
//		node, err := p2p.NewNode(context.Background(), 4000 + i, []string{}, service.GetMarshalizerService(), 3)
//		assert.Nil(t, err)
//
//		node.OnMsgRecv = recv
//
//		nodes = append(nodes, *node)
//	}
//
//	cp := p2p.NewClusterParameter("127.0.0.1", 4000, 4009)
//
//	time.Sleep(time.Second)
//
//	for _, node := range nodes{
//		node.Bootstrap(context.Background(), []p2p.ClusterParameter{*cp})
//	}
//
//	time.Sleep(time.Second)
//
//	for _, node := range nodes {
//		conns := node.P2pNode.Network().Conns()
//
//		fmt.Printf("Node %s is connected to: \n", node.P2pNode.ID().Pretty())
//
//		for _, conn := range conns {
//			fmt.Printf("\t- %s\n", conn.RemotePeer().Pretty())
//		}
//	}
//
//	nodes[0].BroadcastString("ABC", []string{})
//
//
//
//
//
//	time.Sleep(time.Second * 2)
//
//	for _, node := range nodes{
//		node.P2pNode.Close()
//	}
//
//}
