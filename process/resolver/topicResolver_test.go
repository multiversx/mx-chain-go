package resolver_test

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/resolver"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

//-------NewTopicResolver

func TestNewTopicResolver_NilMessengerShouldErr(t *testing.T) {
	t.Parallel()

	tr, err := resolver.NewTopicResolver("test", nil, &mock.MarshalizerMock{})

	assert.Equal(t, process.ErrNilMessenger, err)
	assert.Nil(t, tr)
}

func TestNewTopicResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	tr, err := resolver.NewTopicResolver("test", &mock.MessengerStub{}, nil)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, tr)
}

func TestNewTopicResolver_HasTopicValidatorShouldErr(t *testing.T) {
	t.Parallel()

	topicName := "test"

	ms := &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return topicName+resolver.RequestTopicSuffix == name
		},
	}
	tr, err := resolver.NewTopicResolver(topicName, ms, &mock.MarshalizerMock{})

	assert.Equal(t, process.ErrValidatorAlreadySet, err)
	assert.Nil(t, tr)
}

func TestNewTopicResolver_CreateTopicErrorsShouldErr(t *testing.T) {
	t.Parallel()

	topicName := "test"
	errCreateTopic := errors.New("create topic error")

	ms := &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		HasTopicCalled: func(name string) bool {
			return false
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			if name == topicName+resolver.RequestTopicSuffix {
				return errCreateTopic
			}
			return nil
		},
	}
	tr, err := resolver.NewTopicResolver(topicName, ms, &mock.MarshalizerMock{})

	assert.Equal(t, errCreateTopic, err)
	assert.Nil(t, tr)
}

func TestNewTopicResolver_OkValsWithCreateTopicShouldWork(t *testing.T) {
	t.Parallel()

	topicName := "test"

	createWasCalled := false
	setWasCalled := false

	ms := &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		HasTopicCalled: func(name string) bool {
			return false
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			if name == topicName+resolver.RequestTopicSuffix {
				createWasCalled = true
			}
			return nil
		},
		SetTopicValidatorCalled: func(topic string, handler func(message p2p.MessageP2P) error) error {
			if topic == topicName+resolver.RequestTopicSuffix {
				setWasCalled = true
			}
			return nil
		},
	}
	tr, err := resolver.NewTopicResolver(topicName, ms, &mock.MarshalizerMock{})

	assert.Nil(t, err)
	assert.NotNil(t, tr)
	assert.True(t, setWasCalled && createWasCalled)
}

func TestNewTopicResolver_OkValsWithoutCreateTopicShouldWork(t *testing.T) {
	t.Parallel()

	topicName := "test"

	createWasCalled := false
	setWasCalled := false

	ms := &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		HasTopicCalled: func(name string) bool {
			return true
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			if name == topicName+resolver.RequestTopicSuffix {
				createWasCalled = true
			}
			return nil
		},
		SetTopicValidatorCalled: func(topic string, handler func(message p2p.MessageP2P) error) error {
			if topic == topicName+resolver.RequestTopicSuffix {
				setWasCalled = true
			}
			return nil
		},
	}
	tr, err := resolver.NewTopicResolver(topicName, ms, &mock.MarshalizerMock{})

	assert.Nil(t, err)
	assert.NotNil(t, tr)
	assert.True(t, setWasCalled && !createWasCalled)
}

func TestNewTopicResolver_SetTopicValidatorErrorsShouldErr(t *testing.T) {
	t.Parallel()

	topicName := "test"
	errSetTopicValidator := errors.New("set topic validator error")

	ms := &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		HasTopicCalled: func(name string) bool {
			return true
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			return nil
		},
		SetTopicValidatorCalled: func(topic string, handler func(message p2p.MessageP2P) error) error {
			return errSetTopicValidator
		},
	}
	tr, err := resolver.NewTopicResolver(topicName, ms, &mock.MarshalizerMock{})

	assert.Equal(t, errSetTopicValidator, err)
	assert.Nil(t, tr)
}

//------- RequestValidator

