package heartbeatV2

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("inteagrationtests/node/heartbeatv2")
var keygen = signing.NewKeyGenerator(mcl.NewSuiteBLS12())

func generateTestKeyPay() *integrationTests.TestKeyPair {
	sk, pk := keygen.GeneratePair()
	return &integrationTests.TestKeyPair{
		Sk: sk,
		Pk: pk,
	}
}

func TestHeartbeatV2_AllPeersSendMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	interactingNodes := 3
	nodes := make([]*integrationTests.TestHeartbeatNode, interactingNodes)
	p2pConfig := integrationTests.CreateP2PConfigWithNoDiscovery()
	for i := 0; i < interactingNodes; i++ {
		key := &integrationTests.TestNodeKeys{
			MainKey: generateTestKeyPay(),
		}

		nodes[i] = integrationTests.NewTestHeartbeatNode(t, 3, 0, interactingNodes, p2pConfig, 60, key, nil)
	}
	assert.Equal(t, interactingNodes, len(nodes))

	connectNodes(nodes, interactingNodes)

	// Wait for messages to broadcast
	time.Sleep(time.Second * 15)

	for i := 0; i < len(nodes); i++ {
		nodes[i].Close()
	}

	// Check sent messages
	maxMessageAgeAllowed := time.Second * 5
	checkMessages(t, nodes, maxMessageAgeAllowed)
}

func TestHeartbeatV2_AllPeersSendMessagesOnMultikeyMode(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numHandledKeys := 6
	handledKeys := make([]*integrationTests.TestKeyPair, 0, numHandledKeys)
	for i := 0; i < numHandledKeys; i++ {
		key := generateTestKeyPay()
		handledKeys = append(handledKeys, key)

		pkBytes, _ := key.Pk.ToByteArray()
		log.Debug("generated handled key", "pubkey", pkBytes)
	}
	validators := []*integrationTests.TestKeyPair{handledKeys[1], handledKeys[2], handledKeys[4], handledKeys[5]}

	p2pConfig := integrationTests.CreateP2PConfigWithNoDiscovery()
	node0Keys := &integrationTests.TestNodeKeys{
		MainKey:     handledKeys[0],
		HandledKeys: []*integrationTests.TestKeyPair{handledKeys[1], handledKeys[2]},
	}
	node0 := integrationTests.NewTestHeartbeatNode(t, 3, 0, 2, p2pConfig, 60, node0Keys, validators)

	node1Keys := &integrationTests.TestNodeKeys{
		MainKey:     handledKeys[3],
		HandledKeys: []*integrationTests.TestKeyPair{handledKeys[4], handledKeys[5]},
	}
	node1 := integrationTests.NewTestHeartbeatNode(t, 3, 0, 2, p2pConfig, 60, node1Keys, validators)

	nodes := []*integrationTests.TestHeartbeatNode{node0, node1}
	connectNodes(nodes, len(nodes))

	// Wait for messages to broadcast
	time.Sleep(time.Second * 15)

	for i := 0; i < len(nodes); i++ {
		nodes[i].Close()
	}

	// Check sent messages
	maxMessageAgeAllowed := time.Second * 5

	checkMessages(t, nodes, maxMessageAgeAllowed)
}

func TestHeartbeatV2_PeerJoiningLate(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	interactingNodes := 3
	nodes := make([]*integrationTests.TestHeartbeatNode, interactingNodes)
	p2pConfig := integrationTests.CreateP2PConfigWithNoDiscovery()
	for i := 0; i < interactingNodes; i++ {
		key := &integrationTests.TestNodeKeys{
			MainKey: generateTestKeyPay(),
		}

		nodes[i] = integrationTests.NewTestHeartbeatNode(t, 3, 0, interactingNodes, p2pConfig, 60, key, nil)
	}
	assert.Equal(t, interactingNodes, len(nodes))

	connectNodes(nodes, interactingNodes)

	// Wait for messages to broadcast
	time.Sleep(time.Second * 15)

	// Check sent messages
	maxMessageAgeAllowed := time.Second * 5
	checkMessages(t, nodes, maxMessageAgeAllowed)

	// Add new delayed node which requests messages
	key := &integrationTests.TestNodeKeys{
		MainKey: generateTestKeyPay(),
	}
	delayedNode := integrationTests.NewTestHeartbeatNode(t, 3, 0, 0, p2pConfig, 60, key, nil)
	nodes = append(nodes, delayedNode)
	connectNodes(nodes, len(nodes))
	// Wait for messages to broadcast and requests to finish
	time.Sleep(time.Second * 15)

	for i := 0; i < len(nodes); i++ {
		nodes[i].Close()
	}

	// Check sent messages again - now should have from all peers
	maxMessageAgeAllowed = time.Second * 5 // should not have messages from first Send
	checkMessages(t, nodes, maxMessageAgeAllowed)
}

