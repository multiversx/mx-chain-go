package hashing

// Hasher provides hashing services
type Hasher interface {
	Compute(string) []byte
	EmptyHash() []byte
	Size() int
}
