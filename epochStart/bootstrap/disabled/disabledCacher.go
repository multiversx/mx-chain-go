package disabled

type cacher struct {
}

// NewCacher returns a new instance of cacher
func NewCacher() *cacher {
	return &cacher{}
}

func (d *cacher) Clear() {
}

func (d *cacher) Put(key []byte, value interface{}) bool {
	return true
}

func (d *cacher) Get(key []byte) (value interface{}, ok bool) {
	return nil, false
}

func (d *cacher) Has(key []byte) bool {
	panic("implement me")
}

func (d *cacher) Peek(key []byte) (value interface{}, ok bool) {
	panic("implement me")
}

func (d *cacher) HasOrAdd(key []byte, value interface{}) (ok, evicted bool) {
	panic("implement me")
}

func (d *cacher) Remove(key []byte) {
	panic("implement me")
}

func (d *cacher) RemoveOldest() {
	panic("implement me")
}

func (d *cacher) Keys() [][]byte {
	panic("implement me")
}

func (d *cacher) Len() int {
	panic("implement me")
}

func (d *cacher) MaxSize() int {
	panic("implement me")
}

func (d *cacher) RegisterHandler(func(key []byte)) {
	panic("implement me")
}

func (d *cacher) IsInterfaceNil() bool {
	panic("implement me")
}
