package topicResolverSender_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers/topicResolverSender"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultHashes = [][]byte{[]byte("hash")}

func createMockArgTopicResolverSender() topicResolverSender.ArgTopicResolverSender {
	return topicResolverSender.ArgTopicResolverSender{
		Messenger:                   &mock.MessageHandlerStub{},
		TopicName:                   "topic",
		PeerListCreator:             &mock.PeerListCreatorStub{},
		Marshalizer:                 &mock.MarshalizerMock{},
		Randomizer:                  &mock.IntRandomizerStub{},
		TargetShardId:               0,
		OutputAntiflooder:           &mock.P2PAntifloodHandlerStub{},
		NumIntraShardPeers:          2,
		NumCrossShardPeers:          2,
		NumFullHistoryPeers:         3,
		CurrentNetworkEpochProvider: &mock.CurrentNetworkEpochProviderStub{},
		SelfShardIdProvider:         mock.NewMultipleShardsCoordinatorMock(),
		PreferredPeersHolder: &p2pmocks.PeersHolderStub{
			GetCalled: func() map[uint32][]core.PeerID {
				return map[uint32][]core.PeerID{}
			},
		},
		PeersRatingHandler: &p2pmocks.PeersRatingHandlerStub{},
	}
}

//------- NewTopicResolverSender

func TestNewTopicResolverSender_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.Messenger = nil
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilMessenger, err)
}

func TestNewTopicResolverSender_NilPeersListCreatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.PeerListCreator = nil
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilPeerListCreator, err)
}

func TestNewTopicResolverSender_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.Marshalizer = nil
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestNewTopicResolverSender_NilRandomizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.Randomizer = nil
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilRandomizer, err)
}

func TestNewTopicResolverSender_NilOutputAntiflooderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.OutputAntiflooder = nil
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilAntifloodHandler, err)
}

func TestNewTopicResolverSender_NilCurrentNetworkEpochProviderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.CurrentNetworkEpochProvider = nil
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilCurrentNetworkEpochProvider, err)
}

func TestNewTopicResolverSender_NilPreferredPeersHolderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.PreferredPeersHolder = nil
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilPreferredPeersHolder, err)
}

func TestNewTopicResolverSender_NilPeersRatingHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.PeersRatingHandler = nil
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilPeersRatingHandler, err)
}

func TestNewTopicResolverSender_NilSelfShardIDProviderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.SelfShardIdProvider = nil
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.Equal(t, dataRetriever.ErrNilSelfShardIDProvider, err)
}

func TestNewTopicResolverSender_InvalidNumIntraShardPeersShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.NumIntraShardPeers = -1
	arg.NumCrossShardPeers = 100
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
}

func TestNewTopicResolverSender_InvalidNumCrossShardPeersShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.NumCrossShardPeers = -1
	arg.NumIntraShardPeers = 100
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
}

func TestNewTopicResolverSender_InvalidNumberOfFullHistoryPeersShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.NumFullHistoryPeers = -1
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
}

func TestNewTopicResolverSender_InvalidNumberOfPeersShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.NumIntraShardPeers = 1
	arg.NumCrossShardPeers = 0
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.True(t, check.IfNil(trs))
	assert.True(t, errors.Is(err, dataRetriever.ErrInvalidValue))
}

func TestNewTopicResolverSender_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.False(t, check.IfNil(trs))
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), trs.TargetShardID())
}

func TestNewTopicResolverSender_OkValsWithNumZeroShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.NumIntraShardPeers = 0
	trs, err := topicResolverSender.NewTopicResolverSender(arg)

	assert.False(t, check.IfNil(trs))
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), trs.TargetShardID())
}

//------- SendOnRequestTopic

func TestTopicResolverSender_SendOnRequestTopicMarshalizerFailsShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")

	arg := createMockArgTopicResolverSender()
	arg.Marshalizer = &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (bytes []byte, e error) {
			return nil, errExpected
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)

	assert.Equal(t, errExpected, err)
}

func TestTopicResolverSender_SendOnRequestTopicNoOneToSendShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	arg.PeerListCreator = &mock.PeerListCreatorStub{
		CrossShardPeerListCalled: func() []core.PeerID {
			return make([]core.PeerID, 0)
		},
		IntraShardPeerListCalled: func() []core.PeerID {
			return make([]core.PeerID, 0)
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)

	assert.True(t, errors.Is(err, dataRetriever.ErrSendRequest))
}

