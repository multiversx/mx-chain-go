package mock

import (
	"encoding/json"

	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/p2p/factory"
)

// MessageSignerMock implements P2PSigningHandler interface but it does nothing as it is disabled
type MessageSignerMock struct {
}

// Verify does nothing
func (ms *MessageSignerMock) Verify(_ p2p.MessageP2P) error {
	return nil
}

// Serialize will serialize the list of p2p messages
func (ms *MessageSignerMock) Serialize(messages []p2p.MessageP2P) ([]byte, error) {
	messagesBytes, err := json.Marshal(messages)
	if err != nil {
		return nil, err
	}

	return messagesBytes, nil
}

// Deserialize will unmarshal into a list of p2p messages
func (ms *MessageSignerMock) Deserialize(messagesBytes []byte) ([]p2p.MessageP2P, error) {
	var messages []*factory.Message
	err := json.Unmarshal(messagesBytes, &messages)
	if err != nil {
		return nil, err
	}

	messages2 := make([]p2p.MessageP2P, 0)
	for _, msg := range messages {
		messages2 = append(messages2, msg)
	}

	return messages2, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ms *MessageSignerMock) IsInterfaceNil() bool {
	return ms == nil
}
