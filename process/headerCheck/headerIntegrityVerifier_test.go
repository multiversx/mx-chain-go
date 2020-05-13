package headerCheck

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/require"
)

func TestNewHeaderIntegrityVerifier(t *testing.T) {
	t.Parallel()

	hdrIntVer := NewHeaderIntegrityVerifier()
	require.False(t, check.IfNil(hdrIntVer))
}

func TestHeaderIntegrityVerifier_VerifySoftwareVersionShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer := NewHeaderIntegrityVerifier()
	err := hdrIntVer.Verify(&block.MetaBlock{}, []byte("cID"))
	require.Equal(t, ErrInvalidSoftwareVersion, err)
}

func TestHeaderIntegrityVerifier_VerifyInvalidReferenceChainIDShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer := NewHeaderIntegrityVerifier()
	err := hdrIntVer.Verify(&block.MetaBlock{SoftwareVersion: []byte("version")}, nil)
	require.Equal(t, ErrInvalidReferenceChainID, err)
}

func TestHeaderIntegrityVerifier_VerifyHdrChainIDAndReferenceChainIDMissmatchShouldErr(t *testing.T) {
	t.Parallel()

	hdrIntVer := NewHeaderIntegrityVerifier()
	mb := &block.MetaBlock{
		SoftwareVersion: []byte("software"),
		ChainID:         []byte("chainID-0"),
	}
	err := hdrIntVer.Verify(mb, []byte("chainID-1"))
	require.True(t, errors.Is(err, ErrInvalidChainID))
}

func TestHeaderIntegrityVerifier_VerifyShouldWork(t *testing.T) {
	t.Parallel()

	expectedChainID := []byte("#chainID")
	hdrIntVer := NewHeaderIntegrityVerifier()
	mb := &block.MetaBlock{
		SoftwareVersion: []byte("software"),
		ChainID:         expectedChainID,
	}
	err := hdrIntVer.Verify(mb, expectedChainID)
	require.NoError(t, err)
}
