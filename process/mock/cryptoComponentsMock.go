package mock

import "github.com/ElrondNetwork/elrond-go/crypto"

// CryptoComponentsMock -
type CryptoComponentsMock struct {
	BlockSig crypto.SingleSigner
	TxSig    crypto.SingleSigner
	MultiSig crypto.MultiSigner
	BlKeyGen crypto.KeyGenerator
	TxKeyGen crypto.KeyGenerator
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
	return ccm.MultiSig
}

// BlockSignKeyGen -
func (ccm *CryptoComponentsMock) BlockSignKeyGen() crypto.KeyGenerator {
	return ccm.BlKeyGen
}

// TxSignKeyGen -
func (ccm *CryptoComponentsMock) TxSignKeyGen() crypto.KeyGenerator {
	return ccm.TxKeyGen
}

// IsInterfaceNil -
func (ccm *CryptoComponentsMock) IsInterfaceNil() bool {
	return ccm == nil
}
