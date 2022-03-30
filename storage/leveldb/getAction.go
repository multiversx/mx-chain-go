package leveldb

import "github.com/ElrondNetwork/elrond-go/storage"

type getAction struct {
	key     []byte
	resChan chan *result
}

func (g *getAction) request(s *SerialDB) {
	data, err := g.doGetRequest(s)

	res := &result{
		value: data,
		err:   err,
	}
	g.resChan <- res
}

func (g *getAction) doGetRequest(s *SerialDB) ([]byte, error) {
	db := s.getDbPointer()
	if db == nil {
		return nil, storage.ErrDBIsClosed
	}

	return db.Get(g.key, nil)
}

func (g *getAction) resultChan() chan *result {
	return g.resChan
}
