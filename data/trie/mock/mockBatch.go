// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
	"github.com/pkg/errors"
)

var errMockBatch = errors.New("MockBatch generic error")

type MockBatch struct {
	db     *MockDatabase
	writes []kv
	size   int
	Fail   bool
}

func (b *MockBatch) Put(key, value []byte) error {
	if b.Fail {
		return errMockBatch
	}

	b.writes = append(b.writes, kv{encoding.CopyBytes(key), encoding.CopyBytes(value), false})
	b.size += len(value)
	return nil
}

func (b *MockBatch) Delete(key []byte) error {
	if b.Fail {
		return errMockBatch
	}

	b.writes = append(b.writes, kv{encoding.CopyBytes(key), nil, true})
	b.size += 1
	return nil
}

func (b *MockBatch) Write() error {
	if b.Fail {
		return errMockBatch
	}

	b.db.lock.Lock()
	defer b.db.lock.Unlock()

	for _, kv := range b.writes {
		if kv.del {
			delete(b.db.db, string(kv.k))
			continue
		}
		b.db.db[string(kv.k)] = kv.v
	}
	return nil
}

func (b *MockBatch) ValueSize() int {
	return b.size
}

func (b *MockBatch) Reset() {
	b.writes = b.writes[:0]
	b.size = 0
}
