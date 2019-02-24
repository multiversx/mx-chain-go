package peerDiscovery

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/rand"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/loadBalancer"
	"github.com/btcsuite/btcd/btcec"
	crypto2 "github.com/libp2p/go-libp2p-crypto"
)

type TestInitializer struct {
}

func (ti *TestInitializer) CreateMessenger(ctx context.Context,
	port int,
	peerDiscoveryType p2p.PeerDiscoveryType) p2p.Messenger {

	r := rand.New(rand.NewSource(int64(port)))
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*crypto2.Secp256k1PrivateKey)(prvKey)

	libP2PMes, err := libp2p.NewNetworkMessenger(
		ctx,
		port,
		sk,
		nil,
		loadBalancer.NewOutgoingPipeLoadBalancer(),
		peerDiscoveryType)

	if err != nil {
		fmt.Println(err.Error())
	}

	return libP2PMes
}
