package libp2p_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/mock"
	"github.com/btcsuite/btcd/btcec"
	ggio "github.com/gogo/protobuf/io"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-protocol"
	"github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const timeout = time.Second * 5

var blankMessageHandler = func(msg p2p.MessageP2P) error {
	return nil
}

func generateHostStub() *mock.ConnectableHostStub {
	return &mock.ConnectableHostStub{
		SetStreamHandlerCalled: func(pid protocol.ID, handler net.StreamHandler) {},
	}
}

func createConnStub(stream net.Stream, id peer.ID, sk crypto.PrivKey, remotePeer peer.ID) *mock.ConnStub {
	return &mock.ConnStub{
		GetStreamsCalled: func() []net.Stream {
			if stream == nil {
				return make([]net.Stream, 0)
			}

			return []net.Stream{stream}
		},
		LocalPeerCalled: func() peer.ID {
			return id
		},
		LocalPrivateKeyCalled: func() crypto.PrivKey {
			return sk
		},
		RemotePeerCalled: func() peer.ID {
			return remotePeer
		},
	}
}

func createLibP2PCredentialsDirectSender() (peer.ID, crypto.PrivKey) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*crypto.Secp256k1PrivateKey)(prvKey)
	id, _ := peer.IDFromPublicKey(sk.GetPublic())

	return id, sk
}

//------- NewDirectSender

func TestNewDirectSender_NilContextShouldErr(t *testing.T) {
	hs := &mock.ConnectableHostStub{}

	ds, err := libp2p.NewDirectSender(nil, hs, func(msg p2p.MessageP2P) error {
		return nil
	})

	assert.Nil(t, ds)
	assert.Equal(t, p2p.ErrNilContext, err)
}

func TestNewDirectSender_NilHostShouldErr(t *testing.T) {
	ds, err := libp2p.NewDirectSender(context.Background(), nil, func(msg p2p.MessageP2P) error {
		return nil
	})

	assert.Nil(t, ds)
	assert.Equal(t, p2p.ErrNilHost, err)
}

func TestNewDirectSender_NilMessageHandlerShouldErr(t *testing.T) {
	ds, err := libp2p.NewDirectSender(context.Background(), generateHostStub(), nil)

	assert.Nil(t, ds)
	assert.Equal(t, p2p.ErrNilDirectSendMessageHandler, err)
}

func TestNewDirectSender_OkValsShouldWork(t *testing.T) {
	ds, err := libp2p.NewDirectSender(context.Background(), generateHostStub(), func(msg p2p.MessageP2P) error {
		return nil
	})

	assert.NotNil(t, ds)
	assert.Nil(t, err)
}

func TestNewDirectSender_OkValsShouldCallSetStreamHandlerWithCorrectValues(t *testing.T) {
	var pidCalled protocol.ID
	var handlerCalled net.StreamHandler

	hs := &mock.ConnectableHostStub{
		SetStreamHandlerCalled: func(pid protocol.ID, handler net.StreamHandler) {
			pidCalled = pid
			handlerCalled = handler
		},
	}

	_, _ = libp2p.NewDirectSender(context.Background(), hs, func(msg p2p.MessageP2P) error {
		return nil
	})

	assert.NotNil(t, handlerCalled)
	assert.Equal(t, libp2p.DirectSendID, pidCalled)
}

//------- ProcessReceivedDirectMessage

func TestDirectSender_ProcessReceivedDirectMessageNilMessageShouldErr(t *testing.T) {
	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		blankMessageHandler,
	)

	err := ds.ProcessReceivedDirectMessage(nil)

	assert.Equal(t, p2p.ErrNilMessage, err)
}

func TestDirectSender_ProcessReceivedDirectMessageNilTopicIdsShouldErr(t *testing.T) {
	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		blankMessageHandler,
	)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pubsub_pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	msg.TopicIDs = nil

	err := ds.ProcessReceivedDirectMessage(msg)

	assert.Equal(t, p2p.ErrNilTopic, err)
}

func TestDirectSender_ProcessReceivedDirectMessageEmptyTopicIdsShouldErr(t *testing.T) {
	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		blankMessageHandler,
	)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pubsub_pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	msg.TopicIDs = make([]string, 0)

	err := ds.ProcessReceivedDirectMessage(msg)

	assert.Equal(t, p2p.ErrEmptyTopicList, err)
}

