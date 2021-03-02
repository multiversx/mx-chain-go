package libp2p_test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/data"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsubpb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getRandomID() []byte {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	sk := (*libp2pCrypto.Secp256k1PrivateKey)(prvKey)
	id, _ := peer.IDFromPublicKey(sk.GetPublic())

	return []byte(id)
}

func TestMessage_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	pMes := &pubsub.Message{}
	m, err := libp2p.NewMessage(pMes, nil)

	assert.True(t, check.IfNil(m))
	assert.True(t, errors.Is(err, p2p.ErrNilMarshalizer))
}

func TestMessage_ShouldErrBecauseOfFromField(t *testing.T) {
	t.Parallel()

	from := []byte("dummy from")
	marshalizer := &testscommon.ProtoMarshalizerMock{}

	topicMessage := &data.TopicMessage{
		Version:   libp2p.CurrentTopicMessageVersion,
		Timestamp: time.Now().Unix(),
		Payload:   []byte("data"),
	}
	buff, _ := marshalizer.Marshal(topicMessage)
	topic := "topic"
	mes := &pubsubpb.Message{
		From:  from,
		Data:  buff,
		Topic: &topic,
	}
	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes, marshalizer)

	assert.True(t, check.IfNil(m))
	assert.NotNil(t, err)
}

func TestMessage_ShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.ProtoMarshalizerMock{}
	topicMessage := &data.TopicMessage{
		Version:   libp2p.CurrentTopicMessageVersion,
		Timestamp: time.Now().Unix(),
		Payload:   []byte("data"),
	}
	buff, _ := marshalizer.Marshal(topicMessage)
	topic := "topic"
	mes := &pubsubpb.Message{
		From:  getRandomID(),
		Data:  buff,
		Topic: &topic,
	}

	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes, marshalizer)

	require.Nil(t, err)
	assert.False(t, check.IfNil(m))
}

func TestMessage_From(t *testing.T) {
	t.Parallel()

	from := getRandomID()
	marshalizer := &testscommon.ProtoMarshalizerMock{}
	topicMessage := &data.TopicMessage{
		Version:   libp2p.CurrentTopicMessageVersion,
		Timestamp: time.Now().Unix(),
		Payload:   []byte("data"),
	}
	buff, _ := marshalizer.Marshal(topicMessage)
	topic := "topic"
	mes := &pubsubpb.Message{
		From:  from,
		Data:  buff,
		Topic: &topic,
	}
	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes, marshalizer)

	require.Nil(t, err)
	assert.Equal(t, m.From(), from)
}

func TestMessage_Peer(t *testing.T) {
	t.Parallel()

	id := getRandomID()
	marshalizer := &testscommon.ProtoMarshalizerMock{}

	topicMessage := &data.TopicMessage{
		Version:   libp2p.CurrentTopicMessageVersion,
		Timestamp: time.Now().Unix(),
		Payload:   []byte("data"),
	}
	buff, _ := marshalizer.Marshal(topicMessage)
	topic := "topic"
	mes := &pubsubpb.Message{
		From:  id,
		Data:  buff,
		Topic: &topic,
	}
	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes, marshalizer)

	require.Nil(t, err)
	assert.Equal(t, core.PeerID(id), m.Peer())
}

func TestMessage_WrongVersionShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.ProtoMarshalizerMock{}

	topicMessage := &data.TopicMessage{
		Version:   libp2p.CurrentTopicMessageVersion + 1,
		Timestamp: time.Now().Unix(),
		Payload:   []byte("data"),
	}
	buff, _ := marshalizer.Marshal(topicMessage)
	topic := "topic"
	mes := &pubsubpb.Message{
		From:  getRandomID(),
		Data:  buff,
		Topic: &topic,
	}

	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes, marshalizer)

	assert.True(t, check.IfNil(m))
	assert.True(t, errors.Is(err, p2p.ErrUnsupportedMessageVersion))
}

func TestMessage_PopulatedPkFieldShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.ProtoMarshalizerMock{}

	topicMessage := &data.TopicMessage{
		Version:   libp2p.CurrentTopicMessageVersion,
		Timestamp: time.Now().Unix(),
		Payload:   []byte("data"),
		Pk:        []byte("p"),
	}
	buff, _ := marshalizer.Marshal(topicMessage)
	topic := "topic"
	mes := &pubsubpb.Message{
		From:  getRandomID(),
		Data:  buff,
		Topic: &topic,
	}

	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes, marshalizer)

	assert.True(t, check.IfNil(m))
	assert.True(t, errors.Is(err, p2p.ErrUnsupportedFields))
}

func TestMessage_PopulatedSigFieldShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.ProtoMarshalizerMock{}

	topicMessage := &data.TopicMessage{
		Version:        libp2p.CurrentTopicMessageVersion,
		Timestamp:      time.Now().Unix(),
		Payload:        []byte("data"),
		SignatureOnPid: []byte("s"),
	}
	buff, _ := marshalizer.Marshal(topicMessage)
	topic := "topic"
	mes := &pubsubpb.Message{
		From:  getRandomID(),
		Data:  buff,
		Topic: &topic,
	}

	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes, marshalizer)

	assert.True(t, check.IfNil(m))
	assert.True(t, errors.Is(err, p2p.ErrUnsupportedFields))
}

func TestMessage_NilTopic(t *testing.T) {
	t.Parallel()

	id := getRandomID()
	marshalizer := &testscommon.ProtoMarshalizerMock{}

	topicMessage := &data.TopicMessage{
		Version:   libp2p.CurrentTopicMessageVersion,
		Timestamp: time.Now().Unix(),
		Payload:   []byte("data"),
	}
	buff, _ := marshalizer.Marshal(topicMessage)
	mes := &pubsubpb.Message{
		From:  id,
		Data:  buff,
		Topic: nil,
	}
	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes, marshalizer)

	assert.Equal(t, p2p.ErrNilTopic, err)
	assert.True(t, check.IfNil(m))
}

func TestMessage_NilMessage(t *testing.T) {
	t.Parallel()

	marshalizer := &testscommon.ProtoMarshalizerMock{}

	m, err := libp2p.NewMessage(nil, marshalizer)

	assert.Equal(t, p2p.ErrNilMessage, err)
	assert.True(t, check.IfNil(m))
}
