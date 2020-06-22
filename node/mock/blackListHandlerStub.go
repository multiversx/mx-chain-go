package mock

import "time"

// BlackListHandlerStub -
type BlackListHandlerStub struct {
	AddCalled         func(key string) error
	AddWithSpanCalled func(key string, span time.Duration) error
	HasCalled         func(key string) bool
	SweepCalled       func()
}

// Sweep -
func (blhs *BlackListHandlerStub) Sweep() {
	if blhs.SweepCalled != nil {
		blhs.SweepCalled()
	}
}

// AddWithSpan -
func (blhs *BlackListHandlerStub) AddWithSpan(key string, span time.Duration) error {
	if blhs.AddWithSpanCalled != nil {
		return blhs.AddWithSpanCalled(key, span)
	}

	return nil
}

// Add -
func (blhs *BlackListHandlerStub) Add(key string) error {
	if blhs.AddCalled != nil {
		return blhs.AddCalled(key)
	}

	return nil
}

// Has -
func (blhs *BlackListHandlerStub) Has(key string) bool {
	if blhs.HasCalled != nil {
		return blhs.HasCalled(key)
	}

	return false
}

// IsInterfaceNil -
func (blhs *BlackListHandlerStub) IsInterfaceNil() bool {
	return blhs == nil
}
