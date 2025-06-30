package txcache

type trackedBlock struct {
	nonce                uint64
	hash                 []byte
	rootHash             []byte
	prevHash             []byte
	breadcrumbsByAddress map[string]*accountBreadcrumb
}

func newTrackedBlock(nonce uint64, blockHash []byte, rootHash []byte, prevHash []byte) *trackedBlock {
	return &trackedBlock{
		nonce:                nonce,
		hash:                 blockHash,
		rootHash:             rootHash,
		prevHash:             prevHash,
		breadcrumbsByAddress: make(map[string]*accountBreadcrumb),
	}
}

func (tb *trackedBlock) createOrUpdateVirtualRecords(
	session SelectionSession,
	skippedSenders map[string]struct{},
	sendersInContinuityWithSessionNonce map[string]struct{},
	accountPreviousBreadcrumb map[string]*accountBreadcrumb,
	virtualAccountsByAddress map[string]*virtualAccountRecord,
) error {
	for address, breadcrumb := range tb.breadcrumbsByAddress {
		_, ok := skippedSenders[address]
		if ok {
			continue
		}

		accountState, err := session.GetAccountState([]byte(address))
		if err != nil {
			log.Debug("selectionTracker.createVirtualSelectionSession",
				"err", err)
			return err
		}

		accountNonce := accountState.GetNonce()

		if !breadcrumb.isContinuous(address, accountNonce,
			sendersInContinuityWithSessionNonce, accountPreviousBreadcrumb) {
			skippedSenders[address] = struct{}{}
			delete(virtualAccountsByAddress, address)
			continue
		}

		breadcrumb.createOrUpdateVirtualRecord(virtualAccountsByAddress, accountState, address)
	}

	return nil
}

func (st *trackedBlock) sameNonce(trackedBlock1 *trackedBlock) bool {
	return st.nonce == trackedBlock1.nonce
}
