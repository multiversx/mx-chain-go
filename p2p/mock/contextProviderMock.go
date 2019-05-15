package mock

import (
	"context"
)

type ContextProviderMock struct {
}

func (*ContextProviderMock) Context() context.Context {
	panic("implement me")
}
