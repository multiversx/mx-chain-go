//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. hashEpoch.proto

package fullHistory

import (
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type hashEpochProcessor struct {
	marshalizer marshal.Marshalizer
	storer      storage.Storer
}

func newHashEpochStorer(storer storage.Storer, marshalizer marshal.Marshalizer) *hashEpochProcessor {
	return &hashEpochProcessor{
		storer:      storer,
		marshalizer: marshalizer,
	}
}

// GetEpoch wil return epoch for provided hash
func (hep *hashEpochProcessor) GetEpoch(hash []byte) (uint32, error) {
	hashEpochBytes, err := hep.storer.Get(hash)
	if err != nil {
		return 0, err
	}

	hashEpochData := &EpochByHash{}
	err = hep.marshalizer.Unmarshal(hashEpochData, hashEpochBytes)
	if err != nil {
		return 0, err
	}

	return hashEpochData.Epoch, nil
}

// SaveEpoch will save epoch for provided hash
func (hep *hashEpochProcessor) SaveEpoch(hash []byte, epoch uint32) error {
	hashEpochData := &EpochByHash{
		Epoch: epoch,
	}
	hashEpochBytes, err := hep.marshalizer.Marshal(hashEpochData)
	if err != nil {
		return err
	}

	return hep.storer.Put(hash, hashEpochBytes)
}
