package mock

// KeyLoaderStub -
type KeyLoaderStub struct {
	LoadKeyCalled func(relativePath string, skIndex int) ([]byte, string, error)
}

// LoadKey -
func (kl *KeyLoaderStub) LoadKey(relativePath string, skIndex int) ([]byte, string, error) {
	if kl.LoadKeyCalled != nil {
		return kl.LoadKeyCalled(relativePath, skIndex)
	}

	return nil, "", nil
}
