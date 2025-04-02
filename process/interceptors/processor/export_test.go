package processor

// RegisterHandler registers a callback function to be notified of incoming headers
func (hip *HdrInterceptorProcessor) RegisteredHandlers() []func(topic string, hash []byte, data interface{}) {
	hip.mutHandlers.Lock()
	defer hip.mutHandlers.Unlock()

	return hip.registeredHandlers
}

// GetEligibleNodesMap -
func (epip *equivalentProofsInterceptorProcessor) GetEligibleNodesMap() map[uint32]map[uint32]map[string]struct{} {
	epip.mutEligibleNodesMap.RLock()
	defer epip.mutEligibleNodesMap.RUnlock()
	eligibleMap := epip.eligibleNodesMap

	return eligibleMap
}
