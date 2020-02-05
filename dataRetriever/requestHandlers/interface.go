package requestHandlers

// HashSliceResolver can request multiple hashes at once
type HashSliceResolver interface {
	RequestDataFromHashArray(hashes [][]byte, epoch uint32) error
	IsInterfaceNil() bool
}
