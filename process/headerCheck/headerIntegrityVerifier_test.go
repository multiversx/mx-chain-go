package headerCheck

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/require"
)

func TestNewHeaderIntegrityVerifier_InvalidReferenceChainIDShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderIntegrityVerifier(nil)
	require.True(t, check.IfNil(hdrIntVer))
	require.Equal(t, ErrInvalidReferenceChainID, err)
}

func TestNewHeaderIntegrityVerifier_ShouldWork(t *testing.T) {
	t.Parallel()

	hdrIntVer, err := NewHeaderIntegrityVerifier([]byte("chainID"))
	require.False(t, check.IfNil(hdrIntVer))
	require.NoError(t, err)
}

func TestHeaderIntegrityVerifier_PopulatedReservedShouldErr(t *testing.T) {
	t.Parallel()

	hdr := &block.MetaBlock{
		Reserved: []byte("r"),
	}
	hdrIntVer, _ := NewHeaderIntegrityVerifier([]byte("chainID"))
	err := hdrIntVer.Verify(hdr)
	require.Equal(t, process.ErrReservedFieldNotSupportedYet, err)
}

func TestHeaderIntegrityVerifier_VerifySoftwareVersionShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, _ := NewHeaderIntegrityVerifier([]byte("chainID"))
	err := hdrIntVer.Verify(&block.MetaBlock{})
	require.Equal(t, ErrInvalidSoftwareVersion, err)
}

func TestHeaderIntegrityVerifier_VerifyHdrChainIDAndReferenceChainIDMissmatchShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer, _ := NewHeaderIntegrityVerifier([]byte("chainID"))
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
	hdrIntVer, _ := NewHeaderIntegrityVerifier(expectedChainID)
	mb := &block.MetaBlock{
		SoftwareVersion: []byte("software"),
		ChainID:         expectedChainID,
	}
	err := hdrIntVer.Verify(mb)
	require.NoError(t, err)
}
