package headerCheck

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type headerIntegrityVerifier struct {
	referenceChainID []byte
	headerVersionHandler factory.HeaderVersionHandler
}

// NewHeaderIntegrityVerifier returns a new instance of a structure capable of verifying the integrity of a provided header
func NewHeaderIntegrityVerifier(
	referenceChainID []byte,
	headerVersionHandler factory.HeaderVersionHandler,
) (*headerIntegrityVerifier, error) {

	if len(referenceChainID) == 0 {
		return nil, ErrInvalidReferenceChainID
	}
	if check.IfNil(headerVersionHandler) {
		return nil, fmt.Errorf("%w, in NewHeaderVersioningHandler", ErrNilHeaderVersionHandler)
	}

	hdrIntVer := &headerIntegrityVerifier{
		referenceChainID: referenceChainID,
		headerVersionHandler: headerVersionHandler,
	}

	return hdrIntVer, nil
}

// GetVersion returns the version by providing the epoch
func (hdrIntVer *headerIntegrityVerifier) GetVersion(epoch uint32) string {
	return hdrIntVer.headerVersionHandler.GetVersion(epoch)
}

// Verify will check the header's fields such as the chain ID or the software version
func (hdrIntVer *headerIntegrityVerifier) Verify(hdr data.HeaderHandler) error {
	if len(hdr.GetReserved()) > 0 {
		return process.ErrReservedFieldNotSupportedYet
	}

	err := hdrIntVer.headerVersionHandler.Verify(hdr)
	if err != nil {
		return err
	}

	return hdrIntVer.checkChainID(hdr)
}

// checkChainID returns nil if the header's chain ID matches the one provided
// otherwise, it will error
func (hdrIntVer *headerIntegrityVerifier) checkChainID(hdr data.HeaderHandler) error {
	if !bytes.Equal(hdrIntVer.referenceChainID, hdr.GetChainID()) {
		return fmt.Errorf(
			"%w, expected: %s, got %s",
			ErrInvalidChainID,
			hex.EncodeToString(hdrIntVer.referenceChainID),
			hex.EncodeToString(hdr.GetChainID()),
		)
	}

	return nil
}

// IsInterfaceNil returns true if the value under the interface is nil
func (hdrIntVer *headerIntegrityVerifier) IsInterfaceNil() bool {
	return hdrIntVer == nil
}
