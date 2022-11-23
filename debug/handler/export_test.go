package handler

func (idh *interceptorDebugHandler) Events() []*event {
	keys := idh.cache.Keys()
	requests := make([]*event, 0, len(keys))
	for _, key := range keys {
		obj, ok := idh.cache.Get(key)
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

func (idh *interceptorDebugHandler) SetTimeHandler(handler func() int64) {
	idh.timestampHandler = handler
}
