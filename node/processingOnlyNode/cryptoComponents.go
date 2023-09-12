package processingOnlyNode

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	cryptoCommon "github.com/multiversx/mx-chain-go/common/crypto"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/factory"
	cryptoComp "github.com/multiversx/mx-chain-go/factory/crypto"
	"github.com/multiversx/mx-chain-go/vm"
)

// ArgsCryptoComponentsHolder holds all arguments needed to create a crypto components holder
type ArgsCryptoComponentsHolder struct {
	Config                  config.Config
	EnableEpochsConfig      config.EnableEpochs
	Preferences             config.Preferences
	CoreComponentsHolder    factory.CoreComponentsHolder
	ValidatorKeyPemFileName string
}

type cryptoComponentsHolder struct {
	publicKey               crypto.PublicKey
	privateKey              crypto.PrivateKey
	p2pPublicKey            crypto.PublicKey
	p2pPrivateKey           crypto.PrivateKey
	p2pSingleSigner         crypto.SingleSigner
	txSingleSigner          crypto.SingleSigner
	blockSigner             crypto.SingleSigner
	multiSignerContainer    cryptoCommon.MultiSignerContainer
	peerSignatureHandler    crypto.PeerSignatureHandler
	blockSignKeyGen         crypto.KeyGenerator
	txSignKeyGen            crypto.KeyGenerator
	p2pKeyGen               crypto.KeyGenerator
	messageSignVerifier     vm.MessageSignVerifier
	consensusSigningHandler consensus.SigningHandler
	managedPeersHolder      common.ManagedPeersHolder
	keysHandler             consensus.KeysHandler
	publicKeyBytes          []byte
	publicKeyString         string
}

