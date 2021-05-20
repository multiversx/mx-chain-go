package headerCheck

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHeaderIntegrityVerifier_InvalidReferenceChainIDShouldErr(t *testing.T) {
	t.Parallel()

	hvh := &testscommon.HeaderVersionHandlerMock{}
	hdrIntVer, err := NewHeaderIntegrityVerifier(
		nil,
		hvh,
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.Equal(t, ErrInvalidReferenceChainID, err)
}

func TestNewHeaderIntegrityVerifierr_NilVersionHandlerShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderIntegrityVerifier(
		[]byte("chainID"),
		nil,
	)
	require.True(t, check.IfNil(hdrIntVer))
	require.True(t, errors.Is(err, ErrNilHeaderVersionHandler))
}

func TestNewHeaderIntegrityVerifier_ShouldWork(t *testing.T) {
	t.Parallel()

	hvh := &testscommon.HeaderVersionHandlerMock{}
	hdrIntVer, err := NewHeaderIntegrityVerifier(
		[]byte("chainID"),
		hvh,
	)
	require.False(t, check.IfNil(hdrIntVer))
	require.NoError(t, err)
}

func TestHeaderIntegrityVerifier_PopulatedReservedShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.MetaBlock{
		Reserved: []byte("r"),
	}
	hvh := &testscommon.HeaderVersionHandlerMock{}
	hdrIntVer, _ := NewHeaderIntegrityVerifier(
		[]byte("chainID"),
		hvh,
	)
	err := hdrIntVer.Verify(hdr)
	require.Equal(t, process.ErrReservedFieldNotSupportedYet, err)
}

func TestHeaderIntegrityVerifier_VerifyHdrChainIDAndReferenceChainIDMismatchShouldErr(t *testing.T) {
	t.Parallel()

	hvh := &testscommon.HeaderVersionHandlerMock{}
	hdrIntVer, _ := NewHeaderIntegrityVerifier(
		[]byte("chainID"),
		hvh,
	)
	mb := &block.MetaBlock{
		SoftwareVersion: []byte("software"),
		ChainID:         []byte("different-chainID"),
	}
	err := hdrIntVer.Verify(mb)
	require.True(t, errors.Is(err, ErrInvalidChainID))
}

func TestHeaderIntegrityVerifier_VerifyShouldWork(t *testing.T) {
	t.Parallel()

	expectedChainID := []byte("#chainID")
	hvh := &testscommon.HeaderVersionHandlerMock{}
	hdrIntVer, _ := NewHeaderIntegrityVerifier(
		expectedChainID,
		hvh,
	)
	mb := &block.MetaBlock{
		SoftwareVersion: []byte("software"),
		ChainID:         expectedChainID,
	}
	err := hdrIntVer.Verify(mb)
	require.NoError(t, err)
}

func TestHeaderIntegrityVerifier_GetVersionShouldWork(t *testing.T) {
	t.Parallel()

	hvh := &testscommon.HeaderVersionHandlerMock{
		GetVersionCalled: func(epoch uint32) string {
			return "v1"
		},
	}
	hdrIntVer, _ := NewHeaderIntegrityVerifier(
		[]byte("chainID"),
		hvh,
	)

	assert.Equal(t, "v1", hdrIntVer.GetVersion(1))
}
