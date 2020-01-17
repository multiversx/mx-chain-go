package storageUnit

type nilStorer struct {
}

// NewNilStorer will return a nil storer
func NewNilStorer() *nilStorer {
	return new(nilStorer)
}

func (ns *nilStorer) GetFromEpoch(_ []byte, _ uint32) ([]byte, error) {
	return nil, nil
}

func (ns *nilStorer) HasInEpoch(_ []byte, _ uint32) error {
	return nil
}

func (ns *nilStorer) Put(_, _ []byte) error {
	return nil
}

func (ns *nilStorer) Get(_ []byte) ([]byte, error) {
	return nil, nil
}

func (ns *nilStorer) Has(_ []byte) error {
	return nil
}

func (ns *nilStorer) Remove(_ []byte) error {
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
