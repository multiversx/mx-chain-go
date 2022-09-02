package heartbeat

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/heartbeat"
	mock2 "github.com/ElrondNetwork/elrond-go/heartbeat/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl"
	mclsig "github.com/ElrondNetwork/elrond-go-crypto/signing/mcl/singlesig"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/heartbeat/process"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
	statusHandlerMock "github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

var log = logger.GetOrCreate("integrationtests/node")

var handlers []vmcommon.EpochSubscriberHandler

const (
	stepDelay                 = time.Second / 10
	durationBetweenHeartbeats = time.Second * 5
	providedEpoch             = uint32(11)
)

// TestHeartbeatMonitorWillUpdateAnInactivePeer test what happen if a peer out of 2 stops being responsive on heartbeat status
// The active monitor should change it's active flag to false when a new heartbeat message has arrived.
func TestHeartbeatMonitorWillUpdateAnInactivePeer(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	interactingNodes := 3
	nodes := make([]p2p.Messenger, interactingNodes)

	maxUnresposiveTime := time.Second * 10
	monitor := createMonitor(maxUnresposiveTime)

	senders, pks := prepareNodes(nodes, monitor, interactingNodes, "nodeName")

	defer func() {
		for _, n := range nodes {
			_ = n.Close()
		}
	}()

	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(integrationTests.P2pBootstrapDelay)

	fmt.Println("Sending first messages from both public keys...")
	err := senders[0].SendHeartbeat()
	log.LogIfError(err)

	err = senders[1].SendHeartbeat()
	log.LogIfError(err)

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

func TestHeartbeatMonitorWillNotUpdateTooLongHeartbeatMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	length := 129
	buff := make([]byte, length)

	for i := 0; i < length; i++ {
		buff[i] = byte(97)
	}
	bigNodeName := string(buff)

	interactingNodes := 3
	nodes := make([]p2p.Messenger, interactingNodes)

	maxUnresposiveTime := time.Second * 10
	monitor := createMonitor(maxUnresposiveTime)

	senders, pks := prepareNodes(nodes, monitor, interactingNodes, bigNodeName)

	defer func() {
		for _, n := range nodes {
			_ = n.Close()
		}
	}()

	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(integrationTests.P2pBootstrapDelay)

	fmt.Println("Sending first messages with long name...")
	_ = senders[1].SendHeartbeat()

	time.Sleep(stepDelay)

	secondPK := pks[1]

	pkHeartBeats := monitor.GetHeartbeats()

	assert.True(t, isPkActive(pkHeartBeats, secondPK))
	expectedLen := 128
	assert.True(t, isMessageCorrectLen(pkHeartBeats, secondPK, expectedLen))
}

func TestHeartbeatV2_DeactivationOfHeartbeat(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	interactingNodes := 3
	nodes := make([]*integrationTests.TestHeartbeatNode, interactingNodes)
	p2pConfig := integrationTests.CreateP2PConfigWithNoDiscovery()
	for i := 0; i < interactingNodes; i++ {
		nodes[i] = integrationTests.NewTestHeartbeatNode(t, 3, 0, interactingNodes, p2pConfig, 60)
	}
	assert.Equal(t, interactingNodes, len(nodes))

	messengers := make([]p2p.Messenger, interactingNodes)
	for i := 0; i < interactingNodes; i++ {
		messengers[i] = nodes[i].Messenger
	}

	maxUnresposiveTime := time.Second * 10
	monitor := createMonitor(maxUnresposiveTime)
	senders, _ := prepareNodes(messengers, monitor, interactingNodes, "nodeName")

	// Start sending heartbeats
	timer := time.NewTimer(durationBetweenHeartbeats)
	defer timer.Stop()
	go startSendingHeartbeats(t, senders, timer)

	// Wait for first messages
	time.Sleep(time.Second * 6)

	heartbeats := monitor.GetHeartbeats()
	assert.False(t, heartbeats[0].IsActive) // first one is the monitor which is inactive

	for _, hb := range heartbeats[1:] {
		assert.True(t, hb.IsActive)
	}

	// Stop sending heartbeats
	for _, handler := range handlers {
		handler.EpochConfirmed(providedEpoch+1, 0)
	}

	// Wait enough time to make sure some heartbeats should have been sent
	time.Sleep(time.Second * 15)

	// Check sent messages
	maxHbV2DurationAllowed := time.Second * 5
	checkMessages(t, nodes, monitor, maxHbV2DurationAllowed)
}

