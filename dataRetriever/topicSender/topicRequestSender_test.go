package topicsender_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	topicsender "github.com/multiversx/mx-chain-go/dataRetriever/topicSender"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgBaseTopicSender() topicsender.ArgBaseTopicSender {
	return topicsender.ArgBaseTopicSender{
		MainMessenger:        &p2pmocks.MessengerStub{},
		FullArchiveMessenger: &p2pmocks.MessengerStub{},
		TopicName:            "topic",
		OutputAntiflooder:    &mock.P2PAntifloodHandlerStub{},
		MainPreferredPeersHolder: &p2pmocks.PeersHolderStub{
			GetCalled: func() map[uint32][]core.PeerID {
				return map[uint32][]core.PeerID{}
			},
		},
		FullArchivePreferredPeersHolder: &p2pmocks.PeersHolderStub{
			GetCalled: func() map[uint32][]core.PeerID {
				return map[uint32][]core.PeerID{}
			},
		},
		TargetShardId: 0,
	}
}

func createMockArgTopicRequestSender() topicsender.ArgTopicRequestSender {
	return topicsender.ArgTopicRequestSender{
		ArgBaseTopicSender:          createMockArgBaseTopicSender(),
		Marshaller:                  &mock.MarshalizerMock{},
		Randomizer:                  &mock.IntRandomizerStub{},
		PeerListCreator:             &mock.PeerListCreatorStub{},
		NumIntraShardPeers:          2,
		NumCrossShardPeers:          2,
		NumFullHistoryPeers:         2,
		CurrentNetworkEpochProvider: &mock.CurrentNetworkEpochProviderStub{},
		SelfShardIdProvider:         mock.NewMultipleShardsCoordinatorMock(),
		PeersRatingHandler:          &p2pmocks.PeersRatingHandlerStub{},
	}
}

