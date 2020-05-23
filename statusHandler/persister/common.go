package persister

import "math/big"

// GetUint64 will try to convert an interface type in a uint64
// in case of failure wil return 0
func GetUint64(data interface{}) uint64 {
	value, ok := data.(uint64)
	if !ok {
		return 0
	}

	return value
}

// GetString will try to convert an interface type in a string
// in case of failure wil return 0
func GetString(data interface{}) string {
	value, ok := data.(string)
	if !ok {
		return ""
	}

	return value
}

// GetBigIntFromString will try to convert an interface type in a string
// and the corresponding *big.Int
func GetBigIntFromString(data interface{}) *big.Int {
	value, ok := data.(string)
	if !ok {
		return big.NewInt(0)
	}

	bigIntValue, ok := big.NewInt(0).SetString(value, 10)
	if !ok {
		return big.NewInt(0)
	}

	return bigIntValue
}
