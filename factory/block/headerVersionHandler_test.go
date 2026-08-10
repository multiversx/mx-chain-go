package block

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var versionsCorrectlyConstructed = []config.VersionByEpochs{
	{
		StartEpoch: 0,
		Version:    "*",
	},
	{
		StartEpoch: 1,
		Version:    "v1",
	},
	{
		StartEpoch: 5,
		Version:    "v2",
	},
}

const defaultVersion = "default"

func TestNewHeaderIntegrityVerifierr_InvalidVersionElementOnEpochValuesEqualShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersionHandler(
		[]config.VersionByEpochs{
			{
				StartEpoch: 0,
				Version:    "",
			},
			{
				StartEpoch: 0,
				Version:    "",
			},
		},
		defaultVersion,
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrInvalidVersionOnEpochValues))
}

func TestNewHeaderIntegrityVerifier_InvalidVersionElementOnStringTooLongShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersionHandler(
		[]config.VersionByEpochs{
			{
				StartEpoch: 0,
				Version:    strings.Repeat("a", common.MaxSoftwareVersionLengthInBytes+1),
			},
		},
		defaultVersion,
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrInvalidVersionStringTooLong))
}

func TestNewHeaderIntegrityVerifierr_InvalidDefaultVersionShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersionHandler(
		versionsCorrectlyConstructed,
		"",
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrInvalidSoftwareVersion))
}

func TestNewHeaderIntegrityVerifier_EmptyListShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersionHandler(
		make([]config.VersionByEpochs, 0),
		defaultVersion,
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrEmptyVersionsByEpochsList))
}

func TestNewHeaderIntegrityVerifier_ZerothElementIsNotOnEpochZeroShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersionHandler(
		[]config.VersionByEpochs{
			{
				StartEpoch: 1,
				Version:    "",
			},
		},
		defaultVersion,
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrInvalidVersionOnEpochValues))
}

func TestNewHeaderIntegrityVerifier_ShouldWork(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersionHandler(
		versionsCorrectlyConstructed,
		defaultVersion,
	)
	require.False(t, check.IfNil(hdrIntVer))
	require.NoError(t, err)
}

func TestHeaderIntegrityVerifier_PopulatedReservedShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.MetaBlock{
		Reserved: []byte("r"),
	}
	hdrIntVer, _ := NewHeaderVersionHandler(
		make([]config.VersionByEpochs, 0),
		defaultVersion,
	)
	err := hdrIntVer.Verify(hdr)
	require.Equal(t, process.ErrReservedFieldInvalid, err)
}

func TestHeaderIntegrityVerifier_VerifySoftwareVersionEmptyVersionInHeaderShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, _ := NewHeaderVersionHandler(
		make([]config.VersionByEpochs, 0),
		defaultVersion,
	)
	err := hdrIntVer.Verify(&block.MetaBlock{})
	require.True(t, errors.Is(err, ErrInvalidSoftwareVersion))
}

func TestHeaderIntegrityVerifierr_VerifySoftwareVersionWrongVersionShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, _ := NewHeaderVersionHandler(
		[]config.VersionByEpochs{
			{
				StartEpoch: 0,
				Version:    "v1",
			},
			{
				StartEpoch: 1,
				Version:    "v2",
			},
		},
		defaultVersion,
	)
	err := hdrIntVer.Verify(
		&block.MetaBlock{
			ChainID:         []byte("chainID"),
			SoftwareVersion: []byte("v3"),
			Epoch:           1,
		},
	)
	require.True(t, errors.Is(err, ErrSoftwareVersionMismatch))
}

func TestHeaderIntegrityVerifier_VerifySoftwareVersionWildcardShouldWork(t *testing.T) {
	t.Parallel()

	hdrIntVer, _ := NewHeaderVersionHandler(
		[]config.VersionByEpochs{
			{
				StartEpoch: 0,
				Version:    "v1",
			},
			{
				StartEpoch: 1,
				Version:    "*",
			},
		},
		defaultVersion,
	)
	err := hdrIntVer.Verify(
		&block.MetaBlock{
			ChainID:         []byte("chainID"),
			SoftwareVersion: []byte("v3"),
			Epoch:           1,
		},
	)

	assert.Nil(t, err)
}

func TestHeaderIntegrityVerifier_VerifyShouldWork(t *testing.T) {
	t.Parallel()

	expectedChainID := []byte("#chainID")
	hdrIntVer, _ := NewHeaderVersionHandler(
		versionsCorrectlyConstructed,
		"software",
	)
	mb := &block.MetaBlock{
		SoftwareVersion: []byte("software"),
		ChainID:         expectedChainID,
	}
	err := hdrIntVer.Verify(mb)
	require.NoError(t, err)
}

func TestHeaderIntegrityVerifier_VerifyNotWildcardShouldWork(t *testing.T) {
	t.Parallel()

	hdrIntVer, _ := NewHeaderVersionHandler(
		versionsCorrectlyConstructed,
		"software",
	)
	mb := &block.MetaBlock{
		SoftwareVersion: []byte("v1"),
		Epoch:           1,
	}
	err := hdrIntVer.Verify(mb)
	require.NoError(t, err)
}

func TestHeaderIntegrityVerifier_GetVersionShouldWork(t *testing.T) {
	t.Parallel()

	hdrIntVer, _ := NewHeaderVersionHandler(
		versionsCorrectlyConstructed,
		defaultVersion,
	)

	assert.Equal(t, defaultVersion, hdrIntVer.GetVersion(0, 1))
	assert.Equal(t, "v1", hdrIntVer.GetVersion(1, 1))
	assert.Equal(t, "v1", hdrIntVer.GetVersion(2, 1))
	assert.Equal(t, "v1", hdrIntVer.GetVersion(3, 1))
	assert.Equal(t, "v1", hdrIntVer.GetVersion(4, 1))
	assert.Equal(t, "v2", hdrIntVer.GetVersion(5, 1))
	assert.Equal(t, "v2", hdrIntVer.GetVersion(6, 1))
	assert.Equal(t, "v2", hdrIntVer.GetVersion(1000, 1))
	assert.Equal(t, "v2", hdrIntVer.GetVersion(1200, 1))
}

func TestHeaderIntegrityVerifier_NewHeaderActivatedByRound(t *testing.T) {
	t.Parallel()

	versionsByEpoch := []config.VersionByEpochs{
		{
			StartEpoch: 0,
			Version:    "*",
		},
		{
			StartEpoch: 1,
			Version:    "v1",
		},
		{
			StartEpoch: 5,
			StartRound: 150,
			Version:    "v2",
		},
	}

	hdrIntVer, _ := NewHeaderVersionHandler(
		versionsByEpoch,
		defaultVersion,
	)

	require.Equal(t, defaultVersion, hdrIntVer.GetVersion(0, 1))
	require.Equal(t, "v1", hdrIntVer.GetVersion(5, 10))
	require.Equal(t, "v2", hdrIntVer.GetVersion(5, 150))
}
