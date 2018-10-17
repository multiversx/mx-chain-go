package p2p

import (
	"crypto/ecdsa"
	"math/rand"

	"github.com/btcsuite/btcd/btcec"
	cr "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/pkg/errors"
)

// ConnectParams is used to instantiate a Messenger object
// (contains required data by the Messenger struct)
type ConnectParams struct {
	ID      peer.ID
	PrivKey cr.PrivKey
	PubKey  cr.PubKey
	Port    int
}

// GeneratePrivPubKeys will generate a new private key by using the port
// as a seed for the random generation object
// SHOULD BE USED ONLY IN TESTING!!!
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

// GenerateIDFromPubKey will set the params.ID to a hash of the params.PubKey
func (params *ConnectParams) GenerateIDFromPubKey() {
	params.ID, _ = peer.IDFromPublicKey(params.PubKey)
}

// NewConnectParamsFromPort will generate a new ConnectParams object by using the port
// as a seed for the random generation object
// SHOULD BE USED ONLY IN TESTING!!!
func NewConnectParamsFromPort(port int) (*ConnectParams, error) {
	if port < 0 || port > 65535 {
		return nil, errors.New("port outside [0, 65535]")
	}

	params := new(ConnectParams)

	params.Port = port
	params.GeneratePrivPubKeys(port)
	params.GenerateIDFromPubKey()

	return params, nil
}

// NewConnectParams is used to generate a new ConnectParams. This is the proper
// way to initialize the object. The private key provided is used for
// data and channel encryption and can be used for authentication of messages
func NewConnectParams(port int, privKey cr.PrivKey) (*ConnectParams, error) {
	if port < 0 || port > 65535 {
		return nil, errors.New("port outside [0, 65535]")
	}

	params := new(ConnectParams)

	params.Port = port
	params.PrivKey = privKey
	params.PubKey = privKey.GetPublic()
	params.GenerateIDFromPubKey()

	return params, nil
}
