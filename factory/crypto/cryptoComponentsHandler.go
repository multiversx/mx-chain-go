package crypto

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	cryptoCommon "github.com/multiversx/mx-chain-go/common/crypto"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/vm"
)

var _ factory.ComponentHandler = (*managedCryptoComponents)(nil)
var _ factory.CryptoParamsHolder = (*managedCryptoComponents)(nil)
var _ factory.CryptoComponentsHolder = (*managedCryptoComponents)(nil)
var _ factory.CryptoComponentsHandler = (*managedCryptoComponents)(nil)

// CryptoComponentsHandlerArgs holds the arguments required to create a crypto components handler
type CryptoComponentsHandlerArgs CryptoComponentsFactoryArgs

// managedCryptoComponents creates the crypto components handler that can create, close and access the crypto components
type managedCryptoComponents struct {
	*cryptoComponents
	cryptoComponentsFactory *cryptoComponentsFactory
	mutCryptoComponents     sync.RWMutex
}

// NewManagedCryptoComponents creates a new Crypto components handler
func NewManagedCryptoComponents(ccf *cryptoComponentsFactory) (*managedCryptoComponents, error) {
	if ccf == nil {
		return nil, errors.ErrNilCryptoComponentsFactory
	}

	return &managedCryptoComponents{
		cryptoComponents:        nil,
		cryptoComponentsFactory: ccf,
	}, nil
}

// Create creates the crypto components
func (mcc *managedCryptoComponents) Create() error {
	cc, err := mcc.cryptoComponentsFactory.Create()
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrCryptoComponentsFactoryCreate, err)
	}

	mcc.mutCryptoComponents.Lock()
	mcc.cryptoComponents = cc
	mcc.mutCryptoComponents.Unlock()

	return nil
}

// Close closes the managed crypto components
func (mcc *managedCryptoComponents) Close() error {
	mcc.mutCryptoComponents.Lock()
	defer mcc.mutCryptoComponents.Unlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	err := mcc.cryptoComponents.Close()
	if err != nil {
		return err
	}
	mcc.cryptoComponents = nil

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (mcc *managedCryptoComponents) CheckSubcomponents() error {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return errors.ErrNilCryptoComponents
	}
	if check.IfNil(mcc.cryptoComponents.publicKey) {
		return errors.ErrNilPublicKey
	}
	if check.IfNil(mcc.cryptoComponents.p2pPublicKey) {
		return errors.ErrNilP2pPublicKey
	}
	if check.IfNil(mcc.cryptoComponents.privateKey) {
		return errors.ErrNilPrivateKey
	}
	if check.IfNil(mcc.cryptoComponents.p2pPrivateKey) {
		return errors.ErrNilP2pPrivateKey
	}
	if check.IfNil(mcc.cryptoComponents.txSingleSigner) {
		return errors.ErrNilTxSigner
	}
	if check.IfNil(mcc.cryptoComponents.blockSingleSigner) {
		return errors.ErrNilBlockSigner
	}
	if check.IfNil(mcc.cryptoComponents.p2pSingleSigner) {
		return errors.ErrNilP2pSigner
	}
	if check.IfNil(mcc.cryptoComponents.multiSignerContainer) {
		return errors.ErrNilMultiSigner
	}
	if check.IfNil(mcc.cryptoComponents.peerSignHandler) {
		return errors.ErrNilPeerSignHandler
	}
	if check.IfNil(mcc.cryptoComponents.blockSignKeyGen) {
		return errors.ErrNilBlockSignKeyGen
	}
	if check.IfNil(mcc.cryptoComponents.txSignKeyGen) {
		return errors.ErrNilTxSignKeyGen
	}
	if check.IfNil(mcc.cryptoComponents.p2pKeyGen) {
		return errors.ErrNilP2pKeyGen
	}
	if check.IfNil(mcc.cryptoComponents.messageSignVerifier) {
		return errors.ErrNilMessageSignVerifier
	}
	if check.IfNil(mcc.cryptoComponents.managedPeersHolder) {
		return errors.ErrNilManagedPeersHolder
	}

	return nil
}

// PublicKey returns the configured validator public key
func (mcc *managedCryptoComponents) PublicKey() crypto.PublicKey {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoParams.publicKey
}

// PrivateKey returns the configured validator private key
func (mcc *managedCryptoComponents) PrivateKey() crypto.PrivateKey {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoParams.privateKey
}

// P2pPrivateKey returns the configured p2p private key
func (mcc *managedCryptoComponents) P2pPrivateKey() crypto.PrivateKey {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.p2pCryptoParams.p2pPrivateKey
}

// P2pPublicKey returns the configured p2p public key
func (mcc *managedCryptoComponents) P2pPublicKey() crypto.PublicKey {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.p2pCryptoParams.p2pPublicKey
}

// PublicKeyString returns the configured validator public key as string
func (mcc *managedCryptoComponents) PublicKeyString() string {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return ""
	}

	return mcc.cryptoParams.publicKeyString
}

