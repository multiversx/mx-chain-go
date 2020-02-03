package mock

type MockDB struct {
}

func (MockDB) Put(_, _ []byte) error {
	return nil
}

func (MockDB) Get(_ []byte) ([]byte, error) {
	return []byte{}, nil
}

func (MockDB) Has(_ []byte) error {
	return nil
}

func (MockDB) Init() error {
	return nil
}

func (MockDB) Close() error {
	return nil
}

func (MockDB) Remove(_ []byte) error {
	return nil
}

func (MockDB) Destroy() error {
	return nil
}

func (MockDB) DestroyClosed() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s MockDB) IsInterfaceNil() bool {
	return false
}
