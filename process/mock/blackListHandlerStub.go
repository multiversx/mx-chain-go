package mock

import "time"

// BlackListHandlerStub -
type BlackListHandlerStub struct {
	AddCalled    func(key string) error
	UpsertCalled func(key string, span time.Duration) error
	HasCalled    func(key string) bool
	SweepCalled  func()
	LenCalled    func() int
}

// Add -
func (blhs *BlackListHandlerStub) Add(key string) error {
	if blhs.AddCalled == nil {
		return nil
	}

	return blhs.AddCalled(key)
}

// Upsert -
func (blhs *BlackListHandlerStub) Upsert(key string, span time.Duration) error {
	if blhs.UpsertCalled == nil {
		return nil
	}

	return blhs.UpsertCalled(key, span)
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

// Len -
func (blhs *BlackListHandlerStub) Len() int {
	if blhs.LenCalled == nil {
		return 0
	}

	return blhs.LenCalled()
}

// IsInterfaceNil -
func (blhs *BlackListHandlerStub) IsInterfaceNil() bool {
	return blhs == nil
}
