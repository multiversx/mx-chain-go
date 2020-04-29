package resolver

func (ir *interceptorResolver) Events() []*event {
	ir.mutCriticalArea.Lock()
	defer ir.mutCriticalArea.Unlock()

	keys := ir.cache.Keys()
	requests := make([]*event, 0, len(keys))
	for _, key := range keys {
		obj, ok := ir.cache.Get(key)
		if !ok {
			continue
		}

		req, ok := obj.(*event)
		if !ok {
			continue
		}

		requests = append(requests, req)
	}

	return requests
}

func (ir *interceptorResolver) SetTimehandler(handler func() int64) {
	ir.timestampHandler = handler
}
