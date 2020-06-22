package mock

import "time"

// BlackListHandlerStub -
type BlackListHandlerStub struct {
	AddCalled         func(key string) error
	AddWithSpanCalled func(key string, span time.Duration) error
	HasCalled         func(key string) bool
	SweepCalled       func()
}

// Add -
func (blhs *BlackListHandlerStub) Add(key string) error {
	if blhs.AddCalled == nil {
		return nil
	}

	return blhs.AddCalled(key)
}

// AddWithSpan -
func (blhs *BlackListHandlerStub) AddWithSpan(key string, span time.Duration) error {
	if blhs.AddWithSpanCalled == nil {
		return nil
	}

	return blhs.AddWithSpanCalled(key, span)
}

// Has -
func (blhs *BlackListHandlerStub) Has(key string) bool {
	if blhs.HasCalled == nil {
		return false
	}

	return blhs.HasCalled(key)
}

// Sweep -
func (blhs *BlackListHandlerStub) Sweep() {
	if blhs.SweepCalled == nil {
		return
	}

	blhs.SweepCalled()
}

// IsInterfaceNil -
func (blhs *BlackListHandlerStub) IsInterfaceNil() bool {
	return blhs == nil
}
