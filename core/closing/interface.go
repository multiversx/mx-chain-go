package closing

// Closer closes all stuff released by an object
type Closer interface {
	Close() error
}

// IntRandomizer interface provides functionality over generating integer numbers
type IntRandomizer interface {
	Intn(n int) int
	IsInterfaceNil() bool
}
