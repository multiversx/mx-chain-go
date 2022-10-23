package common

// GetClosedUnbufferedChannel returns an instance of a 'chan struct{}' that is already closed
func GetClosedUnbufferedChannel() chan struct{} {
	ch := make(chan struct{})
	close(ch)

	return ch
}

// GetErrorFromChanNonBlocking will get the error from channel
func GetErrorFromChanNonBlocking(errChan chan error) error {
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}
