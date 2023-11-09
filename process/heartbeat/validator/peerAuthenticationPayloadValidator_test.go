package validator

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/stretchr/testify/assert"
)

func TestNewPeerAuthenticationPayloadValidator(t *testing.T) {
	t.Parallel()

	t.Run("invalid expiry duration should error", func(t *testing.T) {
		t.Parallel()

		valsToTest := int64(100)

		for i := int64(1); i <= valsToTest; i++ {
			validator, err := NewPeerAuthenticationPayloadValidator(minDurationInSec - i)
			assert.True(t, check.IfNil(validator))
			assert.Equal(t, process.ErrInvalidExpiryTimespan, err)
		}
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		valsToTest := int64(100)

		for i := int64(0); i < valsToTest; i++ {
			validator, err := NewPeerAuthenticationPayloadValidator(minDurationInSec + i)
			assert.False(t, check.IfNil(validator))
			assert.Nil(t, err)
		}
	})
}

func TestPeerAuthenticationPayloadValidator_ValidateTimestamp(t *testing.T) {
	t.Parallel()

	t.Run("should work with time.Now handler", func(t *testing.T) {
		t.Parallel()

		currentTime := time.Now()
		validator, _ := NewPeerAuthenticationPayloadValidator(minDurationInSec)
		assert.Nil(t, validator.ValidateTimestamp(currentTime.Unix()))
	})
	t.Run("payload time stamp is exactly the minim accepted", func(t *testing.T) {
		t.Parallel()

		currentTime := time.Now()
		validator, _ := NewPeerAuthenticationPayloadValidator(minDurationInSec)
		validator.getTimeHandler = func() time.Time {
			return currentTime.Add(time.Second * 1120)
		}
		minimumAccepted := currentTime.Add(time.Second * (1120 - minDurationInSec))
		assert.Nil(t, validator.ValidateTimestamp(minimumAccepted.Unix()))
	})
	t.Run("payload time stamp is less than minim accepted", func(t *testing.T) {
		t.Parallel()

		currentTime := time.Now()
		validator, _ := NewPeerAuthenticationPayloadValidator(minDurationInSec)
		validator.getTimeHandler = func() time.Time {
			return currentTime.Add(time.Second * 1120)
		}
		minimumAccepted := currentTime.Add(time.Second * (1120 - minDurationInSec - 1))
		assert.True(t, errors.Is(validator.ValidateTimestamp(minimumAccepted.Unix()), process.ErrMessageExpired))
	})
	t.Run("payload time stamp is exactly the maximum accepted", func(t *testing.T) {
		t.Parallel()

		currentTime := time.Now()
		validator, _ := NewPeerAuthenticationPayloadValidator(minDurationInSec)
		validator.getTimeHandler = func() time.Time {
			return currentTime.Add(time.Second * 1120)
		}
		minimumAccepted := currentTime.Add(time.Second * (1120 + payloadExpiryThresholdInSec))
		assert.Nil(t, validator.ValidateTimestamp(minimumAccepted.Unix()))
	})
	t.Run("payload time stamp is higher than maximum accepted", func(t *testing.T) {
		t.Parallel()

		currentTime := time.Now()
		validator, _ := NewPeerAuthenticationPayloadValidator(minDurationInSec)
		validator.getTimeHandler = func() time.Time {
			return currentTime.Add(time.Second * 1120)
		}
		minimumAccepted := currentTime.Add(time.Second * (1120 + payloadExpiryThresholdInSec + 1))
		assert.True(t, errors.Is(validator.ValidateTimestamp(minimumAccepted.Unix()), process.ErrMessageExpired))
	})
}
