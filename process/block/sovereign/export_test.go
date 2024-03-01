package sovereign

func NewOperationData(topics [][]byte) *operationData {
	return newOperationData(topics)
}

func GetEventData(data []byte) (*eventData, error) {
	return getEventData(data)
}

func SerializeOperationData(opData *operationData) []byte {
	return serializeOperationData(opData)
}
