package antiflood

func (d *debugger) GetData(key []byte) *event {
	obj, ok := d.cache.Get(key)
	if !ok {
		return nil
	}

	return obj.(*event)
}
