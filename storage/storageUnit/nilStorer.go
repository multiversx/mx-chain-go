package storageUnit

type nilStorer struct {
}

// NewNilStorer will return a nil storer
func NewNilStorer() *nilStorer {
	return new(nilStorer)
}

func (ns *nilStorer) GetFromEpoch(key []byte, epoch uint32) ([]byte, error) {
	return nil, nil
}

func (ns *nilStorer) HasInEpoch(key []byte, epoch uint32) error {
	return nil
}

func (ns *nilStorer) SearchFirst(key []byte) ([]byte, error) {
	return nil, nil
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
	return ns == nil
}