// CreateCryptoComponentsHolder will create a new instance of cryptoComponentsHolder
func CreateCryptoComponentsHolder(args ArgsCryptoComponentsHolder) (factory.CryptoComponentsHolder, error) {
	instance := &cryptoComponentsHolder{}

	cryptoComponentsHandlerArgs := cryptoComp.CryptoComponentsFactoryArgs{
		Config:                               args.Config,
		EnableEpochs:                         args.EnableEpochsConfig,
		PrefsConfig:                          args.Preferences,
		CoreComponentsHolder:                 args.CoreComponentsHolder,
		KeyLoader:                            core.NewKeyLoader(),
		ActivateBLSPubKeyMessageVerification: true,
		IsInImportMode:                       false,
		ImportModeNoSigCheck:                 false,
		NoKeyProvided:                        false,

		P2pKeyPemFileName:           "",
		ValidatorKeyPemFileName:     args.ValidatorKeyPemFileName,
		AllValidatorKeysPemFileName: "",
		SkIndex:                     0,
	}

	cryptoComponentsFactory, err := cryptoComp.NewCryptoComponentsFactory(cryptoComponentsHandlerArgs)
	if err != nil {
		return nil, fmt.Errorf("NewCryptoComponentsFactory failed: %w", err)
	}

	managedCryptoComponents, err := cryptoComp.NewManagedCryptoComponents(cryptoComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedCryptoComponents.Create()
	if err != nil {
		return nil, err
	}

	instance.publicKey = managedCryptoComponents.PublicKey()
	instance.privateKey = managedCryptoComponents.PrivateKey()
	instance.publicKeyBytes, err = instance.publicKey.ToByteArray()
	if err != nil {
		return nil, err
	}
	instance.publicKeyString, err = args.CoreComponentsHolder.ValidatorPubKeyConverter().Encode(instance.publicKeyBytes)
	if err != nil {
		return nil, err
	}

	instance.p2pPublicKey = managedCryptoComponents.P2pPublicKey()
	instance.p2pPrivateKey = managedCryptoComponents.P2pPrivateKey()
	instance.p2pSingleSigner = managedCryptoComponents.P2pSingleSigner()
	instance.blockSigner = managedCryptoComponents.BlockSigner()
	instance.txSingleSigner = managedCryptoComponents.TxSingleSigner()
	instance.multiSignerContainer = managedCryptoComponents.MultiSignerContainer()
	instance.peerSignatureHandler = managedCryptoComponents.PeerSignatureHandler()
	instance.blockSignKeyGen = managedCryptoComponents.BlockSignKeyGen()
	instance.txSignKeyGen = managedCryptoComponents.TxSignKeyGen()
	instance.p2pKeyGen = managedCryptoComponents.P2pKeyGen()
	instance.messageSignVerifier = managedCryptoComponents.MessageSignVerifier()
	instance.consensusSigningHandler = managedCryptoComponents.ConsensusSigningHandler()
	instance.managedPeersHolder = managedCryptoComponents.ManagedPeersHolder()
	instance.keysHandler = managedCryptoComponents.KeysHandler()

	return instance, nil
}

// PublicKey will return the public key
func (c *cryptoComponentsHolder) PublicKey() crypto.PublicKey {
	return c.publicKey
}

// PrivateKey will return the private key
func (c *cryptoComponentsHolder) PrivateKey() crypto.PrivateKey {
	return c.privateKey
}

// PublicKeyString will return the private key string
func (c *cryptoComponentsHolder) PublicKeyString() string {
	return c.publicKeyString
}

// PublicKeyBytes will return the public key bytes
func (c *cryptoComponentsHolder) PublicKeyBytes() []byte {
	return c.publicKeyBytes
}

// P2pPublicKey will return the p2p public key
func (c *cryptoComponentsHolder) P2pPublicKey() crypto.PublicKey {
	return c.p2pPublicKey
}

// P2pPrivateKey will return the p2p private key
func (c *cryptoComponentsHolder) P2pPrivateKey() crypto.PrivateKey {
	return c.p2pPrivateKey
}

// P2pSingleSigner will return the p2p single signer
func (c *cryptoComponentsHolder) P2pSingleSigner() crypto.SingleSigner {
	return c.p2pSingleSigner
}

// TxSingleSigner will return the transaction single signer
func (c *cryptoComponentsHolder) TxSingleSigner() crypto.SingleSigner {
	return c.txSingleSigner
}

// BlockSigner will return the block signer
func (c *cryptoComponentsHolder) BlockSigner() crypto.SingleSigner {
	return c.blockSigner
}

// SetMultiSignerContainer will set the multi signer container
func (c *cryptoComponentsHolder) SetMultiSignerContainer(container cryptoCommon.MultiSignerContainer) error {
	c.multiSignerContainer = container

	return nil
}

// MultiSignerContainer will return the multi signer container
func (c *cryptoComponentsHolder) MultiSignerContainer() cryptoCommon.MultiSignerContainer {
	return c.multiSignerContainer
}

// GetMultiSigner will return the multi signer by epoch
func (c *cryptoComponentsHolder) GetMultiSigner(epoch uint32) (crypto.MultiSigner, error) {
	return c.MultiSignerContainer().GetMultiSigner(epoch)
}

// PeerSignatureHandler will return the peer signature handler
func (c *cryptoComponentsHolder) PeerSignatureHandler() crypto.PeerSignatureHandler {
	return c.peerSignatureHandler
}

// BlockSignKeyGen will return the block signer key generator
func (c *cryptoComponentsHolder) BlockSignKeyGen() crypto.KeyGenerator {
	return c.blockSignKeyGen
}

// TxSignKeyGen will return the transaction sign key generator
func (c *cryptoComponentsHolder) TxSignKeyGen() crypto.KeyGenerator {
	return c.txSignKeyGen
}

// P2pKeyGen will return the p2p key generator
func (c *cryptoComponentsHolder) P2pKeyGen() crypto.KeyGenerator {
	return c.p2pKeyGen
}

// MessageSignVerifier will return the message signature verifier
func (c *cryptoComponentsHolder) MessageSignVerifier() vm.MessageSignVerifier {
	return c.messageSignVerifier
}

// ConsensusSigningHandler will return the consensus signing handler
func (c *cryptoComponentsHolder) ConsensusSigningHandler() consensus.SigningHandler {
	return c.consensusSigningHandler
}

// ManagedPeersHolder will return the managed peer holder
func (c *cryptoComponentsHolder) ManagedPeersHolder() common.ManagedPeersHolder {
	return c.managedPeersHolder
}

// KeysHandler will return the keys handler
func (c *cryptoComponentsHolder) KeysHandler() consensus.KeysHandler {
	return c.keysHandler
}

// Clone will clone the cryptoComponentsHolder
func (c *cryptoComponentsHolder) Clone() interface{} {
	return &cryptoComponentsHolder{
		publicKey:               c.PublicKey(),
		privateKey:              c.PrivateKey(),
		p2pPublicKey:            c.P2pPublicKey(),
		p2pPrivateKey:           c.P2pPrivateKey(),
		p2pSingleSigner:         c.P2pSingleSigner(),
		txSingleSigner:          c.TxSingleSigner(),
		blockSigner:             c.BlockSigner(),
		multiSignerContainer:    c.MultiSignerContainer(),
		peerSignatureHandler:    c.PeerSignatureHandler(),
		blockSignKeyGen:         c.BlockSignKeyGen(),
		txSignKeyGen:            c.TxSignKeyGen(),
		p2pKeyGen:               c.P2pKeyGen(),
		messageSignVerifier:     c.MessageSignVerifier(),
		consensusSigningHandler: c.ConsensusSigningHandler(),
		managedPeersHolder:      c.ManagedPeersHolder(),
		keysHandler:             c.KeysHandler(),
		publicKeyBytes:          c.PublicKeyBytes(),
		publicKeyString:         c.PublicKeyString(),
	}
}

func (c *cryptoComponentsHolder) IsInterfaceNil() bool {
	return c == nil
}
