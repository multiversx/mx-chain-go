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
	"errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
)

var errMockDataBase = errors.New("MockDatabase generic error")

// DatabaseMock is used in testing
type DatabaseMock struct {
	db   map[string][]byte
	lock sync.RWMutex
	Fail bool
}

type keyValuePair struct {
	key []byte
	val []byte
	del bool
}

func NewDatabaseMock() *DatabaseMock {
	return &DatabaseMock{
		db: make(map[string][]byte),
	}
}

func (db *DatabaseMock) Init() error {
	if db.Fail {
		return errMockDataBase
	}

	if db == nil {
		db.lock.Lock()
		db.db = make(map[string][]byte, 0)
		db.lock.Unlock()
	}

	return nil
}

func (db *DatabaseMock) Close() error {
	if db.Fail {
		return errMockDataBase
	}

	return nil
}

func (db *DatabaseMock) Destroy() error {
	if db.Fail {
		return errMockDataBase
	}

	db.lock.Lock()
	defer db.lock.Unlock()
	db.db = make(map[string][]byte, 0)
	return nil
}

func (db *DatabaseMock) Put(key []byte, value []byte) error {
	if db.Fail {
		return errMockDataBase
	}

	db.lock.Lock()
	defer db.lock.Unlock()

	db.db[string(key)] = encoding.CopyBytes(value)
	return nil
}

func (db *DatabaseMock) Has(key []byte) (bool, error) {
	if db.Fail {
		return false, errMockDataBase
	}

	db.lock.RLock()
	defer db.lock.RUnlock()

	_, ok := db.db[string(key)]
	return ok, nil
}

func (db *DatabaseMock) Get(key []byte) ([]byte, error) {
	if db.Fail {
		return nil, errMockDataBase
	}

	db.lock.RLock()
	defer db.lock.RUnlock()

	if entry, ok := db.db[string(key)]; ok {
		return encoding.CopyBytes(entry), nil
	}
	return nil, errors.New("not found")
}

func (db *DatabaseMock) Keys() [][]byte {
	db.lock.RLock()
	defer db.lock.RUnlock()

	keys := make([][]byte, 0)
	for key := range db.db {
		keys = append(keys, []byte(key))
	}
	return keys
}

func (db *DatabaseMock) Remove(key []byte) error {
	if db.Fail {
		return errMockDataBase
	}

	db.lock.Lock()
	defer db.lock.Unlock()

	delete(db.db, string(key))
	return nil
}

func (db *DatabaseMock) NewBatch() trie.Batch {
	return &BatchMock{db: db}
}

func (db *DatabaseMock) Len() int { return len(db.db) }
