package processMocks

import "errors"

type GracePeriodErrStub struct{}

// GetGracePeriodForEpoch always returns an error.
func (GracePeriodErrStub) GetGracePeriodForEpoch(_ uint32) (uint32, error) {
	return 0, errors.New("epochChangeGracePeriodHandler forced error")
}

// IsInterfaceNil -
func (GracePeriodErrStub) IsInterfaceNil() bool { return false }
