package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

type PrivateKeyStub struct {
	ToByteArrayHandler    func() ([]byte, error)
	SignHandler           func(message []byte, signer crypto.SingleSigner) ([]byte, error)
	GeneratePublicHandler func() crypto.PublicKey
	SuiteHandler          func() crypto.Suite
	ScalarHandler         func() crypto.Scalar
}

func (sk PrivateKeyStub) ToByteArray() ([]byte, error) {
	return sk.ToByteArrayHandler()
}

func (sk PrivateKeyStub) Sign(message []byte, signer crypto.SingleSigner) ([]byte, error) {
	return sk.SignHandler(message, signer)
}

func (sk PrivateKeyStub) GeneratePublic() crypto.PublicKey {
	return sk.GeneratePublicHandler()
}

func (sk PrivateKeyStub) Suite() crypto.Suite {
	return sk.SuiteHandler()
}

func (sk PrivateKeyStub) Scalar() crypto.Scalar {
	return sk.ScalarHandler()
}
