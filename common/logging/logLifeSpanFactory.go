package logging

import (
	"fmt"
	"time"
)

const (
	epochType    = "epoch"
	secondType   = "second"
	megabyteType = "megabyte"
)

// LogLifeSpanFactoryArgs contains the data needed for the creation of a logLifeSpanFactory
type LogLifeSpanFactoryArgs struct {
	EpochStartNotifierWithConfirm EpochStartNotifierWithConfirm
	LifeSpanType                  string
	RecreateEvery                 int
}

type typeLogLifeSpanFactory struct {
}

// NewTypeLogLifeSpanFactory creates a new factory for log life spans
func NewTypeLogLifeSpanFactory() *typeLogLifeSpanFactory {
	return &typeLogLifeSpanFactory{}
}

// CreateLogLifeSpanner is a factory method for creating log life spanners
func (llsf *typeLogLifeSpanFactory) CreateLogLifeSpanner(args LogLifeSpanFactoryArgs) (LogLifeSpanner, error) {
	switch args.LifeSpanType {
	case epochType:
		{
			els, err := newEpochsLifeSpanner(args.EpochStartNotifierWithConfirm, uint32(args.RecreateEvery))
			if err != nil {
				return nil, fmt.Errorf("%w, with error: %s", ErrCreateLogLifeSpanner, err.Error())
			}
			return els, nil
		}
	case secondType:
		{
			sls, err := newSecondsLifeSpanner(time.Second * time.Duration(args.RecreateEvery))
			if err != nil {
				return nil, fmt.Errorf("%w, with error: %s", ErrCreateLogLifeSpanner, err.Error())
			}
			return sls, nil
		}
	case megabyteType:
		{
			sls, err := newSizeLifeSpanner(&fileSizeChecker{}, uint32(args.RecreateEvery), time.Minute)
			if err != nil {
				return nil, fmt.Errorf("%w, with error: %s", ErrCreateLogLifeSpanner, err.Error())
			}
			return sls, nil
		}
	}

	return nil, ErrUnsupportedLogLifeSpanType
}
