package crypto

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	disabledCrypto "github.com/ElrondNetwork/elrond-go-crypto/signing/disabled"
	disabledSig "github.com/ElrondNetwork/elrond-go-crypto/signing/disabled/singlesig"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519/singlesig"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl"
	mclSig "github.com/ElrondNetwork/elrond-go-crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/secp256k1"
	secp256k1SinglerSig "github.com/ElrondNetwork/elrond-go-crypto/signing/secp256k1/singlesig"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	cryptoCommon "github.com/ElrondNetwork/elrond-go/common/crypto"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/peerSignatureHandler"
	"github.com/ElrondNetwork/elrond-go/genesis/process/disabled"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageunit"
	"github.com/ElrondNetwork/elrond-go/vm"
	systemVM "github.com/ElrondNetwork/elrond-go/vm/process"
)

const disabledSigChecking = "disabled"

// CryptoComponentsFactoryArgs holds the arguments needed for creating crypto components
type CryptoComponentsFactoryArgs struct {
	ValidatorKeyPemFileName              string
	SkIndex                              int
	Config                               config.Config
	EnableEpochs                         config.EnableEpochs
	CoreComponentsHolder                 factory.CoreComponentsHolder
	KeyLoader                            factory.KeyLoaderHandler
	ActivateBLSPubKeyMessageVerification bool
	IsInImportMode                       bool
	ImportModeNoSigCheck                 bool
	NoKeyProvided                        bool
	P2pKeyPemFileName                    string
}

type cryptoComponentsFactory struct {
	consensusType                        string
	validatorKeyPemFileName              string
	skIndex                              int
	config                               config.Config
	enableEpochs                         config.EnableEpochs
	coreComponentsHolder                 factory.CoreComponentsHolder
	activateBLSPubKeyMessageVerification bool
	keyLoader                            factory.KeyLoaderHandler
	isInImportMode                       bool
	importModeNoSigCheck                 bool
	noKeyProvided                        bool
	p2pKeyPemFileName                    string
}

// cryptoParams holds the node public/private key data
type cryptoParams struct {
	publicKey       crypto.PublicKey
	privateKey      crypto.PrivateKey
	publicKeyString string
	publicKeyBytes  []byte
	privateKeyBytes []byte
}

// p2pCryptoParams holds the p2p public/private key data
type p2pCryptoParams struct {
	p2pPublicKey  crypto.PublicKey
	p2pPrivateKey crypto.PrivateKey
}

// cryptoComponents struct holds the crypto components
type cryptoComponents struct {
	txSingleSigner       crypto.SingleSigner
	blockSingleSigner    crypto.SingleSigner
	p2pSingleSigner      crypto.SingleSigner
	multiSignerContainer cryptoCommon.MultiSignerContainer
	peerSignHandler      crypto.PeerSignatureHandler
	blockSignKeyGen      crypto.KeyGenerator
	txSignKeyGen         crypto.KeyGenerator
	p2pKeyGen            crypto.KeyGenerator
	messageSignVerifier  vm.MessageSignVerifier
	cryptoParams
	p2pCryptoParams
}

var log = logger.GetOrCreate("factory")

// NewCryptoComponentsFactory returns a new crypto components factory
func NewCryptoComponentsFactory(args CryptoComponentsFactoryArgs) (*cryptoComponentsFactory, error) {
	if check.IfNil(args.CoreComponentsHolder) {
		return nil, errors.ErrNilCoreComponents
	}
	if len(args.ValidatorKeyPemFileName) == 0 {
		return nil, errors.ErrNilPath
	}
	if args.KeyLoader == nil {
		return nil, errors.ErrNilKeyLoader
	}

	ccf := &cryptoComponentsFactory{
		consensusType:                        args.Config.Consensus.Type,
		validatorKeyPemFileName:              args.ValidatorKeyPemFileName,
		skIndex:                              args.SkIndex,
		config:                               args.Config,
		coreComponentsHolder:                 args.CoreComponentsHolder,
		activateBLSPubKeyMessageVerification: args.ActivateBLSPubKeyMessageVerification,
		keyLoader:                            args.KeyLoader,
		isInImportMode:                       args.IsInImportMode,
		importModeNoSigCheck:                 args.ImportModeNoSigCheck,
		enableEpochs:                         args.EnableEpochs,
		noKeyProvided:                        args.NoKeyProvided,
		p2pKeyPemFileName:                    args.P2pKeyPemFileName,
	}

	return ccf, nil
}

