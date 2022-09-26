package crypto

import (
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	cryptoCommon "github.com/ElrondNetwork/elrond-go/common/crypto"
)

// GetSkPk -
func (ccf *cryptoComponentsFactory) GetSkPk() ([]byte, []byte, error) {
	return ccf.getSkPk()
}

// CreateSingleSigner -
func (ccf *cryptoComponentsFactory) CreateSingleSigner(importModeNoSigCheck bool) (crypto.SingleSigner, error) {
	return ccf.createSingleSigner(importModeNoSigCheck)
}

// CreateDummyCryptoParams
func (ccf *cryptoComponentsFactory) CreateDummyCryptoParams() *cryptoParams {
	return &cryptoParams{}
}

// CreateCryptoParams -
func (ccf *cryptoComponentsFactory) CreateCryptoParams(blockSignKeyGen crypto.KeyGenerator) (*cryptoParams, error) {
	return ccf.createCryptoParams(blockSignKeyGen)
}

// CreateMultiSignerContainer -
func (ccf *cryptoComponentsFactory) CreateMultiSignerContainer(
	blSignKeyGen crypto.KeyGenerator,
	importModeNoSigCheck bool,
) (cryptoCommon.MultiSignerContainer, error) {
	return ccf.createMultiSignerContainer(blSignKeyGen, importModeNoSigCheck)
}

// GetSuite -
func (ccf *cryptoComponentsFactory) GetSuite() (crypto.Suite, error) {
	return ccf.getSuite()
}
