package processHandler

import (
	"context"
)

// ContextProcessHandler defines a wrapper struct that can launch wrapped, stop-able go routines
type ContextProcessHandler struct {
	ctx        context.Context
	cancelFunc func()
}

// NewContextProcessHandler creates a new ContextProcessHandler instance
func NewContextProcessHandler(ctx context.Context) (*ContextProcessHandler, error) {
	ph := &ContextProcessHandler{}
	ph.ctx, ph.cancelFunc = context.WithCancel(ctx)

	return ph, nil
}

// StartGoRoutine starts the provided handler on a goroutine passing containing context
func (cph *ContextProcessHandler) StartGoRoutine(handler func(ctx context.Context)) {
	go handler(cph.ctx)
}

func (cph *ContextProcessHandler) Context() context.Context {
	return cph.ctx
}

// Close implements Closer interface allowing the launched go routines terminate
func (cph *ContextProcessHandler) Close() error {
	cph.cancelFunc()

	return nil
}
