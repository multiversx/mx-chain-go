package mock

import "context"

// KadDhtHandlerStub -
type KadDhtHandlerStub struct {
	BootstrapCalled func(ctx context.Context) error
}

// Bootstrap -
func (kdhs *KadDhtHandlerStub) Bootstrap(ctx context.Context) error {
	if kdhs.BootstrapCalled != nil {
		return kdhs.BootstrapCalled(ctx)
	}

	return nil
}
