package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/storage"
)

// TimeCacheStub -
type TimeCacheStub struct {
	AddCalled                     func(key string) error
	UpsertCalled                  func(key string, span time.Duration) error
	HasCalled                     func(key string) bool
	SweepCalled                   func()
	RegisterEvictionHandlerCalled func(handler storage.EvictionHandler)
}

// Add -
func (tcs *TimeCacheStub) Add(key string) error {
	if tcs.AddCalled != nil {
		return tcs.AddCalled(key)
	}

	return nil
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

// RegisterEvictionHandler -
func (tcs *TimeCacheStub) RegisterEvictionHandler(handler storage.EvictionHandler) {
	if tcs.RegisterEvictionHandlerCalled != nil {
		tcs.RegisterEvictionHandlerCalled(handler)
	}
}

// IsInterfaceNil -
func (tcs *TimeCacheStub) IsInterfaceNil() bool {
	return tcs == nil
}
