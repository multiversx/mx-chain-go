package common

// GetClosedUnbufferedChannel returns an instance of a 'chan struct{}' that is already closed
func GetClosedUnbufferedChannel() chan struct{} {
	ch := make(chan struct{})
	close(ch)

	return ch
}

// ErrFromChan will get the error from channel
func ErrFromChan(errChan chan error) error {
	for {
		select {
		case err := <-errChan:
			return err
		default:
			return nil
		}
	}
}
