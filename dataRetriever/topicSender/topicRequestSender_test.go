package topicsender_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	topicsender "github.com/multiversx/mx-chain-go/dataRetriever/topicSender"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
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
				// the preferred peer must be a connected candidate: non-member preferred
				// peers are deliberately not injected anymore
				return []core.PeerID{regularPeer0, regularPeer1, pidPreferred}
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

var bandTestHashes = [][]byte{[]byte("hash")}

func createBandArg(
	epochIsRecent bool,
	epochOnMainPeers bool,
	topicPeers []core.PeerID,
	connectedPeers []core.PeerID,
	sentFullArchive *[]core.PeerID,
	sentMain *[]core.PeerID,
) topicsender.ArgTopicRequestSender {
	arg := createMockArgTopicRequestSender()
	arg.CurrentNetworkEpochProvider = &mock.CurrentNetworkEpochProviderStub{
		EpochIsActiveInNetworkCalled: func(epoch uint32) bool {
			return epochIsRecent
		},
		EpochIsAvailableOnMainPeersCalled: func(epoch uint32) bool {
			return epochOnMainPeers
		},
	}
	arg.FullArchiveMessenger = &p2pmocks.MessengerStub{
		ConnectedPeersOnTopicCalled: func(topic string) []core.PeerID {
			return topicPeers
		},
		ConnectedPeersCalled: func() []core.PeerID {
			return connectedPeers
		},
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			*sentFullArchive = append(*sentFullArchive, peerID)
			return nil
		},
	}
	arg.PeerListCreator = &mock.PeerListCreatorStub{
		CrossShardPeerListCalled: func() []core.PeerID {
			return []core.PeerID{"cross0", "cross1"}
		},
		IntraShardPeerListCalled: func() []core.PeerID {
			return []core.PeerID{"intra0"}
		},
	}
	arg.MainMessenger = &p2pmocks.MessengerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			*sentMain = append(*sentMain, peerID)
			return nil
		},
	}

	return arg
}

func countOccurrences(peers []core.PeerID, target core.PeerID) int {
	num := 0
	for _, peer := range peers {
		if peer == target {
			num++
		}
	}

	return num
}

func TestTopicRequestSender_ThreeBandRouting(t *testing.T) {
	t.Parallel()

	t.Run("band 1: recent epoch sends main only, wider predicate not consulted", func(t *testing.T) {
		t.Parallel()

		sentFullArchive := make([]core.PeerID, 0)
		sentMain := make([]core.PeerID, 0)
		arg := createBandArg(true, false, nil, nil, &sentFullArchive, &sentMain)
		arg.CurrentNetworkEpochProvider = &mock.CurrentNetworkEpochProviderStub{
			EpochIsActiveInNetworkCalled: func(epoch uint32) bool {
				return true
			},
			EpochIsAvailableOnMainPeersCalled: func(epoch uint32) bool {
				assert.Fail(t, "wider predicate should not be consulted for a recent epoch")
				return true
			},
		}
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, bandTestHashes)

		assert.Nil(t, err)
		assert.Empty(t, sentFullArchive)
		assert.Equal(t, 3, len(sentMain)) // 2 cross + 1 intra
	})
	t.Run("band 2: main with full budget plus exactly one full-archive insurance send", func(t *testing.T) {
		t.Parallel()

		topicPeers := []core.PeerID{"fa0", "fa1", "fa2"}
		sentFullArchive := make([]core.PeerID, 0)
		sentMain := make([]core.PeerID, 0)
		numReqIntra, numReqCross := 0, 0
		arg := createBandArg(false, true, topicPeers, topicPeers, &sentFullArchive, &sentMain)
		trs, _ := topicsender.NewTopicRequestSender(arg)
		_ = trs.SetDebugHandler(&mock.DebugHandler{
			LogRequestedDataCalled: func(topic string, hash [][]byte, intra int, cross int) {
				numReqIntra, numReqCross = intra, cross
			},
		})

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, bandTestHashes)

		assert.Nil(t, err)
		assert.Equal(t, 1, len(sentFullArchive))
		assert.Equal(t, 3, len(sentMain))
		// debug accounting keeps the full-archive count inside the intra aggregate
		assert.Equal(t, 2, numReqIntra) // 1 main intra + 1 full-archive
		assert.Equal(t, 2, numReqCross)
	})
	t.Run("band 3: full-archive only, full budget, exploration slot outside the topic view", func(t *testing.T) {
		t.Parallel()

		topicPeers := []core.PeerID{"fa0", "fa1", "fa2"}
		connectedPeers := []core.PeerID{"fa0", "fa1", "fa2", "hidden"}
		sentFullArchive := make([]core.PeerID, 0)
		sentMain := make([]core.PeerID, 0)
		arg := createBandArg(false, false, topicPeers, connectedPeers, &sentFullArchive, &sentMain)
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, bandTestHashes)

		assert.Nil(t, err)
		assert.Empty(t, sentMain)
		assert.Equal(t, 2, len(sentFullArchive)) // NumFullHistoryPeers
		assert.Equal(t, 1, countOccurrences(sentFullArchive, "hidden"))
	})
	t.Run("band 3: no exploration reservation when the topic view covers all connected peers", func(t *testing.T) {
		t.Parallel()

		topicPeers := []core.PeerID{"fa0", "fa1", "fa2"}
		sentFullArchive := make([]core.PeerID, 0)
		sentMain := make([]core.PeerID, 0)
		arg := createBandArg(false, false, topicPeers, topicPeers, &sentFullArchive, &sentMain)
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, bandTestHashes)

		assert.Nil(t, err)
		assert.Empty(t, sentMain)
		assert.Equal(t, 2, len(sentFullArchive))
		for _, peer := range sentFullArchive {
			assert.Equal(t, 1, countOccurrences(topicPeers, peer))
		}
	})
	t.Run("band 3: falls back to main network when no full-archive send succeeds", func(t *testing.T) {
		t.Parallel()

		sentFullArchive := make([]core.PeerID, 0)
		sentMain := make([]core.PeerID, 0)
		arg := createBandArg(false, false, nil, nil, &sentFullArchive, &sentMain)
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, bandTestHashes)

		assert.Nil(t, err)
		assert.Empty(t, sentFullArchive)
		assert.Equal(t, 3, len(sentMain))
	})
}

