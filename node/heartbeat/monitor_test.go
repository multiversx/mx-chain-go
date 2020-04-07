package heartbeat_test

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat/storage"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

var fromConnectedPeerId = p2p.PeerID("from connected peer Id")

func createMockP2PAntifloodHandler() *mock.P2PAntifloodHandlerStub {
	return &mock.P2PAntifloodHandlerStub{
		CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
			return nil
		},
		CanProcessMessageOnTopicCalled: func(peer p2p.PeerID, topic string) error {
			return nil
		},
	}
}

func createMockStorer() heartbeat.HeartbeatStorageHandler {
	return &mock.HeartbeatStorerStub{
		UpdateGenesisTimeCalled: func(genesisTime time.Time) error {
			return nil
		},
		LoadHbmiDTOCalled: func(pubKey string) (*heartbeat.HeartbeatDTO, error) {
			return nil, errors.New("not found")
		},
		LoadKeysCalled: func() ([][]byte, error) {
			return nil, nil
		},
		SavePubkeyDataCalled: func(pubkey []byte, heartbeat *heartbeat.HeartbeatDTO) error {
			return nil
		},
		SaveKeysCalled: func(peersSlice [][]byte) error {
			return nil
		},
	}
}

func createMockArgHeartbeatMonitor() heartbeat.ArgHeartbeatMonitor {
	return heartbeat.ArgHeartbeatMonitor{
		Marshalizer:                 &mock.MarshalizerMock{},
		MaxDurationPeerUnresponsive: 1,
		PubKeysMap:                  map[uint32][]string{0: {""}},
		GenesisTime:                 time.Now(),
		MessageHandler:              &mock.MessageHandlerStub{},
		Storer:                      createMockStorer(),
		PeerTypeProvider: &mock.PeerTypeProviderStub{
			ComputeForPubKeyCalled: func(pubKey []byte) (core.PeerType, uint32, error) {
				if string(pubKey) == "pk0" {
					return "", 0, nil
				}

				return "", 1, nil
			},
		},
		Timer:                mock.NewMockTimer(),
		AntifloodHandler:     createMockP2PAntifloodHandler(),
		HardforkTrigger:      &mock.HardforkTriggerStub{},
		PeerBlackListHandler: &mock.BlackListHandlerStub{},
	}
}

//------- NewMonitor

func TestNewMonitor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatMonitor()
	arg.Marshalizer = nil
	mon, err := heartbeat.NewMonitor(arg)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrNilMarshalizer, err)
}

func TestNewMonitor_EmptyPublicKeyListShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatMonitor()
	arg.PubKeysMap = make(map[uint32][]string)
	mon, err := heartbeat.NewMonitor(arg)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrEmptyPublicKeysMap, err)
}

func TestNewMonitor_NilMessageHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatMonitor()
	arg.MessageHandler = nil
	mon, err := heartbeat.NewMonitor(arg)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrNilMessageHandler, err)
}

func TestNewMonitor_NilHeartbeatStorerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatMonitor()
	arg.Storer = nil
	mon, err := heartbeat.NewMonitor(arg)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrNilHeartbeatStorer, err)
}

func TestNewMonitor_NilPeerTypeProviderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatMonitor()
	arg.PeerTypeProvider = nil
	mon, err := heartbeat.NewMonitor(arg)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrNilPeerTypeProvider, err)
}

func TestNewMonitor_NilTimeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatMonitor()
	arg.Timer = nil
	mon, err := heartbeat.NewMonitor(arg)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrNilTimer, err)
}

func TestNewMonitor_NilAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatMonitor()
	arg.AntifloodHandler = nil
	mon, err := heartbeat.NewMonitor(arg)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrNilAntifloodHandler, err)
}

func TestNewMonitor_NilHardforkTriggerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatMonitor()
	arg.HardforkTrigger = nil
	mon, err := heartbeat.NewMonitor(arg)

	assert.Nil(t, mon)
	assert.Equal(t, heartbeat.ErrNilHardforkTrigger, err)
}

func TestNewMonitor_NilPeerBlackListHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatMonitor()
	arg.PeerBlackListHandler = nil
	mon, err := heartbeat.NewMonitor(arg)

	assert.Nil(t, mon)
	assert.True(t, errors.Is(err, heartbeat.ErrNilBlackListHandler))
}

