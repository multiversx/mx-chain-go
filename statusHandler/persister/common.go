package persister

// GetUint64 will try to convert an interface type in a uint64
// in case of failure wil return 0
func GetUint64(data interface{}) uint64 {
	value, ok := data.(float64)
	if !ok {
		return 0
	}

	return uint64(value)
}