func TestTopicResolverSender_SendOnRequestTopicShouldWorkAndNotSendToFullHistoryNodes(t *testing.T) {
	t.Parallel()

	pID1 := core.PeerID("peer1")
	pID2 := core.PeerID("peer2")
	sentToPid1 := false
	sentToPid2 := false

	arg := createMockArgTopicResolverSender()
	arg.Messenger = &mock.MessageHandlerStub{
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
	arg.PeerListCreator = &mock.PeerListCreatorStub{
		CrossShardPeerListCalled: func() []core.PeerID {
			return []core.PeerID{pID1}
		},
		IntraShardPeerListCalled: func() []core.PeerID {
			return []core.PeerID{pID2}
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)

	assert.Nil(t, err)
	assert.True(t, sentToPid1)
	assert.True(t, sentToPid2)
}

func TestTopicResolverSender_SendOnRequestTopicShouldWorkAndSendToFullHistoryNodes(t *testing.T) {
	t.Parallel()

	pIDfullHistory := core.PeerID("full history peer")
	sentToFullHistoryPeer := false

	arg := createMockArgTopicResolverSender()
	arg.Messenger = &mock.MessageHandlerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			if bytes.Equal(peerID.Bytes(), pIDfullHistory.Bytes()) {
				sentToFullHistoryPeer = true
			}

			return nil
		},
	}
	arg.PeerListCreator = &mock.PeerListCreatorStub{
		FullHistoryListCalled: func() []core.PeerID {
			return []core.PeerID{pIDfullHistory}
		},
	}
	arg.CurrentNetworkEpochProvider = &mock.CurrentNetworkEpochProviderStub{
		EpochIsActiveInNetworkCalled: func(epoch uint32) bool {
			return false
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)
	assert.Nil(t, err)
	assert.True(t, sentToFullHistoryPeer)
}

func TestTopicResolverSender_SendOnRequestTopicShouldWorkAndSendToPreferredPeers(t *testing.T) {
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

	arg := createMockArgTopicResolverSender()
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
	arg.PreferredPeersHolder = &p2pmocks.PeersHolderStub{
		GetCalled: func() map[uint32][]core.PeerID {
			return map[uint32][]core.PeerID{
				selfShardID:   preferredPeersShard0,
				targetShardID: preferredPeersShard1,
			}
		},
	}
	arg.NumCrossShardPeers = 5
	arg.NumIntraShardPeers = 5
	arg.Messenger = &mock.MessageHandlerStub{
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
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)
	assert.Nil(t, err)
	assert.Equal(t, 1, countPrefPeersSh1)
}

func TestTopicResolverSender_SendOnRequestTopicShouldWorkAndSendToCrossPreferredPeerFirst(t *testing.T) {
	t.Parallel()

	targetShardID := uint32(37)
	pIDPreferred := core.PeerID("preferred peer")
	numTimesSent := 0
	regularPeer0, regularPeer1 := core.PeerID("peer0"), core.PeerID("peer1")
	sentToPreferredPeer := false

	arg := createMockArgTopicResolverSender()
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
	arg.PreferredPeersHolder = &p2pmocks.PeersHolderStub{
		GetCalled: func() map[uint32][]core.PeerID {
			return map[uint32][]core.PeerID{
				targetShardID: {pIDPreferred},
			}
		},
	}

	arg.Messenger = &mock.MessageHandlerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			if bytes.Equal(peerID.Bytes(), pIDPreferred.Bytes()) {
				sentToPreferredPeer = true
				require.Zero(t, numTimesSent)
			}

			numTimesSent++
			return nil
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)
	assert.Nil(t, err)
	assert.True(t, sentToPreferredPeer)
}

func TestTopicResolverSender_SendOnRequestTopicShouldWorkAndSendToIntraPreferredPeerFirst(t *testing.T) {
	t.Parallel()

	selfShardID := uint32(37)
	pIDPreferred := core.PeerID("preferred peer")
	numTimesSent := 0
	regularPeer0, regularPeer1 := core.PeerID("peer0"), core.PeerID("peer1")
	sentToPreferredPeer := false

	arg := createMockArgTopicResolverSender()
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
	arg.PreferredPeersHolder = &p2pmocks.PeersHolderStub{
		GetCalled: func() map[uint32][]core.PeerID {
			return map[uint32][]core.PeerID{
				selfShardID: {pIDPreferred},
			}
		},
	}

	arg.Messenger = &mock.MessageHandlerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			if bytes.Equal(peerID.Bytes(), pIDPreferred.Bytes()) {
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

	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)
	assert.Nil(t, err)
	assert.True(t, sentToPreferredPeer)
}

