package data

// TrimHeaderHandlerSlice creates a copy of the provided slice without the excess capacity
func TrimHeaderHandlerSlice(in []HeaderHandler) []HeaderHandler {
	if len(in) == 0 {
		return []HeaderHandler{}
	}
	ret := make([]HeaderHandler, len(in))
	copy(ret, in)
	return ret
}
