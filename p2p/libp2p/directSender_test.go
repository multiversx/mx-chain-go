package libp2p_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/ElrondNetwork/go-libp2p-pubsub"
	pb "github.com/ElrondNetwork/go-libp2p-pubsub/pb"
	"github.com/btcsuite/btcd/btcec"
	ggio "github.com/gogo/protobuf/io"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/stretchr/testify/assert"
)

const timeout = time.Second * 5
const testMaxSize = 1 << 21

var blankMessageHandler = func(msg *pubsub.Message, fromConnectedPeer core.PeerID) error {
	return nil
}

func generateHostStub() *mock.ConnectableHostStub {
	return &mock.ConnectableHostStub{
		SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {},
	}
}

func createConnStub(stream network.Stream, id peer.ID, sk libp2pCrypto.PrivKey, remotePeer peer.ID) *mock.ConnStub {
	return &mock.ConnStub{
		GetStreamsCalled: func() []network.Stream {
			if stream == nil {
				return make([]network.Stream, 0)
			}

			return []network.Stream{stream}
		},
		LocalPeerCalled: func() peer.ID {
			return id
		},
		LocalPrivateKeyCalled: func() libp2pCrypto.PrivKey {
			return sk
		},
		RemotePeerCalled: func() peer.ID {
			return remotePeer
		},
	}
}

func createLibP2PCredentialsDirectSender() (peer.ID, libp2pCrypto.PrivKey) {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	sk := (*libp2pCrypto.Secp256k1PrivateKey)(prvKey)
	id, _ := peer.IDFromPublicKey(sk.GetPublic())

	return id, sk
}

func TestNewDirectSender(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	t.Run("nil context", func(t *testing.T) {
		t.Parallel()

		var ctx context.Context = nil
		ds, err := libp2p.NewDirectSender(
			ctx,
			&mock.ConnectableHostStub{},
			blankMessageHandler,
			&mock.P2PSignerStub{},
		)

		assert.True(t, check.IfNil(ds))
		assert.Equal(t, p2p.ErrNilContext, err)
	})
	t.Run("nil host", func(t *testing.T) {
		t.Parallel()

		ds, err := libp2p.NewDirectSender(
			context.Background(),
			nil,
			blankMessageHandler,
			&mock.P2PSignerStub{},
		)

		assert.True(t, check.IfNil(ds))
		assert.Equal(t, p2p.ErrNilHost, err)
	})
	t.Run("nil message handler", func(t *testing.T) {
		t.Parallel()

		ds, err := libp2p.NewDirectSender(
			context.Background(),
			generateHostStub(),
			nil,
			&mock.P2PSignerStub{},
		)

		assert.True(t, check.IfNil(ds))
		assert.Equal(t, p2p.ErrNilDirectSendMessageHandler, err)
	})
	t.Run("nil signer", func(t *testing.T) {
		t.Parallel()

		ds, err := libp2p.NewDirectSender(
			context.Background(),
			generateHostStub(),
			blankMessageHandler,
			nil,
		)

		assert.True(t, check.IfNil(ds))
		assert.Equal(t, p2p.ErrNilP2PSigner, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ds, err := libp2p.NewDirectSender(
			context.Background(),
			generateHostStub(),
			blankMessageHandler,
			&mock.P2PSignerStub{},
		)

		assert.False(t, check.IfNil(ds))
		assert.Nil(t, err)
	})
}

func TestNewDirectSender_OkValsShouldCallSetStreamHandlerWithCorrectValues(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	var pidCalled protocol.ID
	var handlerCalled network.StreamHandler

	hs := &mock.ConnectableHostStub{
		SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {
			pidCalled = pid
			handlerCalled = handler
		},
	}

	_, _ = libp2p.NewDirectSender(
		context.Background(),
		hs,
		blankMessageHandler,
		&mock.P2PSignerStub{},
	)

	assert.NotNil(t, handlerCalled)
	assert.Equal(t, libp2p.DirectSendID, pidCalled)
}

// ------- ProcessReceivedDirectMessage

func TestDirectSender_ProcessReceivedDirectMessageNilMessageShouldErr(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		blankMessageHandler,
		&mock.P2PSignerStub{},
	)

	err := ds.ProcessReceivedDirectMessage(nil, "peer id")

	assert.Equal(t, p2p.ErrNilMessage, err)
}

func TestDirectSender_ProcessReceivedDirectMessageNilTopicIdsShouldErr(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		blankMessageHandler,
		&mock.P2PSignerStub{},
	)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	msg.Topic = nil

	err := ds.ProcessReceivedDirectMessage(msg, id)

	assert.Equal(t, p2p.ErrNilTopic, err)
}

