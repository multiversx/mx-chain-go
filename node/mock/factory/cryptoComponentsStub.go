package factory

import (
	"errors"
	"sync"

	"github.com/ElrondNetwork/elrond-go-crypto"
	cryptoCommon "github.com/ElrondNetwork/elrond-go/common/crypto"
	"github.com/ElrondNetwork/elrond-go/vm"
)

// CryptoComponentsMock -
type CryptoComponentsMock struct {
	PubKey            crypto.PublicKey
	PrivKey           crypto.PrivateKey
	PubKeyString      string
	PrivKeyBytes      []byte
	PubKeyBytes       []byte
	BlockSig          crypto.SingleSigner
	TxSig             crypto.SingleSigner
	MultiSigContainer cryptoCommon.MultiSignerContainer
	PeerSignHandler   crypto.PeerSignatureHandler
	BlKeyGen          crypto.KeyGenerator
	TxKeyGen          crypto.KeyGenerator
	MsgSigVerifier    vm.MessageSignVerifier
	mutMultiSig       sync.RWMutex
}

// Create -
func (ccm *CryptoComponentsMock) Create() error {
	return nil
}

// Close -
func (ccm *CryptoComponentsMock) Close() error {
	return nil
}

// CheckSubcomponents -
func (ccm *CryptoComponentsMock) CheckSubcomponents() error {
	return nil
}

// PublicKey -
func (ccm *CryptoComponentsMock) PublicKey() crypto.PublicKey {
	return ccm.PubKey
}

// PrivateKey -
func (ccm *CryptoComponentsMock) PrivateKey() crypto.PrivateKey {
	return ccm.PrivKey
}

// PublicKeyString -
func (ccm *CryptoComponentsMock) PublicKeyString() string {
	return ccm.PubKeyString
}

// PublicKeyBytes -
func (ccm *CryptoComponentsMock) PublicKeyBytes() []byte {
	return ccm.PubKeyBytes
}

// PrivateKeyBytes -
func (ccm *CryptoComponentsMock) PrivateKeyBytes() []byte {
	return ccm.PrivKeyBytes
}

// BlockSigner -
func (ccm *CryptoComponentsMock) BlockSigner() crypto.SingleSigner {
	return ccm.BlockSig
}

// TxSingleSigner -
func (ccm *CryptoComponentsMock) TxSingleSigner() crypto.SingleSigner {
	return ccm.TxSig
}

// MultiSignerContainer -
func (ccm *CryptoComponentsMock) MultiSignerContainer() cryptoCommon.MultiSignerContainer {
	ccm.mutMultiSig.RLock()
	defer ccm.mutMultiSig.RUnlock()

	return ccm.MultiSigContainer
}

// SetMultiSignerContainer -
func (ccm *CryptoComponentsMock) SetMultiSignerContainer(ms cryptoCommon.MultiSignerContainer) error {
	ccm.mutMultiSig.Lock()
	ccm.MultiSigContainer = ms
	ccm.mutMultiSig.Unlock()

	return nil
}

// GetMultiSigner -
func (ccm *CryptoComponentsMock) GetMultiSigner(epoch uint32) (crypto.MultiSigner, error) {
	ccm.mutMultiSig.RLock()
	defer ccm.mutMultiSig.RUnlock()

	if ccm.MultiSigContainer == nil {
		return nil, errors.New("nil multi sig container")
	}

	return ccm.MultiSigContainer.GetMultiSigner(epoch)
}

// PeerSignatureHandler -
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

// MessageSignVerifier -
func (ccm *CryptoComponentsMock) MessageSignVerifier() vm.MessageSignVerifier {
	return ccm.MsgSigVerifier
}

// Clone -
func (ccm *CryptoComponentsMock) Clone() interface{} {
	return &CryptoComponentsMock{
		PubKey:            ccm.PubKey,
		PrivKey:           ccm.PrivKey,
		PubKeyString:      ccm.PubKeyString,
		PrivKeyBytes:      ccm.PrivKeyBytes,
		PubKeyBytes:       ccm.PubKeyBytes,
		BlockSig:          ccm.BlockSig,
		TxSig:             ccm.TxSig,
		MultiSigContainer: ccm.MultiSigContainer,
		PeerSignHandler:   ccm.PeerSignHandler,
		BlKeyGen:          ccm.BlKeyGen,
		TxKeyGen:          ccm.TxKeyGen,
		MsgSigVerifier:    ccm.MsgSigVerifier,
		mutMultiSig:       sync.RWMutex{},
	}
}

// String -
func (ccm *CryptoComponentsMock) String() string {
	return "CryptoComponentsMock"
}

// IsInterfaceNil -
func (ccm *CryptoComponentsMock) IsInterfaceNil() bool {
	return ccm == nil
}
