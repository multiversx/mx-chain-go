package main

import (
	"fmt"

	libp2pCrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

type p2pConverter struct{}

// NewP2pConverter creates a new instance of p2p converter
func NewP2pConverter() *p2pConverter {
	return &p2pConverter{}
}

// Len return zero
func (p *p2pConverter) Len() int {
	return 0
}

// Decode does nothing
func (p *p2pConverter) Decode(humanReadable string) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

// Encode encodes a byte array representing public key as peer ID string
func (p *p2pConverter) Encode(pkBytes []byte) string {
	pubKey, err := libp2pCrypto.UnmarshalSecp256k1PublicKey(pkBytes)
	if err != nil {
		return ""
	}

	id, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		return ""
	}

	return id.Pretty()
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *p2pConverter) IsInterfaceNil() bool {
	return p == nil
}
