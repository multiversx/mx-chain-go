package containers

// AddInterface -
func (a *accountDBSyncers) AddInterface(key string, val interface{}) error {
	a.objects.Insert(key, val)

	return nil
}

// AddInterface -
func (t *trieSyncers) AddInterface(key string, val interface{}) error {
	t.objects.Insert(key, val)

	return nil
}
