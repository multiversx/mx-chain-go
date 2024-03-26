package genericMocks

import (
	factorySovereign "github.com/multiversx/mx-chain-go/factory/sovereign"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	sovereignMock "github.com/multiversx/mx-chain-go/testscommon/sovereign"
)

// TopicsCheckerFactoryMock -
type TopicsCheckerFactoryMock struct {
	CreateTopicsCheckerCalled func() factorySovereign.TopicsCheckerCreator
}

// CreateTopicsChecker -
func (tc *TopicsCheckerFactoryMock) CreateTopicsChecker() sovereign.TopicsCheckerHandler {
	if tc.CreateTopicsCheckerCalled != nil {
		return tc.CreateTopicsChecker()
	}
	return &sovereignMock.TopicsCheckerMock{}
}

// IsInterfaceNil -
func (tc *TopicsCheckerFactoryMock) IsInterfaceNil() bool {
	return tc == nil
}
