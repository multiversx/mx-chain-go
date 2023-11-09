package mock

import (
	"errors"
	"sync"

	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	cryptoCommon "github.com/multiversx/mx-chain-go/common/crypto"
)

// CryptoComponentsMock -
type CryptoComponentsMock struct {
	PrivKey           crypto.PrivateKey
	PubKey            crypto.PublicKey
	BlockSig          crypto.SingleSigner
	TxSig             crypto.SingleSigner
	MultiSigContainer cryptoCommon.MultiSignerContainer
	PeerSignHandler   crypto.PeerSignatureHandler
	BlKeyGen          crypto.KeyGenerator
	TxKeyGen          crypto.KeyGenerator
	ManagedPeers      common.ManagedPeersHolder
	mutCrypto         sync.RWMutex
}

// PrivateKey -
func (ccm *CryptoComponentsMock) PrivateKey() crypto.PrivateKey {
	return ccm.PrivKey
}

// PublicKey -
func (ccm *CryptoComponentsMock) PublicKey() crypto.PublicKey {
	return ccm.PubKey
}

// BlockSigner -
func (ccm *CryptoComponentsMock) BlockSigner() crypto.SingleSigner {
	return ccm.BlockSig
}

// TxSingleSigner -
func (ccm *CryptoComponentsMock) TxSingleSigner() crypto.SingleSigner {
	return ccm.TxSig
}

// GetMultiSigner -
func (ccm *CryptoComponentsMock) GetMultiSigner(epoch uint32) (crypto.MultiSigner, error) {
	ccm.mutCrypto.RLock()
	defer ccm.mutCrypto.RUnlock()

	if ccm.MultiSigContainer == nil {
		return nil, errors.New("multisigner container is nil")
	}

	return ccm.MultiSigContainer.GetMultiSigner(epoch)
}

// MultiSignerContainer -
func (ccm *CryptoComponentsMock) MultiSignerContainer() cryptoCommon.MultiSignerContainer {
	ccm.mutCrypto.RLock()
	defer ccm.mutCrypto.RUnlock()

	return ccm.MultiSigContainer
}

// SetMultiSignerContainer -
func (ccm *CryptoComponentsMock) SetMultiSignerContainer(msc cryptoCommon.MultiSignerContainer) error {
	ccm.mutCrypto.Lock()
	defer ccm.mutCrypto.Unlock()

	ccm.MultiSigContainer = msc
	return nil
}

// PeerSignatureHandler -
func (ccm *CryptoComponentsMock) PeerSignatureHandler() crypto.PeerSignatureHandler {
	return ccm.PeerSignHandler
}

// BlockSignKeyGen -
func (ccm *CryptoComponentsMock) BlockSignKeyGen() crypto.KeyGenerator {
	return ccm.BlKeyGen
}

// TxSignKeyGen -
func (ccm *CryptoComponentsMock) TxSignKeyGen() crypto.KeyGenerator {
	return ccm.TxKeyGen
}

// ManagedPeersHolder -
func (ccm *CryptoComponentsMock) ManagedPeersHolder() common.ManagedPeersHolder {
	return ccm.ManagedPeers
}

// Clone -
func (ccm *CryptoComponentsMock) Clone() interface{} {
	return &CryptoComponentsMock{
		PubKey:            ccm.PubKey,
		BlockSig:          ccm.BlockSig,
		TxSig:             ccm.TxSig,
		MultiSigContainer: ccm.MultiSigContainer,
		PeerSignHandler:   ccm.PeerSignHandler,
		BlKeyGen:          ccm.BlKeyGen,
		TxKeyGen:          ccm.TxKeyGen,
		mutCrypto:         sync.RWMutex{},
	}
}

// IsInterfaceNil -
func (ccm *CryptoComponentsMock) IsInterfaceNil() bool {
	return ccm == nil
}