func TestTopicResolver_RequestValidatorNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	topicName := "test"

	ms := &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		HasTopicCalled: func(name string) bool {
			return true
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			return nil
		},
		SetTopicValidatorCalled: func(topic string, handler func(message p2p.MessageP2P) error) error {
			return nil
		},
	}
	tr, _ := resolver.NewTopicResolver(topicName, ms, &mock.MarshalizerMock{})

	err := tr.RequestValidator(nil)
	assert.Equal(t, process.ErrNilMessage, err)
}

func TestTopicResolver_RequestValidatorNilDataInMessageShouldErr(t *testing.T) {
	t.Parallel()

	topicName := "test"

	ms := &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		HasTopicCalled: func(name string) bool {
			return true
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			return nil
		},
		SetTopicValidatorCalled: func(topic string, handler func(message p2p.MessageP2P) error) error {
			return nil
		},
	}
	tr, _ := resolver.NewTopicResolver(topicName, ms, &mock.MarshalizerMock{})

	msg := &mock.P2PMessageMock{}

	err := tr.RequestValidator(msg)
	assert.Equal(t, process.ErrNilDataToProcess, err)
}

func TestTopicResolver_RequestValidatorMarshalizerFailsAtUnmarshalingShouldErr(t *testing.T) {
	t.Parallel()

	topicName := "test"

	ms := &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		HasTopicCalled: func(name string) bool {
			return true
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			return nil
		},
		SetTopicValidatorCalled: func(topic string, handler func(message p2p.MessageP2P) error) error {
			return nil
		},
	}

	errMarshalizer := errors.New("marshalizer error")

	tr, _ := resolver.NewTopicResolver(topicName, ms, &mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return errMarshalizer
		},
	})

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	err := tr.RequestValidator(msg)
	assert.Equal(t, errMarshalizer, err)
}

func TestTopicResolver_RequestValidatorNilResolveRequestShouldErr(t *testing.T) {
	t.Parallel()

	topicName := "test"

	ms := &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		HasTopicCalled: func(name string) bool {
			return true
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			return nil
		},
		SetTopicValidatorCalled: func(topic string, handler func(message p2p.MessageP2P) error) error {
			return nil
		},
	}

	tr, _ := resolver.NewTopicResolver(topicName, ms, &mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return nil
		},
	})

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	err := tr.RequestValidator(msg)
	assert.Equal(t, process.ErrNilResolverHandler, err)
}

func TestTopicResolver_RequestValidatorResolverHandlerErrorsShouldErr(t *testing.T) {
	t.Parallel()

	topicName := "test"

	ms := &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		HasTopicCalled: func(name string) bool {
			return true
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			return nil
		},
		SetTopicValidatorCalled: func(topic string, handler func(message p2p.MessageP2P) error) error {
			return nil
		},
	}

	tr, _ := resolver.NewTopicResolver(topicName, ms, &mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return nil
		},
	})

	errResolver := errors.New("resolver handler error")

	tr.SetResolverHandler(func(rd process.RequestData) (bytes []byte, e error) {
		return nil, errResolver
	})

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
	}

	err := tr.RequestValidator(msg)
	assert.Equal(t, errResolver, err)
}

func TestTopicResolver_RequestValidatorResolverHandlerReturnsBuffShouldWork(t *testing.T) {
	t.Parallel()

	topicName := "test"

	var sentBuffer []byte
	var sentTopic string
	var sentPeerID p2p.PeerID

	ms := &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		HasTopicCalled: func(name string) bool {
			return true
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			return nil
		},
		SetTopicValidatorCalled: func(topic string, handler func(message p2p.MessageP2P) error) error {
			return nil
		},
		SendDirectToConnectedPeerCalled: func(topic string, buff []byte, peerID p2p.PeerID) error {
			sentBuffer = buff
			sentTopic = topic
			sentPeerID = peerID
			return nil
		},
	}

	tr, _ := resolver.NewTopicResolver(topicName, ms, &mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return nil
		},
	})

	buffResolver := []byte("resolver's buffer containg requested data")

	tr.SetResolverHandler(func(rd process.RequestData) (bytes []byte, e error) {
		return buffResolver, nil
	})

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
		PeerField: p2p.PeerID("requestor"),
	}

	err := tr.RequestValidator(msg)
	assert.Nil(t, err)
	assert.Equal(t, buffResolver, sentBuffer)
	assert.Equal(t, msg.Peer(), sentPeerID)
	assert.Equal(t, topicName, sentTopic)
}

