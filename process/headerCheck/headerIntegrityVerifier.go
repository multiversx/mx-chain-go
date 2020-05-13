package headerCheck

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

type headerIntegrityVerifier struct {
}

// NewHeaderIntegrityVerifier returns a new instance of headerIntegrityVerifier
func NewHeaderIntegrityVerifier() *headerIntegrityVerifier {
	return &headerIntegrityVerifier{}
}

// Verify will check the header's fields such as the chain ID or the software version
func (h *headerIntegrityVerifier) Verify(hdr data.HeaderHandler, referenceChainID []byte) error {
	err := h.checkSoftwareVersion(hdr)
	if err != nil {
		return err
	}

	return h.checkChainID(hdr, referenceChainID)
}

// CheckSoftwareVersion returns nil if the software version has the correct length
func (h *headerIntegrityVerifier) checkSoftwareVersion(hdr data.HeaderHandler) error {
	if len(hdr.GetSoftwareVersion()) == 0 || len(hdr.GetSoftwareVersion()) > core.MaxSoftwareVersionLengthInBytes {
		return ErrInvalidSoftwareVersion
	}
	return nil
}

// CheckChainID returns nil if the header's chain ID matches the one provided
// otherwise, it will error
func (h *headerIntegrityVerifier) checkChainID(hdr data.HeaderHandler, reference []byte) error {
	if len(reference) == 0 {
		return ErrInvalidReferenceChainID
	}
	if !bytes.Equal(reference, hdr.GetChainID()) {
		return fmt.Errorf(
			"%w, expected: %s, got %s",
			ErrInvalidChainID,
			hex.EncodeToString(reference),
			hex.EncodeToString(hdr.GetChainID()),
		)
	}

	return nil
}

// IsInterfaceNil returns true if the value under the interface is nil
func (h *headerIntegrityVerifier) IsInterfaceNil() bool {
	return h == nil
}