func TestTopicRequestSender_TwoPassSelection(t *testing.T) {
	t.Parallel()

	t.Run("topic view smaller than the budget fills from the remaining connected peers", func(t *testing.T) {
		t.Parallel()

		topicPeers := []core.PeerID{"fa0"}
		connectedPeers := []core.PeerID{"fa0", "fa1", "fa2"}
		sentFullArchive := make([]core.PeerID, 0)
		sentMain := make([]core.PeerID, 0)
		arg := createBandArg(false, false, topicPeers, connectedPeers, &sentFullArchive, &sentMain)
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, bandTestHashes)

		assert.Nil(t, err)
		assert.Equal(t, 2, len(sentFullArchive))
		assert.Equal(t, 1, countOccurrences(sentFullArchive, "fa0"))
	})
	t.Run("empty topic view degenerates to all connected peers", func(t *testing.T) {
		t.Parallel()

		connectedPeers := []core.PeerID{"fa0", "fa1", "fa2"}
		sentFullArchive := make([]core.PeerID, 0)
		sentMain := make([]core.PeerID, 0)
		arg := createBandArg(false, false, nil, connectedPeers, &sentFullArchive, &sentMain)
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, bandTestHashes)

		assert.Nil(t, err)
		assert.Empty(t, sentMain)
		assert.Equal(t, 2, len(sentFullArchive))
	})
}

func TestTopicRequestSender_PreferredPeerHandling(t *testing.T) {
	t.Parallel()

	pidPreferred := core.PeerID("preferred")
	setPreferred := func(arg *topicsender.ArgTopicRequestSender) {
		arg.FullArchivePreferredPeersHolder = &p2pmocks.PeersHolderStub{
			GetCalled: func() map[uint32][]core.PeerID {
				return map[uint32][]core.PeerID{0: {pidPreferred}}
			},
		}
	}

	t.Run("preferred topic peer is contacted exactly once and consumes one slot", func(t *testing.T) {
		t.Parallel()

		topicPeers := []core.PeerID{pidPreferred, "fa0", "fa1"}
		sentFullArchive := make([]core.PeerID, 0)
		sentMain := make([]core.PeerID, 0)
		arg := createBandArg(false, false, topicPeers, topicPeers, &sentFullArchive, &sentMain)
		setPreferred(&arg)
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, bandTestHashes)

		assert.Nil(t, err)
		assert.Equal(t, 2, len(sentFullArchive))
		assert.Equal(t, 1, countOccurrences(sentFullArchive, pidPreferred))
		assert.Equal(t, pidPreferred, sentFullArchive[0]) // injected first
	})
	t.Run("failed preferred send does not consume the budget", func(t *testing.T) {
		t.Parallel()

		topicPeers := []core.PeerID{pidPreferred, "fa0", "fa1"}
		sentFullArchive := make([]core.PeerID, 0)
		sentMain := make([]core.PeerID, 0)
		arg := createBandArg(false, false, topicPeers, topicPeers, &sentFullArchive, &sentMain)
		setPreferred(&arg)
		arg.FullArchiveMessenger = &p2pmocks.MessengerStub{
			ConnectedPeersOnTopicCalled: func(topic string) []core.PeerID {
				return topicPeers
			},
			ConnectedPeersCalled: func() []core.PeerID {
				return topicPeers
			},
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				if peerID == pidPreferred {
					return errors.New("send failed")
				}
				sentFullArchive = append(sentFullArchive, peerID)
				return nil
			},
		}
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, bandTestHashes)

		assert.Nil(t, err)
		assert.Equal(t, 2, len(sentFullArchive)) // full budget still used on the other peers
		assert.Zero(t, countOccurrences(sentFullArchive, pidPreferred))
	})
	t.Run("single-slot insurance keeps the preferred peer as ordinary candidate with full rating coverage", func(t *testing.T) {
		t.Parallel()

		topicPeers := []core.PeerID{pidPreferred, "fa0"}
		sentFullArchive := make([]core.PeerID, 0)
		sentMain := make([]core.PeerID, 0)
		fullArchiveCoverage := 0
		arg := createBandArg(false, true, topicPeers, topicPeers, &sentFullArchive, &sentMain)
		setPreferred(&arg)
		arg.PeersRatingHandler = &p2pmocks.PeersRatingHandlerStub{
			GetTopRatedPeersFromListCalled: func(peers []core.PeerID, numOfPeers int) []core.PeerID {
				if countOccurrences(peers, pidPreferred) == 1 {
					// the full-archive pass: the preferred peer was NOT stripped from the list
					fullArchiveCoverage = numOfPeers
				}
				return peers
			},
		}
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, bandTestHashes)

		assert.Nil(t, err)
		assert.Equal(t, 1, len(sentFullArchive))
		assert.Equal(t, len(topicPeers), fullArchiveCoverage) // coverage = len(candidates), not 1
	})
}