func TestTopicResolver_RequestValidatorResolverHandlerReturnsErrorShouldErr(t *testing.T) {
	t.Parallel()

	topicName := "test"

	errSendDirect := errors.New("send direct error")

	ms := &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		HasTopicCalled: func(name string) bool {
			return true
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			return nil
		},
		SetTopicValidatorCalled: func(topic string, handler func(message p2p.MessageP2P) error) error {
			return nil
		},
		SendDirectToConnectedPeerCalled: func(topic string, buff []byte, peerID p2p.PeerID) error {
			return errSendDirect
		},
	}

	tr, _ := resolver.NewTopicResolver(topicName, ms, &mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return nil
		},
	})

	buffResolver := []byte("resolver's buffer containg requested data")

	tr.SetResolverHandler(func(rd process.RequestData) (bytes []byte, e error) {
		return buffResolver, nil
	})

	msg := &mock.P2PMessageMock{
		DataField: make([]byte, 0),
		PeerField: p2p.PeerID("requestor"),
	}

	err := tr.RequestValidator(msg)
	assert.Equal(t, errSendDirect, err)
}

// ------- RequestData

func TestNewTopicResolver_RequestDataMarshalizerFailsAtMarsahlingShouldErr(t *testing.T) {
	t.Parallel()

	topicName := "test"

	ms := &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		HasTopicCalled: func(name string) bool {
			return true
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			return nil
		},
		SetTopicValidatorCalled: func(topic string, handler func(message p2p.MessageP2P) error) error {
			return nil
		},
	}

	errMarshalizer := errors.New("marshalizer error")

	tr, _ := resolver.NewTopicResolver(topicName, ms, &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (bytes []byte, e error) {
			return nil, errMarshalizer
		},
	})

	err := tr.RequestData(process.RequestData{})
	assert.Equal(t, errMarshalizer, err)
}

func TestNewTopicResolver_RequestDataSendTo0ConnectedPeersShouldErr(t *testing.T) {
	t.Parallel()

	topicName := "test"

	ms := &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		HasTopicCalled: func(name string) bool {
			return true
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			return nil
		},
		SetTopicValidatorCalled: func(topic string, handler func(message p2p.MessageP2P) error) error {
			return nil
		},
		ConnectedPeersCalled: func() []p2p.PeerID {
			return make([]p2p.PeerID, 0)
		},
	}

	tr, _ := resolver.NewTopicResolver(topicName, ms, &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (bytes []byte, e error) {
			return make([]byte, 0), nil
		},
	})

	err := tr.RequestData(process.RequestData{})
	assert.Equal(t, process.ErrNoConnectedPeerToSendRequest, err)
}

func TestNewTopicResolver_RequestDataSendTo1ConnectedPeersShouldWork(t *testing.T) {
	t.Parallel()

	topicName := "test"

	peer1 := p2p.PeerID("peer ID 1")
	flagSentToPeer1 := false
	buffToSend := []byte("buff to send")

	ms := &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		HasTopicCalled: func(name string) bool {
			return true
		},
		CreateTopicCalled: func(name string, createPipeForTopic bool) error {
			return nil
		},
		SetTopicValidatorCalled: func(topic string, handler func(message p2p.MessageP2P) error) error {
			return nil
		},
		ConnectedPeersCalled: func() []p2p.PeerID {
			return []p2p.PeerID{peer1}
		},
		SendDirectToConnectedPeerCalled: func(topic string, buff []byte, peerID p2p.PeerID) error {
			if !bytes.Equal(buffToSend, buff) {
				return nil
			}
			if topic != topicName+resolver.RequestTopicSuffix {
				return nil
			}
			if peerID == peer1 {
				flagSentToPeer1 = true
			}

			return nil
		},
	}

	tr, _ := resolver.NewTopicResolver(topicName, ms, &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (bytes []byte, e error) {
			return buffToSend, nil
		},
	})

	err := tr.RequestData(process.RequestData{})
	assert.Equal(t, nil, err)
	assert.True(t, flagSentToPeer1)
}