func TestDirectSender_ProcessReceivedDirectMessageAlreadySeenMsgShouldErr(t *testing.T) {
	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		blankMessageHandler,
	)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pubsub_pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	msg.TopicIDs = []string{"topic"}

	msgId := string(msg.GetFrom()) + string(msg.GetSeqno())
	ds.SeenMessages().Add(msgId)

	err := ds.ProcessReceivedDirectMessage(msg)

	assert.Equal(t, p2p.ErrAlreadySeenMessage, err)
}

func TestDirectSender_ProcessReceivedDirectMessageShouldWork(t *testing.T) {
	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		blankMessageHandler,
	)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pubsub_pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	msg.TopicIDs = []string{"topic"}

	err := ds.ProcessReceivedDirectMessage(msg)

	assert.Nil(t, err)
}

func TestDirectSender_ProcessReceivedDirectMessageShouldCallMessageHandler(t *testing.T) {
	wasCalled := false

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		func(msg p2p.MessageP2P) error {
			wasCalled = true
			return nil
		},
	)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pubsub_pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	msg.TopicIDs = []string{"topic"}

	_ = ds.ProcessReceivedDirectMessage(msg)

	assert.True(t, wasCalled)
}

func TestDirectSender_ProcessReceivedDirectMessageShouldReturnHandlersError(t *testing.T) {
	checkErr := errors.New("checking error")

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		func(msg p2p.MessageP2P) error {
			return checkErr
		},
	)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pubsub_pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	msg.TopicIDs = []string{"topic"}

	err := ds.ProcessReceivedDirectMessage(msg)

	assert.Equal(t, checkErr, err)
}

//------- SendDirectToConnectedPeer

func TestDirectSender_SendDirectToConnectedPeerBufferToLargeShouldErr(t *testing.T) {
	netw := &mock.NetworkStub{}

	id, sk := createLibP2PCredentialsDirectSender()
	remotePeer := peer.ID("remote peer")

	stream := mock.NewStreamMock()
	stream.SetProtocol(libp2p.DirectSendID)

	cs := createConnStub(stream, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []net.Conn {
		return []net.Conn{cs}
	}

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		&mock.ConnectableHostStub{
			SetStreamHandlerCalled: func(pid protocol.ID, handler net.StreamHandler) {},
			NetworkCalled: func() net.Network {
				return netw
			},
		},
		blankMessageHandler,
	)

	messageTooLarge := bytes.Repeat([]byte{65}, libp2p.MaxSendBuffSize)

	err := ds.Send("topic", messageTooLarge, p2p.PeerID(cs.RemotePeer()))

	assert.Equal(t, p2p.ErrMessageTooLarge, err)
}

func TestDirectSender_SendDirectToConnectedPeerNotConnectedPeerShouldErr(t *testing.T) {
	netw := &mock.NetworkStub{
		ConnsToPeerCalled: func(p peer.ID) []net.Conn {
			return make([]net.Conn, 0)
		},
	}

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		&mock.ConnectableHostStub{
			SetStreamHandlerCalled: func(pid protocol.ID, handler net.StreamHandler) {},
			NetworkCalled: func() net.Network {
				return netw
			},
		},
		blankMessageHandler,
	)

	err := ds.Send("topic", []byte("data"), "not connected peer")

	assert.Equal(t, p2p.ErrPeerNotDirectlyConnected, err)
}

func TestDirectSender_SendDirectToConnectedPeerNewStreamErrorsShouldErr(t *testing.T) {
	t.Parallel()

	netw := &mock.NetworkStub{}

	hs := &mock.ConnectableHostStub{
		SetStreamHandlerCalled: func(pid protocol.ID, handler net.StreamHandler) {},
		NetworkCalled: func() net.Network {
			return netw
		},
	}

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		hs,
		blankMessageHandler,
	)

	id, sk := createLibP2PCredentialsDirectSender()
	remotePeer := peer.ID("remote peer")
	errNewStream := errors.New("new stream error")

	cs := createConnStub(nil, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []net.Conn {
		return []net.Conn{cs}
	}

	hs.NewStreamCalled = func(ctx context.Context, p peer.ID, pids ...protocol.ID) (net.Stream, error) {
		return nil, errNewStream
	}

	data := []byte("data")
	topic := "topic"
	err := ds.Send(topic, data, p2p.PeerID(cs.RemotePeer()))

	assert.Equal(t, errNewStream, err)
}

