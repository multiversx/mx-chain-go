package heartbeat_test

import (
	"encoding/hex"
	"encoding/json"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewMonitor

func TestNewMonitor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	mon, err := heartbeat.NewMonitor(
		nil,
		0,
		map[uint32][]string{0: {""}},
		&mock.StorerMock{},
		time.Now(),
		&mock.MessageHandlerStub{},
	)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrNilMarshalizer, err)
}

func TestNewMonitor_EmptyPublicKeyListShouldErr(t *testing.T) {
	t.Parallel()

	mon, err := heartbeat.NewMonitor(
		&mock.MarshalizerMock{},
		0,
		make(map[uint32][]string),
		&mock.StorerMock{},
		time.Now(),
		&mock.MessageHandlerStub{},
	)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrEmptyPublicKeysMap, err)
}

func TestNewMonitor_NilMessageHandlerShouldErr(t *testing.T) {
	t.Parallel()

	mon, err := heartbeat.NewMonitor(
		&mock.MarshalizerMock{},
		0,
		map[uint32][]string{0: {""}},
		&mock.StorerMock{},
		time.Now(),
		nil,
	)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrNilMessageHandler, err)
}

func TestNewMonitor_OkValsShouldCreatePubkeyMap(t *testing.T) {
	t.Parallel()

	mon, err := heartbeat.NewMonitor(
		&mock.MarshalizerMock{},
		1,
		map[uint32][]string{0: {"pk1", "pk2"}},
		mock.NewStorerMock(),
		time.Now(),
		&mock.MessageHandlerStub{},
	)

	assert.NotNil(t, mon)
	assert.Nil(t, err)
	hbStatus := mon.GetHeartbeats()
	assert.Equal(t, 2, len(hbStatus))
}

func TestNewMonitor_ShouldComputeShardId(t *testing.T) {
	t.Parallel()

	pksPerShards := map[uint32][]string{
		0: {"pk0"},
		1: {"pk1"},
	}

	maxDuration := time.Millisecond
	mon, err := heartbeat.NewMonitor(
		&mock.MarshalizerMock{},
		maxDuration,
		pksPerShards,
		mock.NewStorerMock(),
		time.Now(),
		&mock.MessageHandlerStub{},
	)

	assert.NotNil(t, mon)
	assert.Nil(t, err)
	hbStatus := mon.GetHeartbeats()

	assert.Equal(t, uint32(0), hbStatus[0].ComputedShardID)
	assert.Equal(t, uint32(1), hbStatus[1].ComputedShardID)
}

//------- ProcessReceivedMessage

func TestMonitor_ProcessReceivedMessageShouldWork(t *testing.T) {
	t.Parallel()

	pubKey := "pk1"

	mon, _ := heartbeat.NewMonitor(
		&mock.MarshalizerMock{
			UnmarshalHandler: func(obj interface{}, buff []byte) error {
				(obj.(*heartbeat.Heartbeat)).Pubkey = []byte(pubKey)
				return nil
			},
		},
		time.Second*1000,
		map[uint32][]string{0: {pubKey}},
		mock.NewStorerMock(),
		time.Now(),
		&mock.MessageHandlerStub{
			CreateHeartbeatFromP2pMessageCalled: func(message p2p.MessageP2P) (*heartbeat.Heartbeat, error) {
				var rcvHb heartbeat.Heartbeat
				_ = json.Unmarshal(message.Data(), &rcvHb)
				return &rcvHb, nil
			},
		},
	)

	hb := heartbeat.Heartbeat{
		Pubkey: []byte(pubKey),
	}
	hbBytes, _ := json.Marshal(hb)
	err := mon.ProcessReceivedMessage(&mock.P2PMessageStub{DataField: hbBytes})
	assert.Nil(t, err)

	//a delay is mandatory for the go routine to finish its job
	time.Sleep(time.Second)

	hbStatus := mon.GetHeartbeats()
	assert.Equal(t, 1, len(hbStatus))
	assert.Equal(t, hex.EncodeToString([]byte(pubKey)), hbStatus[0].HexPublicKey)
}

func TestMonitor_ProcessReceivedMessageWithNewPublicKey(t *testing.T) {
	t.Parallel()

	pubKey := "pk1"

	mon, _ := heartbeat.NewMonitor(
		&mock.MarshalizerMock{
			UnmarshalHandler: func(obj interface{}, buff []byte) error {
				(obj.(*heartbeat.Heartbeat)).Pubkey = []byte(pubKey)
				return nil
			},
		},
		time.Second*1000,
		map[uint32][]string{0: {"pk2"}},
		mock.NewStorerMock(),
		time.Now(),
		&mock.MessageHandlerStub{
			CreateHeartbeatFromP2pMessageCalled: func(message p2p.MessageP2P) (*heartbeat.Heartbeat, error) {
				var rcvHb heartbeat.Heartbeat
				_ = json.Unmarshal(message.Data(), &rcvHb)
				return &rcvHb, nil
			},
		},
	)

	hb := heartbeat.Heartbeat{
		Pubkey: []byte(pubKey),
	}
	hbBytes, _ := json.Marshal(hb)
	err := mon.ProcessReceivedMessage(&mock.P2PMessageStub{DataField: hbBytes})
	assert.Nil(t, err)

	//a delay is mandatory for the go routine to finish its job
	time.Sleep(time.Second)

	//there should be 2 heartbeats, because a new one should have been added with pk2
	hbStatus := mon.GetHeartbeats()
	assert.Equal(t, 2, len(hbStatus))
	assert.Equal(t, hex.EncodeToString([]byte(pubKey)), hbStatus[0].HexPublicKey)
}

