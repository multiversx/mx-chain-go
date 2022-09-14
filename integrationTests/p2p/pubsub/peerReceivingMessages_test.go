package peerDisconnecting

import (
	"crypto/ecdsa"
	cryptoRand "crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	messagecheck "github.com/ElrondNetwork/elrond-go/p2p/messageCheck"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var durationTest = 30 * time.Second

type messageProcessorStub struct {
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
}

// ProcessReceivedMessage -
func (mps *messageProcessorStub) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID) error {
	return mps.ProcessReceivedMessageCalled(message)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mps *messageProcessorStub) IsInterfaceNil() bool {
	return mps == nil
}

func TestPeerReceivesTheSameMessageMultipleTimesShouldNotHappen(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfPeers := 20

	//Step 1. Create advertiser
	advertiser := integrationTests.CreateMessengerWithKadDht("")

	//Step 2. Create numOfPeers instances of messenger type and call bootstrap
	peers := make([]p2p.Messenger, numOfPeers)
	for i := 0; i < numOfPeers; i++ {
		node := integrationTests.CreateMessengerWithKadDht(integrationTests.GetConnectableAddress(advertiser))
		peers[i] = node
	}

	//cleanup function that closes all messengers
	defer func() {
		for i := 0; i < numOfPeers; i++ {
			if peers[i] != nil {
				_ = peers[i].Close()
			}
		}

		if advertiser != nil {
			_ = advertiser.Close()
		}
	}()

	chanStop := make(chan struct{})

	//Step 3. Register pubsub validators
	mutMapMessages := sync.Mutex{}
	mapMessages := make(map[int]map[string]struct{})
	testTopic := "test"

	for i := 0; i < numOfPeers; i++ {
		idx := i
		mapMessages[idx] = make(map[string]struct{})
		err := peers[idx].CreateTopic(testTopic, true)
		if err != nil {
			fmt.Println("CreateTopic failed:", err.Error())
			continue
		}

		err = peers[idx].RegisterMessageProcessor(testTopic, "test", &messageProcessorStub{
			ProcessReceivedMessageCalled: func(message p2p.MessageP2P) error {
				time.Sleep(time.Second)

				mutMapMessages.Lock()
				defer mutMapMessages.Unlock()

				msgId := "peer: " + message.Peer().Pretty() + " - seqNo: 0x" + hex.EncodeToString(message.SeqNo())
				_, ok := mapMessages[idx][msgId]
				if ok {
					assert.Fail(t, "message %s received twice", msgId)
					chanStop <- struct{}{}
				}

				mapMessages[idx][msgId] = struct{}{}
				return nil
			},
		})
		if err != nil {
			fmt.Println("RegisterMessageProcessor:", err.Error())
		}
	}

	//Step 4. Call bootstrap on all peers
	err := advertiser.Bootstrap()
	if err != nil {
		fmt.Println("Bootstrap failed:", err.Error())
	}
	for _, p := range peers {
		err = p.Bootstrap()
		if err != nil {
			fmt.Printf("Bootstrap() for peer id %s failed:%s\n", p.ID(), err.Error())
		}
	}
	integrationTests.WaitForBootstrapAndShowConnected(peers, integrationTests.P2pBootstrapDelay)

	//Step 5. Continuously send messages from one peer
	for timeStart := time.Now(); timeStart.Add(durationTest).Unix() > time.Now().Unix(); {
		peers[0].Broadcast(testTopic, []byte("test buff"))
		select {
		case <-chanStop:
			return
		default:
		}
		time.Sleep(time.Millisecond)
	}
}

// TestBroadcastMessageComesFormTheConnectedPeers tests what happens in a network when a message comes through pubsub
// The receiving peer should get the message only from one of the connected peers
func TestBroadcastMessageComesFormTheConnectedPeers(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	topic := "test_topic"
	broadcastMessageDuration := time.Second * 2
	peers, err := integrationTests.CreateFixedNetworkOf8Peers()
	assert.Nil(t, err)

	defer func() {
		integrationTests.ClosePeers(peers)
	}()

	//node 0 is connected only to 1 and 3 (check integrationTests.CreateFixedNetworkOf7Peers function)
	//a broadcast message from 6 should be received on node 0 only through peers 1 and 3

	interceptors, err := createTopicsAndMockInterceptors(peers, topic)
	assert.Nil(t, err)

	fmt.Println("bootstrapping nodes")
	time.Sleep(integrationTests.P2pBootstrapDelay)

	broadcastIdx := 6
	receiverIdx := 0
	shouldReceiveFrom := []int{1, 3}

	broadcastPeer := peers[broadcastIdx]
	fmt.Printf("broadcasting message from pid %s\n", broadcastPeer.ID().Pretty())
	broadcastPeer.Broadcast(topic, []byte("dummy"))
	time.Sleep(broadcastMessageDuration)

	countReceivedMessages := 0
	receiverInterceptor := interceptors[receiverIdx]
	for _, idx := range shouldReceiveFrom {
		connectedPid := peers[idx].ID()
		countReceivedMessages += len(receiverInterceptor.Messages(connectedPid))
	}

	assert.Equal(t, 1, countReceivedMessages)
}

func generatePrivateKey() *libp2pCrypto.Secp256k1PrivateKey {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)

	return (*libp2pCrypto.Secp256k1PrivateKey)(prvKey)
}

func TestP2PMessageSinging(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	p2pSigner, err := libp2p.NewP2PSigner(generatePrivateKey())
	require.Nil(t, err)

	messageVerifierArgs := messagecheck.ArgsMessageVerifier{
		Marshaller: integrationTests.TestTxSignMarshalizer,
		P2PSigner:  p2pSigner,
	}
	messageVerifier, err := messagecheck.NewMessageVerifier(messageVerifierArgs)
	require.Nil(t, err)

	topic := "test_topic"

	peers := make([]p2p.Messenger, 0)

	peer1 := integrationTests.CreateMessengerWithNoDiscovery()
	peers = append(peers, peer1)
	peer2 := integrationTests.CreateMessengerWithNoDiscovery()
	peers = append(peers, peer2)

	err = peer1.ConnectToPeer(peer2.Addresses()[0])
	assert.Nil(t, err)

	defer func() {
		integrationTests.ClosePeers(peers)
	}()

	interceptors, err := createTopicsAndMockInterceptors(peers, topic)
	assert.Nil(t, err)

	err = peer1.SendToConnectedPeer(topic, []byte("dummy"), peer2.ID())
	assert.Nil(t, err)

	time.Sleep(time.Second * 2)

	receiverInterceptor := interceptors[1]
	p2pMessages := receiverInterceptor.Messages(peer1.ID())

	for _, p2pMsg := range p2pMessages {
		err := messageVerifier.Verify(p2pMsg)
		require.Nil(t, err)
	}
}

func createTopicsAndMockInterceptors(peers []p2p.Messenger, topic string) ([]*messageProcessor, error) {
	interceptors := make([]*messageProcessor, len(peers))

	for idx, p := range peers {
		err := p.CreateTopic(topic, true)
		if err != nil {
			return nil, fmt.Errorf("%w, pid: %s", err, p.ID())
		}

		interceptors[idx] = newMessageProcessor()
		err = p.RegisterMessageProcessor(topic, "test", interceptors[idx])
		if err != nil {
			return nil, fmt.Errorf("%w, pid: %s", err, p.ID())
		}
	}

	return interceptors, nil
}
