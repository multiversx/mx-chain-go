package core

// KeyLoader holds the logic for loading a key from a file and an index
type KeyLoader struct {
}

// LoadKey loads the key with the given index found in the pem file from the given relative path.
func (kl *KeyLoader) LoadKey(relativePath string, skIndex int) ([]byte, string, error) {
	return LoadSkPkFromPemFile(relativePath, skIndex)
}
