package mock

// PersisterStub -
type PersisterStub struct {
	PutCalled           func(key, val []byte) error
	GetCalled           func(key []byte) ([]byte, error)
	HasCalled           func(key []byte) error
	CloseCalled         func() error
	RemoveCalled        func(key []byte) error
	DestroyCalled       func() error
	DestroyClosedCalled func() error
	RangeKeysCalled     func(handler func(key []byte, val []byte) bool)
}

// Put -
func (p *PersisterStub) Put(key, val []byte) error {
	if p.PutCalled != nil {
		return p.PutCalled(key, val)
	}

	return nil
}

// Get -
func (p *PersisterStub) Get(key []byte) ([]byte, error) {
	if p.GetCalled != nil {
		return p.GetCalled(key)
	}

	return nil, nil
}

// Has -
func (p *PersisterStub) Has(key []byte) error {
	if p.HasCalled != nil {
		return p.HasCalled(key)
	}

	return nil
}

// Close -
func (p *PersisterStub) Close() error {
	if p.CloseCalled != nil {
		return p.CloseCalled()
	}

	return nil
}

// Remove -
func (p *PersisterStub) Remove(key []byte) error {
	if p.RemoveCalled != nil {
		return p.RemoveCalled(key)
	}

	return nil
}

// Destroy -
func (p *PersisterStub) Destroy() error {
	if p.DestroyCalled != nil {
		return p.DestroyCalled()
	}

	return nil
}

// DestroyClosed -
func (p *PersisterStub) DestroyClosed() error {
	if p.DestroyClosedCalled != nil {
		return p.DestroyClosedCalled()
	}

	return nil
}

// RangeKeys -
func (p *PersisterStub) RangeKeys(handler func(key []byte, val []byte) bool) {
	if p.RangeKeysCalled != nil {
		p.RangeKeysCalled(handler)
	}
}

// IsInterfaceNil -
func (p *PersisterStub) IsInterfaceNil() bool {
	return p == nil
}