func startSendingHeartbeats(t *testing.T, senders []*process.Sender, timer *time.Timer) {
	for {
		timer.Reset(durationBetweenHeartbeats)

		<-timer.C
		for _, sender := range senders {
			err := sender.SendHeartbeat()
			assert.Nil(t, err)
		}
	}
}

func checkMessages(t *testing.T, nodes []*integrationTests.TestHeartbeatNode, monitor *process.Monitor, maxHbV2DurationAllowed time.Duration) {
	heartbeats := monitor.GetHeartbeats()
	for _, hb := range heartbeats {
		assert.False(t, hb.IsActive)
	}

	numOfNodes := len(nodes)
	for i := 0; i < numOfNodes; i++ {
		paCache := nodes[i].DataPool.PeerAuthentications()
		hbCache := nodes[i].DataPool.Heartbeats()

		assert.Equal(t, numOfNodes, paCache.Len())
		assert.Equal(t, numOfNodes, hbCache.Len())

		// Check this node received messages from all peers
		for _, node := range nodes {
			pkBytes, err := node.NodeKeys.Pk.ToByteArray()
			assert.Nil(t, err)

			assert.True(t, paCache.Has(pkBytes))
			assert.True(t, hbCache.Has(node.Messenger.ID().Bytes()))

			// Also check message age
			value, _ := paCache.Get(pkBytes)
			msg := value.(*heartbeat.PeerAuthentication)

			marshaller := integrationTests.TestMarshaller
			payload := &heartbeat.Payload{}
			err = marshaller.Unmarshal(payload, msg.Payload)
			assert.Nil(t, err)

			currentTimestamp := time.Now().Unix()
			messageAge := time.Duration(currentTimestamp - payload.Timestamp)
			assert.True(t, messageAge < maxHbV2DurationAllowed)
		}
	}
}

func prepareNodes(
	nodes []p2p.Messenger,
	monitor *process.Monitor,
	interactingNodes int,
	defaultNodeName string,
) ([]*process.Sender, []crypto.PublicKey) {

	senderIdxs := []int{0, 1}
	topicHeartbeat := "topic"
	senders := make([]*process.Sender, 0)
	pks := make([]crypto.PublicKey, 0)
	handlers = make([]vmcommon.EpochSubscriberHandler, 0)

	for i := 0; i < interactingNodes; i++ {
		if nodes[i] == nil {
			nodes[i] = integrationTests.CreateMessengerWithNoDiscovery()
		}
		_ = nodes[i].CreateTopic(topicHeartbeat, true)

		isSender := integrationTests.IsIntInSlice(i, senderIdxs)
		if isSender {
			sender, pk := createSenderWithName(nodes[i], topicHeartbeat, defaultNodeName)
			senders = append(senders, sender)
			pks = append(pks, pk)
		} else {
			_ = nodes[i].RegisterMessageProcessor(topicHeartbeat, "test", monitor)
		}
	}

	for i := 0; i < len(nodes)-1; i++ {
		for j := i + 1; j < len(nodes); j++ {
			_ = nodes[i].ConnectToPeer(integrationTests.GetConnectableAddress(nodes[j]))
		}
	}

	return senders, pks
}

func checkReceivedMessages(t *testing.T, monitor *process.Monitor, pks []crypto.PublicKey, activeIdxs []int) {
	pkHeartBeats := monitor.GetHeartbeats()

	extraPkInMonitor := 1
	assert.Equal(t, len(pks), len(pkHeartBeats)-extraPkInMonitor)

	for idx, pk := range pks {
		pkShouldBeActive := integrationTests.IsIntInSlice(idx, activeIdxs)
		assert.Equal(t, pkShouldBeActive, isPkActive(pkHeartBeats, pk))
		assert.True(t, isMessageReceived(pkHeartBeats, pk))
	}
}

func isMessageReceived(heartbeats []data.PubKeyHeartbeat, pk crypto.PublicKey) bool {
	pkBytes, _ := pk.ToByteArray()

	for _, hb := range heartbeats {
		isPkMatching := hb.PublicKey == integrationTests.TestValidatorPubkeyConverter.Encode(pkBytes)
		if isPkMatching {
			return true
		}
	}

	return false
}

func isPkActive(heartbeats []data.PubKeyHeartbeat, pk crypto.PublicKey) bool {
	pkBytes, _ := pk.ToByteArray()

	for _, hb := range heartbeats {
		isPkMatchingAndActve := hb.PublicKey == integrationTests.TestValidatorPubkeyConverter.Encode(pkBytes) && hb.IsActive
		if isPkMatchingAndActve {
			return true
		}
	}

	return false
}

