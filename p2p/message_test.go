package p2p_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/stretchr/testify/assert"
)

func TestMessage_Data(t *testing.T) {
	data := []byte("data")

	mes := &pubsub_pb.Message{
		Data: data,
	}
	pMes := &pubsub.Message{Message: mes}
	m := p2p.NewMessage(pMes)

	assert.Equal(t, m.Data(), data)
}

func TestMessage_From(t *testing.T) {
	from := []byte("from")

	mes := &pubsub_pb.Message{
		From: from,
	}
	pMes := &pubsub.Message{Message: mes}
	m := p2p.NewMessage(pMes)

	assert.Equal(t, m.From(), from)
}

func TestMessage_Key(t *testing.T) {
	key := []byte("key")

	mes := &pubsub_pb.Message{
		Key: key,
	}
	pMes := &pubsub.Message{Message: mes}
	m := p2p.NewMessage(pMes)

	assert.Equal(t, m.Key(), key)
}

func TestMessage_SeqNo(t *testing.T) {
	seqNo := []byte("seqNo")

	mes := &pubsub_pb.Message{
		Seqno: seqNo,
	}
	pMes := &pubsub.Message{Message: mes}
	m := p2p.NewMessage(pMes)

	assert.Equal(t, m.SeqNo(), seqNo)
}

func TestMessage_Signature(t *testing.T) {
	sig := []byte("sig")

	mes := &pubsub_pb.Message{
		Signature: sig,
	}
	pMes := &pubsub.Message{Message: mes}
	m := p2p.NewMessage(pMes)

	assert.Equal(t, m.Signature(), sig)
}

func TestMessage_TopicIDs(t *testing.T) {
	topics := []string{"topic1", "topic2"}

	mes := &pubsub_pb.Message{
		TopicIDs: topics,
	}
	pMes := &pubsub.Message{Message: mes}
	m := p2p.NewMessage(pMes)

	assert.Equal(t, m.TopicIDs(), topics)
}
