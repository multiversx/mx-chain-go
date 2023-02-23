package cryptoMocks

import (
	"github.com/multiversx/mx-chain-crypto-go"
)

// PrivateKeyStub provides stubs for a PrivateKey implementation
type PrivateKeyStub struct {
	ToByteArrayStub    func() ([]byte, error)
	GeneratePublicStub func() crypto.PublicKey
	ScalarStub         func() crypto.Scalar
	SuiteStub          func() crypto.Suite
}

// PublicKeyStub provides stubs for a PublicKey implementation
type PublicKeyStub struct {
	ToByteArrayStub func() ([]byte, error)
	PointStub       func() crypto.Point
	SuiteStub       func() crypto.Suite
}

// KeyGenStub mocks a key generation implementation
type KeyGenStub struct {
	GeneratePairStub            func() (crypto.PrivateKey, crypto.PublicKey)
	PrivateKeyFromByteArrayStub func(b []byte) (crypto.PrivateKey, error)
	PublicKeyFromByteArrayStub  func(b []byte) (crypto.PublicKey, error)
	CheckPublicKeyValidStub     func(b []byte) error
	SuiteStub                   func() crypto.Suite
}

// ToByteArray returns the byte array representation of the private key
func (privKey *PrivateKeyStub) ToByteArray() ([]byte, error) {
	if privKey.ToByteArrayStub != nil {
		return privKey.ToByteArrayStub()
	}
	return []byte("private key"), nil
}

// GeneratePublic builds a public key for the current private key
func (privKey *PrivateKeyStub) GeneratePublic() crypto.PublicKey {
	if privKey.GeneratePublicStub != nil {
		return privKey.GeneratePublicStub()
	}
	return &PublicKeyStub{}
}

// Suite returns the Suite (curve data) used for this private key
func (privKey *PrivateKeyStub) Suite() crypto.Suite {
	if privKey.SuiteStub != nil {
		return privKey.SuiteStub()
	}
	return &SuiteMock{}
}

// Scalar returns the Scalar corresponding to this Private Key
func (privKey *PrivateKeyStub) Scalar() crypto.Scalar {
	if privKey.ScalarStub != nil {
		return privKey.ScalarStub()
	}
	return &ScalarMock{}
}

// IsInterfaceNil returns true if there is no value under the interface
func (privKey *PrivateKeyStub) IsInterfaceNil() bool {
	return privKey == nil
}

// ToByteArray returns the byte array representation of the public key
func (pubKey *PublicKeyStub) ToByteArray() ([]byte, error) {
	if pubKey.ToByteArrayStub != nil {
		return pubKey.ToByteArrayStub()
	}
	return []byte("public key"), nil
}

// Suite returns the Suite (curve data) used for this private key
func (pubKey *PublicKeyStub) Suite() crypto.Suite {
	if pubKey.SuiteStub != nil {
		return pubKey.SuiteStub()
	}
	return &SuiteMock{}
}

// Point returns the Point corresponding to this Public Key
func (pubKey *PublicKeyStub) Point() crypto.Point {
	if pubKey.PointStub != nil {
		return pubKey.PointStub()
	}
	return &PointMock{}
}

// IsInterfaceNil returns true if there is no value under the interface
func (pubKey *PublicKeyStub) IsInterfaceNil() bool {
	return pubKey == nil
}

// GeneratePair generates a pair of private and public keys
func (keyGen *KeyGenStub) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	if keyGen.GeneratePairStub != nil {
		return keyGen.GeneratePairStub()
	}
	return &PrivateKeyStub{}, &PublicKeyStub{}
}

// PrivateKeyFromByteArray generates the private key from its byte array representation
func (keyGen *KeyGenStub) PrivateKeyFromByteArray(b []byte) (crypto.PrivateKey, error) {
	if keyGen.PrivateKeyFromByteArrayStub != nil {
		return keyGen.PrivateKeyFromByteArrayStub(b)
	}
	return &PrivateKeyStub{}, nil
}

// PublicKeyFromByteArray generates a public key from its byte array representation
func (keyGen *KeyGenStub) PublicKeyFromByteArray(b []byte) (crypto.PublicKey, error) {
	if keyGen.PublicKeyFromByteArrayStub != nil {
		return keyGen.PublicKeyFromByteArrayStub(b)
	}
	return &PublicKeyStub{}, nil
}

// CheckPublicKeyValid verifies the validity of the public key
func (keyGen *KeyGenStub) CheckPublicKeyValid(b []byte) error {
	if keyGen.CheckPublicKeyValidStub != nil {
		return keyGen.CheckPublicKeyValidStub(b)
	}
	return nil
}

// Suite -
func (keyGen *KeyGenStub) Suite() crypto.Suite {
	if keyGen.SuiteStub != nil {
		return keyGen.SuiteStub()
	}
	return &SuiteMock{}
}

// IsInterfaceNil returns true if there is no value under the interface
func (keyGen *KeyGenStub) IsInterfaceNil() bool {
	return keyGen == nil
}
