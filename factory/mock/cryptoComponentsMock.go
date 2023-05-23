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

// CryptoComponentsMock -
type CryptoComponentsMock struct {
	PubKey                  crypto.PublicKey
	PrivKey                 crypto.PrivateKey
	P2pPubKey               crypto.PublicKey
	P2pPrivKey              crypto.PrivateKey
	P2pSig                  crypto.SingleSigner
	PubKeyString            string
	PubKeyBytes             []byte
	BlockSig                crypto.SingleSigner
	TxSig                   crypto.SingleSigner
	MultiSigContainer       cryptoCommon.MultiSignerContainer
	PeerSignHandler         crypto.PeerSignatureHandler
	BlKeyGen                crypto.KeyGenerator
	TxKeyGen                crypto.KeyGenerator
	P2PKeyGen               crypto.KeyGenerator
	MsgSigVerifier          vm.MessageSignVerifier
	SigHandler              consensus.SigningHandler
	ManagedPeersHolderField common.ManagedPeersHolder
	KeysHandlerField        consensus.KeysHandler
	mutMultiSig             sync.RWMutex
}

// PublicKey -
func (ccm *CryptoComponentsMock) PublicKey() crypto.PublicKey {
	return ccm.PubKey
}

// PrivateKey -
func (ccm *CryptoComponentsMock) PrivateKey() crypto.PrivateKey {
	return ccm.PrivKey
}

// P2pPrivateKey -
func (ccm *CryptoComponentsMock) P2pPrivateKey() crypto.PrivateKey {
	return ccm.P2pPrivKey
}

// P2pPublicKey -
func (ccm *CryptoComponentsMock) P2pPublicKey() crypto.PublicKey {
	return ccm.P2pPubKey
}

// P2pSingleSigner -
func (ccm *CryptoComponentsMock) P2pSingleSigner() crypto.SingleSigner {
	return ccm.P2pSig
}

// PublicKeyString -
func (ccm *CryptoComponentsMock) PublicKeyString() string {
	return ccm.PubKeyString
}

// PublicKeyBytes -
func (ccm *CryptoComponentsMock) PublicKeyBytes() []byte {
	return ccm.PubKeyBytes
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

// P2pKeyGen -
func (ccm *CryptoComponentsMock) P2pKeyGen() crypto.KeyGenerator {
	return ccm.P2PKeyGen
}

// MessageSignVerifier -
func (ccm *CryptoComponentsMock) MessageSignVerifier() vm.MessageSignVerifier {
	return ccm.MsgSigVerifier
}

// ConsensusSigningHandler -
func (ccm *CryptoComponentsMock) ConsensusSigningHandler() consensus.SigningHandler {
	return ccm.SigHandler
}

// ManagedPeersHolder -
func (ccm *CryptoComponentsMock) ManagedPeersHolder() common.ManagedPeersHolder {
	return ccm.ManagedPeersHolderField
}

// KeysHandler -
func (ccm *CryptoComponentsMock) KeysHandler() consensus.KeysHandler {
	return ccm.KeysHandlerField
}

// Clone -
func (ccm *CryptoComponentsMock) Clone() interface{} {
	return &CryptoComponentsMock{
		PubKey:                  ccm.PubKey,
		PrivKey:                 ccm.PrivKey,
		PubKeyString:            ccm.PubKeyString,
		PubKeyBytes:             ccm.PubKeyBytes,
		BlockSig:                ccm.BlockSig,
		TxSig:                   ccm.TxSig,
		MultiSigContainer:       ccm.MultiSigContainer,
		PeerSignHandler:         ccm.PeerSignHandler,
		BlKeyGen:                ccm.BlKeyGen,
		TxKeyGen:                ccm.TxKeyGen,
		P2PKeyGen:               ccm.P2PKeyGen,
		MsgSigVerifier:          ccm.MsgSigVerifier,
		ManagedPeersHolderField: ccm.ManagedPeersHolderField,
		KeysHandlerField:        ccm.KeysHandlerField,
		mutMultiSig:             sync.RWMutex{},
	}
}

// IsInterfaceNil -
func (ccm *CryptoComponentsMock) IsInterfaceNil() bool {
	return ccm == nil
}
