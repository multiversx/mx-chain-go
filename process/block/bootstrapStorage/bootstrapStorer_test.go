package bootstrapStorage_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

var (
	testMarshalizer = &mock.MarshalizerMock{}
)

func TestNewBootstrapStorer_NilStorerShouldErr(t *testing.T) {
	t.Parallel()

	bt, err := bootstrapStorage.NewBootstrapStorer(testMarshalizer, nil)

	assert.Nil(t, bt)
	assert.Equal(t, bootstrapStorage.ErrNilBootStorer, err)
}

func TestNewBootstrapStorer_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	storer := &storageStubs.StorerStub{}
	bt, err := bootstrapStorage.NewBootstrapStorer(nil, storer)

	assert.Nil(t, bt)
	assert.Equal(t, bootstrapStorage.ErrNilMarshalizer, err)
}

func TestNewBootstrapStorer_ShouldWork(t *testing.T) {
	t.Parallel()

	storer := genericMocks.NewStorerMock()
	bt, err := bootstrapStorage.NewBootstrapStorer(testMarshalizer, storer)

	assert.NotNil(t, bt)
	assert.Nil(t, err)
	assert.False(t, bt.IsInterfaceNil())
}

func TestBootstrapStorer_PutAndGet(t *testing.T) {
	t.Parallel()

	numRounds := int64(10)
	round := int64(0)
	storer := genericMocks.NewStorerMock()
	bt, _ := bootstrapStorage.NewBootstrapStorer(testMarshalizer, storer)

	headerInfo := bootstrapStorage.BootstrapHeaderInfo{ShardId: 2, Nonce: 3, Hash: []byte("Hash")}
	dataBoot := bootstrapStorage.BootstrapData{
		LastHeader:                headerInfo,
		LastCrossNotarizedHeaders: []bootstrapStorage.BootstrapHeaderInfo{headerInfo},
		LastSelfNotarizedHeaders:  []bootstrapStorage.BootstrapHeaderInfo{headerInfo},
	}

	err := bt.Put(round, dataBoot)
	assert.Nil(t, err)

	for i := int64(0); i < numRounds; i++ {
		round = i
		err = bt.Put(round, dataBoot)
		assert.Nil(t, err)
	}

	round = bt.GetHighestRound()
	for i := numRounds - 1; i >= 0; i-- {
		dataBoot.LastRound = i - 1
		if i == 0 {
			dataBoot.LastRound = 0
		}
		data, err := bt.Get(round)
		assert.Nil(t, err)
		assert.Equal(t, dataBoot, data)
		round--
	}
}

func TestBootstrapStorer_SaveLastRound(t *testing.T) {
	t.Parallel()

	putWasCalled := false
	roundInStorage := int64(5)
	marshalizer := &mock.MarshalizerMock{}
	storer := &storageStubs.StorerStub{
		PutCalled: func(key, data []byte) error {
			putWasCalled = true
			rn := bootstrapStorage.RoundNum{}
			err := marshalizer.Unmarshal(&rn, data)
			roundInStorage = rn.Num
			if err != nil {
				fmt.Println(err.Error())
			}
			return nil
		},
		GetCalled: func(key []byte) ([]byte, error) {
			return marshalizer.Marshal(&bootstrapStorage.RoundNum{Num: roundInStorage})
		},
	}
	bt, _ := bootstrapStorage.NewBootstrapStorer(marshalizer, storer)

	assert.Equal(t, roundInStorage, bt.GetHighestRound())
	newRound := int64(37)
	err := bt.SaveLastRound(newRound)
	assert.Equal(t, newRound, bt.GetHighestRound())
	assert.Nil(t, err)
	assert.True(t, putWasCalled)
}
