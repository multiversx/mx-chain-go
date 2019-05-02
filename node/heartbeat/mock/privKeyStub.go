package mock

import "github.com/ElrondNetwork/elrond-go-sandbox/crypto"

type PrivKeyStub struct {
	ToByteArrayCalled    func() ([]byte, error)
	SuiteCalled          func() crypto.Suite
	GeneratePublicCalled func() crypto.PublicKey
	ScalarCalled         func() crypto.Scalar
}

func (pks *PrivKeyStub) ToByteArray() ([]byte, error) {
	return pks.ToByteArrayCalled()
}

func (pks *PrivKeyStub) Suite() crypto.Suite {
	return pks.SuiteCalled()
}

func (pks *PrivKeyStub) GeneratePublic() crypto.PublicKey {
	return pks.GeneratePublicCalled()
}

func (pks *PrivKeyStub) Scalar() crypto.Scalar {
	return pks.ScalarCalled()
}
