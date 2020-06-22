package immunitycache

func keysAsStrings(keys [][]byte) []string {
	result := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		result[i] = string(keys[i])
	}

	return result
}

func keysAsBytes(keys []string) [][]byte {
	result := make([][]byte, len(keys))
	for i := 0; i < len(keys); i++ {
		result[i] = []byte(keys[i])
	}

	return result
}