func TestTopicRequestSender_RatingCoverageForwarding(t *testing.T) {
	t.Parallel()

	// two top-rated + one bad-rated topic peers plus one fill peer: the reserved exploration slot
	// caps the topic pass at budget 2, but the rating coverage must still span all 3 topic peers
	// so the bad-rated one keeps a nonzero selection probability
	topicPeers := []core.PeerID{"topRated0", "topRated1", "badRated"}
	connectedPeers := []core.PeerID{"topRated0", "topRated1", "badRated", "fill"}
	sentFullArchive := make([]core.PeerID, 0)
	sentMain := make([]core.PeerID, 0)
	topicPassCoverage := 0
	explorationPassCoverage := 0
	arg := createBandArg(false, false, topicPeers, connectedPeers, &sentFullArchive, &sentMain)
	arg.NumFullHistoryPeers = 3
	arg.PeersRatingHandler = &p2pmocks.PeersRatingHandlerStub{
		GetTopRatedPeersFromListCalled: func(peers []core.PeerID, numOfPeers int) []core.PeerID {
			switch {
			case countOccurrences(peers, "badRated") == 1:
				topicPassCoverage = numOfPeers
			case countOccurrences(peers, "fill") == 1:
				explorationPassCoverage = numOfPeers
			}
			return peers
		},
	}
	trs, _ := topicsender.NewTopicRequestSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, bandTestHashes)

	assert.Nil(t, err)
	assert.Equal(t, len(topicPeers), topicPassCoverage)
	assert.Equal(t, 1, explorationPassCoverage)
	assert.Equal(t, 3, len(sentFullArchive))
	assert.Equal(t, 1, countOccurrences(sentFullArchive, "fill"))
}

func TestTopicRequestSender_ZeroFullHistoryPeersIsDefensive(t *testing.T) {
	t.Parallel()

	t.Run("band 2 with zero budget sends main only", func(t *testing.T) {
		t.Parallel()

		topicPeers := []core.PeerID{"fa0", "fa1"}
		sentFullArchive := make([]core.PeerID, 0)
		sentMain := make([]core.PeerID, 0)
		arg := createBandArg(false, true, topicPeers, topicPeers, &sentFullArchive, &sentMain)
		arg.NumFullHistoryPeers = 0
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, bandTestHashes)

		assert.Nil(t, err)
		assert.Empty(t, sentFullArchive)
		assert.Equal(t, 3, len(sentMain))
	})
	t.Run("band 3 with zero budget falls back to main", func(t *testing.T) {
		t.Parallel()

		topicPeers := []core.PeerID{"fa0", "fa1"}
		sentFullArchive := make([]core.PeerID, 0)
		sentMain := make([]core.PeerID, 0)
		arg := createBandArg(false, false, topicPeers, topicPeers, &sentFullArchive, &sentMain)
		arg.NumFullHistoryPeers = 0
		trs, _ := topicsender.NewTopicRequestSender(arg)

		err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, bandTestHashes)

		assert.Nil(t, err)
		assert.Empty(t, sentFullArchive)
		assert.Equal(t, 3, len(sentMain))
	})
}

func TestTopicRequestSender_PreferredPeerAsSoleCandidateIsStillQueried(t *testing.T) {
	t.Parallel()

	pidPreferred := core.PeerID("preferred")
	topicPeers := []core.PeerID{pidPreferred}
	sentFullArchive := make([]core.PeerID, 0)
	sentMain := make([]core.PeerID, 0)
	arg := createBandArg(false, false, topicPeers, topicPeers, &sentFullArchive, &sentMain)
	arg.NumFullHistoryPeers = 3
	arg.FullArchivePreferredPeersHolder = &p2pmocks.PeersHolderStub{
		GetCalled: func() map[uint32][]core.PeerID {
			return map[uint32][]core.PeerID{0: {pidPreferred}}
		},
	}
	trs, _ := topicsender.NewTopicRequestSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, bandTestHashes)

	assert.Nil(t, err)
	assert.Equal(t, 1, countOccurrences(sentFullArchive, pidPreferred))
	assert.Equal(t, 1, len(sentFullArchive))
}
