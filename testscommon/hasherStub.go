package testscommon

<<<<<<<< HEAD:testscommon/hasherStub.go
// HasherStub -
type HasherStub struct {
	ComputeCalled   func(s string) []byte
	EmptyHashCalled func() []byte
	SizeCalled      func() int
}

// Compute -
func (hash *HasherStub) Compute(s string) []byte {
	if hash.ComputeCalled != nil {
		return hash.ComputeCalled(s)
	}
	return nil
}

// EmptyHash -
func (hash *HasherStub) EmptyHash() []byte {
	if hash.EmptyHashCalled != nil {
		hash.EmptyHashCalled()
========
import "crypto/sha256"

var sha256EmptyHash []byte

// HasherMock that will be used for testing
type HasherMock struct {
}

// Compute will output the SHA's equivalent of the input string
func (sha HasherMock) Compute(s string) []byte {
	h := sha256.New()
	_, _ = h.Write([]byte(s))
	return h.Sum(nil)
}

// EmptyHash will return the equivalent of empty string SHA's
func (sha HasherMock) EmptyHash() []byte {
	if len(sha256EmptyHash) == 0 {
		sha256EmptyHash = sha.Compute("")
>>>>>>>> development:testscommon/hasherMock.go
	}
	return sha256EmptyHash
}

<<<<<<<< HEAD:testscommon/hasherStub.go
// Size -
func (hash *HasherStub) Size() int {
	if hash.SizeCalled != nil {
		return hash.SizeCalled()
	}

	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (hash *HasherStub) IsInterfaceNil() bool {
========
// Size returns the required size in bytes
func (HasherMock) Size() int {
	return sha256.Size
}

// IsInterfaceNil returns true if there is no value under the interface
func (sha HasherMock) IsInterfaceNil() bool {
>>>>>>>> development:testscommon/hasherMock.go
	return false
}