func TestNewTopicRequestSender(t *testing.T) {
	t.Parallel()

	t.Run("nil MainMessenger should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.MainMessenger = nil
		trs, err := topicsender.NewTopicRequestSender(arg)
		assert.True(t, check.IfNil(trs))
		assert.True(t, errors.Is(err, dataRetriever.ErrNilMessenger))
	})
	t.Run("nil FullArchiveMessenger should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.FullArchiveMessenger = nil
		trs, err := topicsender.NewTopicRequestSender(arg)
		assert.True(t, check.IfNil(trs))
		assert.True(t, errors.Is(err, dataRetriever.ErrNilMessenger))
	})
	t.Run("nil OutputAntiflooder should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.OutputAntiflooder = nil
		trs, err := topicsender.NewTopicRequestSender(arg)
		assert.True(t, check.IfNil(trs))
		assert.Equal(t, dataRetriever.ErrNilAntifloodHandler, err)
	})
	t.Run("nil MainPreferredPeersHolder should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.MainPreferredPeersHolder = nil
		trs, err := topicsender.NewTopicRequestSender(arg)
		assert.True(t, check.IfNil(trs))
		assert.True(t, errors.Is(err, dataRetriever.ErrNilPreferredPeersHolder))
	})
	t.Run("nil FullArchivePreferredPeersHolder should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.FullArchivePreferredPeersHolder = nil
		trs, err := topicsender.NewTopicRequestSender(arg)
		assert.True(t, check.IfNil(trs))
		assert.True(t, errors.Is(err, dataRetriever.ErrNilPreferredPeersHolder))
	})
	t.Run("nil Marshaller should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.Marshaller = nil
		trs, err := topicsender.NewTopicRequestSender(arg)
		assert.True(t, check.IfNil(trs))
		assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
	})
	t.Run("nil Randomizer should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.Randomizer = nil
		trs, err := topicsender.NewTopicRequestSender(arg)
		assert.True(t, check.IfNil(trs))
		assert.Equal(t, dataRetriever.ErrNilRandomizer, err)
	})
	t.Run("nil PeerListCreator should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.PeerListCreator = nil
		trs, err := topicsender.NewTopicRequestSender(arg)
		assert.True(t, check.IfNil(trs))
		assert.Equal(t, dataRetriever.ErrNilPeerListCreator, err)
	})
	t.Run("nil CurrentNetworkEpochProvider should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.CurrentNetworkEpochProvider = nil
		trs, err := topicsender.NewTopicRequestSender(arg)
		assert.True(t, check.IfNil(trs))
		assert.Equal(t, dataRetriever.ErrNilCurrentNetworkEpochProvider, err)
	})
	t.Run("nil PeersRatingHandler should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.PeersRatingHandler = nil
		trs, err := topicsender.NewTopicRequestSender(arg)
		assert.True(t, check.IfNil(trs))
		assert.Equal(t, dataRetriever.ErrNilPeersRatingHandler, err)
	})
	t.Run("nil SelfShardIdProvider should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.SelfShardIdProvider = nil
		trs, err := topicsender.NewTopicRequestSender(arg)
		assert.True(t, check.IfNil(trs))
		assert.Equal(t, dataRetriever.ErrNilSelfShardIDProvider, err)
	})
	t.Run("invalid NumIntraShardPeers should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.NumIntraShardPeers = -1
		trs, err := topicsender.NewTopicRequestSender(arg)
		assert.True(t, check.IfNil(trs))
		assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "NumIntraShardPeers"))
	})
	t.Run("invalid NumCrossShardPeers should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.NumCrossShardPeers = -1
		trs, err := topicsender.NewTopicRequestSender(arg)
		assert.True(t, check.IfNil(trs))
		assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "NumCrossShardPeers"))
	})
	t.Run("invalid NumFullHistoryPeers should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.NumFullHistoryPeers = -1
		trs, err := topicsender.NewTopicRequestSender(arg)
		assert.True(t, check.IfNil(trs))
		assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "NumFullHistoryPeers"))
	})
	t.Run("invalid total number of peers should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.NumCrossShardPeers = 0
		arg.NumIntraShardPeers = 0
		trs, err := topicsender.NewTopicRequestSender(arg)
		assert.True(t, check.IfNil(trs))
		assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "NumIntraShardPeers"))
		assert.True(t, strings.Contains(err.Error(), "NumCrossShardPeers"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		trs, err := topicsender.NewTopicRequestSender(createMockArgTopicRequestSender())
		assert.False(t, check.IfNil(trs))
		assert.Nil(t, err)
	})
}

