package mock

type accumulatorMock struct {
	ch chan []interface{}
}

// NewAccumulatorMock returns a mock implementation of the accumulator interface
func NewAccumulatorMock() *accumulatorMock {
	return &accumulatorMock{
		ch: make(chan []interface{}),
	}
}

// AddData -
func (am *accumulatorMock) AddData(data interface{}) {
	am.ch <- []interface{}{data}
}

// OutputChannel -
func (am *accumulatorMock) OutputChannel() <-chan []interface{} {
	return am.ch
}

// IsInterfaceNil returns true if there is no value under the interface
func (am *accumulatorMock) IsInterfaceNil() bool {
	return am == nil
}
