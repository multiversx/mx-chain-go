package api

import "context"

// ServerStub -
type ServerStub struct {
	ListenAndServeCalled func() error
	ShutdownCalled       func(ctx context.Context) error
}

// ListenAndServe -
func (stub *ServerStub) ListenAndServe() error {
	if stub.ListenAndServeCalled != nil {
		return stub.ListenAndServeCalled()
	}
	return nil
}

// Shutdown -
func (stub *ServerStub) Shutdown(ctx context.Context) error {
	if stub.ShutdownCalled != nil {
		return stub.ShutdownCalled(ctx)
	}
	return nil
}
