package incomingHeader

import (
	"fmt"
)

type topicsChecker struct{}

func NewTopicsChecker() *topicsChecker {
	return &topicsChecker{}
}

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

func (tc *topicsChecker) IsInterfaceNil() bool {
	return tc == nil
}
