package crypto

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	disabledCrypto "github.com/multiversx/mx-chain-crypto-go/signing/disabled"
	disabledSig "github.com/multiversx/mx-chain-crypto-go/signing/disabled/singlesig"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	mclSig "github.com/multiversx/mx-chain-crypto-go/signing/mcl/singlesig"
	"github.com/multiversx/mx-chain-crypto-go/signing/secp256k1"
	secp256k1SinglerSig "github.com/multiversx/mx-chain-crypto-go/signing/secp256k1/singlesig"
	"github.com/multiversx/mx-chain-go/common"
	cryptoCommon "github.com/multiversx/mx-chain-go/common/crypto"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/peerSignatureHandler"
	"github.com/multiversx/mx-chain-go/genesis/process/disabled"
	"github.com/multiversx/mx-chain-go/keysManagement"
	p2pFactory "github.com/multiversx/mx-chain-go/p2p/factory"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/vm"
	systemVM "github.com/multiversx/mx-chain-go/vm/process"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const (
	disabledSigChecking = "disabled"
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
	P2pKeyPemFileName                    string
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
	p2pKeyPemFileName                    string
}

// cryptoParams holds the node public/private key data
type cryptoParams struct {
	publicKey          crypto.PublicKey
	privateKey         crypto.PrivateKey
	publicKeyString    string
	publicKeyBytes     []byte
	handledPrivateKeys [][]byte
}

// p2pCryptoParams holds the p2p public/private key data
type p2pCryptoParams struct {
	p2pPublicKey  crypto.PublicKey
	p2pPrivateKey crypto.PrivateKey
}

