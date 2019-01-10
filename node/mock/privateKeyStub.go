package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

type PrivateKeyStub struct {
	ToByteArrayHandler    func() ([]byte, error)
	SignHandler           func(message []byte) ([]byte, error)
	GeneratePublicHandler func() crypto.PublicKey
}

func (sk PrivateKeyStub) ToByteArray() ([]byte, error) {
	return sk.ToByteArrayHandler()
}
func (sk PrivateKeyStub) Sign(message []byte) ([]byte, error) {
	return sk.SignHandler(message)
}
func (sk PrivateKeyStub) GeneratePublic() crypto.PublicKey {
	return sk.GeneratePublicHandler()
}
