package mock

import (
	"time"
)

// ContextMock -
type ContextMock struct {
	DoneFunc     func() <-chan struct{}
	DeadlineFunc func() (time.Time, bool)
	ErrFunc      func() error
	ValueFunc    func(key interface{}) interface{}
}

// Done -
func (c *ContextMock) Done() <-chan struct{} {
	if c.DoneFunc != nil {
		return c.DoneFunc()
	}
	return nil
}

// Deadline -
func (c *ContextMock) Deadline() (time.Time, bool) {
	if c.DeadlineFunc != nil {
		return c.DeadlineFunc()
	}
	return time.Time{}, false
}

// Err -
func (c *ContextMock) Err() error {
	if c.ErrFunc != nil {
		return c.ErrFunc()
	}
	return nil
}

// Value -
func (c *ContextMock) Value(key interface{}) interface{} {
	if c.ValueFunc != nil {
		return c.ValueFunc(key)
	}
	return nil
}
