package p2p

import (
	"crypto/ecdsa"
	"fmt"
	"math"
	mrand "math/rand"
	"net"

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
	r := mrand.New(mrand.NewSource(int64(seed)))

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

func NewConnectParamsFromPortAndIP(port int, IP net.IP) *ConnectParams {
	params := new(ConnectParams)

	seed := 0

	for i := 0; i < len(IP); i++ {
		seed += int(IP[i]) * int(math.Pow(float64(256), float64(len(IP)-1-i)))
	}

	seed *= port

	params.Port = port
	params.GeneratePrivPubKeys(seed)
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
