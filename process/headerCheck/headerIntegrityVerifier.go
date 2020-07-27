package headerCheck

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type headerIntegrityVerifier struct {
	referenceChainID []byte
}

// TODO: maybe merge this struct and header sig verifier in a single component

// NewHeaderIntegrityVerifier returns a new instance of headerIntegrityVerifier
func NewHeaderIntegrityVerifier(referenceChainID []byte) (*headerIntegrityVerifier, error) {
	if len(referenceChainID) == 0 {
		return nil, ErrInvalidReferenceChainID
	}

	return &headerIntegrityVerifier{
		referenceChainID: referenceChainID,
	}, nil
}

// Verify will check the header's fields such as the chain ID or the software version
func (h *headerIntegrityVerifier) Verify(hdr data.HeaderHandler) error {
	if len(hdr.GetReserved()) > 0 {
		return process.ErrReservedFieldNotSupportedYet
	}

	err := h.checkSoftwareVersion(hdr)
	if err != nil {
		return err
	}

	return h.checkChainID(hdr)
}

// checkSoftwareVersion returns nil if the software version has the correct length
func (h *headerIntegrityVerifier) checkSoftwareVersion(hdr data.HeaderHandler) error {
	if len(hdr.GetSoftwareVersion()) == 0 || len(hdr.GetSoftwareVersion()) > core.MaxSoftwareVersionLengthInBytes {
		return ErrInvalidSoftwareVersion
	}
	return nil
}

// checkChainID returns nil if the header's chain ID matches the one provided
// otherwise, it will error
func (h *headerIntegrityVerifier) checkChainID(hdr data.HeaderHandler) error {
	if !bytes.Equal(h.referenceChainID, hdr.GetChainID()) {
		return fmt.Errorf(
			"%w, expected: %s, got %s",
			ErrInvalidChainID,
			hex.EncodeToString(h.referenceChainID),
			hex.EncodeToString(hdr.GetChainID()),
		)
	}

	return nil
}

// IsInterfaceNil returns true if the value under the interface is nil
func (h *headerIntegrityVerifier) IsInterfaceNil() bool {
	return h == nil
}
