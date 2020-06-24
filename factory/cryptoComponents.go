package factory

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519/singlesig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	mclmultisig "github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/multisig"
	mclsig "github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go/genesis/process/disabled"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/vm"
	systemVM "github.com/ElrondNetwork/elrond-go/vm/process"
)

// CryptoComponentsFactoryArgs holds the arguments needed for creating crypto components
type CryptoComponentsFactoryArgs struct {
	ValidatorKeyPemFileName              string
	SkIndex                              int
	Config                               config.Config
	CoreComponentsHolder                 CoreComponentsHolder
	ActivateBLSPubKeyMessageVerification bool
}

type cryptoComponentsFactory struct {
	validatorKeyPemFileName              string
	skIndex                              int
	config                               config.Config
	coreComponentsHolder                 CoreComponentsHolder
	activateBLSPubKeyMessageVerification bool
	keyLoader                            func(string, int) ([]byte, string, error)
}

// CryptoParams holds the node public/private key data
type CryptoParams struct {
	PublicKey       crypto.PublicKey
	PrivateKey      crypto.PrivateKey
	PublicKeyString string
	PublicKeyBytes  []byte
	PrivateKeyBytes []byte
}

// cryptoComponents struct holds the crypto components
type cryptoComponents struct {
	TxSingleSigner      crypto.SingleSigner
	SingleSigner        crypto.SingleSigner
	MultiSigner         crypto.MultiSigner
	BlockSignKeyGen     crypto.KeyGenerator
	TxSignKeyGen        crypto.KeyGenerator
	MessageSignVerifier vm.MessageSignVerifier
	CryptoParams
}

// NewCryptoComponentsFactory returns a new crypto components factory
func NewCryptoComponentsFactory(args CryptoComponentsFactoryArgs) (*cryptoComponentsFactory, error) {
	if check.IfNil(args.CoreComponentsHolder) {
		return nil, ErrNilCoreComponents
	}
	if len(args.ValidatorKeyPemFileName) == 0 {
		return nil, ErrNilPath
	}

	return &cryptoComponentsFactory{
		validatorKeyPemFileName:              args.ValidatorKeyPemFileName,
		skIndex:                              args.SkIndex,
		config:                               args.Config,
		coreComponentsHolder:                 args.CoreComponentsHolder,
		activateBLSPubKeyMessageVerification: args.ActivateBLSPubKeyMessageVerification,
		keyLoader:                            core.LoadSkPkFromPemFile,
	}, nil
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
	singleSigner, err := ccf.createSingleSigner()
	if err != nil {
		return nil, err
	}

	multisigHasher, err := ccf.getMultiSigHasherFromConfig()
	if err != nil {
		return nil, err
	}

	multiSigner, err := ccf.createMultiSigner(multisigHasher, cp, blockSignKeyGen)
	if err != nil {
		return nil, err
	}

	var messageSignVerifier vm.MessageSignVerifier
	if ccf.activateBLSPubKeyMessageVerification {
		messageSignVerifier, err = systemVM.NewMessageSigVerifier(blockSignKeyGen, singleSigner)
		if err != nil {
			return nil, err
		}
	} else {
		messageSignVerifier, err = disabled.NewMessageSignVerifier(blockSignKeyGen)
		if err != nil {
			return nil, err
		}
	}

	return &cryptoComponents{
		TxSingleSigner:      txSingleSigner,
		SingleSigner:        singleSigner,
		MultiSigner:         multiSigner,
		BlockSignKeyGen:     blockSignKeyGen,
		TxSignKeyGen:        txSignKeyGen,
		MessageSignVerifier: messageSignVerifier,
		CryptoParams:        *cp,
	}, nil
}

func (ccf *cryptoComponentsFactory) createSingleSigner() (crypto.SingleSigner, error) {
	switch ccf.config.Consensus.Type {
	case consensus.BlsConsensusType:
		return &mclsig.BlsSingleSigner{}, nil
	default:
		return nil, ErrInvalidConsensusConfig
	}
}

func (ccf *cryptoComponentsFactory) getMultiSigHasherFromConfig() (hashing.Hasher, error) {
	if ccf.config.Consensus.Type == consensus.BlsConsensusType && ccf.config.MultisigHasher.Type != "blake2b" {
		return nil, ErrMultiSigHasherMissmatch
	}

	switch ccf.config.MultisigHasher.Type {
	case "sha256":
		return sha256.Sha256{}, nil
	case "blake2b":
		if ccf.config.Consensus.Type == consensus.BlsConsensusType {
			return &blake2b.Blake2b{HashSize: multisig.BlsHashSize}, nil
		}
		return &blake2b.Blake2b{}, nil
	}

	return nil, ErrMissingMultiHasherConfig
}

func (ccf *cryptoComponentsFactory) createMultiSigner(
	hasher hashing.Hasher,
	cp *CryptoParams,
	blSignKeyGen crypto.KeyGenerator,
) (crypto.MultiSigner, error) {
	switch ccf.config.Consensus.Type {
	case consensus.BlsConsensusType:
		blsSigner := &mclmultisig.BlsMultiSigner{Hasher: hasher}
		return multisig.NewBLSMultisig(blsSigner, []string{string(cp.PublicKeyBytes)}, cp.PrivateKey, blSignKeyGen, uint16(0))
	default:
		return nil, ErrInvalidConsensusConfig
	}
}

func (ccf *cryptoComponentsFactory) getSuite() (crypto.Suite, error) {
	switch ccf.config.Consensus.Type {
	case consensus.BlsConsensusType:
		return mcl.NewSuiteBLS12(), nil
	default:
		return nil, ErrInvalidConsensusConfig
	}
}

func (ccf *cryptoComponentsFactory) createCryptoParams(
	keygen crypto.KeyGenerator,
) (*CryptoParams, error) {

	cryptoParams := &CryptoParams{}
	sk, readPk, err := ccf.getSkPk()
	if err != nil {
		return nil, err
	}

	cryptoParams.PrivateKey, err = keygen.PrivateKeyFromByteArray(sk)
	if err != nil {
		return nil, err
	}

	cryptoParams.PublicKey = cryptoParams.PrivateKey.GeneratePublic()
	if len(readPk) > 0 {

		cryptoParams.PublicKeyBytes, err = cryptoParams.PublicKey.ToByteArray()
		if err != nil {
			return nil, err
		}

		if !bytes.Equal(cryptoParams.PublicKeyBytes, readPk) {
			return nil, ErrPublicKeyMismatch
		}
	}

	validatorKeyConverter := ccf.coreComponentsHolder.ValidatorPubKeyConverter()
	cryptoParams.PublicKeyString = validatorKeyConverter.Encode(cryptoParams.PublicKeyBytes)

	return cryptoParams, nil
}

func (ccf *cryptoComponentsFactory) getSkPk() ([]byte, []byte, error) {
	encodedSk, pkString, err := ccf.keyLoader(ccf.validatorKeyPemFileName, ccf.skIndex)
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