// PublicKeyBytes returns the configured validator public key bytes
func (mcc *managedCryptoComponents) PublicKeyBytes() []byte {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoParams.publicKeyBytes
}

// TxSingleSigner returns the transaction signer
func (mcc *managedCryptoComponents) TxSingleSigner() crypto.SingleSigner {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.txSingleSigner
}

// BlockSigner returns block single signer
func (mcc *managedCryptoComponents) BlockSigner() crypto.SingleSigner {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.blockSingleSigner
}

// P2pSingleSigner returns p2p single signer
func (mcc *managedCryptoComponents) P2pSingleSigner() crypto.SingleSigner {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.p2pSingleSigner
}

// MultiSignerContainer returns the multiSigner container holding the multiSigner versions
func (mcc *managedCryptoComponents) MultiSignerContainer() cryptoCommon.MultiSignerContainer {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()
	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.multiSignerContainer
}

// SetMultiSignerContainer sets the multiSigner container in the crypto components
func (mcc *managedCryptoComponents) SetMultiSignerContainer(ms cryptoCommon.MultiSignerContainer) error {
	if check.IfNil(ms) {
		return errors.ErrNilMultiSignerContainer
	}

	mcc.mutCryptoComponents.Lock()
	mcc.multiSignerContainer = ms
	mcc.mutCryptoComponents.Unlock()

	return nil
}

// GetMultiSigner returns the multiSigner configured in the multiSigner container for the given epoch
func (mcc *managedCryptoComponents) GetMultiSigner(epoch uint32) (crypto.MultiSigner, error) {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil, errors.ErrNilCryptoComponentsHolder
	}

	if mcc.multiSignerContainer == nil {
		return nil, errors.ErrNilMultiSignerContainer
	}

	return mcc.MultiSignerContainer().GetMultiSigner(epoch)
}

// PeerSignatureHandler returns the peer signature handler
func (mcc *managedCryptoComponents) PeerSignatureHandler() crypto.PeerSignatureHandler {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.peerSignHandler
}

// BlockSignKeyGen returns the block signer key generator
func (mcc *managedCryptoComponents) BlockSignKeyGen() crypto.KeyGenerator {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.blockSignKeyGen
}

// TxSignKeyGen returns the transaction signer key generator
func (mcc *managedCryptoComponents) TxSignKeyGen() crypto.KeyGenerator {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.txSignKeyGen
}

// P2pKeyGen returns the p2p key generator
func (mcc *managedCryptoComponents) P2pKeyGen() crypto.KeyGenerator {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.p2pKeyGen
}

// MessageSignVerifier returns the message signature verifier
func (mcc *managedCryptoComponents) MessageSignVerifier() vm.MessageSignVerifier {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.messageSignVerifier
}

// ConsensusSigningHandler returns the consensus signing handler
func (mcc *managedCryptoComponents) ConsensusSigningHandler() consensus.SigningHandler {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.consensusSigningHandler
}

// ManagedPeersHolder returns the managed peers holder
func (mcc *managedCryptoComponents) ManagedPeersHolder() common.ManagedPeersHolder {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.managedPeersHolder
}

// KeysHandler returns the handler that manages keys either in single sign mode or multi key mode
func (mcc *managedCryptoComponents) KeysHandler() consensus.KeysHandler {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.cryptoComponents == nil {
		return nil
	}

	return mcc.cryptoComponents.keysHandler
}

// Clone creates a shallow clone of a managedCryptoComponents
func (mcc *managedCryptoComponents) Clone() interface{} {
	cryptoComp := (*cryptoComponents)(nil)
	if mcc.cryptoComponents != nil {
		cryptoComp = &cryptoComponents{
			txSingleSigner:          mcc.TxSingleSigner(),
			blockSingleSigner:       mcc.BlockSigner(),
			p2pSingleSigner:         mcc.P2pSingleSigner(),
			multiSignerContainer:    mcc.MultiSignerContainer(),
			peerSignHandler:         mcc.PeerSignatureHandler(),
			blockSignKeyGen:         mcc.BlockSignKeyGen(),
			txSignKeyGen:            mcc.TxSignKeyGen(),
			p2pKeyGen:               mcc.P2pKeyGen(),
			messageSignVerifier:     mcc.MessageSignVerifier(),
			consensusSigningHandler: mcc.ConsensusSigningHandler(),
			managedPeersHolder:      mcc.ManagedPeersHolder(),
			keysHandler:             mcc.KeysHandler(),
			cryptoParams:            mcc.cryptoParams,
			p2pCryptoParams:         mcc.p2pCryptoParams,
		}
	}

	return &managedCryptoComponents{
		cryptoComponents:        cryptoComp,
		cryptoComponentsFactory: mcc.cryptoComponentsFactory,
		mutCryptoComponents:     sync.RWMutex{},
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (mcc *managedCryptoComponents) IsInterfaceNil() bool {
	return mcc == nil
}

// String returns the name of the component
func (mcc *managedCryptoComponents) String() string {
	return factory.CryptoComponentsName
}
