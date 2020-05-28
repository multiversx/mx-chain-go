package mock

import "time"

// RequestedItemsHandlerStub -
type RequestedItemsHandlerStub struct {
	AddCalled         func(key string) error
	AddWithSpanCalled func(key string, span time.Duration) error
	HasCalled         func(key string) bool
	SweepCalled       func()
}

// AddWithSpan -
func (rihs *RequestedItemsHandlerStub) AddWithSpan(key string, span time.Duration) error {
	if rihs.AddWithSpanCalled == nil {
		return nil
	}

	return rihs.AddWithSpan(key, span)
}

// Add -
func (rihs *RequestedItemsHandlerStub) Add(key string) error {
	if rihs.AddCalled == nil {
		return nil
	}

	return rihs.AddCalled(key)
}

// Has -
func (rihs *RequestedItemsHandlerStub) Has(key string) bool {
	if rihs.HasCalled == nil {
		return false
	}

	return rihs.HasCalled(key)
}

// Sweep -
func (rihs *RequestedItemsHandlerStub) Sweep() {
	if rihs.SweepCalled == nil {
		return
	}

	rihs.SweepCalled()
}

// IsInterfaceNil -
func (rihs *RequestedItemsHandlerStub) IsInterfaceNil() bool {
	return rihs == nil
}
