package bootstrapStorage_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewBootstrapStorer_NilStorerShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	bt, err := bootstrapStorage.NewBootstrapStorer(marshalizer, nil)

	assert.Nil(t, bt)
	assert.Equal(t, bootstrapStorage.ErrNilBootStorer, err)
}

func TestNewBootstrapStorer_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	storer := &mock.StorerStub{}
	bt, err := bootstrapStorage.NewBootstrapStorer(nil, storer)

	assert.Nil(t, bt)
	assert.Equal(t, bootstrapStorage.ErrNilMarshalizer, err)
}

func TestNewBootstrapStorer_ShouldWork(t *testing.T) {
	t.Parallel()

	storer := &mock.StorerMock{}
	marshalizer := &mock.MarshalizerMock{}
	bt, err := bootstrapStorage.NewBootstrapStorer(marshalizer, storer)

	assert.NotNil(t, bt)
	assert.Nil(t, err)
}

func TestBootstrapStorer_PutAndGet(t *testing.T) {
	t.Parallel()

	numRounds := int64(10)
	round := int64(0)
	storer := mock.NewStorerMock()
	marshalizer := &mock.MarshalizerMock{}
	bt, _ := bootstrapStorage.NewBootstrapStorer(marshalizer, storer)

	headerInfo := bootstrapStorage.BootstrapHeaderInfo{2, 3, []byte("Hash")}
	dataBoot := bootstrapStorage.BootstrapData{
		HeaderInfo:           headerInfo,
		LastNotarizedHeaders: []bootstrapStorage.BootstrapHeaderInfo{headerInfo},
		LastFinals:           []bootstrapStorage.BootstrapHeaderInfo{headerInfo},
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
		round = round - 1
	}
}
