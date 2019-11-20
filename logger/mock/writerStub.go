package mock

type WriterStub struct {
	WriteCalled func(p []byte) (n int, err error)
}

func (ws *WriterStub) Write(p []byte) (n int, err error) {
	return ws.WriteCalled(p)
}
