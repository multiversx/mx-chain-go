package mock

import "time"

// TimeCacheStub -
type TimeCacheStub struct {
	AddCalled    func(key string) error
	UpsertCalled func(key string, span time.Duration) error
	HasCalled    func(key string) bool
	SweepCalled  func()
}

// Upsert -
func (tcs *TimeCacheStub) Upsert(key string, span time.Duration) error {
	if tcs.UpsertCalled == nil {
		return nil
	}

	return tcs.UpsertCalled(key, span)
}

// Add -
func (tcs *TimeCacheStub) Add(key string) error {
	if tcs.AddCalled == nil {
		return nil
	}

	return tcs.AddCalled(key)
}

// Has -
func (tcs *TimeCacheStub) Has(key string) bool {
	if tcs.HasCalled == nil {
		return false
	}

	return tcs.HasCalled(key)
}

// Sweep -
func (tcs *TimeCacheStub) Sweep() {
	if tcs.SweepCalled == nil {
		return
	}

	tcs.SweepCalled()
}

// IsInterfaceNil -
func (tcs *TimeCacheStub) IsInterfaceNil() bool {
	return tcs == nil
}