func TestHeartbeatV2_PeerAuthenticationMessageExpiration(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	interactingNodes := 3
	nodes := make([]*integrationTests.TestHeartbeatNode, interactingNodes)
	p2pConfig := integrationTests.CreateP2PConfigWithNoDiscovery()
	for i := 0; i < interactingNodes; i++ {
		key := &integrationTests.TestNodeKeys{
			MainKey: generateTestKeyPay(),
		}

		nodes[i] = integrationTests.NewTestHeartbeatNode(t, 3, 0, interactingNodes, p2pConfig, 20, key, nil)
	}
	assert.Equal(t, interactingNodes, len(nodes))

	connectNodes(nodes, interactingNodes)

	log.Info("wait for messages to broadcast...")
	time.Sleep(time.Second * 15)

	log.Info("closing the last node")
	lastNode := nodes[interactingNodes-1]
	lastNode.Close()

	log.Info("waiting the messages from the last node expire")
	time.Sleep(time.Second * 30)

	log.Info("first node clears its peer authentication messages data pool after all other nodes stop sending peer authentication messages")
	for i := 1; i < len(nodes)-1; i++ {
		_ = nodes[i].Sender.Close()
	}
	time.Sleep(time.Second * 2)
	nodes[0].DataPool.PeerAuthentications().Clear()

	log.Info("first node requests the peer authentication messages from other peer(s)")

	requestHashes := make([][]byte, 0)
	for i := 1; i < len(nodes); i++ {
		pkBytes, err := nodes[i].NodeKeys.MainKey.Pk.ToByteArray()
		assert.Nil(t, err)

		requestHashes = append(requestHashes, pkBytes)
	}

	nodes[0].RequestHandler.RequestPeerAuthenticationsByHashes(nodes[0].ShardCoordinator.SelfId(), requestHashes)

	time.Sleep(time.Second * 5)

	// first node should have not received the requested message because it is expired
	lastPkBytes := requestHashes[len(requestHashes)-1]
	assert.False(t, nodes[0].RequestedItemsHandler.Has(string(lastPkBytes)))
	assert.Equal(t, interactingNodes-2, nodes[0].DataPool.PeerAuthentications().Len())
}

func TestHeartbeatV2_AllPeersSendMessagesOnAllNetworks(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	interactingNodes := 3
	nodes := make([]*integrationTests.TestHeartbeatNode, interactingNodes)
	p2pConfig := integrationTests.CreateP2PConfigWithNoDiscovery()
	for i := 0; i < interactingNodes; i++ {
		key := &integrationTests.TestNodeKeys{
			MainKey: generateTestKeyPay(),
		}

		nodes[i] = integrationTests.NewTestHeartbeatNode(t, 3, 0, interactingNodes, p2pConfig, 60, key, nil)
	}
	assert.Equal(t, interactingNodes, len(nodes))

	// connect nodes on main network only
	for i := 0; i < interactingNodes-1; i++ {
		for j := i + 1; j < interactingNodes; j++ {
			src := nodes[i]
			dst := nodes[j]
			_ = src.ConnectOnMain(dst)
		}
	}

	// Wait for messages to broadcast
	time.Sleep(time.Second * 15)

	// check peer shard mappers
	// full archive should not be updated at this point
	for i := 0; i < interactingNodes; i++ {
		for j := 0; j < interactingNodes; j++ {
			if i == j {
				continue
			}

			peerInfo := nodes[i].FullArchivePeerShardMapper.GetPeerInfo(nodes[j].MainMessenger.ID())
			assert.Equal(t, core.UnknownPeer, peerInfo.PeerType) // nodes not connected on this network

			peerInfoMain := nodes[i].MainPeerShardMapper.GetPeerInfo(nodes[j].MainMessenger.ID())
			assert.Equal(t, nodes[j].ShardCoordinator.SelfId(), peerInfoMain.ShardID)
			assert.Equal(t, core.ValidatorPeer, peerInfoMain.PeerType) // on main network they are all validators
		}
	}

	// connect nodes on full archive network as well
	for i := 0; i < interactingNodes-1; i++ {
		for j := i + 1; j < interactingNodes; j++ {
			src := nodes[i]
			dst := nodes[j]
			_ = src.ConnectOnFullArchive(dst)
		}
	}

	// Wait for messages to broadcast
	time.Sleep(time.Second * 15)

	// check peer shard mappers
	// full archive should be updated at this point
	for i := 0; i < interactingNodes; i++ {
		for j := 0; j < interactingNodes; j++ {
			if i == j {
				continue
			}

			peerInfo := nodes[i].FullArchivePeerShardMapper.GetPeerInfo(nodes[j].MainMessenger.ID())
			assert.Equal(t, nodes[j].ShardCoordinator.SelfId(), peerInfo.ShardID)
			assert.Equal(t, core.ObserverPeer, peerInfo.PeerType) // observers because the peerAuth is not sent on this network

			peerInfoMain := nodes[i].MainPeerShardMapper.GetPeerInfo(nodes[j].MainMessenger.ID())
			assert.Equal(t, nodes[j].ShardCoordinator.SelfId(), peerInfoMain.ShardID)
			assert.Equal(t, core.ValidatorPeer, peerInfoMain.PeerType)
		}
	}

	for i := 0; i < len(nodes); i++ {
		nodes[i].Close()
	}
}

