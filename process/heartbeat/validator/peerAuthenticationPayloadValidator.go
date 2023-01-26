package validator

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-go/process"
)

const (
	minDurationInSec            = 10
	payloadExpiryThresholdInSec = 10
)

type peerAuthenticationPayloadValidator struct {
	expiryTimespanInSec int64
	getTimeHandler      func() time.Time
}

// NewPeerAuthenticationPayloadValidator creates a new peer authentication payload validator instance
func NewPeerAuthenticationPayloadValidator(expiryTimespanInSec int64) (*peerAuthenticationPayloadValidator, error) {
	if expiryTimespanInSec < minDurationInSec {
		return nil, process.ErrInvalidExpiryTimespan
	}

	return &peerAuthenticationPayloadValidator{
		expiryTimespanInSec: expiryTimespanInSec,
		getTimeHandler:      time.Now,
	}, nil
}

// ValidateTimestamp will return an error if the provided payload timestamp is not valid
func (validator *peerAuthenticationPayloadValidator) ValidateTimestamp(payloadTimestamp int64) error {
	currentTimeStamp := validator.getTimeHandler().Unix()
	minTimestampAllowed := currentTimeStamp - validator.expiryTimespanInSec
	maxTimestampAllowed := currentTimeStamp + payloadExpiryThresholdInSec

	if payloadTimestamp < minTimestampAllowed || payloadTimestamp > maxTimestampAllowed {
		return fmt.Errorf("%w message time stamp: %v, minimum: %v, maximum: %v",
			process.ErrMessageExpired, payloadTimestamp, minTimestampAllowed, maxTimestampAllowed)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (validator *peerAuthenticationPayloadValidator) IsInterfaceNil() bool {
	return validator == nil
}
