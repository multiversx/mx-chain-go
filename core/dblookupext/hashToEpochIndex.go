//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. epochByHash.proto

package dblookupext

import (
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type epochByHashIndex struct {
	marshalizer marshal.Marshalizer
	storer      storage.Storer
}

func newHashToEpochIndex(storer storage.Storer, marshalizer marshal.Marshalizer) *epochByHashIndex {
	return &epochByHashIndex{
		storer:      storer,
		marshalizer: marshalizer,
	}
}

func (i *epochByHashIndex) getEpochByHash(hash []byte) (uint32, error) {
	rawBytes, err := i.storer.Get(hash)
	if err != nil {
		return 0, err
	}

	record := &EpochByHash{}
	err = i.marshalizer.Unmarshal(record, rawBytes)
	if err != nil {
		return 0, err
	}

	return record.Epoch, nil
}

func (i *epochByHashIndex) saveEpochByHash(hash []byte, epoch uint32) error {
	record := &EpochByHash{
		Epoch: epoch,
	}

	rawBytes, err := i.marshalizer.Marshal(record)
	if err != nil {
		return err
	}

	return i.storer.Put(hash, rawBytes)
}
