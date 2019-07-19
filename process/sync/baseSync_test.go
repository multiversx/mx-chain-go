package sync_test

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
)

func GetCacherWithHeaders(
	hdr1 data.HeaderHandler,
	hdr2 data.HeaderHandler,
	hash1 []byte,
	hash2 []byte,
) storage.Cacher {
	sds := &mock.CacherStub{
		RegisterHandlerCalled: func(func(key []byte)) {},
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(key, hash1) {
				return &hdr1, true
			}
			if bytes.Equal(key, hash2) {
				return &hdr2, true
			}

			return nil, false
		},
	}
	return sds
}
