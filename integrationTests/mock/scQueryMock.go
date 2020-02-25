package mock

import (
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// ScQueryMock -
type ScQueryMock struct {
	ExecuteQueryCalled func(query *process.SCQuery) (*vmcommon.VMOutput, error)
}

// ExecuteQuery -
func (s *ScQueryMock) ExecuteQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	if s.ExecuteQueryCalled != nil {
		return s.ExecuteQueryCalled(query)
	}
	return &vmcommon.VMOutput{}, nil
}

// IsInterfaceNil -
func (s *ScQueryMock) IsInterfaceNil() bool {
	return s == nil
}
