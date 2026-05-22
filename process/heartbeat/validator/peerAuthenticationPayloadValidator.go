package validator

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-go/process"
)

const (
	payloadExpiryThresholdInSec = 10
)

type peerAuthenticationPayloadValidator struct {
	getTimeHandler func() time.Time
}

// NewPeerAuthenticationPayloadValidator creates a new peer authentication payload validator instance
func NewPeerAuthenticationPayloadValidator() (*peerAuthenticationPayloadValidator, error) {
	return &peerAuthenticationPayloadValidator{
		getTimeHandler: time.Now,
	}, nil
}

// ValidateTimestamp will return an error if the provided payload timestamp is not valid
func (validator *peerAuthenticationPayloadValidator) ValidateTimestamp(payloadTimestamp int64) error {
	currentTimeStamp := validator.getTimeHandler().Unix()
	maxTimestampAllowed := currentTimeStamp + payloadExpiryThresholdInSec

	if payloadTimestamp > maxTimestampAllowed {
		return fmt.Errorf("%w message time stamp: %v, maximum: %v",
			process.ErrMessageExpired, payloadTimestamp, maxTimestampAllowed)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (validator *peerAuthenticationPayloadValidator) IsInterfaceNil() bool {
	return validator == nil
}
