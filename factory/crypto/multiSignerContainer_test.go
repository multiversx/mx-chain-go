package crypto

import (
	"math/rand"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	disabledMultiSig "github.com/multiversx/mx-chain-crypto-go/signing/disabled/multisig"
	mclMultiSig "github.com/multiversx/mx-chain-crypto-go/signing/mcl/multisig"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewMultiSignerContainer(t *testing.T) {
	t.Parallel()

	args := createDefaultMultiSignerArgs()
	multiSigConfig := createDefaultMultiSignerConfig()

	t.Run("nil multiSigner config should err", func(t *testing.T) {
		multiSigContainer, err := NewMultiSignerContainer(args, nil)

		require.Nil(t, multiSigContainer)
		require.Equal(t, errors.ErrMissingMultiSignerConfig, err)
	})
	t.Run("missing epoch 0 config should err", func(t *testing.T) {
		multiSigConfigClone := append([]config.MultiSignerConfig{}, multiSigConfig...)
		multiSigConfigClone[0].EnableEpoch = 1
		multiSigContainer, err := NewMultiSignerContainer(args, multiSigConfigClone)

		require.Nil(t, multiSigContainer)
		require.Equal(t, errors.ErrMissingEpochZeroMultiSignerConfig, err)
	})
	t.Run("invalid multiSigner type should err", func(t *testing.T) {
		multiSigConfigClone := append([]config.MultiSignerConfig{}, multiSigConfig...)
		multiSigConfigClone[1].Type = "invalid type"
		multiSigContainer, err := NewMultiSignerContainer(args, multiSigConfigClone)

		require.Nil(t, multiSigContainer)
		require.Equal(t, errors.ErrSignerNotSupported, err)
	})
	t.Run("valid params", func(t *testing.T) {
		multiSigContainer, err := NewMultiSignerContainer(args, multiSigConfig)

		require.Nil(t, err)
		require.NotNil(t, multiSigContainer)
	})
}

func TestContainer_GetMultiSigner(t *testing.T) {
	t.Parallel()

	args := createDefaultMultiSignerArgs()
	multiSigConfig := createDefaultMultiSignerConfig()

	t.Run("missing epoch config should err (can only happen if epoch 0 is missing)", func(t *testing.T) {
		multiSigContainer, _ := NewMultiSignerContainer(args, multiSigConfig)
		multiSigContainer.multiSigners[0].epoch = 1

		multiSigner, err := multiSigContainer.GetMultiSigner(0)
		require.Nil(t, multiSigner)
		require.Equal(t, errors.ErrMissingMultiSigner, err)
	})
	t.Run("get multi signer OK", func(t *testing.T) {
		multiSigContainer, _ := NewMultiSignerContainer(args, multiSigConfig)

		for i := uint32(0); i < 10; i++ {
			multiSigner, err := multiSigContainer.GetMultiSigner(i)
			require.Nil(t, err)
			require.Equal(t, multiSigContainer.multiSigners[0].multiSigner, multiSigner)
		}
		for i := uint32(10); i < 30; i++ {
			multiSigner, err := multiSigContainer.GetMultiSigner(i)
			require.Nil(t, err)
			require.Equal(t, multiSigContainer.multiSigners[1].multiSigner, multiSigner)
		}
	})
}

func TestContainer_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var msc *container
	assert.True(t, check.IfNil(msc))

	args := createDefaultMultiSignerArgs()
	multiSigConfig := createDefaultMultiSignerConfig()

	msc, _ = NewMultiSignerContainer(args, multiSigConfig)
	assert.False(t, check.IfNil(msc))
}

