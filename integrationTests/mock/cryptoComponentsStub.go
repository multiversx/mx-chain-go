package mock

import (
	"errors"
	"sync"

	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	cryptoCommon "github.com/multiversx/mx-chain-go/common/crypto"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/vm"
)

// CryptoComponentsStub -
type CryptoComponentsStub struct {
	PubKey                  crypto.PublicKey
	PublicKeyCalled         func() crypto.PublicKey
	PrivKey                 crypto.PrivateKey
	P2pPubKey               crypto.PublicKey
	P2pPrivKey              crypto.PrivateKey
	PubKeyBytes             []byte
	PubKeyString            string
	BlockSig                crypto.SingleSigner
	TxSig                   crypto.SingleSigner
	P2pSig                  crypto.SingleSigner
	MultiSigContainer       cryptoCommon.MultiSignerContainer
	PeerSignHandler         crypto.PeerSignatureHandler
	BlKeyGen                crypto.KeyGenerator
	TxKeyGen                crypto.KeyGenerator
	P2PKeyGen               crypto.KeyGenerator
	MsgSigVerifier          vm.MessageSignVerifier
	ManagedPeersHolderField common.ManagedPeersHolder
	KeysHandlerField        consensus.KeysHandler
	KeysHandlerCalled       func() consensus.KeysHandler
	SigHandler              consensus.SigningHandler
	mutMultiSig             sync.RWMutex
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
	if ccs.PublicKeyCalled != nil {
		return ccs.PublicKeyCalled()
	}
	return ccs.PubKey
}

// PrivateKey -
func (ccs *CryptoComponentsStub) PrivateKey() crypto.PrivateKey {
	return ccs.PrivKey
}

// P2pPrivateKey -
func (ccs *CryptoComponentsStub) P2pPrivateKey() crypto.PrivateKey {
	return ccs.P2pPrivKey
}

// P2pPublicKey -
func (ccs *CryptoComponentsStub) P2pPublicKey() crypto.PublicKey {
	return ccs.P2pPubKey
}

// PublicKeyString -
func (ccs *CryptoComponentsStub) PublicKeyString() string {
	return ccs.PubKeyString
}

// PublicKeyBytes -
func (ccs *CryptoComponentsStub) PublicKeyBytes() []byte {
	return ccs.PubKeyBytes
}

// BlockSigner -
func (ccs *CryptoComponentsStub) BlockSigner() crypto.SingleSigner {
	return ccs.BlockSig
}

// P2pSingleSigner -
func (ccs *CryptoComponentsStub) P2pSingleSigner() crypto.SingleSigner {
	return ccs.P2pSig
}

// TxSingleSigner -
func (ccs *CryptoComponentsStub) TxSingleSigner() crypto.SingleSigner {
	return ccs.TxSig
}

// PeerSignatureHandler -
func (ccs *CryptoComponentsStub) PeerSignatureHandler() crypto.PeerSignatureHandler {
	ccs.mutMultiSig.RLock()
	defer ccs.mutMultiSig.RUnlock()

	return ccs.PeerSignHandler
}

// MultiSignerContainer -
func (ccs *CryptoComponentsStub) MultiSignerContainer() cryptoCommon.MultiSignerContainer {
	ccs.mutMultiSig.RLock()
	defer ccs.mutMultiSig.RUnlock()

	return ccs.MultiSigContainer
}

// SetMultiSignerContainer -
func (ccs *CryptoComponentsStub) SetMultiSignerContainer(ms cryptoCommon.MultiSignerContainer) error {
	ccs.mutMultiSig.Lock()
	ccs.MultiSigContainer = ms
	ccs.mutMultiSig.Unlock()

	return nil
}

// GetMultiSigner -
func (ccs *CryptoComponentsStub) GetMultiSigner(epoch uint32) (crypto.MultiSigner, error) {
	ccs.mutMultiSig.RLock()
	defer ccs.mutMultiSig.RUnlock()

	if ccs.MultiSigContainer == nil {
		return nil, errors.New("nil multi sig container")
	}

	return ccs.MultiSigContainer.GetMultiSigner(epoch)
}

// BlockSignKeyGen -
func (ccs *CryptoComponentsStub) BlockSignKeyGen() crypto.KeyGenerator {
	return ccs.BlKeyGen
}

// TxSignKeyGen -
func (ccs *CryptoComponentsStub) TxSignKeyGen() crypto.KeyGenerator {
	return ccs.TxKeyGen
}

// P2pKeyGen -
func (ccs *CryptoComponentsStub) P2pKeyGen() crypto.KeyGenerator {
	return ccs.P2PKeyGen
}

// MessageSignVerifier -
func (ccs *CryptoComponentsStub) MessageSignVerifier() vm.MessageSignVerifier {
	return ccs.MsgSigVerifier
}

// ConsensusSigningHandler -
func (ccs *CryptoComponentsStub) ConsensusSigningHandler() consensus.SigningHandler {
	return ccs.SigHandler
}

// ManagedPeersHolder -
func (ccs *CryptoComponentsStub) ManagedPeersHolder() common.ManagedPeersHolder {
	return ccs.ManagedPeersHolderField
}

// KeysHandler -
func (ccs *CryptoComponentsStub) KeysHandler() consensus.KeysHandler {
	if ccs.KeysHandlerCalled != nil {
		return ccs.KeysHandlerCalled()
	}
	return ccs.KeysHandlerField
}

// Clone -
func (ccs *CryptoComponentsStub) Clone() interface{} {
	return &CryptoComponentsStub{
		PubKey:                  ccs.PubKey,
		P2pPubKey:               ccs.P2pPubKey,
		PrivKey:                 ccs.PrivKey,
		P2pPrivKey:              ccs.P2pPrivKey,
		PubKeyString:            ccs.PubKeyString,
		PubKeyBytes:             ccs.PubKeyBytes,
		BlockSig:                ccs.BlockSig,
		TxSig:                   ccs.TxSig,
		MultiSigContainer:       ccs.MultiSigContainer,
		PeerSignHandler:         ccs.PeerSignHandler,
		BlKeyGen:                ccs.BlKeyGen,
		TxKeyGen:                ccs.TxKeyGen,
		P2PKeyGen:               ccs.P2PKeyGen,
		MsgSigVerifier:          ccs.MsgSigVerifier,
		ManagedPeersHolderField: ccs.ManagedPeersHolderField,
		KeysHandlerField:        ccs.KeysHandlerField,
		mutMultiSig:             sync.RWMutex{},
	}
}

// String -
func (ccs *CryptoComponentsStub) String() string {
	return "CryptoComponentsStub"
}

// IsInterfaceNil -
func (ccs *CryptoComponentsStub) IsInterfaceNil() bool {
	return ccs == nil
}
