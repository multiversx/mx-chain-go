package containers

func (rc *ResolversContainer) Insert(key string, value interface{}) bool {
	return rc.objects.Insert(key, value)
}
