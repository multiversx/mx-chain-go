package leveldb

import (
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type putBatchAction struct {
	batch   *batch
	resChan chan *result
}

func (p *putBatchAction) request(s *SerialDB) {
	p.resChan <- &result{
		err: p.doPutRequest(s),
	}
}

func (p *putBatchAction) doPutRequest(s *SerialDB) error {
	db := s.getDbPointer()
	if db == nil {
		return storage.ErrDBIsClosed
	}

	wopt := &opt.WriteOptions{
		Sync: true,
	}

	return db.Write(p.batch.batch, wopt)
}

func (p *putBatchAction) resultChan() chan *result {
	return p.resChan
}
