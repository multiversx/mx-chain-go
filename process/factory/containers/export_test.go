package containers

func (ic *InterceptorsContainer) Insert(key string, value interface{}) bool {
	return ic.objects.Insert(key, value)
}

func (rc *ResolversContainer) Insert(key string, value interface{}) bool {
	return rc.objects.Insert(key, value)
}
