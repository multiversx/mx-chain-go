package logging

import (
	"errors"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
)

// LogLifeSpanFactoryArgs contains the data needed for the creation of a logLifeSpanFactory
type LogLifeSpanFactoryArgs struct {
	EpochStartNotifierWithConfirm EpochStartNotifierWithConfirm
	LifeSpanConfig                config.LifeSpanConfig
}

// CreateLogLifeSpanner is a factory method for creating log life spanners
func CreateLogLifeSpanner(args LogLifeSpanFactoryArgs) (LogLifeSpanner, error) {
	config := args.LifeSpanConfig
	switch config.Type {
	case "epoch":
		{
			els, err := newEpochsLifeSpanner(args.EpochStartNotifierWithConfirm, uint32(config.RecreateEvery))
			if err != nil {
				return nil, err
			}
			return els, nil
		}
	case "second":
		{
			sls, err := newSecondsLifeSpanner(time.Second * time.Duration(config.RecreateEvery))
			if err != nil {
				return nil, err
			}
			return sls, nil
		}
	case "MB":
		{
			sls, err := newSizeLifeSpanner(uint32(config.RecreateEvery))
			if err != nil {
				return nil, err
			}
			return sls, nil
		}
	}

	return nil, errors.New("error")
}
