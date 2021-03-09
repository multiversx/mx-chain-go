package mock

import "github.com/ElrondNetwork/elrond-go/crypto"

// PrivateKeyStub -
type PrivateKeyStub struct {
	ToByteArrayCalled    func() ([]byte, error)
	SuiteCalled          func() crypto.Suite
	GeneratePublicCalled func() crypto.PublicKey
	ScalarCalled         func() crypto.Scalar
}

// ToByteArray -
func (pks *PrivateKeyStub) ToByteArray() ([]byte, error) {
	if pks.ToByteArrayCalled != nil {
		return pks.ToByteArrayCalled()
	}

	return make([]byte, 0), nil
}

// Suite -
func (pks *PrivateKeyStub) Suite() crypto.Suite {
	if pks.SuiteCalled != nil {
		return pks.SuiteCalled()
	}

	return nil
}

// IsInterfaceNil -
func (pks *PrivateKeyStub) IsInterfaceNil() bool {
	return pks == nil
}

// GeneratePublic -
func (pks *PrivateKeyStub) GeneratePublic() crypto.PublicKey {
	if pks.GeneratePublicCalled != nil {
		return pks.GeneratePublicCalled()
	}

	return nil
}

// Scalar -
func (pks *PrivateKeyStub) Scalar() crypto.Scalar {
	if pks.ScalarCalled != nil {
		return pks.ScalarCalled()
	}

	return nil
}