func TestMonitor_ProcessReceivedMessageWithNewShardID(t *testing.T) {
	t.Parallel()

	pubKey := []byte("pk1")

	mon, _ := heartbeat.NewMonitor(
		&mock.MarshalizerMock{
			UnmarshalHandler: func(obj interface{}, buff []byte) error {
				var rcvdHb heartbeat.Heartbeat
				_ = json.Unmarshal(buff, &rcvdHb)
				(obj.(*heartbeat.Heartbeat)).Pubkey = rcvdHb.Pubkey
				(obj.(*heartbeat.Heartbeat)).ShardID = rcvdHb.ShardID
				return nil
			},
		},
		time.Second*1000,
		map[uint32][]string{0: {"pk1"}},
		mock.NewStorerMock(),
		time.Now(),
		&mock.MessageHandlerStub{
			CreateHeartbeatFromP2pMessageCalled: func(message p2p.MessageP2P) (*heartbeat.Heartbeat, error) {
				var rcvHb heartbeat.Heartbeat
				_ = json.Unmarshal(message.Data(), &rcvHb)
				return &rcvHb, nil
			},
		},
	)

	// First send from pk1 from shard 0
	hb := &heartbeat.Heartbeat{
		Pubkey:  pubKey,
		ShardID: uint32(0),
	}

	buffToSend, err := json.Marshal(hb)
	assert.Nil(t, err)

	err = mon.ProcessReceivedMessage(&mock.P2PMessageStub{DataField: buffToSend})
	assert.Nil(t, err)

	//a delay is mandatory for the go routine to finish its job
	time.Sleep(time.Second)

	hbStatus := mon.GetHeartbeats()

	assert.Equal(t, uint32(0), hbStatus[0].ReceivedShardID)

	// now we send a new heartbeat which will contain a new shard id
	hb = &heartbeat.Heartbeat{
		Pubkey:  pubKey,
		ShardID: uint32(1),
	}

	buffToSend, err = json.Marshal(hb)

	assert.Nil(t, err)

	err = mon.ProcessReceivedMessage(&mock.P2PMessageStub{DataField: buffToSend})

	time.Sleep(1 * time.Second)

	hbStatus = mon.GetHeartbeats()

	// check if shard ID is changed at the same status
	assert.Equal(t, uint32(1), hbStatus[0].ReceivedShardID)
	assert.Equal(t, 1, len(hbStatus))
}

func TestMonitor_ProcessReceivedMessageShouldSetPeerInactive(t *testing.T) {
	t.Parallel()

	pubKey1 := "pk1-should-stay-online"
	pubKey2 := "pk2-should-go-offline"

	mon, _ := heartbeat.NewMonitor(
		&mock.MarshalizerMock{
			UnmarshalHandler: func(obj interface{}, buff []byte) error {
				var rcvdHb heartbeat.Heartbeat
				_ = json.Unmarshal(buff, &rcvdHb)
				(obj.(*heartbeat.Heartbeat)).Pubkey = rcvdHb.Pubkey
				(obj.(*heartbeat.Heartbeat)).ShardID = rcvdHb.ShardID
				return nil
			},
		},
		time.Millisecond*5,
		map[uint32][]string{0: {pubKey1, pubKey2}},
		mock.NewStorerMock(),
		time.Now(),
		&mock.MessageHandlerStub{
			CreateHeartbeatFromP2pMessageCalled: func(message p2p.MessageP2P) (*heartbeat.Heartbeat, error) {
				var rcvHb heartbeat.Heartbeat
				_ = json.Unmarshal(message.Data(), &rcvHb)
				return &rcvHb, nil
			},
		},
	)

	// First send from pk1
	err := sendHbMessageFromPubKey(pubKey1, mon)
	assert.Nil(t, err)

	// Send from pk2
	err = sendHbMessageFromPubKey(pubKey2, mon)
	assert.Nil(t, err)

	// set pk2 to inactive as max inactive time is lower
	time.Sleep(6 * time.Millisecond)

	// Check that both are added
	hbStatus := mon.GetHeartbeats()
	assert.Equal(t, 2, len(hbStatus))

	// Now send a message from pk1 in order to see that pk2 is not active anymore
	err = sendHbMessageFromPubKey(pubKey1, mon)
	assert.Nil(t, err)

	time.Sleep(5 * time.Millisecond)

	hbStatus = mon.GetHeartbeats()

	// check if pk1 is still on
	assert.True(t, hbStatus[0].IsActive)
	// check if pk2 was set to offline by pk1
	assert.False(t, hbStatus[1].IsActive)
}

func sendHbMessageFromPubKey(pubKey string, mon *heartbeat.Monitor) error {
	hb := &heartbeat.Heartbeat{
		Pubkey: []byte(pubKey),
	}
	buffToSend, _ := json.Marshal(hb)
	err := mon.ProcessReceivedMessage(&mock.P2PMessageStub{DataField: buffToSend})
	return err
}
