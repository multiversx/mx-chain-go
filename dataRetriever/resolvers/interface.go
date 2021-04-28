package resolvers

type baseStorageResolver interface {
	getFromStorage(key []byte, epoch uint32) ([]byte, error)
	searchFirst(key []byte) ([]byte, error)
}
