package mock

import "time"

// TimeCacheStub -
type TimeCacheStub struct {
	AddCalled         func(key string) error
	AddWithSpanCalled func(key string, span time.Duration) error
	UpdateCalled      func(key string, span time.Duration) error
	HasCalled         func(key string) bool
	SweepCalled       func()
}

// Add -
func (tcs *TimeCacheStub) Add(key string) error {
	if tcs.AddCalled != nil {
		return tcs.AddCalled(key)
	}

	return nil
}

// AddWithSpan -
func (tcs *TimeCacheStub) AddWithSpan(key string, span time.Duration) error {
	if tcs.AddWithSpanCalled != nil {
		return tcs.AddWithSpanCalled(key, span)
	}

	return nil
}

// Update -
func (tcs *TimeCacheStub) Update(key string, span time.Duration) error {
	if tcs.UpdateCalled != nil {
		return tcs.UpdateCalled(key, span)
	}

	return nil
}

// Has -
func (tcs *TimeCacheStub) Has(key string) bool {
	if tcs.HasCalled != nil {
		return tcs.HasCalled(key)
	}

	return false
}

// Sweep -
func (tcs *TimeCacheStub) Sweep() {
	if tcs.SweepCalled != nil {
		tcs.SweepCalled()
	}
}

// IsInterfaceNil -
func (tcs *TimeCacheStub) IsInterfaceNil() bool {
	return tcs == nil
}
