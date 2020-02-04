package mock

import "sync"

// WsConnStub -
type WsConnStub struct {
	mutHandlers        sync.Mutex
	closeCalled        func() error
	readMessageCalled  func() (messageType int, p []byte, err error)
	writeMessageCalled func(messageType int, data []byte) error
}

// Close -
func (wcs *WsConnStub) Close() error {
	wcs.mutHandlers.Lock()
	defer wcs.mutHandlers.Unlock()

	return wcs.closeCalled()
}

// ReadMessage -
func (wcs *WsConnStub) ReadMessage() (messageType int, p []byte, err error) {
	wcs.mutHandlers.Lock()
	defer wcs.mutHandlers.Unlock()

	return wcs.readMessageCalled()
}

// WriteMessage -
func (wcs *WsConnStub) WriteMessage(messageType int, data []byte) error {
	wcs.mutHandlers.Lock()
	defer wcs.mutHandlers.Unlock()

	return wcs.writeMessageCalled(messageType, data)
}

// SetReadMessageHandler -
func (wcs *WsConnStub) SetReadMessageHandler(f func() (messageType int, p []byte, err error)) {
	wcs.mutHandlers.Lock()
	defer wcs.mutHandlers.Unlock()

	wcs.readMessageCalled = f
}

// SetWriteMessageHandler -
func (wcs *WsConnStub) SetWriteMessageHandler(f func(messageType int, data []byte) error) {
	wcs.mutHandlers.Lock()
	defer wcs.mutHandlers.Unlock()

	wcs.writeMessageCalled = f
}

// SetCloseHandler -
func (wcs *WsConnStub) SetCloseHandler(f func() error) {
	wcs.mutHandlers.Lock()
	defer wcs.mutHandlers.Unlock()

	wcs.closeCalled = f
}
