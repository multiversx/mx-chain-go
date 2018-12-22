package mock

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/davecgh/go-spew/spew"
)

var errMockTrie = errors.New("TrieMock generic error")

// TrieMock that will be used for testing
type TrieMock struct {
	mutData sync.RWMutex
	keys    []string
	data    map[string][]byte

	Fail         bool
	FailRecreate bool
	FailGet      bool
	TTFGet       int
	FailUpdate   bool
	TTFUpdate    int
	mdbwc        trie.DBWriteCacher
}

// NewMockTrie creates a new TrieMock obj
func NewMockTrie() *TrieMock {
	mt := TrieMock{mutData: sync.RWMutex{}, data: make(map[string][]byte, 0), keys: make([]string, 0), Fail: false}
	mt.mdbwc = NewMockDBWriteCacher()

	return &mt
}

// NewMockTrieWithDBW creates a new TrieMock obj on the same cacher object
func NewMockTrieWithDBW(cacher trie.DBWriteCacher) *TrieMock {
	mt := TrieMock{mutData: sync.RWMutex{}, data: make(map[string][]byte, 0), keys: make([]string, 0), Fail: false}
	mt.mdbwc = cacher

	return &mt
}

// SetCacheLimit is an empty func
func (mt *TrieMock) SetCacheLimit(l uint16) {
	//nothing
}

// Get returns the data from the 'trie'
func (mt *TrieMock) Get(key []byte) ([]byte, error) {
	if mt.Fail {
		return nil, errMockTrie
	}

	if mt.FailGet {
		if mt.TTFGet <= 0 {
			return nil, errMockTrie
		}

		mt.TTFGet--
	}

	mt.mutData.RLock()
	defer mt.mutData.RUnlock()

	val, ok := mt.data[string(key)]

	if !ok {
		return nil, nil
	}

	return val, nil
}

// Update inserts or modifies the data in the 'trie'
func (mt *TrieMock) Update(key, value []byte) error {
	if mt.Fail {
		return errMockTrie
	}

	if mt.FailUpdate {
		if mt.TTFUpdate <= 0 {
			return errMockTrie
		}

		mt.TTFUpdate--
	}

	mt.mutData.Lock()
	defer mt.mutData.Unlock()

	_, ok := mt.data[string(key)]
	if !ok {
		mt.keys = append(mt.keys, string(key))
	}

	mt.data[string(key)] = value

	if len(value) == 0 {
		delete(mt.data, string(key))
		idx := -1
		for i := 0; i < len(mt.keys); i++ {
			if mt.keys[i] == string(key) {
				idx = i
				break
			}
		}

		mt.keys = append(mt.keys[:idx], mt.keys[idx+1:]...)
	}
	return nil
}

// Delete erases the data from the 'trie'
func (mt *TrieMock) Delete(key []byte) error {
	panic("Delete will not be implemented!")
}

// Root returns the hash of the entire 'trie'
func (mt *TrieMock) Root() []byte {
	if len(mt.data) == 0 {
		return make([]byte, encoding.HashLength)
	}

	mt.mutData.RLock()
	defer mt.mutData.RUnlock()
	buff := bytes.Buffer{}

	for i := 0; i < len(mt.keys); i++ {
		val := mt.data[mt.keys[i]]

		buff.Write([]byte(mt.keys[i]))
		buff.Write(val)
	}

	hash := HasherMock{}
	return hash.Compute(buff.String())
}

// Commit make changes to be permanent (calls the commit on underlining cacher)
func (mt *TrieMock) Commit(onleaf trie.LeafCallback) (root []byte, err error) {
	if mt.Fail {
		return nil, errMockTrie
	}

	mock := mt.mdbwc.(*DBWriteCacherMock)
	mock.AppendMockTrie(mt.Copy().(*TrieMock))

	return mt.Root(), nil
}

// DBW returns the underling cacher
func (mt *TrieMock) DBW() trie.DBWriteCacher {
	return mt.mdbwc
}

