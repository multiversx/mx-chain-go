package mock

type MockDB struct {
}

func (MockDB) Put(key, val []byte) error {
	return nil
}

func (MockDB) Get(key []byte) ([]byte, error) {
	return []byte{}, nil
}

func (MockDB) Has(key []byte) error {
	return nil
}

func (MockDB) Init() error {
	return nil
}

func (MockDB) Close() error {
	return nil
}

func (MockDB) Remove(key []byte) error {
	return nil
}

func (MockDB) Destroy() error {
	return nil
}
