package leveldb

type result struct {
	value []byte
	err   error
}

type serialQueryer interface {
	request(s *SerialDB)
	resultChan() chan *result
}
