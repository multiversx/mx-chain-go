package endProcess

// ArgEndProcess represents an object that encapsulates the reason and description for ending the process
type ArgEndProcess struct {
	Reason      string
	Description string
}

// GetDummyEndProcessChannel returns a dummy channel of type ArgEndProcess
func GetDummyEndProcessChannel() chan ArgEndProcess {
	ch := make(chan ArgEndProcess, 1)

	return ch
}
