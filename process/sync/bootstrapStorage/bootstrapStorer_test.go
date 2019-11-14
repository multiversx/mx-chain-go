package bootstrapStorage

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapStorer_PutAndGet(t *testing.T) {
	t.Parallel()

	numRounds := int64(10)

	storer := mock.NewStorerMock()
	marshalizer := &mock.MarshalizerMock{}
	bt, _ := NewBootstrapStorer(marshalizer, storer)

	headerInfo := BootstrapHeaderInfo{2, 3, []byte("Hash")}
	dataBoot := BootstrapData{
		Round:                0,
		HeaderInfo:           headerInfo,
		LastNotarizedHeaders: []BootstrapHeaderInfo{headerInfo},
		LastFinal:            []BootstrapHeaderInfo{headerInfo},
	}

	err := bt.Put(dataBoot)
	assert.Nil(t, err)

	for i := int64(0); i < numRounds; i++ {
		dataBoot.Round = i
		err = bt.Put(dataBoot)
		assert.Nil(t, err)
	}

	dataBoot.Round = bt.GetHighestRound()
	for i := numRounds - 1; i >= 0; i-- {
		dataBoot.LastRound = i - 1
		if i == 0 {
			dataBoot.LastRound = 0
		}
		data, err := bt.Get(dataBoot.Round)
		assert.Nil(t, err)
		assert.Equal(t, dataBoot, data)
		dataBoot.Round = dataBoot.Round - 1
	}
}
