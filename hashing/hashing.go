package hashing

// Hasher defines an interface for hashing
type Hasher interface {
	Compute(string) []byte
	EmptyHash() []byte
	Size() int
}

// DefHash holds the current implementation of the hasher interface
var DefHash Hasher

func init() {
	DefHash = &Sha256{}
}