func TestNewMonitor_OkValsShouldCreatePubkeyMap(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeartbeatMonitor()
	arg.PubKeysMap = map[uint32][]string{0: {"pk1", "pk2"}}
	mon, err := heartbeat.NewMonitor(arg)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(mon))

	hbStatus := mon.GetHeartbeats()
	assert.Equal(t, 2, len(hbStatus))
}

func TestNewMonitor_ShouldComputeShardId(t *testing.T) {
	t.Parallel()

	pksPerShards := map[uint32][]string{
		0: {"pk0"},
		1: {"pk1"},
	}

	arg := createMockArgHeartbeatMonitor()
	arg.MaxDurationPeerUnresponsive = time.Millisecond
	arg.PubKeysMap = pksPerShards
	mon, err := heartbeat.NewMonitor(arg)

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

	arg := createMockArgHeartbeatMonitor()
	arg.Marshalizer = &mock.MarshalizerMock{
		UnmarshalHandler: func(obj interface{}, buff []byte) error {
			(obj.(*heartbeat.Heartbeat)).Pubkey = []byte(pubKey)
			return nil
		},
	}
	arg.MaxDurationPeerUnresponsive = time.Second * 1000
	arg.PubKeysMap = map[uint32][]string{0: {pubKey}}
	arg.MessageHandler = &mock.MessageHandlerStub{
		CreateHeartbeatFromP2PMessageCalled: func(message p2p.MessageP2P) (*heartbeat.Heartbeat, error) {
			var rcvHb heartbeat.Heartbeat
			_ = json.Unmarshal(message.Data(), &rcvHb)
			return &rcvHb, nil
		},
	}
	mon, _ := heartbeat.NewMonitor(arg)

	hb := heartbeat.Heartbeat{
		Pubkey: []byte(pubKey),
	}
	hbBytes, _ := json.Marshal(hb)
	err := mon.ProcessReceivedMessage(&mock.P2PMessageStub{DataField: hbBytes}, fromConnectedPeerId)
	assert.Nil(t, err)

	//a delay is mandatory for the go routine to finish its job
	time.Sleep(time.Second)

	hbStatus := mon.GetHeartbeats()
	assert.Equal(t, 1, len(hbStatus))
	assert.Equal(t, hex.EncodeToString([]byte(pubKey)), hbStatus[0].HexPublicKey)
}

func TestMonitor_ProcessReceivedMessageProcessTriggerErrorShouldErr(t *testing.T) {
	t.Parallel()

	pubKey := "pk1"
	triggerWasCalled := false
	expectedErr := errors.New("expected error")

	arg := createMockArgHeartbeatMonitor()
	arg.MaxDurationPeerUnresponsive = time.Second * 1000
	arg.PubKeysMap = map[uint32][]string{0: {pubKey}}
	arg.MessageHandler = &mock.MessageHandlerStub{
		CreateHeartbeatFromP2PMessageCalled: func(message p2p.MessageP2P) (*heartbeat.Heartbeat, error) {
			var rcvHb heartbeat.Heartbeat
			_ = json.Unmarshal(message.Data(), &rcvHb)
			return &rcvHb, nil
		},
	}
	arg.HardforkTrigger = &mock.HardforkTriggerStub{
		TriggerReceivedCalled: func(payload []byte, data []byte, pkBytes []byte) (bool, error) {
			triggerWasCalled = true

			return true, expectedErr
		},
	}
	mon, _ := heartbeat.NewMonitor(arg)

	hb := heartbeat.Heartbeat{
		Pubkey: []byte(pubKey),
	}
	hbBytes, _ := json.Marshal(hb)
	err := mon.ProcessReceivedMessage(&mock.P2PMessageStub{DataField: hbBytes}, fromConnectedPeerId)

	//a delay is mandatory for the go routine to finish its job
	time.Sleep(time.Second)

	assert.Equal(t, expectedErr, err)
	assert.True(t, triggerWasCalled)
}

