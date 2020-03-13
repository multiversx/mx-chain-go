package mock

import (
	"context"
)

// ContextProviderMock -
type ContextProviderMock struct {
}

// Context -
func (*ContextProviderMock) Context() context.Context {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *ContextProviderMock) IsInterfaceNil() bool {
	return c == nil
}
