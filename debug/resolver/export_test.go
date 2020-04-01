package resolver

func (ir *interceptorResolver) Requests() []*request {
	ir.mutCriticalArea.Lock()
	defer ir.mutCriticalArea.Unlock()

	keys := ir.cache.Keys()
	requests := make([]*request, 0, len(keys))
	for _, key := range keys {
		obj, ok := ir.cache.Get(key)
		if !ok {
			continue
		}

		req, ok := obj.(*request)
		if !ok {
			continue
		}

		requests = append(requests, req)
	}

	return requests
}
