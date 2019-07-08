package libp2p_test

import (
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/stretchr/testify/assert"
)

func TestNewMutexHolder_InvalidCapacityShouldErr(t *testing.T) {
	t.Parallel()

	mh, err := libp2p.NewMutexHolder(-1)

	assert.Nil(t, mh)
	assert.NotNil(t, err)
}

func TestNewMutexHolder_InvalidCapacityShouldWork(t *testing.T) {
	t.Parallel()

	mh, err := libp2p.NewMutexHolder(10)

	assert.NotNil(t, mh)
	assert.Nil(t, err)
}

func TestMutexHolder_MutexNotFoundShouldCreate(t *testing.T) {
	t.Parallel()

	mh, _ := libp2p.NewMutexHolder(10)
	key := "key"
	mut := mh.Get(key)

	assert.NotNil(t, mut)
	assert.Equal(t, 1, mh.Mutexes().Len())
	addedMutex, _ := mh.Mutexes().Get([]byte(key))
	//pointer testing to not have the situation of creating new mutexes for each getMutex call
	assert.True(t, mut == addedMutex)
}

func TestMutexHolder_OtherObjectInCacheShouldRewriteWithNewMutexAndReturn(t *testing.T) {
	t.Parallel()

	mh, _ := libp2p.NewMutexHolder(10)
	key := "key"
	mh.Mutexes().Put([]byte(key), "not a mutex value")
	mut := mh.Get(key)

	assert.NotNil(t, mut)
	assert.Equal(t, 1, mh.Mutexes().Len())
	addedMutex, _ := mh.Mutexes().Get([]byte(key))
	//pointer testing to not have the situation of creating new mutexes for each getMutex call
	assert.True(t, mut == addedMutex)
}

func TestMutexHolder_MutexFoundShouldReturnIt(t *testing.T) {
	t.Parallel()

	mh, _ := libp2p.NewMutexHolder(10)
	key := "key"
	mut := &sync.Mutex{}
	mh.Mutexes().Put([]byte(key), mut)
	mutRecov := mh.Get(key)

	assert.NotNil(t, mutRecov)
	assert.Equal(t, 1, mh.Mutexes().Len())
	addedMutex, _ := mh.Mutexes().Get([]byte(key))
	//pointer testing to not have the situation of creating new mutexes for each getMutex call
	assert.True(t, mut == addedMutex)
	assert.True(t, mut == mutRecov)
}
