package mock

import "github.com/ElrondNetwork/elrond-go-sandbox/crypto"

type SingleSignerStub struct {
	SignCalled   func(private crypto.PrivateKey, msg []byte) ([]byte, error)
	VerifyCalled func(public crypto.PublicKey, msg []byte, sig []byte) error
}

func (sss *SingleSignerStub) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	return sss.SignCalled(private, msg)
}

func (sss *SingleSignerStub) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	return sss.VerifyCalled(public, msg, sig)
}
