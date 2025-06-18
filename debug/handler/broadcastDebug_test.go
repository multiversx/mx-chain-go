package handler

import (
	"testing"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroadcastDebug_ProcessMultipleMessageTypes(t *testing.T) {
	t.Parallel()

	_ = logger.SetLogLevel("*:DEBUG")

	cfg := config.BroadcastStatisticsConfig{
		Messages: []string{
			"intercepted tx",
			"intercepted miniblock",
			"intercepted header",
			"intercepted heartbeat",
		},
	}

	id := NewBroadcastDebug(cfg)
	require.NotNil(t, id)

	testCases := []struct {
		name        string
		messageType string
		hash        []byte
		originator  string
		fromPeer    string
	}{
		{
			name:        "transaction message",
			messageType: "intercepted tx",
			hash:        []byte("tx_hash_1"),
			originator:  "originator_peer",
			fromPeer:    "connected_peer_1",
		},
		{
			name:        "miniblock message",
			messageType: "intercepted miniblock",
			hash:        []byte("miniblock_hash_1"),
			originator:  "originator_peer",
			fromPeer:    "connected_peer_2",
		},
		{
			name:        "header message",
			messageType: "intercepted header",
			hash:        []byte("header_hash_1"),
			originator:  "originator_peer",
			fromPeer:    "connected_peer_3",
		},
		{
			name:        "heartbeat message",
			messageType: "intercepted heartbeat",
			hash:        []byte("heartbeat_hash_1"),
			originator:  "originator_peer",
			fromPeer:    "connected_peer_4",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMessage := &p2pmocks.P2PMessageMock{
				BroadcastMethodField: p2p.Broadcast,
				FromField:            []byte(tc.originator),
			}
			mockData := &testscommon.InterceptedDataStub{
				TypeCalled: func() string {
					return tc.messageType
				},
				HashCalled: func() []byte {
					return tc.hash
				},
			}

			fromPeerID := core.PeerID(tc.fromPeer)

			id.Process(mockData, mockMessage, fromPeerID)

			hexHash := "74785f686173685f31" // hex encoded "tx_hash_1"
			if tc.messageType == "intercepted miniblock" {
				hexHash = "6d696e69626c6f636b5f686173685f31" // hex encoded "miniblock_hash_1"
			} else if tc.messageType == "intercepted header" {
				hexHash = "6865616465725f686173685f31" // hex encoded "header_hash_1"
			} else if tc.messageType == "intercepted heartbeat" {
				hexHash = "6865617274626561745f686173685f31" // hex encoded "heartbeat_hash_1"
			}

			assert.Contains(t, id.receivedBroadcast, tc.messageType)

			assert.Contains(t, id.receivedBroadcast[tc.messageType], hexHash)

			ev := id.receivedBroadcast[tc.messageType][hexHash]
			assert.Equal(t, fromPeerID.Pretty(), ev.from)
			assert.Equal(t, core.PeerID(tc.originator).Pretty(), ev.originator)
			assert.Equal(t, 1, ev.numReceived)
			assert.Greater(t, ev.firstTimeReceivedMilli, int64(0))
		})
	}

	t.Run("duplicate message processing", func(t *testing.T) {
		mockMessage := &p2pmocks.P2PMessageMock{
			BroadcastMethodField: p2p.Broadcast,
			FromField:            []byte("originator_peer"),
		}

		mockData := &testscommon.InterceptedDataStub{
			TypeCalled: func() string {
				return "intercepted tx"
			},
			HashCalled: func() []byte {
				return []byte("tx_hash_1")
			},
		}

		fromPeerID := core.PeerID("connected_peer_1")

		id.Process(mockData, mockMessage, fromPeerID)

		hexHash := "74785f686173685f31"
		ev := id.receivedBroadcast["intercepted tx"][hexHash]
		assert.Equal(t, 2, ev.numReceived)
	})

	t.Run("message type not in configuration", func(t *testing.T) {
		mockMessage := &p2pmocks.P2PMessageMock{
			BroadcastMethodField: p2p.Broadcast,
			FromField:            []byte("originator_peer"),
		}

		mockData := &testscommon.InterceptedDataStub{
			TypeCalled: func() string {
				return "unknown message type"
			},
			HashCalled: func() []byte {
				return []byte("unknown_hash")
			},
		}

		fromPeerID := core.PeerID("connected_peer_6")
		initialCount := len(id.receivedBroadcast)

		id.Process(mockData, mockMessage, fromPeerID)

		assert.Equal(t, initialCount, len(id.receivedBroadcast))
	})

	t.Run("non-broadcast message", func(t *testing.T) {
		nonBroadcastMessage := &p2pmocks.P2PMessageMock{
			BroadcastMethodField: p2p.Direct,
			FromField:            []byte("originator_peer"),
		}

		mockData := &testscommon.InterceptedDataStub{
			TypeCalled: func() string {
				return "intercepted tx"
			},
			HashCalled: func() []byte {
				return []byte("direct_tx_hash")
			},
		}

		fromPeerID := core.PeerID("connected_peer_7")
		initialCount := len(id.receivedBroadcast["intercepted tx"])

		id.Process(mockData, nonBroadcastMessage, fromPeerID)

		assert.Equal(t, initialCount, len(id.receivedBroadcast["intercepted tx"]))
	})

	id.PrintReceivedTxsBroadcastAndCleanRecords()
}
