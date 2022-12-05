package crypto

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	disabledCrypto "github.com/ElrondNetwork/elrond-go-crypto/signing/disabled"
	disabledSig "github.com/ElrondNetwork/elrond-go-crypto/signing/disabled/singlesig"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519/singlesig"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl"
	mclSig "github.com/ElrondNetwork/elrond-go-crypto/signing/mcl/singlesig"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	cryptoCommon "github.com/ElrondNetwork/elrond-go/common/crypto"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/peerSignatureHandler"
	"github.com/ElrondNetwork/elrond-go/genesis/process/disabled"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/keysManagement"
	p2pFactory "github.com/ElrondNetwork/elrond-go/p2p/factory"
	storageFactory "github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageunit"
	"github.com/ElrondNetwork/elrond-go/vm"
	systemVM "github.com/ElrondNetwork/elrond-go/vm/process"
)

const (
	disabledSigChecking        = "disabled"
	mainMachineRedundancyLevel = 0
)

// CryptoComponentsFactoryArgs holds the arguments needed for creating crypto components
type CryptoComponentsFactoryArgs struct {
	ValidatorKeyPemFileName              string
	AllValidatorKeysPemFileName          string
	SkIndex                              int
	Config                               config.Config
	EnableEpochs                         config.EnableEpochs
	PrefsConfig                          config.Preferences
	CoreComponentsHolder                 factory.CoreComponentsHolder
	KeyLoader                            factory.KeyLoaderHandler
	ActivateBLSPubKeyMessageVerification bool
	IsInImportMode                       bool
	ImportModeNoSigCheck                 bool
	NoKeyProvided                        bool
	CurrentPid                           core.PeerID
}

type cryptoComponentsFactory struct {
	consensusType                        string
	validatorKeyPemFileName              string
	allValidatorKeysPemFileName          string
	skIndex                              int
	config                               config.Config
	enableEpochs                         config.EnableEpochs
	prefsConfig                          config.Preferences
	validatorPubKeyConverter             core.PubkeyConverter
	activateBLSPubKeyMessageVerification bool
	keyLoader                            factory.KeyLoaderHandler
	isInImportMode                       bool
	importModeNoSigCheck                 bool
	noKeyProvided                        bool
	currentPid                           core.PeerID
}

// cryptoParams holds the node public/private key data
type cryptoParams struct {
	publicKey          crypto.PublicKey
	privateKey         crypto.PrivateKey
	publicKeyString    string
	publicKeyBytes     []byte
	privateKeyBytes    []byte
	handledPrivateKeys [][]byte
}

// cryptoComponents struct holds the crypto components
type cryptoComponents struct {
	txSingleSigner       crypto.SingleSigner
	blockSingleSigner    crypto.SingleSigner
	multiSignerContainer cryptoCommon.MultiSignerContainer
	peerSignHandler      crypto.PeerSignatureHandler
	blockSignKeyGen      crypto.KeyGenerator
	txSignKeyGen         crypto.KeyGenerator
	messageSignVerifier  vm.MessageSignVerifier
	managedPeersHolder   heartbeat.ManagedPeersHolder
	keysHandler          consensus.KeysHandler
	cryptoParams
}

var log = logger.GetOrCreate("factory")

