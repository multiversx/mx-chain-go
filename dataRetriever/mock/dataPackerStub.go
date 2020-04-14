package mock

// DataPackerStub -
type DataPackerStub struct {
	PackDataInChunksCalled func(data [][]byte, limit int) ([][]byte, error)
}

// PackDataInChunks -
func (dps *DataPackerStub) PackDataInChunks(data [][]byte, limit int) ([][]byte, error) {
	if dps.PackDataInChunksCalled != nil {
		return dps.PackDataInChunksCalled(data, limit)
	}

	return data, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dps *DataPackerStub) IsInterfaceNil() bool {
	return dps == nil
}
