package crypto

import (
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	disabledMultiSig "github.com/ElrondNetwork/elrond-go-crypto/signing/disabled/multisig"
	mclMultiSig "github.com/ElrondNetwork/elrond-go-crypto/signing/mcl/multisig"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/multisig"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/errors"
)

const (
	blsNonKOSK = "non-KOSK"
	blsKOSK    = "KOSK"
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
	hasher               hashing.Hasher
	cryptoParams         *cryptoParams
	blSignKeyGen         crypto.KeyGenerator
	consensusType        string
	importModeNoSigCheck bool
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
		return nil, errors.ErrNilMultiSigner
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

// GetMultiSigner returns the multiSignerContainer configured for the given epoch
func (c *container) GetMultiSigner(epoch uint32) (crypto.MultiSigner, error) {
	c.mutSigners.RLock()
	defer c.mutSigners.RUnlock()

	for i := len(c.multiSigners) - 1; i >= 0; i-- {
		if epoch >= c.multiSigners[i].epoch {
			return c.multiSigners[i].multiSigner, nil
		}
	}
	return nil, errors.ErrMissingMultiSignerConfig
}

// IsInterfaceNil returns true if the underlying object is nil
func (c *container) IsInterfaceNil() bool {
	return c == nil
}

func createMultiSigner(multiSigType string, args MultiSigArgs) (crypto.MultiSigner, error) {
	if args.importModeNoSigCheck {
		log.Warn("using disabled multi signer because the node is running in import-db 'turbo mode'")
		return &disabledMultiSig.DisabledMultiSig{}, nil
	}

	switch args.consensusType {
	case consensus.BlsConsensusType:
		blsSigner, err := createLowLevelSigner(multiSigType, args.hasher)
		if err != nil {
			return nil, err
		}
		return multisig.NewBLSMultisig(blsSigner, []string{string(args.cryptoParams.publicKeyBytes)}, args.cryptoParams.privateKey, args.blSignKeyGen, uint16(0))
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
	case blsNonKOSK:
		return &mclMultiSig.BlsMultiSigner{Hasher: hasher}, nil
	case blsKOSK:
		return &mclMultiSig.BlsMultiSignerKOSK{}, nil
	default:
		return nil, errors.ErrSignerNotSupported
	}
}

func sortMultiSignerConfig(multiSignerConfig []config.MultiSignerConfig) []config.MultiSignerConfig {
	sortedMultiSignerConfig := append([]config.MultiSignerConfig{}, multiSignerConfig...)
	sort.Slice(sortedMultiSignerConfig, func(i, j int) bool {
		return sortedMultiSignerConfig[i].EnableEpoch < sortedMultiSignerConfig[j].EnableEpoch
	})

	return sortedMultiSignerConfig
}