func TestDirectSender_SendDirectToConnectedPeerExistingStreamShouldSendToStream(t *testing.T) {
	netw := &mock.NetworkStub{}

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		&mock.ConnectableHostStub{
			SetStreamHandlerCalled: func(pid protocol.ID, handler net.StreamHandler) {},
			NetworkCalled: func() net.Network {
				return netw
			},
		},
		blankMessageHandler,
	)

	generatedCounter := ds.Counter()

	id, sk := createLibP2PCredentialsDirectSender()
	remotePeer := peer.ID("remote peer")

	stream := mock.NewStreamMock()
	stream.SetProtocol(libp2p.DirectSendID)

	cs := createConnStub(stream, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []net.Conn {
		return []net.Conn{cs}
	}

	receivedMsg := &pubsub_pb.Message{}
	chanDone := make(chan bool)

	go func(s net.Stream) {
		reader := ggio.NewDelimitedReader(s, 1<<20)
		for {
			err := reader.ReadMsg(receivedMsg)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			chanDone <- true
		}
	}(stream)

	data := []byte("data")
	topic := "topic"
	err := ds.Send(topic, data, p2p.PeerID(cs.RemotePeer()))

	select {
	case <-chanDone:
	case <-time.After(timeout):
		assert.Fail(t, "timeout getting data from stream")
		return
	}

	assert.Nil(t, err)
	assert.Equal(t, receivedMsg.Data, data)
	assert.Equal(t, receivedMsg.TopicIDs[0], topic)
	assert.Equal(t, receivedMsg.Seqno, ds.NextSeqno(&generatedCounter))
}

func TestDirectSender_SendDirectToConnectedPeerNewStreamShouldSendToStream(t *testing.T) {
	netw := &mock.NetworkStub{}

	hs := &mock.ConnectableHostStub{
		SetStreamHandlerCalled: func(pid protocol.ID, handler net.StreamHandler) {},
		NetworkCalled: func() net.Network {
			return netw
		},
	}

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		hs,
		blankMessageHandler,
	)

	generatedCounter := ds.Counter()

	id, sk := createLibP2PCredentialsDirectSender()
	remotePeer := peer.ID("remote peer")

	stream := mock.NewStreamMock()
	stream.SetProtocol(libp2p.DirectSendID)

	cs := createConnStub(stream, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []net.Conn {
		return []net.Conn{cs}
	}

	hs.NewStreamCalled = func(ctx context.Context, p peer.ID, pids ...protocol.ID) (net.Stream, error) {
		if p == remotePeer && pids[0] == libp2p.DirectSendID {
			return stream, nil
		}
		return nil, errors.New("wrong parameters")
	}

	receivedMsg := &pubsub_pb.Message{}
	chanDone := make(chan bool)

	go func(s net.Stream) {
		reader := ggio.NewDelimitedReader(s, 1<<20)
		for {
			err := reader.ReadMsg(receivedMsg)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			chanDone <- true
		}
	}(stream)

	data := []byte("data")
	topic := "topic"
	err := ds.Send(topic, data, p2p.PeerID(cs.RemotePeer()))

	select {
	case <-chanDone:
	case <-time.After(timeout):
		assert.Fail(t, "timeout getting data from stream")
		return
	}

	assert.Nil(t, err)
	assert.Equal(t, receivedMsg.Data, data)
	assert.Equal(t, receivedMsg.TopicIDs[0], topic)
	assert.Equal(t, receivedMsg.Seqno, ds.NextSeqno(&generatedCounter))
}

//------- received mesages tests

func TestDirectSender_ReceivedSentMessageShouldCallMessageHandlerTestFullCycle(t *testing.T) {
	var streamHandler net.StreamHandler
	netw := &mock.NetworkStub{}

	hs := &mock.ConnectableHostStub{
		SetStreamHandlerCalled: func(pid protocol.ID, handler net.StreamHandler) {
			streamHandler = handler
		},
		NetworkCalled: func() net.Network {
			return netw
		},
	}

	var receivedMsg p2p.MessageP2P
	chanDone := make(chan bool)

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		hs,
		func(msg p2p.MessageP2P) error {
			receivedMsg = msg
			chanDone <- true
			return nil
		},
	)

	id, sk := createLibP2PCredentialsDirectSender()
	remotePeer := peer.ID("remote peer")

	stream := mock.NewStreamMock()
	stream.SetProtocol(libp2p.DirectSendID)

	streamHandler(stream)

	cs := createConnStub(stream, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []net.Conn {
		return []net.Conn{cs}
	}

	data := []byte("data")
	topic := "topic"
	_ = ds.Send(topic, data, p2p.PeerID(cs.RemotePeer()))

	select {
	case <-chanDone:
	case <-time.After(timeout):
		assert.Fail(t, "timeout")
		return
	}

	assert.NotNil(t, receivedMsg)
	assert.Equal(t, data, receivedMsg.Data())
	assert.Equal(t, []string{topic}, receivedMsg.TopicIDs())
}
