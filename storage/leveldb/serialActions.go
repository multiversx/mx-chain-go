package leveldb

import (
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type putBatchAct struct {
	batch   *batch
	resChan chan<- error
}

type pairResult struct {
	value []byte
	err   error
}

type serialQueryer interface {
	request(s *SerialDB)
}

type getAct struct {
	key     []byte
	val     []byte
	resChan chan<- *pairResult
}

type delAct struct {
	key     []byte
	resChan chan<- error
}

type hasAct struct {
	key     []byte
	resChan chan<- error
}

func (p *putBatchAct) request(s *SerialDB) {
	wopt := &opt.WriteOptions{
		Sync: true,
	}

	err := s.db.Write(p.batch.batch, wopt)
	p.resChan <- err
}

func (g *getAct) request(s *SerialDB) {
	data, err := s.db.Get(g.key, nil)

	res := &pairResult{
		value: data,
		err:   err,
	}
	g.resChan <- res
}

func (d *delAct) request(s *SerialDB) {
	err := s.db.Delete(d.key, nil)

	d.resChan <- err
}

func (h *hasAct) request(s *SerialDB) {
	has, err := s.db.Has(h.key, nil)

	if err != nil {
		h.resChan <- err
		return
	}

	if has {
		h.resChan <- nil
		return
	}

	h.resChan <- storage.ErrKeyNotFound
}
