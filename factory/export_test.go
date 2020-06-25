package factory

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/hashing"
)

// CreateStorerTemplatePaths -
func (ccf *coreComponentsFactory) CreateStorerTemplatePaths() (string, string) {
	return ccf.createStorerTemplatePaths()
}

// GetSkPk -
func (ccf *cryptoComponentsFactory) GetSkPk() ([]byte, []byte, error) {
	return ccf.getSkPk()
}

// CreateSingleSigner -
func (ccf *cryptoComponentsFactory) CreateSingleSigner() (crypto.SingleSigner, error) {
	return ccf.createSingleSigner()
}

// GetMultiSigHasherFromConfig -
func (ccf *cryptoComponentsFactory) GetMultiSigHasherFromConfig() (hashing.Hasher, error) {
	return ccf.getMultiSigHasherFromConfig()
}

// CreateCryptoParams -
func (ccf *cryptoComponentsFactory) CreateCryptoParams(blockSignKeyGen crypto.KeyGenerator) (*CryptoParams, error) {
	return ccf.createCryptoParams(blockSignKeyGen)
}

// CreateMultiSigner -
func (ccf *cryptoComponentsFactory) CreateMultiSigner(
	h hashing.Hasher, cp *CryptoParams, blSignKeyGen crypto.KeyGenerator,
) (crypto.MultiSigner, error) {
	return ccf.createMultiSigner(h, cp, blSignKeyGen)
}

// GetSuite -
func (ccf *cryptoComponentsFactory) GetSuite() (crypto.Suite, error) {
	return ccf.getSuite()
}

// SetKeyLoader -
func (ccf *cryptoComponentsFactory) SetKeyLoader(keyLoad func(string, int) ([]byte, string, error)) {
	ccf.keyLoader = keyLoad
}

// SetListenAddress -
func (ncf *networkComponentsFactory) SetListenAddress(address string) {
	ncf.listenAddress = address
}

// CreateTries -
func (scf *stateComponentsFactory) CreateTries() (*TriesComponents, error) {
	return scf.createTries()
}
