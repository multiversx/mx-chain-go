package mock

type DataPackerStub struct {
	PackDataInChunksCalled func(data [][]byte, limit int) ([][]byte, error)
}

func (dps *DataPackerStub) PackDataInChunks(data [][]byte, limit int) ([][]byte, error) {
	return dps.PackDataInChunksCalled(data, limit)
}
