package fullHistory

import (
	"github.com/ElrondNetwork/elrond-go/storage"
)

type miniblockHashByTxHashIndex struct {
	storer storage.Storer
}

func newMiniblockHashByTxHashIndex(storer storage.Storer) *miniblockHashByTxHashIndex {
	return &miniblockHashByTxHashIndex{
		storer: storer,
	}
}

func (i *miniblockHashByTxHashIndex) getMiniblockByTx(txHash []byte) ([]byte, error) {
	miniblockHash, err := i.storer.Get(txHash)
	if err != nil {
		return nil, err
	}

	return miniblockHash, nil
}

func (i *miniblockHashByTxHashIndex) putMiniblockByTx(txHash []byte, miniblockHash []byte) error {
	return i.storer.Put(txHash, miniblockHash)
}