// Create will create and return crypto components
func (ccf *cryptoComponentsFactory) Create() (*cryptoComponents, error) {
	suite, err := ccf.getSuite()
	if err != nil {
		return nil, err
	}

	blockSignKeyGen := signing.NewKeyGenerator(suite)
	cp, err := ccf.createCryptoParams(blockSignKeyGen)
	if err != nil {
		return nil, err
	}

	txSignKeyGen := signing.NewKeyGenerator(ed25519.NewEd25519())
	txSingleSigner := &singlesig.Ed25519Signer{}
	processingSingleSigner, err := ccf.createSingleSigner(false)
	if err != nil {
		return nil, err
	}

	interceptSingleSigner, err := ccf.createSingleSigner(ccf.importModeNoSigCheck)
	if err != nil {
		return nil, err
	}

	p2pSingleSigner := &secp256k1SinglerSig.Secp256k1Signer{}

	multiSigner, err := ccf.createMultiSignerContainer(blockSignKeyGen, ccf.importModeNoSigCheck)
	if err != nil {
		return nil, err
	}

	var messageSignVerifier vm.MessageSignVerifier
	if ccf.activateBLSPubKeyMessageVerification {
		messageSignVerifier, err = systemVM.NewMessageSigVerifier(blockSignKeyGen, processingSingleSigner)
		if err != nil {
			return nil, err
		}
	} else {
		messageSignVerifier, err = disabled.NewMessageSignVerifier(blockSignKeyGen)
		if err != nil {
			return nil, err
		}
	}

	cacheConfig := ccf.config.PublicKeyPIDSignature
	cachePkPIDSignature, err := storageunit.NewCache(storageFactory.GetCacherFromConfig(cacheConfig))
	if err != nil {
		return nil, err
	}

	peerSigHandler, err := peerSignatureHandler.NewPeerSignatureHandler(cachePkPIDSignature, interceptSingleSigner, blockSignKeyGen)
	if err != nil {
		return nil, err
	}

	p2pKeyGenerator := signing.NewKeyGenerator(secp256k1.NewSecp256k1())
	p2pCryptoParams, err := ccf.createP2pCryptoParams(p2pKeyGenerator)
	if err != nil {
		return nil, err
	}

	log.Debug("block sign pubkey", "value", cp.publicKeyString)

	return &cryptoComponents{
		txSingleSigner:       txSingleSigner,
		blockSingleSigner:    interceptSingleSigner,
		multiSignerContainer: multiSigner,
		peerSignHandler:      peerSigHandler,
		blockSignKeyGen:      blockSignKeyGen,
		txSignKeyGen:         txSignKeyGen,
		p2pKeyGen:            p2pKeyGenerator,
		messageSignVerifier:  messageSignVerifier,
		cryptoParams:         *cp,
		p2pCryptoParams:      *p2pCryptoParams,
		p2pSingleSigner:      p2pSingleSigner,
	}, nil
}

func (ccf *cryptoComponentsFactory) createSingleSigner(importModeNoSigCheck bool) (crypto.SingleSigner, error) {
	if importModeNoSigCheck {
		log.Warn("using disabled single signer because the node is running in import-db 'turbo mode'")
		return &disabledSig.DisabledSingleSig{}, nil
	}

	switch ccf.consensusType {
	case consensus.BlsConsensusType:
		return &mclSig.BlsSingleSigner{}, nil
	case disabledSigChecking:
		log.Warn("using disabled single signer")
		return &disabledSig.DisabledSingleSig{}, nil
	default:
		return nil, errors.ErrInvalidConsensusConfig
	}
}

func (ccf *cryptoComponentsFactory) createMultiSignerContainer(
	blSignKeyGen crypto.KeyGenerator,
	importModeNoSigCheck bool,
) (cryptoCommon.MultiSignerContainer, error) {

	args := MultiSigArgs{
		MultiSigHasherType:   ccf.config.MultisigHasher.Type,
		BlSignKeyGen:         blSignKeyGen,
		ConsensusType:        ccf.consensusType,
		ImportModeNoSigCheck: importModeNoSigCheck,
	}
	return NewMultiSignerContainer(args, ccf.enableEpochs.BLSMultiSignerEnableEpoch)
}