func TestDirectSender_ProcessReceivedDirectMessageKeyFieldIsNotNilShouldErr(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		blankMessageHandler,
		&mock.P2PSignerStub{},
	)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	topic := "topic"
	msg.Topic = &topic

	t.Run("Key contains a non-empty byte slice", func(t *testing.T) {
		msg.Key = []byte("random key")

		err := ds.ProcessReceivedDirectMessage(msg, id)
		assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for Key field as the node accepts only nil on this field"))
	})
	t.Run("Key contains an empty byte slice", func(t *testing.T) {
		msg.Key = make([]byte, 0)

		err := ds.ProcessReceivedDirectMessage(msg, id)
		assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "for Key field as the node accepts only nil on this field"))
	})
}

func TestDirectSender_ProcessReceivedDirectMessageAbnormalSeqNoFieldShouldErr(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		blankMessageHandler,
		&mock.P2PSignerStub{},
	)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = bytes.Repeat([]byte{0x00}, libp2p.SequenceNumberSize+1)
	msg.From = []byte(id)
	topic := "topic"
	msg.Topic = &topic

	err := ds.ProcessReceivedDirectMessage(msg, id)
	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
	assert.True(t, strings.Contains(err.Error(), "for SeqNo field as the node accepts only a maximum"))
}

func TestDirectSender_ProcessReceivedDirectMessageAlreadySeenMsgShouldErr(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		blankMessageHandler,
		&mock.P2PSignerStub{},
	)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	topic := "topic"
	msg.Topic = &topic

	msgId := string(msg.GetFrom()) + string(msg.GetSeqno())
	ds.SeenMessages().Add(msgId)

	err := ds.ProcessReceivedDirectMessage(msg, id)

	assert.Equal(t, p2p.ErrAlreadySeenMessage, err)
}

func TestDirectSender_ProcessReceivedDirectMessageShouldWork(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		blankMessageHandler,
		&mock.P2PSignerStub{},
	)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pb.Message{}
	msg.Data = []byte("data")
	msg.From = []byte(id)
	topic := "topic"
	msg.Topic = &topic

	t.Run("Seqno contains bytes", func(t *testing.T) {
		msg.Seqno = []byte("111")
		err := ds.ProcessReceivedDirectMessage(msg, id)
		assert.Nil(t, err)
	})
	t.Run("empty Seqno", func(t *testing.T) {
		msg.Seqno = make([]byte, 0)
		err := ds.ProcessReceivedDirectMessage(msg, id)
		assert.Nil(t, err)
	})
	t.Run("max Seqno", func(t *testing.T) {
		msg.Seqno = bytes.Repeat([]byte{0xFF}, libp2p.SequenceNumberSize)
		err := ds.ProcessReceivedDirectMessage(msg, id)
		assert.Nil(t, err)
	})
	t.Run("min Seqno", func(t *testing.T) {
		msg.Seqno = bytes.Repeat([]byte{0x00}, libp2p.SequenceNumberSize)
		err := ds.ProcessReceivedDirectMessage(msg, id)
		assert.Nil(t, err)
	})
}

func TestDirectSender_ProcessReceivedDirectMessageShouldCallMessageHandler(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	wasCalled := false

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		func(msg *pubsub.Message, fromConnectedPeer core.PeerID) error {
			wasCalled = true
			return nil
		},
		&mock.P2PSignerStub{},
	)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	topic := "topic"
	msg.Topic = &topic

	_ = ds.ProcessReceivedDirectMessage(msg, id)

	assert.True(t, wasCalled)
}

func TestDirectSender_ProcessReceivedDirectMessageShouldReturnHandlersError(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	checkErr := errors.New("checking error")

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		func(msg *pubsub.Message, fromConnectedPeer core.PeerID) error {
			return checkErr
		},
		&mock.P2PSignerStub{},
	)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	topic := "topic"
	msg.Topic = &topic

	err := ds.ProcessReceivedDirectMessage(msg, id)

	assert.Equal(t, checkErr, err)
}

// ------- SendDirectToConnectedPeer

func TestDirectSender_SendDirectToConnectedPeerBufferToLargeShouldErr(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	netw := &mock.NetworkStub{}

	id, sk := createLibP2PCredentialsDirectSender()
	remotePeer := peer.ID("remote peer")

	stream := mock.NewStreamMock()
	_ = stream.SetProtocol(libp2p.DirectSendID)

	cs := createConnStub(stream, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []network.Conn {
		return []network.Conn{cs}
	}

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		&mock.ConnectableHostStub{
			SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {},
			NetworkCalled: func() network.Network {
				return netw
			},
		},
		blankMessageHandler,
		&mock.P2PSignerStub{},
	)

	messageTooLarge := bytes.Repeat([]byte{65}, libp2p.MaxSendBuffSize)

	err := ds.Send("topic", messageTooLarge, core.PeerID(cs.RemotePeer()))

	assert.True(t, errors.Is(err, p2p.ErrMessageTooLarge))
}

