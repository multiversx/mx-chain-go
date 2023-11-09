package p2pQuota

import "github.com/multiversx/mx-chain-core-go/core"

type Quota struct {
	*quota
}

func (q *Quota) NumReceived() uint32 {
	return q.numReceivedMessages
}

func (q *Quota) SizeReceived() uint64 {
	return q.sizeReceivedMessages
}

func (q *Quota) NumProcessed() uint32 {
	return q.numProcessedMessages
}

func (q *Quota) SizeProcessed() uint64 {
	return q.sizeProcessedMessages
}

func (pqp *p2pQuotaProcessor) GetQuota(pid core.PeerID) *Quota {
	pqp.mutStatistics.Lock()
	q := pqp.statistics[pid]
	pqp.mutStatistics.Unlock()

	if q == nil {
		return nil
	}

	return &Quota{quota: q}
}
