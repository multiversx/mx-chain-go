package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
)

type StorerMock struct {
	memDb data.DBWriteCacher
}

func NewStorerMock() *StorerMock {
	memDb, _ := mock.NewMemDbMock()
	return &StorerMock{
		memDb: memDb,
	}
}

func (sm *StorerMock) Put(key, data []byte) error {
	return sm.memDb.Put(key, data)
}

func (sm *StorerMock) Get(key []byte) ([]byte, error) {
	return sm.memDb.Get(key)
}

func (sm *StorerMock) Has(key []byte) error {
	_, err := sm.memDb.Get(key)
	return err
}

func (sm *StorerMock) Remove(key []byte) error {
	panic("imp[lement me")
}

func (sm *StorerMock) ClearCache() {
	panic("imp[lement me")
}

func (sm *StorerMock) DestroyUnit() error {
	panic("imp[lement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (sm *StorerMock) IsInterfaceNil() bool {
	if sm == nil {
		return true
	}
	return false
}
