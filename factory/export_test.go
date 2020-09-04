package factory

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
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
	h hashing.Hasher, cp *cryptoParams, blSignKeyGen crypto.KeyGenerator,
) (crypto.MultiSigner, error) {
	return ccf.createMultiSigner(h, cp, blSignKeyGen)
}

// GetSuite -
func (ccf *cryptoComponentsFactory) GetSuite() (crypto.Suite, error) {
	return ccf.getSuite()
}

// GetFactory
func (cc *managedCryptoComponents) GetFactory() *cryptoComponentsFactory {
	return cc.cryptoComponentsFactory
}

// SetListenAddress -
func (ncf *networkComponentsFactory) SetListenAddress(address string) {
	ncf.listenAddress = address
}

// CreateTries -
func (scf *stateComponentsFactory) CreateTries() (state.TriesHolder, map[string]data.StorageManager, error) {
	return scf.createTries()
}
