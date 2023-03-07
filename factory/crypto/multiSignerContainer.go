package crypto

import (
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	disabledMultiSig "github.com/multiversx/mx-chain-crypto-go/signing/disabled/multisig"
	mclMultiSig "github.com/multiversx/mx-chain-crypto-go/signing/mcl/multisig"
	"github.com/multiversx/mx-chain-crypto-go/signing/multisig"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/errors"
)

const (
	blsNoKOSK = "no-KOSK"
	blsKOSK   = "KOSK"
)

type epochMultiSigner struct {
	epoch       uint32
	multiSigner crypto.MultiSigner
}

type container struct {
	multiSigners []*epochMultiSigner
	mutSigners   sync.RWMutex
}

// MultiSigArgs holds the arguments for creating the multiSignerContainer container
type MultiSigArgs struct {
	MultiSigHasherType   string
	BlSignKeyGen         crypto.KeyGenerator
	ConsensusType        string
	ImportModeNoSigCheck bool
}

// NewMultiSignerContainer creates the multiSignerContainer container
func NewMultiSignerContainer(args MultiSigArgs, multiSignerConfig []config.MultiSignerConfig) (*container, error) {
	if len(multiSignerConfig) == 0 {
		return nil, errors.ErrMissingMultiSignerConfig
	}

	c := &container{
		multiSigners: make([]*epochMultiSigner, len(multiSignerConfig)),
	}

	sortedMultiSignerConfig := sortMultiSignerConfig(multiSignerConfig)
	if sortedMultiSignerConfig[0].EnableEpoch != 0 {
		return nil, errors.ErrMissingEpochZeroMultiSignerConfig
	}

	for i, mConfig := range sortedMultiSignerConfig {
		multiSigner, err := createMultiSigner(mConfig.Type, args)
		if err != nil {
			return nil, err
		}

		c.multiSigners[i] = &epochMultiSigner{
			multiSigner: multiSigner,
			epoch:       mConfig.EnableEpoch,
		}
	}

	return c, nil
}

// GetMultiSigner returns the multiSigner configured for the given epoch
func (c *container) GetMultiSigner(epoch uint32) (crypto.MultiSigner, error) {
	c.mutSigners.RLock()
	defer c.mutSigners.RUnlock()

	for i := len(c.multiSigners) - 1; i >= 0; i-- {
		if epoch >= c.multiSigners[i].epoch {
			return c.multiSigners[i].multiSigner, nil
		}
	}
	return nil, errors.ErrMissingMultiSigner
}

// IsInterfaceNil returns true if the underlying object is nil
func (c *container) IsInterfaceNil() bool {
	return c == nil
}

func createMultiSigner(multiSigType string, args MultiSigArgs) (crypto.MultiSigner, error) {
	if args.ImportModeNoSigCheck {
		log.Warn("using disabled multi signer because the node is running in import-db 'turbo mode'")
		return &disabledMultiSig.DisabledMultiSig{}, nil
	}

	switch args.ConsensusType {
	case consensus.BlsConsensusType:
		hasher, err := getMultiSigHasherFromConfig(args)
		if err != nil {
			return nil, err
		}
		blsSigner, err := createLowLevelSigner(multiSigType, hasher)
		if err != nil {
			return nil, err
		}
		return multisig.NewBLSMultisig(blsSigner, args.BlSignKeyGen)
	case disabledSigChecking:
		log.Warn("using disabled multi signer")
		return &disabledMultiSig.DisabledMultiSig{}, nil
	default:
		return nil, errors.ErrInvalidConsensusConfig
	}
}

func createLowLevelSigner(multiSigType string, hasher hashing.Hasher) (crypto.LowLevelSignerBLS, error) {
	if check.IfNil(hasher) {
		return nil, errors.ErrNilHasher
	}

	switch multiSigType {
	case blsNoKOSK:
		return &mclMultiSig.BlsMultiSigner{Hasher: hasher}, nil
	case blsKOSK:
		return &mclMultiSig.BlsMultiSignerKOSK{}, nil
	default:
		return nil, errors.ErrSignerNotSupported
	}
}

func getMultiSigHasherFromConfig(args MultiSigArgs) (hashing.Hasher, error) {
	if args.ConsensusType == consensus.BlsConsensusType && args.MultiSigHasherType != "blake2b" {
		return nil, errors.ErrMultiSigHasherMissmatch
	}

	switch args.MultiSigHasherType {
	case "sha256":
		return sha256.NewSha256(), nil
	case "blake2b":
		if args.ConsensusType == consensus.BlsConsensusType {
			return blake2b.NewBlake2bWithSize(mclMultiSig.HasherOutputSize)
		}
		return blake2b.NewBlake2b(), nil
	}

	return nil, errors.ErrMissingMultiHasherConfig
}

func sortMultiSignerConfig(multiSignerConfig []config.MultiSignerConfig) []config.MultiSignerConfig {
	sortedMultiSignerConfig := append([]config.MultiSignerConfig{}, multiSignerConfig...)
	sort.Slice(sortedMultiSignerConfig, func(i, j int) bool {
		return sortedMultiSignerConfig[i].EnableEpoch < sortedMultiSignerConfig[j].EnableEpoch
	})

	return sortedMultiSignerConfig
}
