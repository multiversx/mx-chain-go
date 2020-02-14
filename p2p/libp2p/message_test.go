package libp2p_test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/btcsuite/btcd/btcec"
	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pubsubpb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/stretchr/testify/assert"
)

func TestMessage_ShouldErrBecauseOfFromField(t *testing.T) {
	from := []byte("dummy from")

	mes := &pubsubpb.Message{
		From: from,
	}

	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes)
	assert.Nil(t, m)
	assert.NotNil(t, err)
}

func TestMessage_ShouldWork(t *testing.T) {
	data := []byte("data")

	mes := &pubsubpb.Message{
		From: getRandomID(),
		Data: data,
	}

	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(m))
}

func TestMessage_Data(t *testing.T) {
	data := []byte("data")

	mes := &pubsubpb.Message{
		From: getRandomID(),
		Data: data,
	}
	pMes := &pubsub.Message{Message: mes}
	m, _ := libp2p.NewMessage(pMes)

	assert.Equal(t, m.Data(), data)
}

func TestMessage_From(t *testing.T) {
	from := getRandomID()

	mes := &pubsubpb.Message{
		From: from,
	}
	pMes := &pubsub.Message{Message: mes}
	m, err := libp2p.NewMessage(pMes)
	assert.Nil(t, err)

	assert.Equal(t, m.From(), from)
}

func TestMessage_Key(t *testing.T) {
	key := []byte("key")

	mes := &pubsubpb.Message{
		From: getRandomID(),
		Key:  key,
	}
	pMes := &pubsub.Message{Message: mes}
	m, _ := libp2p.NewMessage(pMes)

	assert.Equal(t, m.Key(), key)
}

func TestMessage_SeqNo(t *testing.T) {
	seqNo := []byte("seqNo")

	mes := &pubsubpb.Message{
		From:  getRandomID(),
		Seqno: seqNo,
	}
	pMes := &pubsub.Message{Message: mes}
	m, _ := libp2p.NewMessage(pMes)

	assert.Equal(t, m.SeqNo(), seqNo)
}

func TestMessage_Signature(t *testing.T) {
	sig := []byte("sig")

	mes := &pubsubpb.Message{
		From:      getRandomID(),
		Signature: sig,
	}
	pMes := &pubsub.Message{Message: mes}
	m, _ := libp2p.NewMessage(pMes)

	assert.Equal(t, m.Signature(), sig)
}

func TestMessage_TopicIDs(t *testing.T) {
	topics := []string{"topic1", "topic2"}

	mes := &pubsubpb.Message{
		From:     getRandomID(),
		TopicIDs: topics,
	}
	pMes := &pubsub.Message{Message: mes}
	m, _ := libp2p.NewMessage(pMes)

	assert.Equal(t, m.TopicIDs(), topics)
}

func TestMessage_Peer(t *testing.T) {
	id := getRandomID()

	mes := &pubsubpb.Message{
		From: id,
	}
	pMes := &pubsub.Message{Message: mes}

	m, _ := libp2p.NewMessage(pMes)

	assert.Equal(t, p2p.PeerID(id), m.Peer())
}

func getRandomID() []byte {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	sk := (*libp2pCrypto.Secp256k1PrivateKey)(prvKey)
	id, _ := peer.IDFromPublicKey(sk.GetPublic())

	return []byte(id)
}
