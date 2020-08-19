package headerCheck

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
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

func TestNewHeaderVersioningHandler_InvalidReferenceChainIDShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersioningHandler(
		nil,
		make([]config.VersionByEpochs, 0),
		defaultVersion,
		&testscommon.CacherStub{},
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
				Version:    "",
			},
			{
				StartEpoch: 0,
				Version:    "",
			},
		},
		defaultVersion,
		&testscommon.CacherStub{},
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrInvalidVersionOnEpochValues))
}

func TestNewHeaderVersioningHandler_InvalidVersionElementOnStringTooLongShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersioningHandler(
		[]byte("chainID"),
		[]config.VersionByEpochs{
			{
				StartEpoch: 0,
				Version:    strings.Repeat("a", core.MaxSoftwareVersionLengthInBytes+1),
			},
		},
		defaultVersion,
		&testscommon.CacherStub{},
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrInvalidVersionStringTooLong))
}

func TestNewHeaderVersioningHandler_InvalidDefaultVersionShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersioningHandler(
		[]byte("chainID"),
		versionsCorrectlyConstructed,
		defaultVersion,
		nil,
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrNilCacher))
}

func TestNewHeaderVersioningHandler_NilCacherShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersioningHandler(
		[]byte("chainID"),
		versionsCorrectlyConstructed,
		"",
		&testscommon.CacherStub{},
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrInvalidSoftwareVersion))
}

func TestNewHeaderVersioningHandler_EmptyListShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersioningHandler(
		[]byte("chainID"),
		make([]config.VersionByEpochs, 0),
		"",
		&testscommon.CacherStub{},
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrEmptyVersionsByEpochsList))
}

func TestNewHeaderVersioningHandler_ZerothElementIsNotOnEpochZeroShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersioningHandler(
		[]byte("chainID"),
		[]config.VersionByEpochs{
			{
				StartEpoch: 1,
				Version:    "",
			},
		},
		"",
		&testscommon.CacherStub{},
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrInvalidVersionOnEpochValues))
}

func TestNewHeaderVersioningHandler_ShouldWork(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderVersioningHandler(
		[]byte("chainID"),
		versionsCorrectlyConstructed,
		defaultVersion,
		&testscommon.CacherStub{},
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
		&testscommon.CacherStub{},
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
		&testscommon.CacherStub{},
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
				Version:    "v1",
			},
			{
				StartEpoch: 1,
				Version:    "v2",
			},
		},
		defaultVersion,
		&testscommon.CacherStub{},
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
				Version:    "v1",
			},
			{
				StartEpoch: 1,
				Version:    "*",
			},
		},
		defaultVersion,
		&testscommon.CacherStub{},
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
		versionsCorrectlyConstructed,
		"software",
		&testscommon.CacherStub{},
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
		versionsCorrectlyConstructed,
		"software",
		&testscommon.CacherStub{},
	)
	mb := &block.MetaBlock{
		SoftwareVersion: []byte("software"),
		ChainID:         expectedChainID,
	}
	err := hdrIntVer.Verify(mb)
	require.NoError(t, err)
}

func TestHeaderVersioningHandler_VerifyNotWildcardShouldWork(t *testing.T) {
	t.Parallel()

	expectedChainID := []byte("#chainID")
	hdrIntVer, _ := NewHeaderVersioningHandler(
		expectedChainID,
		versionsCorrectlyConstructed,
		"software",
		&testscommon.CacherStub{},
	)
	mb := &block.MetaBlock{
		SoftwareVersion: []byte("v1"),
		ChainID:         expectedChainID,
		Epoch:           1,
	}
	err := hdrIntVer.Verify(mb)
	require.NoError(t, err)
}

func TestHeaderVersioningHandler_GetVersionShouldWork(t *testing.T) {
	t.Parallel()

	numPutCalls := uint32(0)
	hdrIntVer, _ := NewHeaderVersioningHandler(
		[]byte("chainID"),
		versionsCorrectlyConstructed,
		defaultVersion,
		&testscommon.CacherStub{
			PutCalled: func(key []byte, value interface{}, sizeInBytes int) bool {
				atomic.AddUint32(&numPutCalls, 1)
				epoch := binary.BigEndian.Uint32(key)
				switch epoch {
				case 0:
					assert.Equal(t, "*", value.(string))
				case 1:
					assert.Equal(t, "v1", value.(string))
				case 2:
					assert.Equal(t, "v1", value.(string))
				case 3:
					assert.Equal(t, "v1", value.(string))
				case 4:
					assert.Equal(t, "v1", value.(string))
				case 5:
					assert.Equal(t, "v2", value.(string))
				case 6:
					assert.Equal(t, "v2", value.(string))
				case 1000:
					assert.Equal(t, "v2", value.(string))
				case 1200:
					assert.Equal(t, "v2", value.(string))
				default:
					assert.Fail(t, fmt.Sprintf("unexpected case for epoch %d", epoch))
				}

				return false
			},
		},
	)

	assert.Equal(t, defaultVersion, hdrIntVer.GetVersion(0))
	assert.Equal(t, "v1", hdrIntVer.GetVersion(1))
	assert.Equal(t, "v1", hdrIntVer.GetVersion(2))
	assert.Equal(t, "v1", hdrIntVer.GetVersion(3))
	assert.Equal(t, "v1", hdrIntVer.GetVersion(4))
	assert.Equal(t, "v2", hdrIntVer.GetVersion(5))
	assert.Equal(t, "v2", hdrIntVer.GetVersion(6))
	assert.Equal(t, "v2", hdrIntVer.GetVersion(1000))
	assert.Equal(t, "v2", hdrIntVer.GetVersion(1200))
	assert.Equal(t, uint32(9), atomic.LoadUint32(&numPutCalls))
}

func TestHeaderVersioningHandler_ExistsInInternalCacheShouldReturn(t *testing.T) {
	t.Parallel()

	cachedVersion := "cached version"
	hdrIntVer, _ := NewHeaderVersioningHandler(
		[]byte("chainID"),
		versionsCorrectlyConstructed,
		defaultVersion,
		&testscommon.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return cachedVersion, true
			},
		},
	)

	assert.Equal(t, cachedVersion, hdrIntVer.GetVersion(0))
	assert.Equal(t, cachedVersion, hdrIntVer.GetVersion(1))
	assert.Equal(t, cachedVersion, hdrIntVer.GetVersion(2))
	assert.Equal(t, cachedVersion, hdrIntVer.GetVersion(500))
	assert.Equal(t, cachedVersion, hdrIntVer.GetVersion(999))
	assert.Equal(t, cachedVersion, hdrIntVer.GetVersion(1000))
	assert.Equal(t, cachedVersion, hdrIntVer.GetVersion(1200))
}