func TestTopicResolverSender_SendOnRequestTopicShouldWorkAndSkipAntifloodChecksForPreferredPeers(t *testing.T) {
	t.Parallel()

	selfShardID := uint32(37)
	pIDPreferred := core.PeerID("preferred peer")
	regularPeer0, regularPeer1 := core.PeerID("peer0"), core.PeerID("peer1")
	targetShardID := uint32(55)

	sentToPreferredPeer := false

	arg := createMockArgTopicResolverSender()
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
	arg.PreferredPeersHolder = &p2pmocks.PeersHolderStub{
		GetCalled: func() map[uint32][]core.PeerID {
			return map[uint32][]core.PeerID{
				targetShardID: {pIDPreferred},
			}
		},
		ContainsCalled: func(peerID core.PeerID) bool {
			return peerID == pIDPreferred
		},
	}

	arg.Messenger = &mock.MessageHandlerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			if peerID == pIDPreferred {
				sentToPreferredPeer = true
			}
			return nil
		},
	}
	arg.OutputAntiflooder = &mock.P2PAntifloodHandlerStub{
		CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			if fromConnectedPeer == pIDPreferred {
				require.Fail(t, "CanProcessMessage should have not be called for preferred peer")
			}

			return nil
		},
	}

	selfShardIDProvider := mock.NewMultipleShardsCoordinatorMock()
	selfShardIDProvider.CurrentShard = selfShardID
	arg.SelfShardIdProvider = selfShardIDProvider

	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)
	require.NoError(t, err)
	require.True(t, sentToPreferredPeer)
}

func TestTopicResolverSender_SendOnRequestTopicShouldNotSendToPreferredPeerFirstIfOnlyOnePeerToRequest(t *testing.T) {
	t.Parallel()

	pIDPreferred := core.PeerID("preferred peer")
	numTimesSent := 0
	regularPeer0, regularPeer1 := core.PeerID("peer0"), core.PeerID("peer1")
	sentToPreferredPeer := false

	arg := createMockArgTopicResolverSender()
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
	arg.PreferredPeersHolder = &p2pmocks.PeersHolderStub{
		GetCalled: func() map[uint32][]core.PeerID {
			return map[uint32][]core.PeerID{
				37: {pIDPreferred},
			}
		},
	}

	arg.Messenger = &mock.MessageHandlerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			if bytes.Equal(peerID.Bytes(), pIDPreferred.Bytes()) {
				sentToPreferredPeer = true
				require.Zero(t, numTimesSent)
			}

			numTimesSent++
			return nil
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)
	assert.Nil(t, err)
	assert.False(t, sentToPreferredPeer)
}

func TestTopicResolverSender_SendOnRequestShouldStopAfterSendingToRequiredNum(t *testing.T) {
	t.Parallel()

	pIDs := []core.PeerID{"pid1", "pid2", "pid3", "pid4", "pid5"}

	numSent := 0
	arg := createMockArgTopicResolverSender()
	arg.Messenger = &mock.MessageHandlerStub{
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
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)

	assert.Nil(t, err)
	assert.Equal(t, arg.NumCrossShardPeers+arg.NumIntraShardPeers, numSent)
}

func TestTopicResolverSender_SendOnRequestNoIntraShardShouldNotCallIntraShard(t *testing.T) {
	t.Parallel()

	pIDs := []core.PeerID{"pid1", "pid2", "pid3", "pid4", "pid5"}
	pIDNotCalled := core.PeerID("pid not called")

	numSent := 0
	arg := createMockArgTopicResolverSender()
	arg.Messenger = &mock.MessageHandlerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			if peerID == pIDNotCalled {
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
			return []core.PeerID{pIDNotCalled}
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)

	assert.Nil(t, err)
	assert.Equal(t, arg.NumCrossShardPeers, numSent)
}

func TestTopicResolverSender_SendOnRequestNoCrossShardShouldNotCallCrossShard(t *testing.T) {
	t.Parallel()

	pIDs := []core.PeerID{"pid1", "pid2", "pid3", "pid4", "pid5"}
	pIDNotCalled := core.PeerID("pid not called")

	numSent := 0
	arg := createMockArgTopicResolverSender()
	arg.Messenger = &mock.MessageHandlerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			if peerID == pIDNotCalled {
				assert.Fail(t, fmt.Sprintf("should not have called pid %s", peerID))
			}
			numSent++

			return nil
		},
	}
	arg.NumCrossShardPeers = 0
	arg.PeerListCreator = &mock.PeerListCreatorStub{
		CrossShardPeerListCalled: func() []core.PeerID {
			return []core.PeerID{pIDNotCalled}
		},
		IntraShardPeerListCalled: func() []core.PeerID {
			return pIDs
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)

	assert.Nil(t, err)
	assert.Equal(t, arg.NumIntraShardPeers, numSent)
}

