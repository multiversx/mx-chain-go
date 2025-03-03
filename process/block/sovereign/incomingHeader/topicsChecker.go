package incomingHeader

import (
	"fmt"

	"github.com/multiversx/mx-chain-go/process/block/sovereign/incomingHeader/dto"
)

type topicsChecker struct{}

// NewTopicsChecker creates a topics checker which is able to validate topics
func NewTopicsChecker() *topicsChecker {
	return &topicsChecker{}
}

// CheckValidity will receive the topics and validate them
func (tc *topicsChecker) CheckValidity(topics [][]byte) error {
	// TODO: Check each param validity (e.g. check that topic[0] == valid address)
	if len(topics) < dto.MinTopicsInTransferEvent || len(topics[2:])%dto.NumTransferTopics != 0 {
		log.Error("incomingHeaderHandler.createIncomingSCRs",
			"error", dto.ErrInvalidNumTopicsIncomingEvent,
			"num topics", len(topics),
			"topics", topics)

		return fmt.Errorf("%w for %s; num topics = %d", dto.ErrInvalidNumTopicsIncomingEvent, "eventIDDepositIncomingTransfer", len(topics))
	}

	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (tc *topicsChecker) IsInterfaceNil() bool {
	return tc == nil
}
