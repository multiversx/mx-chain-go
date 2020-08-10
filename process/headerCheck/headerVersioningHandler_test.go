package headerCheck

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var versionsCorrectlyConstructed = []config.VersionByEpochs{
	{
		StartEpoch: 0,
		EndEpoch:   1,
		Version:    "*",
	},
	{
		StartEpoch: 1,
		EndEpoch:   2,
		Version:    "v1",
	},
	{
		StartEpoch: 2,
		EndEpoch:   1000,
		Version:    "v2",
	},
}

const defaultVersion = "default"

func TestNewHeaderVersioningHandler_InvalidReferenceChainIDShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersioningHandler(
		nil,
		make([]config.VersionByEpochs, 0),
		defaultVersion,
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.Equal(t, ErrInvalidReferenceChainID, err)
}

func TestNewHeaderVersioningHandler_InvalidVersionElementOnEpochValuesEqualShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersioningHandler(
		[]byte("chainID"),
		[]config.VersionByEpochs{
			{
				StartEpoch: 0,
				EndEpoch:   0,
				Version:    "",
			},
		},
		defaultVersion,
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrInvalidVersionOnEpochValues))
}

func TestNewHeaderVersioningHandler_InvalidVersionElementOnEpochValuesLessShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersioningHandler(
		[]byte("chainID"),
		[]config.VersionByEpochs{
			{
				StartEpoch: 1,
				EndEpoch:   0,
				Version:    "",
			},
		},
		defaultVersion,
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrInvalidVersionOnEpochValues))
}

func TestNewHeaderVersioningHandler_InvalidVersionElementOnGapsShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersioningHandler(
		[]byte("chainID"),
		[]config.VersionByEpochs{
			{
				StartEpoch: 0,
				EndEpoch:   1,
				Version:    "",
			},
			{
				StartEpoch: 2,
				EndEpoch:   3,
				Version:    "",
			},
		},
		defaultVersion,
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrInvalidVersionFoundGaps))
}

func TestNewHeaderVersioningHandler_InvalidVersionElementOnInterlaceShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersioningHandler(
		[]byte("chainID"),
		[]config.VersionByEpochs{
			{
				StartEpoch: 0,
				EndEpoch:   2,
				Version:    "",
			},
			{
				StartEpoch: 1,
				EndEpoch:   3,
				Version:    "",
			},
		},
		defaultVersion,
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrInvalidVersionFoundGaps))
}

func TestNewHeaderVersioningHandler_InvalidVersionElementOnStringTooLongShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersioningHandler(
		[]byte("chainID"),
		[]config.VersionByEpochs{
			{
				StartEpoch: 0,
				EndEpoch:   100,
				Version:    strings.Repeat("a", core.MaxSoftwareVersionLengthInBytes+1),
			},
		},
		defaultVersion,
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrInvalidVersionStringTooLong))
}

func TestNewHeaderVersioningHandler_InvalidDefaultVersionShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersioningHandler(
		[]byte("chainID"),
		versionsCorrectlyConstructed,
		"",
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrInvalidSoftwareVersion))
}

func TestNewHeaderVersioningHandler_ShouldWork(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersioningHandler(
		[]byte("chainID"),
		versionsCorrectlyConstructed,
		defaultVersion,
	)
	require.False(t, check.IfNil(hdrIntVer))
	require.NoError(t, err)
}

func TestHeaderVersioningHandler_PopulatedReservedShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.MetaBlock{
		Reserved: []byte("r"),
	}
	hdrIntVer, _ := NewHeaderVersioningHandler(
		[]byte("chainID"),
		make([]config.VersionByEpochs, 0),
		defaultVersion,
	)
	err := hdrIntVer.Verify(hdr)
	require.Equal(t, process.ErrReservedFieldNotSupportedYet, err)
}

func TestHeaderVersioningHandler_VerifySoftwareVersionEmptyVersionInHeaderShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, _ := NewHeaderVersioningHandler(
		[]byte("chainID"),
		make([]config.VersionByEpochs, 0),
		defaultVersion,
	)
	err := hdrIntVer.Verify(&block.MetaBlock{})
	require.True(t, errors.Is(err, ErrInvalidSoftwareVersion))
}

func TestHeaderVersioningHandler_VerifySoftwareVersionWrongVersionShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, _ := NewHeaderVersioningHandler(
		[]byte("chainID"),
		[]config.VersionByEpochs{
			{
				StartEpoch: 0,
				EndEpoch:   1,
				Version:    "v1",
			},
			{
				StartEpoch: 1,
				EndEpoch:   2,
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

func TestHeaderVersioningHandler_VerifySoftwareVersionWildcardShouldWork(t *testing.T) {
	t.Parallel()

	hdrIntVer, _ := NewHeaderVersioningHandler(
		[]byte("chainID"),
		[]config.VersionByEpochs{
			{
				StartEpoch: 0,
				EndEpoch:   1,
				Version:    "v1",
			},
			{
				StartEpoch: 1,
				EndEpoch:   2,
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

func TestHeaderVersioningHandler_VerifyHdrChainIDAndReferenceChainIDMismatchShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, _ := NewHeaderVersioningHandler(
		[]byte("chainID"),
		make([]config.VersionByEpochs, 0),
		"software",
	)
	mb := &block.MetaBlock{
		SoftwareVersion: []byte("software"),
		ChainID:         []byte("different-chainID"),
	}
	err := hdrIntVer.Verify(mb)
	require.True(t, errors.Is(err, ErrInvalidChainID))
}

func TestHeaderVersioningHandler_VerifyShouldWork(t *testing.T) {
	t.Parallel()

	expectedChainID := []byte("#chainID")
	hdrIntVer, _ := NewHeaderVersioningHandler(
		expectedChainID,
		make([]config.VersionByEpochs, 0),
		"software",
	)
	mb := &block.MetaBlock{
		SoftwareVersion: []byte("software"),
		ChainID:         expectedChainID,
	}
	err := hdrIntVer.Verify(mb)
	require.NoError(t, err)
}

func TestHeaderVersioningHandler_GetVersionShouldWork(t *testing.T) {
	t.Parallel()

	hdrIntVer, _ := NewHeaderVersioningHandler(
		[]byte("chainID"),
		versionsCorrectlyConstructed,
		defaultVersion,
	)

	assert.Equal(t, defaultVersion, hdrIntVer.GetVersion(0))
	assert.Equal(t, "v1", hdrIntVer.GetVersion(1))
	assert.Equal(t, "v2", hdrIntVer.GetVersion(2))
	assert.Equal(t, "v2", hdrIntVer.GetVersion(500))
	assert.Equal(t, "v2", hdrIntVer.GetVersion(999))
	assert.Equal(t, defaultVersion, hdrIntVer.GetVersion(1000))
	assert.Equal(t, defaultVersion, hdrIntVer.GetVersion(1200))
}