func (ccf *cryptoComponentsFactory) getSuite() (crypto.Suite, error) {
	switch ccf.config.Consensus.Type {
	case consensus.BlsConsensusType:
		return mcl.NewSuiteBLS12(), nil
	case disabledSigChecking:
		log.Warn("using disabled multi signer")
		return disabledCrypto.NewDisabledSuite(), nil
	default:
		return nil, errors.ErrInvalidConsensusConfig
	}
}

func (ccf *cryptoComponentsFactory) createCryptoParams(
	keygen crypto.KeyGenerator,
) (*cryptoParams, error) {

	shouldGenerateCryptoParams := ccf.isInImportMode || ccf.noKeyProvided
	if shouldGenerateCryptoParams {
		return ccf.generateCryptoParams(keygen)
	}

	return ccf.readCryptoParams(keygen)
}

func (ccf *cryptoComponentsFactory) readCryptoParams(keygen crypto.KeyGenerator) (*cryptoParams, error) {
	cp := &cryptoParams{}
	sk, readPk, err := ccf.getSkPk()
	if err != nil {
		return nil, err
	}

	cp.privateKey, err = keygen.PrivateKeyFromByteArray(sk)
	if err != nil {
		return nil, err
	}

	cp.publicKey = cp.privateKey.GeneratePublic()
	if len(readPk) > 0 {
		cp.publicKeyBytes, err = cp.publicKey.ToByteArray()
		if err != nil {
			return nil, err
		}

		if !bytes.Equal(cp.publicKeyBytes, readPk) {
			return nil, errors.ErrPublicKeyMismatch
		}
	}

	validatorKeyConverter := ccf.coreComponentsHolder.ValidatorPubKeyConverter()
	cp.publicKeyString = validatorKeyConverter.Encode(cp.publicKeyBytes)

	return cp, nil
}

func (ccf *cryptoComponentsFactory) generateCryptoParams(keygen crypto.KeyGenerator) (*cryptoParams, error) {
	var message string
	if ccf.noKeyProvided {
		message = "with no-key flag enabled"
	} else {
		message = "in import mode"
	}

	log.Warn(fmt.Sprintf("the node is %s! Will generate a fresh new BLS key", message))
	cp := &cryptoParams{}
	cp.privateKey, cp.publicKey = keygen.GeneratePair()

	var err error
	cp.publicKeyBytes, err = cp.publicKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	validatorKeyConverter := ccf.coreComponentsHolder.ValidatorPubKeyConverter()
	cp.publicKeyString = validatorKeyConverter.Encode(cp.publicKeyBytes)

	return cp, nil
}

func (ccf *cryptoComponentsFactory) getSkPk() ([]byte, []byte, error) {
	encodedSk, pkString, err := ccf.keyLoader.LoadKey(ccf.validatorKeyPemFileName, ccf.skIndex)
	if err != nil {
		return nil, nil, err
	}

	skBytes, err := hex.DecodeString(string(encodedSk))
	if err != nil {
		return nil, nil, fmt.Errorf("%w for encoded secret key", err)
	}

	validatorKeyConverter := ccf.coreComponentsHolder.ValidatorPubKeyConverter()
	pkBytes, err := validatorKeyConverter.Decode(pkString)
	if err != nil {
		return nil, nil, fmt.Errorf("%w for encoded public key %s", err, pkString)
	}

	return skBytes, pkBytes, nil
}

func (ccf *cryptoComponentsFactory) createP2pCryptoParams(
	keygen crypto.KeyGenerator,
) (*p2pCryptoParams, error) {
	cp := &p2pCryptoParams{}

	p2pPrivateKeyBytes, err := common.GetSkBytesFromP2pKey(ccf.p2pKeyPemFileName)
	if err != nil {
		return nil, err
	}

	if len(p2pPrivateKeyBytes) == 0 {
		cp.p2pPrivateKey, cp.p2pPublicKey = keygen.GeneratePair()

		log.Info("p2p private key: generated a new private key for p2p signing")

		return cp, nil
	}

	cp.p2pPrivateKey, err = keygen.PrivateKeyFromByteArray(p2pPrivateKeyBytes)
	if err != nil {
		return nil, err
	}

	log.Info("p2p private key: using the provided private key for p2p signing")

	cp.p2pPublicKey = cp.p2pPrivateKey.GeneratePublic()

	return cp, nil
}

// Close closes all underlying components that need closing
func (cc *cryptoComponents) Close() error {
	return nil
}
