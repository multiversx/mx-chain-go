package mock

import (
	"bytes"
	"errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/trie/encoding"
	"sync"
)

var errMockTrie = errors.New("TrieMock generic error")

type TrieMock struct {
	mutData      sync.RWMutex
	data         map[string][]byte
	Fail         bool
	FailRecreate bool
	mdbwc        trie.DBWriteCacher
}

func NewMockTrie() *TrieMock {
	mt := TrieMock{mutData: sync.RWMutex{}, data: make(map[string][]byte, 0), Fail: false}
	mt.mdbwc = NewMockDBWriteCacher()

	return &mt
}

func NewMockTrieWithDBW(cacher trie.DBWriteCacher) *TrieMock {
	mt := TrieMock{mutData: sync.RWMutex{}, data: make(map[string][]byte, 0), Fail: false}
	mt.mdbwc = cacher

	return &mt
}

func (mt *TrieMock) SetCacheLimit(l uint16) {
	//nothing
}

func (mt *TrieMock) Get(key []byte) ([]byte, error) {
	if mt.Fail {
		return nil, errMockTrie
	}

	mt.mutData.RLock()
	defer mt.mutData.RUnlock()

	val, ok := mt.data[string(key)]

	if !ok {
		return nil, nil
	}

	return val, nil
}

func (mt *TrieMock) Update(key, value []byte) error {
	if mt.Fail {
		return errMockTrie
	}

	mt.mutData.Lock()
	defer mt.mutData.Unlock()

	mt.data[string(key)] = value
	return nil
}

func (mt *TrieMock) Delete(key []byte) error {
	if mt.Fail {
		return errMockTrie
	}

	mt.mutData.Lock()
	defer mt.mutData.Unlock()

	delete(mt.data, string(key))
	return nil
}

func (mt *TrieMock) Root() []byte {
	if len(mt.data) == 0 {
		return make([]byte, encoding.HashLength)
	}

	buff := bytes.Buffer{}

	for key, val := range mt.data {
		buff.Write([]byte(key))
		buff.Write(val)
	}

	hash := HasherMock{}
	return hash.Compute(buff.String())
}

func (mt *TrieMock) Commit(onleaf trie.LeafCallback) (root []byte, err error) {
	if mt.Fail {
		return nil, errMockTrie
	}

	mock := mt.mdbwc.(*DBWriteCacherMock)
	mock.AppendMockTrie(mt)

	return mt.Root(), nil
}

func (mt *TrieMock) DBW() trie.DBWriteCacher {
	return mt.mdbwc
}

func (mt *TrieMock) Recreate(root []byte, dbw trie.DBWriteCacher) (trie.Trier, error) {
	if mt.Fail {
		return nil, errMockTrie
	}

	if mt.FailRecreate {
		return nil, errMockTrie
	}

	mock := dbw.(*DBWriteCacherMock)
	return mock.RetrieveMockTrie(root)
}

func (mt *TrieMock) Copy() trie.Trier {
	return nil
}

type DBWriteCacherMock struct {
	mutTries sync.RWMutex
	tries    map[string]trie.Trier
}

func NewMockDBWriteCacher() *DBWriteCacherMock {
	mdbwc := DBWriteCacherMock{}
	mdbwc.mutTries = sync.RWMutex{}
	mdbwc.tries = make(map[string]trie.Trier)
	return &mdbwc
}

func (mdbw *DBWriteCacherMock) PersistDB() trie.PersisterBatcher {
	panic("implement me")
}

func (mdbw *DBWriteCacherMock) InsertBlob(hash []byte, blob []byte) {
	panic("implement me")
}

func (mdbw *DBWriteCacherMock) Node(hash []byte) ([]byte, error) {
	panic("implement me")
}

func (mdbw *DBWriteCacherMock) Reference(child []byte, parent []byte) {
	panic("implement me")
}

func (mdbw *DBWriteCacherMock) Dereference(root []byte) {
	panic("implement me")
}

func (mdbw *DBWriteCacherMock) Cap(limit float64) error {
	panic("implement me")
}

func (mdbw *DBWriteCacherMock) Commit(node []byte, report bool) error {
	panic("implement me")
}

func (mdbw *DBWriteCacherMock) Size() (float64, float64) {
	panic("implement me")
}

func (mdbw *DBWriteCacherMock) InsertWithLock(hash []byte, blob []byte, node trie.Node) {

}

func (mdbw *DBWriteCacherMock) CachedNode(hash []byte, cachegen uint16) trie.Node {
	panic("implement me")
}

func (mdbw *DBWriteCacherMock) AppendMockTrie(mockTrie *TrieMock) {
	mdbw.mutTries.Lock()
	defer mdbw.mutTries.Unlock()

	mdbw.tries[string(mockTrie.Root())] = mockTrie
}

func (mdbw *DBWriteCacherMock) RetrieveMockTrie(root []byte) (trie.Trier, error) {
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

	return val, nil
}
