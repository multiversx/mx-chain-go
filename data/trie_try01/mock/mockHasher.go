package mock

import (
	"crypto/sha256"
)

type MockHasher struct {
	empty []byte
}

func (mh *MockHasher) Compute(data string) []byte {
	h := sha256.New()
	h.Write([]byte(data))
	hashed := h.Sum(nil)
	return hashed
}

func (mh *MockHasher) Size() int{
	return sha256.Size
}

func (mh *MockHasher) EmptyHash() []byte{
	if len(mh.empty) == 0{
		mh.empty = mh.Compute("")
	}

	return mh.empty
}