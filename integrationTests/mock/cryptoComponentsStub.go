package mock

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/vm"
)

// CryptoComponentsStub -
type CryptoComponentsStub struct {
	PubKey          crypto.PublicKey
	PrivKey         crypto.PrivateKey
	PubKeyString    string
	PrivKeyBytes    []byte
	PubKeyBytes     []byte
	BlockSig        crypto.SingleSigner
	TxSig           crypto.SingleSigner
	MultiSig        crypto.MultiSigner
	PeerSignHandler crypto.PeerSignatureHandler
	BlKeyGen        crypto.KeyGenerator
	TxKeyGen        crypto.KeyGenerator
	MsgSigVerifier  vm.MessageSignVerifier
	mutMultiSig     sync.RWMutex
}

// Create -
func (ccs *CryptoComponentsStub) Create() error {
	return nil
}

// Close -
func (ccs *CryptoComponentsStub) Close() error {
	return nil
}

// CheckSubcomponents -
func (ccs *CryptoComponentsStub) CheckSubcomponents() error {
	return nil
}

// PublicKey -
func (ccs *CryptoComponentsStub) PublicKey() crypto.PublicKey {
	return ccs.PubKey
}

// PrivateKey -
func (ccs *CryptoComponentsStub) PrivateKey() crypto.PrivateKey {
	return ccs.PrivKey
}

// PublicKeyString -
func (ccs *CryptoComponentsStub) PublicKeyString() string {
	return ccs.PubKeyString
}

// PublicKeyBytes -
func (ccs *CryptoComponentsStub) PublicKeyBytes() []byte {
	return ccs.PubKeyBytes
}

// PrivateKeyBytes -
func (ccs *CryptoComponentsStub) PrivateKeyBytes() []byte {
	return ccs.PrivKeyBytes
}

// BlockSigner -
func (ccs *CryptoComponentsStub) BlockSigner() crypto.SingleSigner {
	return ccs.BlockSig
}

// TxSingleSigner -
func (ccs *CryptoComponentsStub) TxSingleSigner() crypto.SingleSigner {
	return ccs.TxSig
}

// MultiSigner -
func (ccs *CryptoComponentsStub) MultiSigner() crypto.MultiSigner {
	ccs.mutMultiSig.RLock()
	defer ccs.mutMultiSig.RUnlock()

	return ccs.MultiSig
}

// PeerSignatureHandler -
func (ccs *CryptoComponentsStub) PeerSignatureHandler() crypto.PeerSignatureHandler {
	ccs.mutMultiSig.RLock()
	defer ccs.mutMultiSig.RUnlock()

	return ccs.PeerSignHandler
}

// SetMultiSigner -
func (ccs *CryptoComponentsStub) SetMultiSigner(ms crypto.MultiSigner) error {
	ccs.mutMultiSig.Lock()
	ccs.MultiSig = ms
	ccs.mutMultiSig.Unlock()

	return nil
}

// BlockSignKeyGen -
func (ccs *CryptoComponentsStub) BlockSignKeyGen() crypto.KeyGenerator {
	return ccs.BlKeyGen
}

// TxSignKeyGen -
func (ccs *CryptoComponentsStub) TxSignKeyGen() crypto.KeyGenerator {
	return ccs.TxKeyGen
}

// MessageSignVerifier -
func (ccs *CryptoComponentsStub) MessageSignVerifier() vm.MessageSignVerifier {
	return ccs.MsgSigVerifier
}

// Clone -
func (ccs *CryptoComponentsStub) Clone() interface{} {
	return &CryptoComponentsStub{
		PubKey:          ccs.PubKey,
		PrivKey:         ccs.PrivKey,
		PubKeyString:    ccs.PubKeyString,
		PrivKeyBytes:    ccs.PrivKeyBytes,
		PubKeyBytes:     ccs.PubKeyBytes,
		BlockSig:        ccs.BlockSig,
		TxSig:           ccs.TxSig,
		MultiSig:        ccs.MultiSig,
		PeerSignHandler: ccs.PeerSignHandler,
		BlKeyGen:        ccs.BlKeyGen,
		TxKeyGen:        ccs.TxKeyGen,
		MsgSigVerifier:  ccs.MsgSigVerifier,
		mutMultiSig:     sync.RWMutex{},
	}
}

// IsInterfaceNil -
func (ccs *CryptoComponentsStub) IsInterfaceNil() bool {
	return ccs == nil
}