func TestTopicResolverSender_SendOnRequestTopicErrorsShouldReturnError(t *testing.T) {
	t.Parallel()

	pID1 := core.PeerID("peer1")
	sentToPid1 := false

	expectedErr := errors.New("expected error")
	arg := createMockArgTopicResolverSender()
	arg.Messenger = &mock.MessageHandlerStub{
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
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SendOnRequestTopic(&dataRetriever.RequestData{}, defaultHashes)

	assert.True(t, errors.Is(err, dataRetriever.ErrSendRequest))
	assert.True(t, sentToPid1)
}

//------- Send

func TestTopicResolverSender_SendOutputAntiflooderErrorsShouldNotSendButError(t *testing.T) {
	t.Parallel()

	pID1 := core.PeerID("peer1")
	buffToSend := []byte("buff")

	expectedErr := errors.New("can not send to peer")
	arg := createMockArgTopicResolverSender()
	arg.Messenger = &mock.MessageHandlerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			assert.Fail(t, "send shouldn't have been called")

			return nil
		},
	}
	arg.OutputAntiflooder = &mock.P2PAntifloodHandlerStub{
		CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			if fromConnectedPeer == pID1 {
				return expectedErr
			}

			assert.Fail(t, "wrong peer provided, should have been called with the destination peer")
			return nil
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.Send(buffToSend, pID1)

	assert.True(t, errors.Is(err, expectedErr))
}

func TestTopicResolverSender_SendShouldNotCheckAntifloodForPreferred(t *testing.T) {
	t.Parallel()

	pID1 := core.PeerID("peer1")
	buffToSend := []byte("buff")
	sendWasCalled := false

	arg := createMockArgTopicResolverSender()
	arg.Messenger = &mock.MessageHandlerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			sendWasCalled = true
			return nil
		},
	}
	arg.OutputAntiflooder = &mock.P2PAntifloodHandlerStub{
		CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			require.Fail(t, "CanProcessMessage should have not be called for preferred peer")

			return nil
		},
	}
	arg.PreferredPeersHolder = &p2pmocks.PeersHolderStub{
		ContainsCalled: func(peerID core.PeerID) bool {
			return peerID == pID1
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.Send(buffToSend, pID1)
	require.NoError(t, err)
	require.True(t, sendWasCalled)
}

func TestTopicResolverSender_SendShouldWork(t *testing.T) {
	t.Parallel()

	pID1 := core.PeerID("peer1")
	sentToPid1 := false
	buffToSend := []byte("buff")

	arg := createMockArgTopicResolverSender()
	arg.Messenger = &mock.MessageHandlerStub{
		SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
			if bytes.Equal(peerID.Bytes(), pID1.Bytes()) &&
				bytes.Equal(buff, buffToSend) {
				sentToPid1 = true
			}

			return nil
		},
	}
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.Send(buffToSend, pID1)

	assert.Nil(t, err)
	assert.True(t, sentToPid1)
}

func TestTopicResolverSender_Topic(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	assert.Equal(t, arg.TopicName+topicResolverSender.TopicRequestSuffix, trs.RequestTopic())
}

func TestTopicResolverSender_ResolverDebugHandler(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	handler := &mock.ResolverDebugHandler{}

	err := trs.SetResolverDebugHandler(handler)
	assert.Nil(t, err)

	assert.True(t, handler == trs.ResolverDebugHandler()) //pointer testing
}

func TestTopicResolverSender_SetResolverDebugHandlerNilShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	err := trs.SetResolverDebugHandler(nil)
	assert.Equal(t, dataRetriever.ErrNilResolverDebugHandler, err)
}

func TestTopicResolverSender_NumPeersToQueryr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTopicResolverSender()
	trs, _ := topicResolverSender.NewTopicResolverSender(arg)

	intra := 1123
	cross := 2143

	trs.SetNumPeersToQuery(intra, cross)
	recoveredIntra, recoveredCross := trs.NumPeersToQuery()

	assert.Equal(t, intra, recoveredIntra)
	assert.Equal(t, cross, recoveredCross)
}
