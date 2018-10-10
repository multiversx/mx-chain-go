package p2p

import (
	"crypto/ecdsa"
	"fmt"
	"math/rand"

	"github.com/btcsuite/btcd/btcec"
	cr "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
)

type ConnectParams struct {
	ID      peer.ID
	PrivKey cr.PrivKey
	PubKey  cr.PubKey
	Addr    ma.Multiaddr
	Port    int
}

func (params *ConnectParams) GeneratePrivPubKeys(seed int) {
	r := rand.New(rand.NewSource(int64(seed)))

	prvKey, err := ecdsa.GenerateKey(btcec.S256(), r)

	if err != nil {
		panic(err)
	}

	k := (*cr.Secp256k1PrivateKey)(prvKey)

	params.PrivKey = k
	params.PubKey = k.GetPublic()
}

func (params *ConnectParams) GenerateIDFromPubKey() {
	params.ID, _ = peer.IDFromPublicKey(params.PubKey)
}

func NewConnectParamsFromPort(port int) *ConnectParams {
	params := new(ConnectParams)

	params.Port = port
	params.GeneratePrivPubKeys(port)
	params.GenerateIDFromPubKey()
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

	if err != nil {
		panic(err)
	}

	params.Addr = addr

	return params
}

func NewConnectParams(ipAddr string, port int, privKey cr.PrivKey) *ConnectParams {
	params := new(ConnectParams)

	params.Port = port
	params.PrivKey = privKey
	params.PubKey = privKey.GetPublic()
	params.GenerateIDFromPubKey()
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", ipAddr, port))

	if err != nil {
		panic(err)
	}

	params.Addr = addr

	return params
}
