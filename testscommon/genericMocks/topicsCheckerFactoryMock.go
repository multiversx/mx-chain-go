package genericMocks

import (
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	sovereign2 "github.com/multiversx/mx-chain-go/testscommon/sovereign"
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
	return &sovereign2.TopicsCheckerMock{}
}

// IsInterfaceNil -
func (tc *TopicsCheckerFactoryMock) IsInterfaceNil() bool {
	return tc == nil
}
