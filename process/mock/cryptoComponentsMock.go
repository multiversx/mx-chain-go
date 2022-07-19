package mock

import (
	"errors"
	"sync"

	"github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/process"
)

// CryptoComponentsMock -
type CryptoComponentsMock struct {
	BlockSig          crypto.SingleSigner
	TxSig             crypto.SingleSigner
	MultiSigContainer process.MultiSignerContainer
	PeerSignHandler   crypto.PeerSignatureHandler
	BlKeyGen          crypto.KeyGenerator
	TxKeyGen          crypto.KeyGenerator
	PubKey            crypto.PublicKey
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

// SetMultiSignerContainer -
func (ccm *CryptoComponentsMock) SetMultiSignerContainer(multiSignerContainer process.MultiSignerContainer) error {
	ccm.mutMultiSig.Lock()
	ccm.MultiSigContainer = multiSignerContainer
	ccm.mutMultiSig.Unlock()
	return nil
}

// GetMultiSigner -
func (ccm *CryptoComponentsMock) GetMultiSigner(epoch uint32) (crypto.MultiSigner, error) {
	if ccm.MultiSigContainer == nil {
		return nil, errors.New("nil multi signer")
	}

	return ccm.MultiSigContainer.GetMultiSigner(epoch)
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
