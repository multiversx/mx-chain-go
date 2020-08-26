package storageResolvers

type trieNodeResolver struct {
	*storageResolver
}

// NewTrieNodeResolver returns a new trie node resolver instance. This instance is mocked as it is not supported when
// trying to request from storage.
func NewTrieNodeResolver() *trieNodeResolver {
	return &trieNodeResolver{}
}

// RequestDataFromHash returns nil
func (tnRes *trieNodeResolver) RequestDataFromHash(_ []byte, _ uint32) error {
	return nil
}

// RequestDataFromHashArray returns nil
func (tnRes *trieNodeResolver) RequestDataFromHashArray(_ [][]byte, _ uint32) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tnRes *trieNodeResolver) IsInterfaceNil() bool {
	return tnRes == nil
}
