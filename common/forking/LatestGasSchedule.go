package forking

import "errors"

var gasSchedule *gasScheduleNotifier

type LatestGasSchedule struct {
	*gasScheduleNotifier
}

// NewLatestGasSchedule creates a new instance of a LatestGasSchedule component
func NewLatestGasSchedule() (*LatestGasSchedule, error) {
	if gasSchedule == nil {
		return nil, errors.New("gas schedule is nil")
	}

	gasSchedule.EpochConfirmed(2000, 0)

	return &LatestGasSchedule{gasSchedule}, nil
}
