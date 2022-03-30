package leveldb

import "github.com/ElrondNetwork/elrond-go/storage"

type hasAction struct {
	key     []byte
	resChan chan *result
}

func (h *hasAction) request(s *SerialDB) {
	has, err := h.doHasRequest(s)

	if err != nil {
		h.resChan <- &result{
			err: err,
		}
		return
	}

	if has {
		h.resChan <- &result{}
		return
	}

	h.resChan <- &result{
		err: storage.ErrKeyNotFound,
	}
}

func (h *hasAction) doHasRequest(s *SerialDB) (bool, error) {
	db := s.getDbPointer()
	if db == nil {
		return false, storage.ErrDBIsClosed
	}

	return db.Has(h.key, nil)
}

func (h *hasAction) resultChan() chan *result {
	return h.resChan
}
