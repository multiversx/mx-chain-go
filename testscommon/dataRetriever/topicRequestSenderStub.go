package dataRetriever

import "github.com/ElrondNetwork/elrond-go/dataRetriever"

// TopicRequestSenderStub -
type TopicRequestSenderStub struct {
	SendOnRequestTopicCalled      func(rd *dataRetriever.RequestData, originalHashes [][]byte) error
	SetNumPeersToQueryCalled      func(intra int, cross int)
	GetNumPeersToQueryCalled      func() (int, int)
	RequestTopicCalled            func() string
	TargetShardIDCalled           func() uint32
	SetResolverDebugHandlerCalled func(handler dataRetriever.ResolverDebugHandler) error
	ResolverDebugHandlerCalled    func() dataRetriever.ResolverDebugHandler
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

// SetResolverDebugHandler -
func (stub *TopicRequestSenderStub) SetResolverDebugHandler(handler dataRetriever.ResolverDebugHandler) error {
	if stub.SetResolverDebugHandlerCalled != nil {
		return stub.SetResolverDebugHandlerCalled(handler)
	}
	return nil
}

// ResolverDebugHandler -
func (stub *TopicRequestSenderStub) ResolverDebugHandler() dataRetriever.ResolverDebugHandler {
	if stub.ResolverDebugHandlerCalled != nil {
		return stub.ResolverDebugHandlerCalled()
	}
	return nil
}

// IsInterfaceNil -
func (stub *TopicRequestSenderStub) IsInterfaceNil() bool {
	return stub == nil
}
