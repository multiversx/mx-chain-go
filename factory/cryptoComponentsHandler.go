package factory

import (
	"context"
	"sync"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/vm"
)

var _ ComponentHandler = (*ManagedCryptoComponents)(nil)
var _ CryptoParamsHolder = (*ManagedCryptoComponents)(nil)
var _ CryptoComponentsHolder = (*ManagedCryptoComponents)(nil)
var _ CryptoComponentsHandler = (*ManagedCryptoComponents)(nil)

// CryptoComponentsHandlerArgs holds the arguments required to create a crypto components handler
type CryptoComponentsHandlerArgs CryptoComponentsFactoryArgs

// ManagedCryptoComponents creates the crypto components handler that can create, close and access the crypto components
type ManagedCryptoComponents struct {
	*CryptoComponents
	cryptoComponentsFactory *cryptoComponentsFactory
	cancelFunc              func()
	mutCryptoComponents     sync.RWMutex
}

// NewManagedCryptoComponents creates a new Crypto components handler
func NewManagedCryptoComponents(args CryptoComponentsHandlerArgs) (*ManagedCryptoComponents, error) {
	ccf, err := NewCryptoComponentsFactory(CryptoComponentsFactoryArgs(args))
	if err != nil {
		return nil, err
	}

	return &ManagedCryptoComponents{
		CryptoComponents:        nil,
		cryptoComponentsFactory: ccf,
	}, nil
}

// Close closes the managed crypto components
func (mcc *ManagedCryptoComponents) Close() error {
	mcc.mutCryptoComponents.Lock()
	mcc.cancelFunc()
	mcc.cancelFunc = nil
	mcc.CryptoComponents = nil
	mcc.mutCryptoComponents.Unlock()

	return nil
}

// Create creates the crypto components
func (mcc *ManagedCryptoComponents) Create() error {
	cryptoComponents, err := mcc.cryptoComponentsFactory.Create()
	if err != nil {
		return err
	}

	mcc.mutCryptoComponents.Lock()
	mcc.CryptoComponents = cryptoComponents
	_, mcc.cancelFunc = context.WithCancel(context.Background())
	mcc.mutCryptoComponents.Unlock()

	return nil
}

// PublicKey returns the configured validator public key
func (mcc *ManagedCryptoComponents) PublicKey() crypto.PublicKey {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.CryptoComponents == nil {
		return nil
	}

	return mcc.CryptoParams.PublicKey
}

// PrivateKey returns the configured validator private key
func (mcc *ManagedCryptoComponents) PrivateKey() crypto.PrivateKey {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.CryptoComponents == nil {
		return nil
	}

	return mcc.CryptoParams.PrivateKey
}

// PublicKeyString returns the configured validator public key as string
func (mcc *ManagedCryptoComponents) PublicKeyString() string {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.CryptoComponents == nil {
		return ""
	}

	return mcc.CryptoParams.PublicKeyString
}

// PublicKeyBytes returns the configured validator public key bytes
func (mcc *ManagedCryptoComponents) PublicKeyBytes() []byte {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.CryptoComponents == nil {
		return nil
	}

	return mcc.CryptoParams.PublicKeyBytes
}

// PrivateKeyBytes returns the configured validator private key bytes
func (mcc *ManagedCryptoComponents) PrivateKeyBytes() []byte {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.CryptoComponents == nil {
		return nil
	}

	return mcc.CryptoParams.PrivateKeyBytes
}

// TxSingleSigner returns the transaction signer
func (mcc *ManagedCryptoComponents) TxSingleSigner() crypto.SingleSigner {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.CryptoComponents == nil {
		return nil
	}

	return mcc.CryptoComponents.TxSingleSigner
}

// SingleSigner returns block single signer
func (mcc *ManagedCryptoComponents) SingleSigner() crypto.SingleSigner {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.CryptoComponents == nil {
		return nil
	}

	return mcc.CryptoComponents.SingleSigner
}

// MultiSigner returns the block multi-signer
func (mcc *ManagedCryptoComponents) MultiSigner() crypto.MultiSigner {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.CryptoComponents == nil {
		return nil
	}

	return mcc.CryptoComponents.MultiSigner
}

// BlockSignKeyGen returns the block signer key generator
func (mcc *ManagedCryptoComponents) BlockSignKeyGen() crypto.KeyGenerator {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.CryptoComponents == nil {
		return nil
	}

	return mcc.CryptoComponents.BlockSignKeyGen
}

// TxSignKeyGen returns the transaction signer key generator
func (mcc *ManagedCryptoComponents) TxSignKeyGen() crypto.KeyGenerator {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.CryptoComponents == nil {
		return nil
	}

	return mcc.CryptoComponents.TxSignKeyGen
}

// MessageSignVerifier returns the message signature verifier
func (mcc *ManagedCryptoComponents) MessageSignVerifier() vm.MessageSignVerifier {
	mcc.mutCryptoComponents.RLock()
	defer mcc.mutCryptoComponents.RUnlock()

	if mcc.CryptoComponents == nil {
		return nil
	}

	return mcc.CryptoComponents.MessageSignVerifier
}

// IsInterfaceNil returns true if there is no value under the interface
func (mcc *ManagedCryptoComponents) IsInterfaceNil() bool {
	return mcc == nil
}
