package mock

type WsConnStub struct {
	CloseCalled        func() error
	ReadMessageCalled  func() (messageType int, p []byte, err error)
	WriteMessageCalled func(messageType int, data []byte) error
}

func (wcs *WsConnStub) Close() error {
	return wcs.CloseCalled()
}

func (wcs *WsConnStub) ReadMessage() (messageType int, p []byte, err error) {
	return wcs.ReadMessageCalled()
}

func (wcs *WsConnStub) WriteMessage(messageType int, data []byte) error {
	return wcs.WriteMessageCalled(messageType, data)
}
