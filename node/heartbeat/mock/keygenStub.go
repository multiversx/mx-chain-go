package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

type KeyGeneratorStub struct {
	GeneratePairCalled            func() (crypto.PrivateKey, crypto.PublicKey)
	PrivateKeyFromByteArrayCalled func(b []byte) (crypto.PrivateKey, error)
	PublicKeyFromByteArrayCalled  func(b []byte) (crypto.PublicKey, error)
	SuiteCalled                   func() crypto.Suite
}

func (ks *KeyGeneratorStub) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	return ks.GeneratePairCalled()
}

func (ks *KeyGeneratorStub) PrivateKeyFromByteArray(b []byte) (crypto.PrivateKey, error) {
	return ks.PrivateKeyFromByteArrayCalled(b)
}

func (ks *KeyGeneratorStub) PublicKeyFromByteArray(b []byte) (crypto.PublicKey, error) {
	return ks.PublicKeyFromByteArrayCalled(b)
}

func (ks *KeyGeneratorStub) Suite() crypto.Suite {
	return ks.SuiteCalled()
}
