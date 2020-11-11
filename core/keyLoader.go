package core

// KeyLoader holds the logic for loading a key from a file and an index
type KeyLoader struct {
}

func (kl *KeyLoader) LoadKey(relativePath string, skIndex int) ([]byte, string, error) {
	return LoadSkPkFromPemFile(relativePath, skIndex)
}
