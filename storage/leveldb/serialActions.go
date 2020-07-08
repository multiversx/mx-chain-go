package leveldb

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type putBatchAct struct {
	batch   *batch
	resChan chan<- *pairResult
}

type pairResult struct {
	value                 []byte
	err                   error
	timestampBeforeAccess time.Time
	timestampAfterAccess  time.Time
}

type serialQueryer interface {
	request(s *SerialDB)
}

type getAct struct {
	key     []byte
	resChan chan<- *pairResult
}

type hasAct struct {
	key     []byte
	resChan chan<- *pairResult
}

func (p *putBatchAct) request(s *SerialDB) {
	wopt := &opt.WriteOptions{
		Sync: true,
	}

	timeBefore := time.Now()
	err := s.db.Write(p.batch.batch, wopt)
	timeAfter := time.Now()

	res := &pairResult{
		value:                 nil,
		err:                   err,
		timestampBeforeAccess: timeBefore,
		timestampAfterAccess:  timeAfter,
	}

	p.resChan <- res
}

func (g *getAct) request(s *SerialDB) {
	timeBefore := time.Now()
	data, err := s.db.Get(g.key, nil)
	timeAfter := time.Now()

	res := &pairResult{
		value:                 data,
		err:                   err,
		timestampBeforeAccess: timeBefore,
		timestampAfterAccess:  timeAfter,
	}

	g.resChan <- res
}

func (h *hasAct) request(s *SerialDB) {
	timeBefore := time.Now()
	has, err := s.db.Has(h.key, nil)
	timeAfter := time.Now()

	res := &pairResult{
		value:                 nil,
		err:                   err,
		timestampBeforeAccess: timeBefore,
		timestampAfterAccess:  timeAfter,
	}

	if err != nil {
		h.resChan <- res
		return
	}

	if has {
		res.err = nil
		h.resChan <- res
		return
	}

	res.err = storage.ErrKeyNotFound
	h.resChan <- res
}
