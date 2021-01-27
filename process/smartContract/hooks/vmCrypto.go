package hooks

import (
	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"golang.org/x/crypto/ripemd160" //nolint TODO:deprecated
)

// VMCryptoHook is a wrapper used in vm implementation
type VMCryptoHook struct {
}

// NewVMCryptoHook creates a new instance of a vm crypto hook
func NewVMCryptoHook() *VMCryptoHook {
	return &VMCryptoHook{}
}

// Sha256 returns a sha 256 hash of the input string. Should return in hex format.
func (vmch *VMCryptoHook) Sha256(data []byte) ([]byte, error) {
	return sha256.Sha256{}.Compute(string(data)), nil
}

// Keccak256 returns a keccak 256 hash of the input string. Should return in hex format.
func (vmch *VMCryptoHook) Keccak256(data []byte) ([]byte, error) {
	return keccak.Keccak{}.Compute(string(data)), nil
}

// Ripemd160 is a legacy hash and should not be used for new applications
func (vmch *VMCryptoHook) Ripemd160(data []byte) ([]byte, error) {
	hash := ripemd160.New()
	_, err := hash.Write(data)
	if err != nil {
		return nil, err
	}

	result := hash.Sum(nil)
	return result, nil
}

// Ecrecover calculates the corresponding Ethereum address for the public key which created the given signature
// https://ewasm.readthedocs.io/en/mkdocs/system_contracts/
func (vmch *VMCryptoHook) Ecrecover(_ []byte, _ []byte, _ []byte, _ []byte) ([]byte, error) {
	return nil, ErrNotImplemented
}

// IsInterfaceNil returns true if the underlying object is nil
func (vmch *VMCryptoHook) IsInterfaceNil() bool {
	return vmch == nil
}
