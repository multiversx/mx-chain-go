package storageUnit

type nilStorer struct {
}

// NewNilStorer will return a nil storer
func NewNilStorer() *nilStorer {
	return new(nilStorer)
}

func (ns *nilStorer) Put(key, data []byte) error {
	return nil
}

func (ns *nilStorer) Get(key []byte) ([]byte, error) {
	return nil, nil
}

func (ns *nilStorer) Has(key []byte) error {
	return nil
}

func (ns *nilStorer) Remove(key []byte) error {
	return nil
}

func (ns *nilStorer) ClearCache() {
}

func (ns *nilStorer) DestroyUnit() error {
	return nil
}

func (ns *nilStorer) IsInterfaceNil() bool {
	if ns == nil {
		return true
	}
	return false
}