// NewCryptoComponentsFactory returns a new crypto components factory
func NewCryptoComponentsFactory(args CryptoComponentsFactoryArgs) (*cryptoComponentsFactory, error) {
	if check.IfNil(args.CoreComponentsHolder) {
		return nil, errors.ErrNilCoreComponents
	}
	if check.IfNil(args.CoreComponentsHolder.ValidatorPubKeyConverter()) {
		return nil, errors.ErrNilPubKeyConverter
	}
	if len(args.ValidatorKeyPemFileName) == 0 {
		return nil, errors.ErrNilPath
	}
	if args.KeyLoader == nil {
		return nil, errors.ErrNilKeyLoader
	}
	if len(args.CurrentPid) == 0 {
		return nil, errors.ErrEmptyPeerID
	}

	ccf := &cryptoComponentsFactory{
		consensusType:                        args.Config.Consensus.Type,
		validatorKeyPemFileName:              args.ValidatorKeyPemFileName,
		skIndex:                              args.SkIndex,
		config:                               args.Config,
		prefsConfig:                          args.PrefsConfig,
		validatorPubKeyConverter:             args.CoreComponentsHolder.ValidatorPubKeyConverter(),
		activateBLSPubKeyMessageVerification: args.ActivateBLSPubKeyMessageVerification,
		keyLoader:                            args.KeyLoader,
		isInImportMode:                       args.IsInImportMode,
		importModeNoSigCheck:                 args.ImportModeNoSigCheck,
		enableEpochs:                         args.EnableEpochs,
		noKeyProvided:                        args.NoKeyProvided,
		currentPid:                           args.CurrentPid,
		allValidatorKeysPemFileName:          args.AllValidatorKeysPemFileName,
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

	// TODO: refactor the logic for isMainMachine
	redundancyLevel := int(ccf.prefsConfig.Preferences.RedundancyLevel)
	isMainMachine := redundancyLevel == mainMachineRedundancyLevel
	argsManagedPeersHolder := keysManagement.ArgsManagedPeersHolder{
		KeyGenerator:                     blockSignKeyGen,
		P2PIdentityGenerator:             p2pFactory.NewIdentityGenerator(),
		IsMainMachine:                    isMainMachine,
		MaxRoundsWithoutReceivedMessages: redundancyLevel,
		PrefsConfig:                      ccf.prefsConfig,
	}
	managedPeersHolder, err := keysManagement.NewManagedPeersHolder(argsManagedPeersHolder)
	if err != nil {
		return nil, err
	}

	for _, skBytes := range cp.handledPrivateKeys {
		errAddManagedPeer := managedPeersHolder.AddManagedPeer(skBytes)
		if errAddManagedPeer != nil {
			return nil, errAddManagedPeer
		}
	}

	log.Debug("block sign pubkey", "value", cp.publicKeyString)

	argsKeysHandler := keysManagement.ArgsKeysHandler{
		ManagedPeersHolder: managedPeersHolder,
		PrivateKey:         cp.privateKey,
		Pid:                ccf.currentPid,
	}
	keysHandler, err := keysManagement.NewKeysHandler(argsKeysHandler)
	if err != nil {
		return nil, err
	}

	return &cryptoComponents{
		txSingleSigner:       txSingleSigner,
		blockSingleSigner:    interceptSingleSigner,
		multiSignerContainer: multiSigner,
		peerSignHandler:      peerSigHandler,
		blockSignKeyGen:      blockSignKeyGen,
		txSignKeyGen:         txSignKeyGen,
		messageSignVerifier:  messageSignVerifier,
		managedPeersHolder:   managedPeersHolder,
		keysHandler:          keysHandler,
		cryptoParams:         *cp,
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

	handledPrivateKeys, err := ccf.processAllHandledKeys(keygen)
	if err != nil {
		return nil, err
	}

	if ccf.isInImportMode {
		if len(handledPrivateKeys) > 0 {
			return nil, fmt.Errorf("invalid node configuration: import-db mode and allValidatorsKeys.pem file provided")
		}

		return ccf.generateCryptoParams(keygen, "in import mode", handledPrivateKeys)
	}
	if ccf.noKeyProvided {
		return ccf.generateCryptoParams(keygen, "with no-key flag enabled", handledPrivateKeys)
	}
	if len(handledPrivateKeys) > 0 {
		return ccf.generateCryptoParams(keygen, "running with a provided allValidatorsKeys.pem", handledPrivateKeys)
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

	cp.publicKeyString = ccf.validatorPubKeyConverter.Encode(cp.publicKeyBytes)

	return cp, nil
}

func (ccf *cryptoComponentsFactory) generateCryptoParams(
	keygen crypto.KeyGenerator,
	reason string,
	handledPrivateKeys [][]byte,
) (*cryptoParams, error) {
	log.Warn(fmt.Sprintf("the node is %s! Will generate a fresh new BLS key", reason))
	cp := &cryptoParams{}
	cp.privateKey, cp.publicKey = keygen.GeneratePair()

	var err error
	cp.publicKeyBytes, err = cp.publicKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	cp.publicKeyString = ccf.validatorPubKeyConverter.Encode(cp.publicKeyBytes)
	cp.handledPrivateKeys = handledPrivateKeys

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

	pkBytes, err := ccf.validatorPubKeyConverter.Decode(pkString)
	if err != nil {
		return nil, nil, fmt.Errorf("%w for encoded public key %s", err, pkString)
	}

	return skBytes, pkBytes, nil
}

func (ccf *cryptoComponentsFactory) processAllHandledKeys(keygen crypto.KeyGenerator) ([][]byte, error) {
	privateKeys, publicKeys, err := ccf.keyLoader.LoadAllKeys(ccf.allValidatorKeysPemFileName)
	if err != nil {
		log.Debug("allValidatorsKeys could not be loaded", "reason", err)
		return make([][]byte, 0), nil
	}

	if len(privateKeys) != len(publicKeys) {
		return nil, fmt.Errorf("key loading error for the allValidatorsKeys file: mismatch number of private and public keys")
	}

	handledPrivateKeys := make([][]byte, 0, len(privateKeys))
	for i, pkString := range publicKeys {
		sk := privateKeys[i]
		processedSkBytes, errCheck := ccf.processPrivatePublicKey(keygen, sk, pkString, i)
		if errCheck != nil {
			return nil, errCheck
		}

		log.Debug("loaded handled node key", "public key", pkString)
		handledPrivateKeys = append(handledPrivateKeys, processedSkBytes)
	}

	return handledPrivateKeys, nil
}

func (ccf *cryptoComponentsFactory) processPrivatePublicKey(keygen crypto.KeyGenerator, encodedSk []byte, pkString string, index int) ([]byte, error) {
	skBytes, err := hex.DecodeString(string(encodedSk))
	if err != nil {
		return nil, fmt.Errorf("%w for encoded secret key, key index %d", err, index)
	}

	pkBytes, err := ccf.validatorPubKeyConverter.Decode(pkString)
	if err != nil {
		return nil, fmt.Errorf("%w for encoded public key %s, key index %d", err, pkString, index)
	}

	sk, err := keygen.PrivateKeyFromByteArray(skBytes)
	if err != nil {
		return nil, fmt.Errorf("%w secret key, key index %d", err, index)
	}

	pk := sk.GeneratePublic()
	pkGeneratedBytes, err := pk.ToByteArray()
	if err != nil {
		return nil, fmt.Errorf("%w while generating public key bytes, key index %d", err, index)
	}

	if !bytes.Equal(pkGeneratedBytes, pkBytes) {
		return nil, fmt.Errorf("public keys mismatch, read %s, generated %s, key index %d",
			pkString,
			ccf.validatorPubKeyConverter.Encode(pkBytes),
			index,
		)
	}

	return skBytes, nil
}

// Close closes all underlying components that need closing
func (cc *cryptoComponents) Close() error {
	return nil
}
