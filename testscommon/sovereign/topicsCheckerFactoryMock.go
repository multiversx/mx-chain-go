package sovereign

import (
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
)

// TopicsCheckerFactoryMock -
type TopicsCheckerFactoryMock struct {
	CreateTopicsCheckerCalled func() sovereign.TopicsCheckerCreator
}

// CreateTopicsChecker -
func (tc *TopicsCheckerFactoryMock) CreateTopicsChecker() sovereign.TopicsCheckerHandler {
	if tc.CreateTopicsCheckerCalled != nil {
		return tc.CreateTopicsChecker()
	}
	return &TopicsCheckerMock{}
}

// IsInterfaceNil -
func (tc *TopicsCheckerFactoryMock) IsInterfaceNil() bool {
	return tc == nil
}
