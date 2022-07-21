package crypto

import (
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
)

// GetSkPk -
func (ccf *cryptoComponentsFactory) GetSkPk() ([]byte, []byte, error) {
	return ccf.getSkPk()
}

// CreateSingleSigner -
func (ccf *cryptoComponentsFactory) CreateSingleSigner(importModeNoSigCheck bool) (crypto.SingleSigner, error) {
	return ccf.createSingleSigner(importModeNoSigCheck)
}

// GetMultiSigHasherFromConfig -
func (ccf *cryptoComponentsFactory) GetMultiSigHasherFromConfig() (hashing.Hasher, error) {
	return ccf.getMultiSigHasherFromConfig()
}

// CreateDummyCryptoParams
func (ccf *cryptoComponentsFactory) CreateDummyCryptoParams() *cryptoParams {
	return &cryptoParams{}
}

// CreateCryptoParams -
func (ccf *cryptoComponentsFactory) CreateCryptoParams(blockSignKeyGen crypto.KeyGenerator) (*cryptoParams, error) {
	return ccf.createCryptoParams(blockSignKeyGen)
}

// CreateMultiSigner -
func (ccf *cryptoComponentsFactory) CreateMultiSigner(
	h hashing.Hasher, cp *cryptoParams, blSignKeyGen crypto.KeyGenerator, importModeNoSigCheck bool,
) (crypto.MultiSigner, error) {
	return ccf.createMultiSigner(h, cp, blSignKeyGen, importModeNoSigCheck)
}

// GetSuite -
func (ccf *cryptoComponentsFactory) GetSuite() (crypto.Suite, error) {
	return ccf.getSuite()
}