func TestDirectSender_SendDirectToConnectedPeerNotConnectedPeerShouldErr(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	netw := &mock.NetworkStub{
		ConnsToPeerCalled: func(p peer.ID) []network.Conn {
			return make([]network.Conn, 0)
		},
	}

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		&mock.ConnectableHostStub{
			SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {},
			NetworkCalled: func() network.Network {
				return netw
			},
		},
		blankMessageHandler,
		&mock.P2PSignerStub{},
	)

	err := ds.Send("topic", []byte("data"), "not connected peer")

	assert.Equal(t, p2p.ErrPeerNotDirectlyConnected, err)
}

func TestDirectSender_SendDirectToConnectedPeerNewStreamErrorsShouldErr(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	netw := &mock.NetworkStub{}

	hs := &mock.ConnectableHostStub{
		SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {},
		NetworkCalled: func() network.Network {
			return netw
		},
	}

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		hs,
		blankMessageHandler,
		&mock.P2PSignerStub{},
	)

	id, sk := createLibP2PCredentialsDirectSender()
	remotePeer := peer.ID("remote peer")
	errNewStream := errors.New("new stream error")

	cs := createConnStub(nil, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []network.Conn {
		return []network.Conn{cs}
	}

	hs.NewStreamCalled = func(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error) {
		return nil, errNewStream
	}

	data := []byte("data")
	topic := "topic"
	err := ds.Send(topic, data, core.PeerID(cs.RemotePeer()))

	assert.Equal(t, errNewStream, err)
}

func TestDirectSender_SendDirectToConnectedPeerSignFails(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	netw := &mock.NetworkStub{}

	expectedErr := errors.New("expected error")
	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		&mock.ConnectableHostStub{
			SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {},
			NetworkCalled: func() network.Network {
				return netw
			},
		},
		blankMessageHandler,
		&mock.P2PSignerStub{
			SignCalled: func(payload []byte) ([]byte, error) {
				return nil, expectedErr
			},
		},
	)

	id, sk := createLibP2PCredentialsDirectSender()
	remotePeer := peer.ID("remote peer")

	stream := mock.NewStreamMock()
	_ = stream.SetProtocol(libp2p.DirectSendID)

	cs := createConnStub(stream, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []network.Conn {
		return []network.Conn{cs}
	}

	data := []byte("data")
	topic := "topic"
	err := ds.Send(topic, data, core.PeerID(cs.RemotePeer()))

	assert.Equal(t, expectedErr, err)
}

func TestDirectSender_SendDirectToConnectedPeerExistingStreamShouldSendToStream(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	netw := &mock.NetworkStub{}

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		&mock.ConnectableHostStub{
			SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {},
			NetworkCalled: func() network.Network {
				return netw
			},
		},
		blankMessageHandler,
		&mock.P2PSignerStub{},
	)

	id, sk := createLibP2PCredentialsDirectSender()
	remotePeer := peer.ID("remote peer")

	stream := mock.NewStreamMock()
	err := stream.SetProtocol(libp2p.DirectSendID)
	assert.Nil(t, err)

	cs := createConnStub(stream, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []network.Conn {
		return []network.Conn{cs}
	}

	receivedMsg := &pb.Message{}
	chanDone := make(chan bool)

	go func(s network.Stream) {
		reader := ggio.NewDelimitedReader(s, testMaxSize)
		for {
			errRead := reader.ReadMsg(receivedMsg)
			if errRead != nil {
				fmt.Println(errRead.Error())
				return
			}

			chanDone <- true
		}
	}(stream)

	data := []byte("data")
	topic := "topic"
	err = ds.Send(topic, data, core.PeerID(cs.RemotePeer()))
	assert.Nil(t, err)

	select {
	case <-chanDone:
	case <-time.After(timeout):
		assert.Fail(t, "timeout getting data from stream")
		return
	}

	assert.Nil(t, err)
	assert.Equal(t, data, receivedMsg.Data)
	assert.Equal(t, topic, *receivedMsg.Topic)
}