func connectNodes(nodes []*integrationTests.TestHeartbeatNode, interactingNodes int) {
	for i := 0; i < interactingNodes-1; i++ {
		for j := i + 1; j < interactingNodes; j++ {
			src := nodes[i]
			dst := nodes[j]
			_ = src.ConnectOnMain(dst)
			_ = src.ConnectOnFullArchive(dst)
		}
	}
}

func checkMessages(t *testing.T, nodes []*integrationTests.TestHeartbeatNode, maxMessageAgeAllowed time.Duration) {
	validators := make([][]byte, 0)
	observers := make([][]byte, 0)
	for _, node := range nodes {
		pkBuff, err := node.NodeKeys.MainKey.Pk.ToByteArray()
		require.Nil(t, err)

		if len(node.NodeKeys.HandledKeys) == 0 {
			// single mode, the node's public key should be a validator
			validators = append(validators, pkBuff)
			continue
		}

		// multikey mode, all handled keys are validators
		validators = append(validators, getHandledPublicKeys(t, node.NodeKeys)...)
		observers = append(observers, pkBuff)
	}

	for _, node := range nodes {
		paCache := node.DataPool.PeerAuthentications()
		hbCache := node.DataPool.Heartbeats()

		assert.Equal(t, len(validators), paCache.Len())
		assert.Equal(t, len(observers)+len(validators), hbCache.Len())

		checkPeerAuthenticationMessages(t, paCache, validators, maxMessageAgeAllowed)
		checkHeartbeatMessages(t, hbCache)
	}
}

func getHandledPublicKeys(tb testing.TB, keys *integrationTests.TestNodeKeys) [][]byte {
	publicKeys := make([][]byte, 0, len(keys.HandledKeys))
	for _, key := range keys.HandledKeys {
		pkBuff, err := key.Pk.ToByteArray()
		require.Nil(tb, err)

		publicKeys = append(publicKeys, pkBuff)
	}

	return publicKeys
}

func checkPeerAuthenticationMessages(tb testing.TB, cache storage.Cacher, validators [][]byte, maxMessageAgeAllowed time.Duration) {
	for _, validatorPk := range validators {
		value, found := cache.Get(validatorPk)
		require.True(tb, found)
		msg := value.(*heartbeat.PeerAuthentication)

		marshaller := integrationTests.TestMarshaller
		payload := &heartbeat.Payload{}
		err := marshaller.Unmarshal(payload, msg.Payload)
		assert.Nil(tb, err)

		// check message age
		currentTimestamp := time.Now().Unix()
		messageAge := time.Duration(currentTimestamp - payload.Timestamp)
		assert.True(tb, messageAge < maxMessageAgeAllowed)
	}
}

func checkHeartbeatMessages(tb testing.TB, cache storage.Cacher) {
	allHbKeys := cache.Keys()
	for _, key := range allHbKeys {
		hbData, found := cache.Get(key)
		assert.True(tb, found)

		hbMessage := hbData.(*heartbeat.HeartbeatV2)

		assert.Equal(tb, integrationTests.DefaultIdentity, hbMessage.Identity)
		assert.Contains(tb, hbMessage.NodeDisplayName, integrationTests.DefaultNodeName)
	}
}
