package antiflood

func (qfp *quotaFloodPreventer) SetGlobalQuotaValues(maxMessages uint32, size uint64) {
	qfp.mutOperation.Lock()
	qfp.globalQuota.numReceivedMessages = maxMessages
	qfp.globalQuota.sizeReceivedMessages = size
	qfp.mutOperation.Unlock()
}
