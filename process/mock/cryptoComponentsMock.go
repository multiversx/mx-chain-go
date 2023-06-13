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
	BlockSig          crypto.SingleSigner
	TxSig             crypto.SingleSigner
	MultiSigContainer cryptoCommon.MultiSignerContainer
	PeerSignHandler   crypto.PeerSignatureHandler
	BlKeyGen          crypto.KeyGenerator
	TxKeyGen          crypto.KeyGenerator
	PubKey            crypto.PublicKey
	PrivKey           crypto.PrivateKey
	ManagedPeers      common.ManagedPeersHolder
	mutMultiSig       sync.RWMutex
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
	ccm.mutMultiSig.RLock()
	defer ccm.mutMultiSig.RUnlock()

	if ccm.MultiSigContainer == nil {
		return nil, errors.New("multisigner container is nil")
	}

	return ccm.MultiSigContainer.GetMultiSigner(epoch)
}

// MultiSignerContainer -
func (ccm *CryptoComponentsMock) MultiSignerContainer() cryptoCommon.MultiSignerContainer {
	ccm.mutMultiSig.RLock()
	defer ccm.mutMultiSig.RUnlock()

	return ccm.MultiSigContainer
}

// SetMultiSignerContainer -
func (ccm *CryptoComponentsMock) SetMultiSignerContainer(msc cryptoCommon.MultiSignerContainer) error {
	ccm.mutMultiSig.Lock()
	defer ccm.mutMultiSig.Unlock()

	ccm.MultiSigContainer = msc
	return nil
}

// PeerSignatureHandler returns the peer signature handler
func (ccm *CryptoComponentsMock) PeerSignatureHandler() crypto.PeerSignatureHandler {
	ccm.mutMultiSig.RLock()
	defer ccm.mutMultiSig.RUnlock()

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

// PublicKey -
func (ccm *CryptoComponentsMock) PublicKey() crypto.PublicKey {
	return ccm.PubKey
}

// PrivateKey -
func (ccm *CryptoComponentsMock) PrivateKey() crypto.PrivateKey {
	return ccm.PrivKey
}

// ManagedPeersHolder -
func (ccm *CryptoComponentsMock) ManagedPeersHolder() common.ManagedPeersHolder {
	return ccm.ManagedPeers
}

// Clone -
func (ccm *CryptoComponentsMock) Clone() interface{} {
	return &CryptoComponentsMock{
		BlockSig:          ccm.BlockSig,
		TxSig:             ccm.TxSig,
		MultiSigContainer: ccm.MultiSigContainer,
		PeerSignHandler:   ccm.PeerSignHandler,
		BlKeyGen:          ccm.BlKeyGen,
		TxKeyGen:          ccm.TxKeyGen,
		PubKey:            ccm.PubKey,
		mutMultiSig:       sync.RWMutex{},
	}
}

// IsInterfaceNil -
func (ccm *CryptoComponentsMock) IsInterfaceNil() bool {
	return ccm == nil
}