// ------- SelectRandomPeers

func TestSelectRandomPeers_ConnectedPeersLen0ShoudRetEmpty(t *testing.T) {
	t.Parallel()

	selectedPeers := resolver.SelectRandomPeers(make([]p2p.PeerID, 0), 0, nil)

	assert.Equal(t, 0, len(selectedPeers))
}

func TestSelectRandomPeers_ConnectedPeersLenSmallerThanRequiredShoudRetListTest1(t *testing.T) {
	t.Parallel()

	connectedPeers := []p2p.PeerID{p2p.PeerID("peer 1"), p2p.PeerID("peer 2")}

	selectedPeers := resolver.SelectRandomPeers(connectedPeers, 3, nil)

	assert.Equal(t, connectedPeers, selectedPeers)
}

func TestSelectRandomPeers_ConnectedPeersLenSmallerThanRequiredShoudRetListTest2(t *testing.T) {
	t.Parallel()

	connectedPeers := []p2p.PeerID{p2p.PeerID("peer 1"), p2p.PeerID("peer 2")}

	selectedPeers := resolver.SelectRandomPeers(connectedPeers, 2, nil)

	assert.Equal(t, connectedPeers, selectedPeers)
}

func TestSelectRandomPeers_ConnectedPeersTestRandomizerRepeat0ThreeTimes(t *testing.T) {
	t.Parallel()

	connectedPeers := []p2p.PeerID{p2p.PeerID("peer 1"), p2p.PeerID("peer 2"), p2p.PeerID("peer 3")}

	valuesGenerated := []int{0, 0, 0, 1}
	idxGenerated := 0

	mr := &mock.IntRandomizerMock{
		IntnCalled: func(n int) int {
			val := valuesGenerated[idxGenerated]
			idxGenerated++
			return val
		},
	}

	selectedPeers := resolver.SelectRandomPeers(connectedPeers, 2, mr)

	//since iterating a map does not guarantee the order, we have to search in any combination possible
	foundPeer0 := false
	foundPeer1 := false

	for i := 0; i < len(selectedPeers); i++ {
		if selectedPeers[i] == connectedPeers[0] {
			foundPeer0 = true
		}
		if selectedPeers[i] == connectedPeers[1] {
			foundPeer1 = true
		}
	}

	assert.True(t, foundPeer0 && foundPeer1)
	assert.Equal(t, 2, len(selectedPeers))
}

func TestSelectRandomPeers_ConnectedPeersTestRandomizerRepeat2TwoTimes(t *testing.T) {
	t.Parallel()

	connectedPeers := []p2p.PeerID{p2p.PeerID("peer 1"), p2p.PeerID("peer 2"), p2p.PeerID("peer 3")}

	valuesGenerated := []int{2, 2, 0}
	idxGenerated := 0

	mr := &mock.IntRandomizerMock{
		IntnCalled: func(n int) int {
			val := valuesGenerated[idxGenerated]
			idxGenerated++
			return val
		},
	}

	selectedPeers := resolver.SelectRandomPeers(connectedPeers, 2, mr)

	//since iterating a map does not guarantee the order, we have to search in any combination possible
	foundPeer0 := false
	foundPeer2 := false

	for i := 0; i < len(selectedPeers); i++ {
		if selectedPeers[i] == connectedPeers[0] {
			foundPeer0 = true
		}
		if selectedPeers[i] == connectedPeers[2] {
			foundPeer2 = true
		}
	}

	assert.True(t, foundPeer0 && foundPeer2)
	assert.Equal(t, 2, len(selectedPeers))
}