func TestMonitor_ProcessReceivedMessageWithNewPublicKey(t *testing.T) {
	t.Parallel()

	pubKey := "pk1"

	arg := createMockArgHeartbeatMonitor()
	arg.Marshalizer = &mock.MarshalizerMock{
		UnmarshalHandler: func(obj interface{}, buff []byte) error {
			(obj.(*heartbeat.Heartbeat)).Pubkey = []byte(pubKey)
			return nil
		},
	}
	arg.MaxDurationPeerUnresponsive = time.Second * 1000
	arg.PubKeysMap = map[uint32][]string{0: {"pk2"}}
	arg.MessageHandler = &mock.MessageHandlerStub{
		CreateHeartbeatFromP2PMessageCalled: func(message p2p.MessageP2P) (*heartbeat.Heartbeat, error) {
			var rcvHb heartbeat.Heartbeat
			_ = json.Unmarshal(message.Data(), &rcvHb)
			return &rcvHb, nil
		},
	}
	mon, _ := heartbeat.NewMonitor(arg)

	hb := heartbeat.Heartbeat{
		Pubkey: []byte(pubKey),
	}
	hbBytes, _ := json.Marshal(hb)
	err := mon.ProcessReceivedMessage(&mock.P2PMessageStub{DataField: hbBytes}, fromConnectedPeerId)
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

	arg := createMockArgHeartbeatMonitor()
	arg.Marshalizer = &mock.MarshalizerMock{
		UnmarshalHandler: func(obj interface{}, buff []byte) error {
			var rcvdHb heartbeat.Heartbeat
			_ = json.Unmarshal(buff, &rcvdHb)
			(obj.(*heartbeat.Heartbeat)).Pubkey = rcvdHb.Pubkey
			(obj.(*heartbeat.Heartbeat)).ShardID = rcvdHb.ShardID
			return nil
		},
	}
	arg.MaxDurationPeerUnresponsive = time.Second * 1000
	arg.PubKeysMap = map[uint32][]string{0: {"pk1"}}
	arg.MessageHandler = &mock.MessageHandlerStub{
		CreateHeartbeatFromP2PMessageCalled: func(message p2p.MessageP2P) (*heartbeat.Heartbeat, error) {
			var rcvHb heartbeat.Heartbeat
			_ = json.Unmarshal(message.Data(), &rcvHb)
			return &rcvHb, nil
		},
	}

	mon, _ := heartbeat.NewMonitor(arg)

	// First send from pk1 from shard 0
	hb := &heartbeat.Heartbeat{
		Pubkey:  pubKey,
		ShardID: uint32(0),
	}

	buffToSend, err := json.Marshal(hb)
	assert.Nil(t, err)

	err = mon.ProcessReceivedMessage(&mock.P2PMessageStub{DataField: buffToSend}, fromConnectedPeerId)
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

	err = mon.ProcessReceivedMessage(&mock.P2PMessageStub{DataField: buffToSend}, fromConnectedPeerId)
	assert.Nil(t, err)

	time.Sleep(1 * time.Second)

	hbStatus = mon.GetHeartbeats()

	// check if shard ID is changed at the same status
	assert.Equal(t, uint32(1), hbStatus[0].ReceivedShardID)
	assert.Equal(t, 1, len(hbStatus))
}

