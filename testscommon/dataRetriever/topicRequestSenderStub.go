package dataRetriever

import "github.com/multiversx/mx-chain-go/dataRetriever"

// TopicRequestSenderStub -
type TopicRequestSenderStub struct {
	SendOnRequestTopicCalled func(rd *dataRetriever.RequestData, originalHashes [][]byte) error
	SetNumPeersToQueryCalled func(intra int, cross int)
	GetNumPeersToQueryCalled func() (int, int)
	RequestTopicCalled       func() string
	TargetShardIDCalled      func() uint32
	SetDebugHandlerCalled    func(handler dataRetriever.DebugHandler) error
	DebugHandlerCalled       func() dataRetriever.DebugHandler
}

// SendOnRequestTopic -
func (stub *TopicRequestSenderStub) SendOnRequestTopic(rd *dataRetriever.RequestData, originalHashes [][]byte) error {
	if stub.SendOnRequestTopicCalled != nil {
		return stub.SendOnRequestTopicCalled(rd, originalHashes)
	}
	return nil
}

// SetNumPeersToQuery -
func (stub *TopicRequestSenderStub) SetNumPeersToQuery(intra int, cross int) {
	if stub.SetNumPeersToQueryCalled != nil {
		stub.SetNumPeersToQueryCalled(intra, cross)
	}
}

// NumPeersToQuery -
func (stub *TopicRequestSenderStub) NumPeersToQuery() (int, int) {
	if stub.GetNumPeersToQueryCalled != nil {
		return stub.GetNumPeersToQueryCalled()
	}
	return 0, 0
}

// RequestTopic -
func (stub *TopicRequestSenderStub) RequestTopic() string {
	if stub.RequestTopicCalled != nil {
		return stub.RequestTopicCalled()
	}
	return ""
}

// TargetShardID -
func (stub *TopicRequestSenderStub) TargetShardID() uint32 {
	if stub.TargetShardIDCalled != nil {
		return stub.TargetShardIDCalled()
	}
	return 0
}

// SetDebugHandler -
func (stub *TopicRequestSenderStub) SetDebugHandler(handler dataRetriever.DebugHandler) error {
	if stub.SetDebugHandlerCalled != nil {
		return stub.SetDebugHandlerCalled(handler)
	}
	return nil
}

// DebugHandler -
func (stub *TopicRequestSenderStub) DebugHandler() dataRetriever.DebugHandler {
	if stub.DebugHandlerCalled != nil {
		return stub.DebugHandlerCalled()
	}
	return nil
}

// IsInterfaceNil -
func (stub *TopicRequestSenderStub) IsInterfaceNil() bool {
	return stub == nil
}
