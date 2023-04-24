package mock

import "errors"

// KeyLoaderStub -
type KeyLoaderStub struct {
	LoadKeyCalled     func(relativePath string, skIndex int) ([]byte, string, error)
	LoadAllKeysCalled func(path string) ([][]byte, []string, error)
}

// LoadKey -
func (kl *KeyLoaderStub) LoadKey(relativePath string, skIndex int) ([]byte, string, error) {
	if kl.LoadKeyCalled != nil {
		return kl.LoadKeyCalled(relativePath, skIndex)
	}

	return nil, "", nil
}

// LoadAllKeys -
func (kl *KeyLoaderStub) LoadAllKeys(path string) ([][]byte, []string, error) {
	if kl.LoadAllKeysCalled != nil {
		return kl.LoadAllKeysCalled(path)
	}

	return nil, nil, errors.New("not implemented")
}

// IsInterfaceNil -
func (kl *KeyLoaderStub) IsInterfaceNil() bool {
	return kl == nil
}