// Recreate returns a new 'trie' based on the root hash and underlying cacher
func (mt *TrieMock) Recreate(root []byte, dbw trie.DBWriteCacher) (trie.PatriciaMerkelTree, error) {
	if mt.Fail {
		return nil, errMockTrie
	}

	if mt.FailRecreate {
		return nil, errMockTrie
	}

	mock := dbw.(*DBWriteCacherMock)
	return mock.RetrieveMockTrie(root)
}

// Copy returns a new 'trie'
func (mt *TrieMock) Copy() trie.PatriciaMerkelTree {
	newTrie := NewMockTrieWithDBW(mt.mdbwc)

	mt.mutData.RLock()
	defer mt.mutData.RUnlock()

	for k, v := range mt.data {
		newTrie.data[k] = v
	}

	keys := make([]string, len(mt.keys))
	copy(keys, mt.keys)
	newTrie.keys = keys

	rootC := newTrie.Root()
	rootS := mt.Root()

	if !bytes.Equal(rootC, rootS) {
		fmt.Println("New")
		spew.Dump(newTrie)
		fmt.Println("Old")
		spew.Dump(mt)
		panic(fmt.Sprintf("root mismatch new: %d, old: %v", rootC, rootS))
	}

	return newTrie
}

// DBWriteCacherMock that will be used for testing
type DBWriteCacherMock struct {
	mutTries sync.RWMutex
	tries    map[string]trie.PatriciaMerkelTree
}

// NewMockDBWriteCacher returns a new DBWriteCacherMock
func NewMockDBWriteCacher() *DBWriteCacherMock {
	mdbwc := DBWriteCacherMock{}
	mdbwc.mutTries = sync.RWMutex{}
	mdbwc.tries = make(map[string]trie.PatriciaMerkelTree)
	return &mdbwc
}

// PersistDB is an empty func
func (mdbw *DBWriteCacherMock) Storer() storage.Storer {
	panic("implement me")
}

// InsertBlob is an empty func
func (mdbw *DBWriteCacherMock) InsertBlob(hash []byte, blob []byte) {
	panic("implement me")
}

// Node is an empty func
func (mdbw *DBWriteCacherMock) Node(hash []byte) ([]byte, error) {
	panic("implement me")
}

// Reference is an empty func
func (mdbw *DBWriteCacherMock) Reference(child []byte, parent []byte) {
	panic("implement me")
}

// Dereference is an empty func
func (mdbw *DBWriteCacherMock) Dereference(root []byte) {
	panic("implement me")
}

// Cap is an empty func
func (mdbw *DBWriteCacherMock) Cap(limit float64) error {
	panic("implement me")
}

// Commit is an empty func
func (mdbw *DBWriteCacherMock) Commit(node []byte, report bool) error {
	panic("implement me")
}

// Size is an empty func
func (mdbw *DBWriteCacherMock) Size() (float64, float64) {
	panic("implement me")
}

// InsertWithLock is an empty func
func (mdbw *DBWriteCacherMock) InsertWithLock(hash []byte, blob []byte, node trie.Node) {

}

// CachedNode is an empty func
func (mdbw *DBWriteCacherMock) CachedNode(hash []byte, cachegen uint16) trie.Node {
	panic("implement me")
}

// AppendMockTrie adds a new 'trie'
func (mdbw *DBWriteCacherMock) AppendMockTrie(mockTrie *TrieMock) {
	mdbw.mutTries.Lock()
	defer mdbw.mutTries.Unlock()

	mdbw.tries[string(mockTrie.Root())] = mockTrie
}

// RetrieveMockTrie retrieves a stored 'trie' or nil if not found
func (mdbw *DBWriteCacherMock) RetrieveMockTrie(root []byte) (trie.PatriciaMerkelTree, error) {
	mdbw.mutTries.Lock()
	defer mdbw.mutTries.Unlock()

	val, ok := mdbw.tries[string(root)]

	if !ok {
		if !bytes.Equal(root, make([]byte, encoding.HashLength)) &&
			len(root) > 0 {
			return nil, errors.New("root not found")
		}

		mt := NewMockTrieWithDBW(mdbw)
		return mt, nil
	}

	return val.Copy(), nil
}
