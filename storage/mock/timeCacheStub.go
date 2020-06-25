package mock

import "time"

// TimeCacheStub -
type TimeCacheStub struct {
	UpsertCalled func(key string, span time.Duration) error
	HasCalled    func(key string) bool
	SweepCalled  func()
}

// Upsert -
func (tcs *TimeCacheStub) Upsert(key string, span time.Duration) error {
	if tcs.UpsertCalled != nil {
		return tcs.UpsertCalled(key, span)
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
