package validator

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/stretchr/testify/assert"
)

func TestPeerAuthenticationPayloadValidator_ValidateTimestamp(t *testing.T) {
	t.Parallel()

	t.Run("should work with time.Now handler", func(t *testing.T) {
		t.Parallel()

		currentTime := time.Now()
		validator, _ := NewPeerAuthenticationPayloadValidator()
		assert.False(t, check.IfNil(validator))
		assert.Nil(t, validator.ValidateTimestamp(currentTime.Unix()))
	})
	t.Run("payload time stamp is exactly the maximum accepted", func(t *testing.T) {
		t.Parallel()

		currentTime := time.Now()
		validator, _ := NewPeerAuthenticationPayloadValidator()
		validator.getTimeHandler = func() time.Time {
			return currentTime.Add(time.Second * 1120)
		}
		minimumAccepted := currentTime.Add(time.Second * (1120 + payloadExpiryThresholdInSec))
		assert.Nil(t, validator.ValidateTimestamp(minimumAccepted.Unix()))
	})
	t.Run("payload time stamp is higher than maximum accepted", func(t *testing.T) {
		t.Parallel()

		currentTime := time.Now()
		validator, _ := NewPeerAuthenticationPayloadValidator()
		validator.getTimeHandler = func() time.Time {
			return currentTime.Add(time.Second * 1120)
		}
		minimumAccepted := currentTime.Add(time.Second * (1120 + payloadExpiryThresholdInSec + 1))
		assert.True(t, errors.Is(validator.ValidateTimestamp(minimumAccepted.Unix()), process.ErrMessageExpired))
	})
}