func isMessageCorrectLen(heartbeats []data.PubKeyHeartbeat, pk crypto.PublicKey, expectedLen int) bool {
	pkBytes, _ := pk.ToByteArray()

	for _, hb := range heartbeats {
		isPkMatching := hb.PublicKey == integrationTests.TestValidatorPubkeyConverter.Encode(pkBytes)
		if isPkMatching {
			return len(hb.NodeDisplayName) == expectedLen
		}
	}

	return false
}

func createSenderWithName(messenger p2p.Messenger, topic string, nodeName string) (*process.Sender, crypto.PublicKey) {
	suite := mcl.NewSuiteBLS12()
	signer := &mclsig.BlsSingleSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	sk, pk := keyGen.GeneratePair()
	version := "v01"

	argSender := process.ArgHeartbeatSender{
		PeerMessenger:        messenger,
		PeerSignatureHandler: &mock2.PeerSignatureHandler{Signer: signer},
		PrivKey:              sk,
		Marshalizer:          integrationTests.TestMarshalizer,
		Topic:                topic,
		ShardCoordinator:     &sharding.OneShardCoordinator{},
		PeerTypeProvider:     &mock.PeerTypeProviderStub{},
		StatusHandler:        &statusHandlerMock.AppStatusHandlerStub{},
		VersionNumber:        version,
		NodeDisplayName:      nodeName,
		HardforkTrigger:      &testscommon.HardforkTriggerStub{},
		CurrentBlockProvider: &testscommon.ChainHandlerStub{},
		RedundancyHandler:    &mock.RedundancyHandlerStub{},
		EpochNotifier: &epochNotifier.EpochNotifierStub{
			RegisterNotifyHandlerCalled: func(handler vmcommon.EpochSubscriberHandler) {
				handlers = append(handlers, handler)
			},
		},
		HeartbeatDisableEpoch: providedEpoch,
	}

	sender, _ := process.NewSender(argSender)
	return sender, pk
}

func createMonitor(maxDurationPeerUnresponsive time.Duration) *process.Monitor {
	suite := mcl.NewSuiteBLS12()
	singlesigner := &mclsig.BlsSingleSigner{}
	keyGen := signing.NewKeyGenerator(suite)
	marshalizer := &marshal.GogoProtoMarshalizer{}

	mp, _ := process.NewMessageProcessor(
		&mock2.PeerSignatureHandler{Signer: singlesigner, KeyGen: keyGen},
		marshalizer,
		&p2pmocks.NetworkShardingCollectorStub{},
	)

	argMonitor := process.ArgHeartbeatMonitor{
		Marshalizer:                 integrationTests.TestMarshalizer,
		MaxDurationPeerUnresponsive: maxDurationPeerUnresponsive,
		PubKeysMap:                  map[uint32][]string{0: {""}},
		GenesisTime:                 time.Now(),
		MessageHandler:              mp,
		Storer: &mock.HeartbeatStorerStub{
			UpdateGenesisTimeCalled: func(genesisTime time.Time) error {
				return nil
			},
			LoadHeartBeatDTOCalled: func(pubKey string) (*data.HeartbeatDTO, error) {
				return nil, errors.New("not found")
			},
			LoadKeysCalled: func() ([][]byte, error) {
				return nil, nil
			},
			SavePubkeyDataCalled: func(pubkey []byte, heartbeat *data.HeartbeatDTO) error {
				return nil
			},
			SaveKeysCalled: func(peersSlice [][]byte) error {
				return nil
			},
		},
		PeerTypeProvider: &mock.PeerTypeProviderStub{},
		Timer:            &process.RealTimer{},
		AntifloodHandler: &mock.P2PAntifloodHandlerStub{
			CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
				return nil
			},
		},
		HardforkTrigger:                    &testscommon.HardforkTriggerStub{},
		ValidatorPubkeyConverter:           integrationTests.TestValidatorPubkeyConverter,
		HeartbeatRefreshIntervalInSec:      1,
		HideInactiveValidatorIntervalInSec: 600,
		AppStatusHandler:                   &statusHandlerMock.AppStatusHandlerStub{},
		EpochNotifier:                      &epochNotifier.EpochNotifierStub{},
		HeartbeatDisableEpoch:              providedEpoch,
	}

	monitor, _ := process.NewMonitor(argMonitor)

	return monitor
}