func TestContainer_createMultiSigner(t *testing.T) {
	t.Parallel()

	t.Run("create disabled multi signer", func(t *testing.T) {
		args := createDefaultMultiSignerArgs()
		args.ImportModeNoSigCheck = true
		multiSigType := "KOSK"
		multiSigner, err := createMultiSigner(multiSigType, args)
		require.Nil(t, err)
		_, ok := multiSigner.(*disabledMultiSig.DisabledMultiSig)
		require.True(t, ok)
	})
	t.Run("invalid consensus config", func(t *testing.T) {
		args := createDefaultMultiSignerArgs()
		args.ConsensusType = "invalid"
		multiSigType := "KOSK"
		multiSigner, err := createMultiSigner(multiSigType, args)
		require.Nil(t, multiSigner)
		require.Equal(t, errors.ErrInvalidConsensusConfig, err)
	})
	t.Run("bls consensus type invalid hasher config", func(t *testing.T) {
		args := createDefaultMultiSignerArgs()
		args.ConsensusType = consensus.BlsConsensusType
		args.MultiSigHasherType = "sha256"
		multiSigType := "KOSK"
		multiSigner, err := createMultiSigner(multiSigType, args)
		require.Nil(t, multiSigner)
		require.Equal(t, errors.ErrMultiSigHasherMissmatch, err)
	})
	t.Run("bls consensus type signer not supported", func(t *testing.T) {
		args := createDefaultMultiSignerArgs()
		args.ConsensusType = consensus.BlsConsensusType
		args.MultiSigHasherType = "blake2b"
		multiSigType := "not supported"
		multiSigner, err := createMultiSigner(multiSigType, args)
		require.Nil(t, multiSigner)
		require.Equal(t, errors.ErrSignerNotSupported, err)
	})
	t.Run("bls consensus type KOSK OK", func(t *testing.T) {
		args := createDefaultMultiSignerArgs()
		args.ConsensusType = consensus.BlsConsensusType
		args.MultiSigHasherType = "blake2b"
		multiSigType := blsKOSK
		multiSigner, err := createMultiSigner(multiSigType, args)
		require.Nil(t, err)
		require.NotNil(t, multiSigner)
	})
	t.Run("bls consensus type no-KOSK OK", func(t *testing.T) {
		args := createDefaultMultiSignerArgs()
		args.ConsensusType = consensus.BlsConsensusType
		args.MultiSigHasherType = "blake2b"
		multiSigType := blsNoKOSK
		multiSigner, err := createMultiSigner(multiSigType, args)
		require.Nil(t, err)
		require.NotNil(t, multiSigner)
	})
	t.Run("disabledSigChecking", func(t *testing.T) {
		args := createDefaultMultiSignerArgs()
		args.ConsensusType = disabledSigChecking
		multiSigType := blsNoKOSK
		multiSigner, err := createMultiSigner(multiSigType, args)
		require.Nil(t, err)
		require.NotNil(t, multiSigner)

		_, ok := multiSigner.(*disabledMultiSig.DisabledMultiSig)
		require.True(t, ok)
	})
}

func TestContainer_createLowLevelSigner(t *testing.T) {
	t.Parallel()

	hasher := &hashingMocks.HasherMock{}
	t.Run("nil hasher should err", func(t *testing.T) {
		llSig, err := createLowLevelSigner(blsKOSK, nil)
		require.Nil(t, llSig)
		require.Equal(t, errors.ErrNilHasher, err)
	})
	t.Run("not supported multiSig type should err", func(t *testing.T) {
		llSig, err := createLowLevelSigner("not supported", hasher)
		require.Nil(t, llSig)
		require.Equal(t, errors.ErrSignerNotSupported, err)
	})
	t.Run("multiSig of type no KOSK", func(t *testing.T) {
		llSig, err := createLowLevelSigner(blsNoKOSK, hasher)
		require.Nil(t, err)
		_, ok := llSig.(*mclMultiSig.BlsMultiSigner)
		require.True(t, ok)
	})
	t.Run("multiSig of type KOSK", func(t *testing.T) {
		llSig, err := createLowLevelSigner(blsKOSK, hasher)
		require.Nil(t, err)
		_, ok := llSig.(*mclMultiSig.BlsMultiSignerKOSK)
		require.True(t, ok)
	})
}

