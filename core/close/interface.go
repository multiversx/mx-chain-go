package close

// Closer closes all stuff released by an object
type Closer interface {
	Close() error
}