func TestTopicResolverSender_SendOnRequestTopic(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	var defaultHashes = [][]byte{[]byte("hash")}

	t.Run("marshal fails", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.Marshaller = &mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) (bytes []byte, e error) {
				return nil, expectedErr
			},
		}
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)

		assert.Equal(t, expectedErr, err)
	})
	t.Run("no peers should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgTopicRequestSender()
		arg.PeerListCreator = &mock.PeerListCreatorStub{
			CrossShardPeerListCalled: func() []core.PeerID {
				return make([]core.PeerID, 0)
			},
			IntraShardPeerListCalled: func() []core.PeerID {
				return make([]core.PeerID, 0)
			},
		}
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)

		assert.True(t, errors.Is(err, dataRetriever.ErrSendRequest))
	})
	t.Run("should work and not send to full history", func(t *testing.T) {
		t.Parallel()

		pID1 := core.PeerID("peer1")
		pID2 := core.PeerID("peer2")
		sentToPid1 := false
		sentToPid2 := false

		arg := createMockArgTopicRequestSender()
		arg.MainMessenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				if bytes.Equal(peerID.Bytes(), pID1.Bytes()) {
					sentToPid1 = true
				}
				if bytes.Equal(peerID.Bytes(), pID2.Bytes()) {
					sentToPid2 = true
				}

				return nil
			},
		}
		arg.FullArchiveMessenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				assert.Fail(t, "should have not been called")
				return nil
			},
		}
		arg.PeerListCreator = &mock.PeerListCreatorStub{
			CrossShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{pID1}
			},
			IntraShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{pID2}
			},
		}
		decreaseCalledCounter := 0
		arg.PeersRatingHandler = &p2pmocks.PeersRatingHandlerStub{
			DecreaseRatingCalled: func(pid core.PeerID) {
				decreaseCalledCounter++
				if !bytes.Equal(pid.Bytes(), pID1.Bytes()) && !bytes.Equal(pid.Bytes(), pID2.Bytes()) {
					assert.Fail(t, "should be one of the provided pids")
				}
			},
		}
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)

		assert.Nil(t, err)
		assert.True(t, sentToPid1)
		assert.True(t, sentToPid2)
		assert.Equal(t, 2, decreaseCalledCounter)
	})
	t.Run("should work and send to full history", func(t *testing.T) {
		t.Parallel()

		pIDfullHistory := core.PeerID("full history peer")
		sentToFullHistoryPeer := false

		arg := createMockArgTopicRequestSender()
		arg.MainMessenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				assert.Fail(t, "should have not been called")

				return nil
			},
		}
		arg.FullArchiveMessenger = &p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{pIDfullHistory}
			},
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				if bytes.Equal(peerID.Bytes(), pIDfullHistory.Bytes()) {
					sentToFullHistoryPeer = true
				}

				return nil
			},
		}
		arg.CurrentNetworkEpochProvider = &mock.CurrentNetworkEpochProviderStub{
			EpochIsActiveInNetworkCalled: func(epoch uint32) bool {
				return false
			},
		}
		decreaseCalledCounter := 0
		arg.PeersRatingHandler = &p2pmocks.PeersRatingHandlerStub{
			DecreaseRatingCalled: func(pid core.PeerID) {
				decreaseCalledCounter++
				assert.Equal(t, pIDfullHistory, pid)
			},
		}
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)
		assert.Nil(t, err)
		assert.True(t, sentToFullHistoryPeer)
		assert.Equal(t, 1, decreaseCalledCounter)
	})
	t.Run("should work and send to preferred regular peers", func(t *testing.T) {
		t.Parallel()

		selfShardID := uint32(0)
		targetShardID := uint32(1)
		countPrefPeersSh0 := 0
		preferredPeersShard0 := make([]core.PeerID, 0)
		for idx := 0; idx < 5; idx++ {
			preferredPeersShard0 = append(preferredPeersShard0, core.PeerID(fmt.Sprintf("prefPIDsh0-%d", idx)))
		}

		countPrefPeersSh1 := 0
		preferredPeersShard1 := make([]core.PeerID, 0)
		for idx := 0; idx < 5; idx++ {
			preferredPeersShard1 = append(preferredPeersShard1, core.PeerID(fmt.Sprintf("prefPIDsh1-%d", idx)))
		}
		regularPeer0, regularPeer1 := core.PeerID("peer0"), core.PeerID("peer1")

		arg := createMockArgTopicRequestSender()
		arg.TargetShardId = targetShardID

		selfShardIDProvider := mock.NewMultipleShardsCoordinatorMock()
		selfShardIDProvider.CurrentShard = selfShardID
		arg.SelfShardIdProvider = selfShardIDProvider

		arg.PeerListCreator = &mock.PeerListCreatorStub{
			CrossShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{regularPeer0}
			},
			IntraShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{regularPeer1}
			},
		}
		arg.MainPreferredPeersHolder = &p2pmocks.PeersHolderStub{
			GetCalled: func() map[uint32][]core.PeerID {
				return map[uint32][]core.PeerID{
					selfShardID:   preferredPeersShard0,
					targetShardID: preferredPeersShard1,
				}
			},
		}
		arg.NumCrossShardPeers = 5
		arg.NumIntraShardPeers = 5
		arg.MainMessenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				if strings.HasPrefix(string(peerID), "prefPIDsh0") {
					countPrefPeersSh0++
				}

				if strings.HasPrefix(string(peerID), "prefPIDsh1") {
					countPrefPeersSh1++
				}

				return nil
			},
		}
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)
		assert.Nil(t, err)
		assert.Equal(t, 1, countPrefPeersSh1)
	})
	t.Run("should work and send to preferred regular cross peer first", func(t *testing.T) {
		t.Parallel()

		targetShardID := uint32(37)
		pidPreferred := core.PeerID("preferred peer")
		numTimesSent := 0
		regularPeer0, regularPeer1 := core.PeerID("peer0"), core.PeerID("peer1")
		sentToPreferredPeer := false

		arg := createMockArgTopicRequestSender()
		arg.TargetShardId = targetShardID
		arg.NumCrossShardPeers = 5
		arg.PeerListCreator = &mock.PeerListCreatorStub{
			CrossShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{regularPeer0, regularPeer1, regularPeer0, regularPeer1}
			},
			IntraShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{}
			},
		}
		arg.MainPreferredPeersHolder = &p2pmocks.PeersHolderStub{
			GetCalled: func() map[uint32][]core.PeerID {
				return map[uint32][]core.PeerID{
					targetShardID: {pidPreferred},
				}
			},
		}

		arg.MainMessenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				if bytes.Equal(peerID.Bytes(), pidPreferred.Bytes()) {
					sentToPreferredPeer = true
					require.Zero(t, numTimesSent)
				}

				numTimesSent++
				return nil
			},
		}
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)
		assert.Nil(t, err)
		assert.True(t, sentToPreferredPeer)
	})
	t.Run("should work and send to preferred regular intra peer first", func(t *testing.T) {
		t.Parallel()

		selfShardID := uint32(37)
		pidPreferred := core.PeerID("preferred peer")
		numTimesSent := 0
		regularPeer0, regularPeer1 := core.PeerID("peer0"), core.PeerID("peer1")
		sentToPreferredPeer := false

		arg := createMockArgTopicRequestSender()
		arg.TargetShardId = 0
		arg.NumCrossShardPeers = 5
		arg.PeerListCreator = &mock.PeerListCreatorStub{
			CrossShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{}
			},
			IntraShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{regularPeer0, regularPeer1, regularPeer0, regularPeer1}
			},
		}
		arg.MainPreferredPeersHolder = &p2pmocks.PeersHolderStub{
			GetCalled: func() map[uint32][]core.PeerID {
				return map[uint32][]core.PeerID{
					selfShardID: {pidPreferred},
				}
			},
		}

		arg.MainMessenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				if bytes.Equal(peerID.Bytes(), pidPreferred.Bytes()) {
					sentToPreferredPeer = true
					require.Zero(t, numTimesSent)
				}

				numTimesSent++
				return nil
			},
		}

		selfShardIDProvider := mock.NewMultipleShardsCoordinatorMock()
		selfShardIDProvider.CurrentShard = selfShardID
		arg.SelfShardIdProvider = selfShardIDProvider

		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)
		assert.Nil(t, err)
		assert.True(t, sentToPreferredPeer)
	})
	t.Run("should work and send to preferred full archive first", func(t *testing.T) {
		t.Parallel()

		selfShardID := uint32(37)
		pidPreferred := core.PeerID("preferred peer")
		sentToPreferredPeer := false
		regularPeer0, regularPeer1 := core.PeerID("peer0"), core.PeerID("peer1")

		arg := createMockArgTopicRequestSender()
		arg.NumFullHistoryPeers = 2
		arg.CurrentNetworkEpochProvider = &mock.CurrentNetworkEpochProviderStub{
			EpochIsActiveInNetworkCalled: func(epoch uint32) bool {
				return false
			},
		}
		arg.FullArchivePreferredPeersHolder = &p2pmocks.PeersHolderStub{
			GetCalled: func() map[uint32][]core.PeerID {
				return map[uint32][]core.PeerID{
					selfShardID: {pidPreferred},
				}
			},
		}
		arg.FullArchiveMessenger = &p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{regularPeer0, regularPeer1}
			},
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				if bytes.Equal(peerID.Bytes(), pidPreferred.Bytes()) {
					sentToPreferredPeer = true
				}

				return nil
			},
		}
		arg.MainMessenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				assert.Fail(t, "should not have been called")

				return nil
			},
		}

		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)
		assert.Nil(t, err)
		assert.True(t, sentToPreferredPeer)
	})
	t.Run("should work and skip antiflood checks for preferred peers", func(t *testing.T) {
		t.Parallel()

		selfShardID := uint32(37)
		pidPreferred := core.PeerID("preferred peer")
		regularPeer0, regularPeer1 := core.PeerID("peer0"), core.PeerID("peer1")
		targetShardID := uint32(55)

		sentToPreferredPeer := false

		arg := createMockArgTopicRequestSender()
		arg.TargetShardId = targetShardID
		arg.NumCrossShardPeers = 5
		arg.PeerListCreator = &mock.PeerListCreatorStub{
			CrossShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{regularPeer0, regularPeer1, regularPeer0, regularPeer1}
			},
			IntraShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{}
			},
		}
		arg.MainPreferredPeersHolder = &p2pmocks.PeersHolderStub{
			GetCalled: func() map[uint32][]core.PeerID {
				return map[uint32][]core.PeerID{
					targetShardID: {pidPreferred},
				}
			},
			ContainsCalled: func(peerID core.PeerID) bool {
				return peerID == pidPreferred
			},
		}

		arg.MainMessenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				if peerID == pidPreferred {
					sentToPreferredPeer = true
				}
				return nil
			},
		}
		arg.OutputAntiflooder = &mock.P2PAntifloodHandlerStub{
			CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
				if fromConnectedPeer == pidPreferred {
					require.Fail(t, "CanProcessMessage should have not be called for preferred peer")
				}

				return nil
			},
		}

		selfShardIDProvider := mock.NewMultipleShardsCoordinatorMock()
		selfShardIDProvider.CurrentShard = selfShardID
		arg.SelfShardIdProvider = selfShardIDProvider

		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)
		require.NoError(t, err)
		require.True(t, sentToPreferredPeer)
	})
	t.Run("should not send to preferred peer if only one peer to request", func(t *testing.T) {
		pidPreferred := core.PeerID("preferred peer")
		numTimesSent := 0
		regularPeer0, regularPeer1 := core.PeerID("peer0"), core.PeerID("peer1")
		sentToPreferredPeer := false

		arg := createMockArgTopicRequestSender()
		arg.TargetShardId = 1
		arg.NumCrossShardPeers = 1
		arg.PeerListCreator = &mock.PeerListCreatorStub{
			CrossShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{regularPeer0, regularPeer1, regularPeer0, regularPeer1}
			},
			IntraShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{}
			},
		}
		arg.MainPreferredPeersHolder = &p2pmocks.PeersHolderStub{
			GetCalled: func() map[uint32][]core.PeerID {
				return map[uint32][]core.PeerID{
					37: {pidPreferred},
				}
			},
		}

		arg.MainMessenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				if bytes.Equal(peerID.Bytes(), pidPreferred.Bytes()) {
					sentToPreferredPeer = true
					require.Zero(t, numTimesSent)
				}

				numTimesSent++
				return nil
			},
		}
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)
		assert.Nil(t, err)
		assert.False(t, sentToPreferredPeer)
	})
	t.Run("should stop after sending to required num", func(t *testing.T) {
		t.Parallel()

		pIDs := []core.PeerID{"pid1", "pid2", "pid3", "pid4", "pid5"}

		numSent := 0
		arg := createMockArgTopicRequestSender()
		arg.MainMessenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				numSent++

				return nil
			},
		}
		arg.PeerListCreator = &mock.PeerListCreatorStub{
			CrossShardPeerListCalled: func() []core.PeerID {
				return pIDs
			},
			IntraShardPeerListCalled: func() []core.PeerID {
				return pIDs
			},
		}
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)

		assert.Nil(t, err)
		assert.Equal(t, arg.NumCrossShardPeers+arg.NumIntraShardPeers, numSent)
	})
	t.Run("should not call intra shard peers", func(t *testing.T) {
		t.Parallel()

		pIDs := []core.PeerID{"pid1", "pid2", "pid3", "pid4", "pid5"}
		pidNotCalled := core.PeerID("pid not called")

		numSent := 0
		arg := createMockArgTopicRequestSender()
		arg.MainMessenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				if peerID == pidNotCalled {
					assert.Fail(t, fmt.Sprintf("should not have called pid %s", peerID))
				}
				numSent++

				return nil
			},
		}
		arg.NumIntraShardPeers = 0
		arg.PeerListCreator = &mock.PeerListCreatorStub{
			CrossShardPeerListCalled: func() []core.PeerID {
				return pIDs
			},
			IntraShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{pidNotCalled}
			},
		}
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)

		assert.Nil(t, err)
		assert.Equal(t, arg.NumCrossShardPeers, numSent)
	})
	t.Run("should not call cross shard", func(t *testing.T) {
		t.Parallel()

		pIDs := []core.PeerID{"pid1", "pid2", "pid3", "pid4", "pid5"}
		pidNotCalled := core.PeerID("pid not called")

		numSent := 0
		arg := createMockArgTopicRequestSender()
		arg.MainMessenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				if peerID == pidNotCalled {
					assert.Fail(t, fmt.Sprintf("should not have called pid %s", peerID))
				}
				numSent++

				return nil
			},
		}
		arg.NumCrossShardPeers = 0
		arg.PeerListCreator = &mock.PeerListCreatorStub{
			CrossShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{pidNotCalled}
			},
			IntraShardPeerListCalled: func() []core.PeerID {
				return pIDs
			},
		}
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)

		assert.Nil(t, err)
		assert.Equal(t, arg.NumIntraShardPeers, numSent)
	})
	t.Run("SendToConnectedPeerCalled returns error", func(t *testing.T) {
		t.Parallel()

		pID1 := core.PeerID("peer1")
		sentToPid1 := false

		arg := createMockArgTopicRequestSender()
		arg.MainMessenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				if bytes.Equal(peerID.Bytes(), pID1.Bytes()) {
					sentToPid1 = true
				}

				return expectedErr
			},
		}
		arg.PeerListCreator = &mock.PeerListCreatorStub{
			CrossShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{pID1}
			},
			IntraShardPeerListCalled: func() []core.PeerID {
				return make([]core.PeerID, 0)
			},
		}
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)

		assert.True(t, errors.Is(err, dataRetriever.ErrSendRequest))
		assert.True(t, sentToPid1)
	})
	t.Run("should work and try on both networks", func(t *testing.T) {
		t.Parallel()

		crossPid := core.PeerID("cross peer")
		intraPid := core.PeerID("intra peer")
		cnt := 0

		arg := createMockArgTopicRequestSender()
		arg.MainMessenger = &p2pmocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				cnt++

				return nil
			},
		}
		arg.PeerListCreator = &mock.PeerListCreatorStub{
			CrossShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{crossPid}
			},
			IntraShardPeerListCalled: func() []core.PeerID {
				return []core.PeerID{intraPid}
			},
		}
		arg.FullArchiveMessenger = &p2pmocks.MessengerStub{
			ConnectedPeersCalled: func() []core.PeerID {
				return []core.PeerID{} // empty list, so it will fallback to the main network
			},
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				assert.Fail(t, "should have not been called")

				return nil
			},
		}
		arg.CurrentNetworkEpochProvider = &mock.CurrentNetworkEpochProviderStub{
			EpochIsActiveInNetworkCalled: func(epoch uint32) bool {
				return false // force the full archive network
			},
		}
		trs, _ := topicsender.NewTopicRequestSender(arg)
		assert.NotNil(t, trs)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)
		assert.Nil(t, err)
		assert.Equal(t, 2, cnt)
	})
}

func TestTopicRequestSender_NumPeersToQuery(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicRequestSender()
	trs, _ := topicsender.NewTopicRequestSender(arg)

	intra := 1123
	cross := 2143

	trs.SetNumPeersToQuery(intra, cross)
	recoveredIntra, recoveredCross := trs.NumPeersToQuery()

	assert.Equal(t, intra, recoveredIntra)
	assert.Equal(t, cross, recoveredCross)
}