func TestContainer_getMultiSigHasherFromConfig(t *testing.T) {
	t.Parallel()

	t.Run("mismatch config consensus type and hasher type", func(t *testing.T) {
		args := createDefaultMultiSignerArgs()
		args.ConsensusType = consensus.BlsConsensusType
		args.MultiSigHasherType = "sha256"
		hasher, err := getMultiSigHasherFromConfig(args)
		require.Nil(t, hasher)
		require.Equal(t, errors.ErrMultiSigHasherMissmatch, err)
	})
	t.Run("sha256 config", func(t *testing.T) {
		args := createDefaultMultiSignerArgs()
		args.ConsensusType = "dummy config"
		args.MultiSigHasherType = "sha256"
		hasher, err := getMultiSigHasherFromConfig(args)
		require.Nil(t, err)
		require.NotNil(t, hasher)
	})
	t.Run("invalid hasher config", func(t *testing.T) {
		args := createDefaultMultiSignerArgs()
		args.ConsensusType = "dummy config"
		args.MultiSigHasherType = "unknown"
		hasher, err := getMultiSigHasherFromConfig(args)
		require.Nil(t, hasher)
		require.Equal(t, errors.ErrMissingMultiHasherConfig, err)
	})
	t.Run("blake2b config and bls consensus", func(t *testing.T) {
		args := createDefaultMultiSignerArgs()
		args.ConsensusType = consensus.BlsConsensusType
		args.MultiSigHasherType = "blake2b"
		hasher, err := getMultiSigHasherFromConfig(args)
		require.Nil(t, err)
		require.NotNil(t, hasher)
	})
	t.Run("blake2b config and non-bls consensus", func(t *testing.T) {
		args := createDefaultMultiSignerArgs()
		args.ConsensusType = "dummy config"
		args.MultiSigHasherType = "blake2b"
		hasher, err := getMultiSigHasherFromConfig(args)
		require.Nil(t, err)
		require.NotNil(t, hasher)
	})
}

func TestContainer_sortMultiSignerConfig(t *testing.T) {
	multiSignersOrderedConfig := []config.MultiSignerConfig{
		{
			EnableEpoch: 2,
			Type:        "KOSK",
		},
		{
			EnableEpoch: 10,
			Type:        "no-KOSK",
		},
		{
			EnableEpoch: 100,
			Type:        "BN",
		},
		{
			EnableEpoch: 200,
			Type:        "DUMMY",
		},
	}

	for i := 0; i < 20; i++ {
		shuffledConfig := append([]config.MultiSignerConfig{}, multiSignersOrderedConfig...)
		rand.Shuffle(len(shuffledConfig), func(i, j int) {
			shuffledConfig[i], shuffledConfig[j] = shuffledConfig[j], shuffledConfig[i]
		})
		sortedConfig := sortMultiSignerConfig(shuffledConfig)
		require.Equal(t, multiSignersOrderedConfig, sortedConfig)
	}
}

func Test_getMultiSigHasherFromConfigInvalidHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultMultiSignerArgs()
	args.ConsensusType = ""
	args.MultiSigHasherType = ""

	multiSigHasher, err := getMultiSigHasherFromConfig(args)
	require.Nil(t, multiSigHasher)
	require.Equal(t, errors.ErrMissingMultiHasherConfig, err)
}

func Test_getMultiSigHasherFromConfigMismatchConsensusTypeMultiSigHasher(t *testing.T) {
	t.Parallel()

	args := createDefaultMultiSignerArgs()
	args.MultiSigHasherType = "sha256"

	multiSigHasher, err := getMultiSigHasherFromConfig(args)
	require.Nil(t, multiSigHasher)
	require.Equal(t, errors.ErrMultiSigHasherMissmatch, err)
}

func Test_getMultiSigHasherFromConfigOK(t *testing.T) {
	t.Parallel()

	args := createDefaultMultiSignerArgs()
	args.ConsensusType = "bls"
	args.MultiSigHasherType = "blake2b"

	multiSigHasher, err := getMultiSigHasherFromConfig(args)
	require.Nil(t, err)
	require.NotNil(t, multiSigHasher)
}

func createDefaultMultiSignerArgs() MultiSigArgs {
	return MultiSigArgs{
		MultiSigHasherType:   "blake2b",
		BlSignKeyGen:         &cryptoMocks.KeyGenStub{},
		ConsensusType:        "bls",
		ImportModeNoSigCheck: false,
	}
}

func createDefaultMultiSignerConfig() []config.MultiSignerConfig {
	return []config.MultiSignerConfig{
		{
			EnableEpoch: 0,
			Type:        "no-KOSK",
		},
		{
			EnableEpoch: 10,
			Type:        "KOSK",
		},
	}
}
