package crypto

import (
	"crypto/ecdsa"

	"github.com/ElrondNetwork/elrond-go-core/core"
	randFactory "github.com/ElrondNetwork/elrond-go/p2p/libp2p/rand/factory"
	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

const emptySeed = ""

type identityGenerator struct {
}

// NewIdentityGenerator creates a new identity generator
func NewIdentityGenerator() *identityGenerator {
	return &identityGenerator{}
}

// CreateRandomP2PIdentity creates a valid random p2p identity to sign messages on the behalf of other identity
func (generator *identityGenerator) CreateRandomP2PIdentity() ([]byte, core.PeerID, error) {
	sk, err := generator.CreateP2PPrivateKey(emptySeed)
	if err != nil {
		return nil, "", err
	}

	skBuff, err := crypto.MarshalPrivateKey(sk)
	if err != nil {
		return nil, "", err
	}

	pid, err := peer.IDFromPublicKey(sk.GetPublic())
	if err != nil {
		return nil, "", err
	}

	return skBuff, core.PeerID(pid), nil
}

// CreateP2PPrivateKey will create a new P2P private key based on the provided seed. If the seed is the empty string
// it will use the crypto's random generator to provide a random one. Otherwise, it will create a deterministic private
// key. This is useful when we want a private key that never changes, such as in the network seeders
func (generator *identityGenerator) CreateP2PPrivateKey(seed string) (*crypto.Secp256k1PrivateKey, error) {
	randReader, err := randFactory.NewRandFactory(seed)
	if err != nil {
		return nil, err
	}

	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), randReader)

	return (*crypto.Secp256k1PrivateKey)(prvKey), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (generator *identityGenerator) IsInterfaceNil() bool {
	return generator == nil
}