// cryptoComponents struct holds the crypto components
type cryptoComponents struct {
	txSingleSigner          crypto.SingleSigner
	blockSingleSigner       crypto.SingleSigner
	p2pSingleSigner         crypto.SingleSigner
	multiSignerContainer    cryptoCommon.MultiSignerContainer
	peerSignHandler         crypto.PeerSignatureHandler
	blockSignKeyGen         crypto.KeyGenerator
	txSignKeyGen            crypto.KeyGenerator
	p2pKeyGen               crypto.KeyGenerator
	messageSignVerifier     vm.MessageSignVerifier
	consensusSigningHandler consensus.SigningHandler
	managedPeersHolder      common.ManagedPeersHolder
	keysHandler             consensus.KeysHandler
	cryptoParams
	p2pCryptoParams
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
		p2pKeyPemFileName:                    args.P2pKeyPemFileName,
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
	txSingleSigner := &disabledSig.DisabledSingleSig{}

	processingSingleSigner, err := ccf.createSingleSigner(true)
	if err != nil {
		return nil, err
	}

	interceptSingleSigner, err := ccf.createSingleSigner(true)
	if err != nil {
		return nil, err
	}

	p2pSingleSigner := &secp256k1SinglerSig.Secp256k1Signer{}

	multiSigner, err := ccf.createMultiSignerContainer(blockSignKeyGen, true)
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
	p2pCryptoParamsInstance, err := ccf.createP2pCryptoParams(p2pKeyGenerator)
	if err != nil {
		return nil, err
	}

	redundancyLevel := int(ccf.prefsConfig.Preferences.RedundancyLevel)
	maxRoundsOfInactivity := redundancyLevel * ccf.config.Redundancy.MaxRoundsOfInactivityAccepted
	argsManagedPeersHolder := keysManagement.ArgsManagedPeersHolder{
		KeyGenerator:          blockSignKeyGen,
		P2PKeyGenerator:       p2pKeyGenerator,
		MaxRoundsOfInactivity: maxRoundsOfInactivity,
		PrefsConfig:           ccf.prefsConfig,
		P2PKeyConverter:       p2pFactory.NewP2PKeyConverter(),
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

	currentPid, err := argsManagedPeersHolder.P2PKeyConverter.ConvertPublicKeyToPeerID(p2pCryptoParamsInstance.p2pPublicKey)
	if err != nil {
		return nil, err
	}

	argsKeysHandler := keysManagement.ArgsKeysHandler{
		ManagedPeersHolder: managedPeersHolder,
		PrivateKey:         cp.privateKey,
		Pid:                currentPid,
	}
	keysHandler, err := keysManagement.NewKeysHandler(argsKeysHandler)
	if err != nil {
		return nil, err
	}

	signingHandlerArgs := ArgsSigningHandler{
		PubKeys:              []string{cp.publicKeyString},
		MultiSignerContainer: multiSigner,
		KeyGenerator:         blockSignKeyGen,
		SingleSigner:         interceptSingleSigner,
		KeysHandler:          keysHandler,
	}
	consensusSigningHandler, err := NewSigningHandler(signingHandlerArgs)
	if err != nil {
		return nil, err
	}

	return &cryptoComponents{
		txSingleSigner:          txSingleSigner,
		blockSingleSigner:       interceptSingleSigner,
		multiSignerContainer:    multiSigner,
		peerSignHandler:         peerSigHandler,
		blockSignKeyGen:         blockSignKeyGen,
		txSignKeyGen:            txSignKeyGen,
		p2pKeyGen:               p2pKeyGenerator,
		messageSignVerifier:     messageSignVerifier,
		consensusSigningHandler: consensusSigningHandler,
		managedPeersHolder:      managedPeersHolder,
		keysHandler:             keysHandler,
		cryptoParams:            *cp,
		p2pCryptoParams:         *p2pCryptoParamsInstance,
		p2pSingleSigner:         p2pSingleSigner,
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

	handledKeysInfo := "running in single-key mode"
	if len(handledPrivateKeys) > 0 {
		handledKeysInfo = fmt.Sprintf("running in multi-key mode, managing %d keys", len(handledPrivateKeys))
	}

	if ccf.isInImportMode {
		if len(handledPrivateKeys) > 0 {
			return nil, fmt.Errorf("invalid node configuration: import-db mode and allValidatorsKeys.pem file provided")
		}

		return ccf.generateCryptoParams(keygen, "in import-db mode", make([][]byte, 0))
	}
	cp, err := ccf.readCryptoParams(keygen)
	if err == nil {
		cp.handledPrivateKeys = handledPrivateKeys

		log.Info(fmt.Sprintf("the node loaded the validatorKey.pem file and is %s", handledKeysInfo))

		return cp, nil
	}

	log.Debug("failure while reading the BLS key, will autogenerate one", "error", err)

	return ccf.generateCryptoParams(keygen, handledKeysInfo, handledPrivateKeys)
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

	cp.publicKeyString, err = ccf.validatorPubKeyConverter.Encode(cp.publicKeyBytes)
	if err != nil {
		return nil, err
	}

	return cp, nil
}

func (ccf *cryptoComponentsFactory) generateCryptoParams(
	keygen crypto.KeyGenerator,
	reason string,
	handledPrivateKeys [][]byte,
) (*cryptoParams, error) {
	log.Info(fmt.Sprintf("the node is %s! Will generate a fresh new BLS key", reason))
	cp := &cryptoParams{}
	cp.privateKey, cp.publicKey = keygen.GeneratePair()

	var err error
	cp.publicKeyBytes, err = cp.publicKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	cp.publicKeyString, err = ccf.validatorPubKeyConverter.Encode(cp.publicKeyBytes)
	if err != nil {
		return nil, err
	}
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

func (ccf *cryptoComponentsFactory) createP2pCryptoParams(
	keygen crypto.KeyGenerator,
) (*p2pCryptoParams, error) {
	privKey, pubKey, err := CreateP2pKeyPair(ccf.p2pKeyPemFileName, keygen, log)
	if err != nil {
		return nil, err
	}

	return &p2pCryptoParams{
		p2pPrivateKey: privKey,
		p2pPublicKey:  pubKey,
	}, nil
}

// CreateP2pKeyPair will create a set of key pair for p2p based on provided pem file. If
// the provided key is empty it will generate a new one
func CreateP2pKeyPair(
	keyFileName string,
	keyGen crypto.KeyGenerator,
	log logger.Logger,
) (crypto.PrivateKey, crypto.PublicKey, error) {
	privKeyBytes, err := common.GetSkBytesFromP2pKey(keyFileName)
	if err != nil {
		return nil, nil, err
	}

	if len(privKeyBytes) == 0 {
		privKey, pubKey := keyGen.GeneratePair()

		log.Info("p2p private key: generated a new private key for p2p signing")

		return privKey, pubKey, nil
	}

	privKey, err := keyGen.PrivateKeyFromByteArray(privKeyBytes)
	if err != nil {
		return nil, nil, err
	}

	log.Info("p2p private key: using the provided private key for p2p signing")

	return privKey, privKey.GeneratePublic(), nil
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

func (ccf *cryptoComponentsFactory) processPrivatePublicKey(_ crypto.KeyGenerator, encodedSk []byte, _ string, index int) ([]byte, error) {
	skBytes, err := hex.DecodeString(string(encodedSk))
	if err != nil {
		return nil, fmt.Errorf("%w for encoded secret key, key index %d", err, index)
	}

	return skBytes, nil
}

// Close closes all underlying components that need closing
func (cc *cryptoComponents) Close() error {
	return nil
}
