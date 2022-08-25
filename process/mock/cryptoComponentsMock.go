package mock

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-crypto"
)

// CryptoComponentsMock -
type CryptoComponentsMock struct {
	BlockSig        crypto.SingleSigner
	TxSig           crypto.SingleSigner
	MultiSig        crypto.MultiSigner
	PeerSignHandler crypto.PeerSignatureHandler
	BlKeyGen        crypto.KeyGenerator
	TxKeyGen        crypto.KeyGenerator
	PubKey          crypto.PublicKey
	mutMultiSig     sync.RWMutex
}

// BlockSigner -
func (ccm *CryptoComponentsMock) BlockSigner() crypto.SingleSigner {
	return ccm.BlockSig
}

// TxSingleSigner -
func (ccm *CryptoComponentsMock) TxSingleSigner() crypto.SingleSigner {
	return ccm.TxSig
}

// MultiSigner -
func (ccm *CryptoComponentsMock) MultiSigner() crypto.MultiSigner {
	ccm.mutMultiSig.RLock()
	defer ccm.mutMultiSig.RUnlock()
	return ccm.MultiSig
}

// SetMultiSigner -
func (ccm *CryptoComponentsMock) SetMultiSigner(multiSigner crypto.MultiSigner) error {
	ccm.mutMultiSig.Lock()
	ccm.MultiSig = multiSigner
	ccm.mutMultiSig.Unlock()
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

// Clone -
func (ccm *CryptoComponentsMock) Clone() interface{} {
	return &CryptoComponentsMock{
		BlockSig:        ccm.BlockSig,
		TxSig:           ccm.TxSig,
		MultiSig:        ccm.MultiSig,
		PeerSignHandler: ccm.PeerSignHandler,
		BlKeyGen:        ccm.BlKeyGen,
		TxKeyGen:        ccm.TxKeyGen,
		PubKey:          ccm.PubKey,
		mutMultiSig:     sync.RWMutex{},
	}
}

// IsInterfaceNil -
func (ccm *CryptoComponentsMock) IsInterfaceNil() bool {
	return ccm == nil
}
