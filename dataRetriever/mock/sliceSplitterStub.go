package mock

type SliceSplitterStub struct {
	SendDataInChunksCalled func(data [][]byte, sendHandler func(buff []byte) error, maxPacketSize int) error
}

func (sss *SliceSplitterStub) SendDataInChunks(data [][]byte, sendHandler func(buff []byte) error, maxPacketSize int) error {
	return sss.SendDataInChunksCalled(data, sendHandler, maxPacketSize)
}
