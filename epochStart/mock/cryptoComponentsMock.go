package mock

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/crypto"
)

// CryptoComponentsMock -
type CryptoComponentsMock struct {
	PubKey    crypto.PublicKey
	BlockSig  crypto.SingleSigner
	TxSig     crypto.SingleSigner
	MultiSig  crypto.MultiSigner
	BlKeyGen  crypto.KeyGenerator
	TxKeyGen  crypto.KeyGenerator
	mutCrypto sync.RWMutex
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

// MultiSigner -
func (ccm *CryptoComponentsMock) MultiSigner() crypto.MultiSigner {
	ccm.mutCrypto.RLock()
	defer ccm.mutCrypto.RUnlock()

	return ccm.MultiSig
}

// SetMultiSigner -
func (ccm *CryptoComponentsMock) SetMultiSigner(m crypto.MultiSigner) error {
	ccm.mutCrypto.Lock()
	ccm.MultiSig = m
	ccm.mutCrypto.Unlock()

	return nil
}

// BlockSignKeyGen -
func (ccm *CryptoComponentsMock) BlockSignKeyGen() crypto.KeyGenerator {
	return ccm.BlKeyGen
}

// TxSignKeyGen -
func (ccm *CryptoComponentsMock) TxSignKeyGen() crypto.KeyGenerator {
	return ccm.TxKeyGen
}

// Clone -
func (ccm *CryptoComponentsMock) Clone() interface{} {
	return &CryptoComponentsMock{
		PubKey:    ccm.PubKey,
		BlockSig:  ccm.BlockSig,
		TxSig:     ccm.TxSig,
		MultiSig:  ccm.MultiSig,
		BlKeyGen:  ccm.BlKeyGen,
		TxKeyGen:  ccm.TxKeyGen,
		mutCrypto: sync.RWMutex{},
	}
}

// IsInterfaceNil -
func (ccm *CryptoComponentsMock) IsInterfaceNil() bool {
	return ccm == nil
}
