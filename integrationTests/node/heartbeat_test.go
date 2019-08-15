package node

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

var stepDelay = time.Second

// TestHeartbeatMonitorWillUpdateAnInactivePeer test what happen if a peer out of 2 stops being responsive on heartbeat status
// The active monitor should change it's active flag to false when a new heartbeat message has arrived.
func TestHeartbeatMonitorWillUpdateAnInactivePeer(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)
	maxUnresposiveTime := time.Second * 2

	monitor := createMonitor(maxUnresposiveTime)
	nodes, senders, pks := prepareNodes(advertiserAddr, monitor)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Close()
		}
	}()

	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(time.Second * 5)

	fmt.Println("Sending first messages from both public keys...")
	_ = senders[0].SendHeartbeat()
	_ = senders[1].SendHeartbeat()

	time.Sleep(stepDelay)

	fmt.Println("Checking both public keys are active...")
	checkReceivedMessages(t, monitor, pks, []int{0, 1})

	fmt.Println("Waiting for max unresponsive time...")
	time.Sleep(maxUnresposiveTime)

	fmt.Println("Only first pk will send another message...")
	_ = senders[0].SendHeartbeat()

	time.Sleep(stepDelay)

	fmt.Println("Checking only first pk is active...")
	checkReceivedMessages(t, monitor, pks, []int{0})
}

func prepareNodes(
	advertiserAddr string,
	monitor *heartbeat.Monitor,
) ([]p2p.Messenger, []*heartbeat.Sender, []crypto.PublicKey) {

	senderIdxs := []int{0, 1}
	interactingNodes := 3
	nodes := make([]p2p.Messenger, interactingNodes)
	topicHeartbeat := "topic"
	senders := make([]*heartbeat.Sender, 0)
	pks := make([]crypto.PublicKey, 0)

	for i := 0; i < interactingNodes; i++ {
		nodes[i] = integrationTests.CreateMessengerWithKadDht(context.Background(), advertiserAddr)
		_ = nodes[i].CreateTopic(topicHeartbeat, true)

		isSender := integrationTests.IsIntInSlice(i, senderIdxs)
		if isSender {
			sender, pk := createSender(nodes[i], topicHeartbeat)
			senders = append(senders, sender)
			pks = append(pks, pk)
		} else {
			_ = nodes[i].RegisterMessageProcessor(topicHeartbeat, monitor)
		}

		_ = nodes[i].Bootstrap()
	}

	return nodes, senders, pks
}

func checkReceivedMessages(t *testing.T, monitor *heartbeat.Monitor, pks []crypto.PublicKey, activeIdxs []int) {
	pkHeartBeats := monitor.GetHeartbeats()

	extraPkInMonitor := 1
	assert.Equal(t, len(pks), len(pkHeartBeats)-extraPkInMonitor)

	for idx, pk := range pks {
		pkShouldBeActive := integrationTests.IsIntInSlice(idx, activeIdxs)
		assert.Equal(t, pkShouldBeActive, isPkActive(pkHeartBeats, pk))
		assert.True(t, isMessageReceived(pkHeartBeats, pk))
	}
}

func isMessageReceived(heartbeats []heartbeat.PubKeyHeartbeat, pk crypto.PublicKey) bool {
	pkBytes, _ := pk.ToByteArray()

	for _, hb := range heartbeats {
		isPkMatching := hb.HexPublicKey == hex.EncodeToString(pkBytes)
		if isPkMatching {
			return true
		}
	}

	return false
}

func isPkActive(heartbeats []heartbeat.PubKeyHeartbeat, pk crypto.PublicKey) bool {
	pkBytes, _ := pk.ToByteArray()

	for _, hb := range heartbeats {
		isPkMatchingAndActve := hb.HexPublicKey == hex.EncodeToString(pkBytes) && hb.IsActive
		if isPkMatchingAndActve {
			return true
		}
	}

	return false
}

func createSender(messenger p2p.Messenger, topic string) (*heartbeat.Sender, crypto.PublicKey) {
	suite := kyber.NewBlakeSHA256Ed25519()
	signer := &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()

	version := "v01"
	nodeName := "nodeName"
	sender, _ := heartbeat.NewSender(
		messenger,
		signer,
		sk,
		integrationTests.TestMarshalizer,
		topic,
		&sharding.OneShardCoordinator{},
		version,
		nodeName,
	)

	return sender, pk
}

func createMonitor(maxDurationPeerUnresponsive time.Duration) *heartbeat.Monitor {
	suite := kyber.NewBlakeSHA256Ed25519()
	signer := &singlesig.SchnorrSigner{}
	keyGen := signing.NewKeyGenerator(suite)

	monitor, _ := heartbeat.NewMonitor(
		signer,
		keyGen,
		integrationTests.TestMarshalizer,
		maxDurationPeerUnresponsive,
		map[uint32][]string{0: {""}},
	)

	return monitor
}
