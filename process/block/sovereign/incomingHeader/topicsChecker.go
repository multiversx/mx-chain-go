package incomingHeader

import (
	"fmt"
)

type topicsChecker struct{}

// NewTopicsChecker creates a topics checker which is able to validate topics
func NewTopicsChecker() *topicsChecker {
	return &topicsChecker{}
}

// CheckValidity will receive the topics and validate them
func (tc *topicsChecker) CheckValidity(topics [][]byte) error {
	// TODO: Check each param validity (e.g. check that topic[0] == valid address)
	if len(topics) < minTopicsInTransferEvent || len(topics[2:])%numTransferTopics != 0 {
		log.Error("incomingHeaderHandler.createIncomingSCRs",
			"error", errInvalidNumTopicsIncomingEvent,
			"num topics", len(topics),
			"topics", topics)

		return fmt.Errorf("%w for %s; num topics = %d", errInvalidNumTopicsIncomingEvent, "eventIDDepositIncomingTransfer", len(topics))
	}

	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (tc *topicsChecker) IsInterfaceNil() bool {
	return tc == nil
}
