package mock

import "github.com/multiversx/mx-chain-core-go/data/sovereign"

type DataCodecMock struct {
	SerializeEventDataCalled   func(eventData sovereign.EventData) ([]byte, error)
	DeserializeEventDataCalled func(data []byte) (*sovereign.EventData, error)
	SerializeTokenDataCalled   func(tokenData sovereign.EsdtTokenData) ([]byte, error)
	DeserializeTokenDataCalled func(data []byte) (*sovereign.EsdtTokenData, error)
	GetTokenDataBytesCalled    func(tokenNonce []byte, tokenData []byte) ([]byte, error)
	SerializeOperationCalled   func(operation sovereign.Operation) ([]byte, error)
}

// SerializeEventData -
func (dcm *DataCodecMock) SerializeEventData(eventData sovereign.EventData) ([]byte, error) {
	if dcm.SerializeEventDataCalled != nil {
		return dcm.SerializeEventDataCalled(eventData)
	}

	return nil, nil
}

// DeserializeEventData -
func (dcm *DataCodecMock) DeserializeEventData(data []byte) (*sovereign.EventData, error) {
	if dcm.DeserializeEventDataCalled != nil {
		return dcm.DeserializeEventDataCalled(data)
	}

	return nil, nil
}

// SerializeTokenData -
func (dcm *DataCodecMock) SerializeTokenData(tokenData sovereign.EsdtTokenData) ([]byte, error) {
	if dcm.SerializeTokenDataCalled != nil {
		return dcm.SerializeTokenDataCalled(tokenData)
	}

	return nil, nil
}

// DeserializeTokenData -
func (dcm *DataCodecMock) DeserializeTokenData(data []byte) (*sovereign.EsdtTokenData, error) {
	if dcm.DeserializeTokenDataCalled != nil {
		return dcm.DeserializeTokenDataCalled(data)
	}

	return nil, nil
}

// GetTokenDataBytes -
func (dcm *DataCodecMock) GetTokenDataBytes(tokenNonce []byte, tokenData []byte) ([]byte, error) {
	if dcm.GetTokenDataBytesCalled != nil {
		return dcm.GetTokenDataBytesCalled(tokenNonce, tokenData)
	}

	return nil, nil
}

// SerializeOperation -
func (dcm *DataCodecMock) SerializeOperation(operation sovereign.Operation) ([]byte, error) {
	if dcm.SerializeOperationCalled != nil {
		return dcm.SerializeOperationCalled(operation)
	}

	return nil, nil
}

// IsInterfaceNil -
func (dcm *DataCodecMock) IsInterfaceNil() bool {
	return dcm == nil
}
