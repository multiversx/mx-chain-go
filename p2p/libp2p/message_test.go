package libp2p_test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
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
	m := libp2p.NewMessage(pMes)

	assert.Equal(t, m.Data(), data)
}

func TestMessage_From(t *testing.T) {
	from := []byte("from")

	mes := &pubsub_pb.Message{
		From: from,
	}
	pMes := &pubsub.Message{Message: mes}
	m := libp2p.NewMessage(pMes)

	assert.Equal(t, m.From(), from)
}

func TestMessage_Key(t *testing.T) {
	key := []byte("key")

	mes := &pubsub_pb.Message{
		Key: key,
	}
	pMes := &pubsub.Message{Message: mes}
	m := libp2p.NewMessage(pMes)

	assert.Equal(t, m.Key(), key)
}

func TestMessage_SeqNo(t *testing.T) {
	seqNo := []byte("seqNo")

	mes := &pubsub_pb.Message{
		Seqno: seqNo,
	}
	pMes := &pubsub.Message{Message: mes}
	m := libp2p.NewMessage(pMes)

	assert.Equal(t, m.SeqNo(), seqNo)
}

func TestMessage_Signature(t *testing.T) {
	sig := []byte("sig")

	mes := &pubsub_pb.Message{
		Signature: sig,
	}
	pMes := &pubsub.Message{Message: mes}
	m := libp2p.NewMessage(pMes)

	assert.Equal(t, m.Signature(), sig)
}

func TestMessage_TopicIDs(t *testing.T) {
	topics := []string{"topic1", "topic2"}

	mes := &pubsub_pb.Message{
		TopicIDs: topics,
	}
	pMes := &pubsub.Message{Message: mes}
	m := libp2p.NewMessage(pMes)

	assert.Equal(t, m.TopicIDs(), topics)
}

func TestMessage_Peer(t *testing.T) {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	sk := (*libp2pCrypto.Secp256k1PrivateKey)(prvKey)
	id, _ := peer.IDFromPublicKey(sk.GetPublic())

	mes := &pubsub_pb.Message{
		From: []byte(id),
	}
	pMes := &pubsub.Message{Message: mes}

	m := libp2p.NewMessage(pMes)

	assert.Equal(t, p2p.PeerID(id), m.Peer())
}
