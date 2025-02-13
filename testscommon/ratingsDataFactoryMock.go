package testscommon

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/rating"
)

// RatingsDataFactoryMock -
type RatingsDataFactoryMock struct {
	CreateRatingsDataCalled func(args rating.RatingsDataArg) (process.RatingsInfoHandler, error)
}

// CreateRatingsData -
func (f *RatingsDataFactoryMock) CreateRatingsData(args rating.RatingsDataArg) (process.RatingsInfoHandler, error) {
	if f.CreateRatingsDataCalled != nil {
		return f.CreateRatingsDataCalled(args)
	}

	return &RatingsInfoMock{}, nil
}

// IsInterfaceNil -
func (f *RatingsDataFactoryMock) IsInterfaceNil() bool {
	return f == nil
}