func TestDirectSender_SendDirectToConnectedPeerNewStreamShouldSendToStream(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	netw := &mock.NetworkStub{}

	hs := &mock.ConnectableHostStub{
		SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {},
		NetworkCalled: func() network.Network {
			return netw
		},
	}

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		hs,
		blankMessageHandler,
		&mock.P2PSignerStub{},
	)

	id, sk := createLibP2PCredentialsDirectSender()
	remotePeer := peer.ID("remote peer")

	stream := mock.NewStreamMock()
	_ = stream.SetProtocol(libp2p.DirectSendID)

	cs := createConnStub(stream, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []network.Conn {
		return []network.Conn{cs}
	}

	hs.NewStreamCalled = func(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error) {
		if p == remotePeer && pids[0] == libp2p.DirectSendID {
			return stream, nil
		}
		return nil, errors.New("wrong parameters")
	}

	receivedMsg := &pb.Message{}
	chanDone := make(chan bool)

	go func(s network.Stream) {
		reader := ggio.NewDelimitedReader(s, testMaxSize)
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
	err := ds.Send(topic, data, core.PeerID(cs.RemotePeer()))

	select {
	case <-chanDone:
	case <-time.After(timeout):
		assert.Fail(t, "timeout getting data from stream")
		return
	}

	assert.Nil(t, err)
	assert.Equal(t, data, receivedMsg.Data)
	assert.Equal(t, topic, *receivedMsg.Topic)
}

// ------- received messages tests

func TestDirectSender_ReceivedSentMessageShouldCallMessageHandlerTestFullCycle(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	var streamHandler network.StreamHandler
	netw := &mock.NetworkStub{}

	hs := &mock.ConnectableHostStub{
		SetStreamHandlerCalled: func(pid protocol.ID, handler network.StreamHandler) {
			streamHandler = handler
		},
		NetworkCalled: func() network.Network {
			return netw
		},
	}

	var receivedMsg *pubsub.Message
	chanDone := make(chan bool)

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		hs,
		func(msg *pubsub.Message, fromConnectedPeer core.PeerID) error {
			receivedMsg = msg
			chanDone <- true
			return nil
		},
		&mock.P2PSignerStub{},
	)

	id, sk := createLibP2PCredentialsDirectSender()
	remotePeer := peer.ID("remote peer")

	stream := mock.NewStreamMock()
	stream.SetConn(
		&mock.ConnStub{
			RemotePeerCalled: func() peer.ID {
				return remotePeer
			},
		})
	_ = stream.SetProtocol(libp2p.DirectSendID)

	streamHandler(stream)

	cs := createConnStub(stream, id, sk, remotePeer)

	netw.ConnsToPeerCalled = func(p peer.ID) []network.Conn {
		return []network.Conn{cs}
	}
	cs.LocalPeerCalled = func() peer.ID {
		return cs.RemotePeer()
	}

	data := []byte("data")
	topic := "topic"
	_ = ds.Send(topic, data, core.PeerID(cs.RemotePeer()))

	select {
	case <-chanDone:
	case <-time.After(timeout):
		assert.Fail(t, "timeout")
		return
	}

	assert.NotNil(t, receivedMsg)
	assert.Equal(t, data, receivedMsg.Data)
	assert.Equal(t, topic, *receivedMsg.Topic)
}

func TestDirectSender_ProcessReceivedDirectMessageFromMismatchesFromConnectedPeerShouldErr(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		blankMessageHandler,
		&mock.P2PSignerStub{},
	)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	topic := "topic"
	msg.Topic = &topic

	err := ds.ProcessReceivedDirectMessage(msg, "not the same peer id")

	assert.True(t, errors.Is(err, p2p.ErrInvalidValue))
}

func TestDirectSender_ProcessReceivedDirectMessageSignatureFails(t *testing.T) {
	t.Skip("test")
	t.Parallel()

	verifyCalled := false
	expectedErr := fmt.Errorf("expected error")
	ds, _ := libp2p.NewDirectSender(
		context.Background(),
		generateHostStub(),
		blankMessageHandler,
		&mock.P2PSignerStub{
			VerifyCalled: func(payload []byte, pid core.PeerID, signature []byte) error {
				verifyCalled = true
				return expectedErr
			},
		},
	)

	id, _ := createLibP2PCredentialsDirectSender()

	msg := &pb.Message{}
	msg.Data = []byte("data")
	msg.Seqno = []byte("111")
	msg.From = []byte(id)
	msg.Signature = []byte("signature")
	topic := "topic"
	msg.Topic = &topic

	err := ds.ProcessReceivedDirectMessage(msg, peer.ID(msg.From))

	assert.True(t, errors.Is(err, expectedErr))
	assert.True(t, verifyCalled)
}
