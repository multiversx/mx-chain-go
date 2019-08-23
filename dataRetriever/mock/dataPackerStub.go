package mock

type DataPackerStub struct {
	PackDataInChunksCalled func(data [][]byte, limit int) ([][]byte, error)
}

func (dps *DataPackerStub) PackDataInChunks(data [][]byte, limit int) ([][]byte, error) {
	return dps.PackDataInChunksCalled(data, limit)
}

// IsInterfaceNil returns true if there is no value under the interface
func (dps *DataPackerStub) IsInterfaceNil() bool {
	if dps == nil {
		return true
	}
	return false
}