func TestMonitor_ProcessReceivedMessageShouldSetPeerInactive(t *testing.T) {
	t.Parallel()

	th := mock.NewMockTimer()
	pubKey1 := "pk1-should-stay-online"
	pubKey2 := "pk2-should-go-offline"
	storer, _ := storage.NewHeartbeatDbStorer(mock.NewStorerMock(), &mock.MarshalizerFake{})
	arg := createMockArgHeartbeatMonitor()
	arg.Marshalizer = &mock.MarshalizerMock{
		UnmarshalHandler: func(obj interface{}, buff []byte) error {
			var rcvdHb heartbeat.Heartbeat
			_ = json.Unmarshal(buff, &rcvdHb)
			(obj.(*heartbeat.Heartbeat)).Pubkey = rcvdHb.Pubkey
			(obj.(*heartbeat.Heartbeat)).ShardID = rcvdHb.ShardID
			return nil
		},
	}
	arg.MessageHandler = &mock.MessageHandlerStub{
		CreateHeartbeatFromP2PMessageCalled: func(message p2p.MessageP2P) (*heartbeat.Heartbeat, error) {
			var rcvHb heartbeat.Heartbeat
			_ = json.Unmarshal(message.Data(), &rcvHb)
			return &rcvHb, nil
		},
	}
	arg.MaxDurationPeerUnresponsive = time.Second * 5
	arg.PubKeysMap = map[uint32][]string{0: {pubKey1, pubKey2}}
	arg.Storer = storer
	arg.Timer = th
	mon, _ := heartbeat.NewMonitor(arg)

	// First send from pk1
	err := sendHbMessageFromPubKey(pubKey1, mon)
	assert.Nil(t, err)

	// Send from pk2
	err = sendHbMessageFromPubKey(pubKey2, mon)
	assert.Nil(t, err)

	// set pk2 to inactive as max inactive time is lower
	time.Sleep(10 * time.Millisecond)
	th.IncrementSeconds(6)

	// Check that both are added
	hbStatus := mon.GetHeartbeats()
	assert.Equal(t, 2, len(hbStatus))
	//assert.False(t, hbStatus[1].IsActive)

	// Now send a message from pk1 in order to see that pk2 is not active anymore
	err = sendHbMessageFromPubKey(pubKey1, mon)
	time.Sleep(5 * time.Millisecond)
	assert.Nil(t, err)

	th.IncrementSeconds(4)

	hbStatus = mon.GetHeartbeats()

	// check if pk1 is still on
	assert.True(t, hbStatus[0].IsActive)
	// check if pk2 was set to offline by pk1
	assert.False(t, hbStatus[1].IsActive)
}

func TestMonitor_ProcessReceivedMessageImpersonatedMessageShouldErr(t *testing.T) {
	t.Parallel()

	pubKey := "pk1"
	originator := p2p.PeerID("message originator")

	arg := createMockArgHeartbeatMonitor()
	arg.Marshalizer = &mock.MarshalizerMock{
		UnmarshalHandler: func(obj interface{}, buff []byte) error {
			(obj.(*heartbeat.Heartbeat)).Pubkey = []byte(pubKey)
			return nil
		},
	}
	arg.MaxDurationPeerUnresponsive = time.Second * 1000
	arg.PubKeysMap = map[uint32][]string{0: {"pk2"}}
	arg.MessageHandler = &mock.MessageHandlerStub{
		CreateHeartbeatFromP2PMessageCalled: func(message p2p.MessageP2P) (*heartbeat.Heartbeat, error) {
			var rcvHb heartbeat.Heartbeat
			_ = json.Unmarshal(message.Data(), &rcvHb)
			return &rcvHb, nil
		},
	}
	originatorWasBlacklisted := false
	connectedPeerWasBlacklisted := false
	arg.PeerBlackListHandler = &mock.BlackListHandlerStub{
		AddCalled: func(key string) error {
			if key == originator.Pretty() {
				originatorWasBlacklisted = true
			}
			if key == fromConnectedPeerId.Pretty() {
				connectedPeerWasBlacklisted = true
			}

			return nil
		},
	}
	mon, _ := heartbeat.NewMonitor(arg)

	hb := heartbeat.Heartbeat{
		Pubkey: []byte(pubKey),
	}
	hbBytes, _ := json.Marshal(hb)
	message := &mock.P2PMessageStub{
		DataField: hbBytes,
		PeerField: originator,
	}

	err := mon.ProcessReceivedMessage(message, fromConnectedPeerId)
	assert.True(t, errors.Is(err, heartbeat.ErrHeartbeatPidMismatch))
	assert.True(t, originatorWasBlacklisted)
	assert.True(t, connectedPeerWasBlacklisted)
}

func sendHbMessageFromPubKey(pubKey string, mon *heartbeat.Monitor) error {
	hb := &heartbeat.Heartbeat{
		Pubkey: []byte(pubKey),
	}
	buffToSend, _ := json.Marshal(hb)
	err := mon.ProcessReceivedMessage(&mock.P2PMessageStub{DataField: buffToSend}, fromConnectedPeerId)
	return err
}
