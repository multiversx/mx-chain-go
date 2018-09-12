package p2p

import (
	"fmt"
	ci "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	mrand "math/rand"
)

// NOTE: for real network tests, use ZeroLocalTCPAddress so the kernel
// assigns an unused TCP port. otherwise you may get clashes. This
// function remains here so that p2p/net/mock (which does not touch the
// real network) can assign different addresses to peers.
var ZeroLocalTCPAddress ma.Multiaddr

type P2PParams struct {
	ID      peer.ID
	PrivKey ci.PrivKey
	PubKey  ci.PubKey
	Addr    ma.Multiaddr
	Port    int
}

func init() {
	// initialize ZeroLocalTCPAddress
	maddr, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		panic(err)
	}
	ZeroLocalTCPAddress = maddr
}

func (params *P2PParams) GeneratePrivPubKeys(seed int) {
	r := mrand.New(mrand.NewSource(int64(seed)))

	prvKey, pubKey, _ := ci.GenerateKeyPairWithReader(ci.RSA, 2048, r)

	params.PrivKey = prvKey
	params.PubKey = pubKey
}

func (params *P2PParams) GenerateIDFromPubKey() {
	params.ID, _ = peer.IDFromPublicKey(params.PubKey)
}

func NewP2PParams(port int) *P2PParams {
	params := new(P2PParams)

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
