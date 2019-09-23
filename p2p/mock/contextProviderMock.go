package mock

import (
	"context"
)

type ContextProviderMock struct {
}

func (*ContextProviderMock) Context() context.Context {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *ContextProviderMock) IsInterfaceNil() bool {
	if c == nil {
		return true
	}
	return false
}
