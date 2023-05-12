package processor

// NewPeerAuthenticationRequestsProcessorWithoutGoRoutine -
func NewPeerAuthenticationRequestsProcessorWithoutGoRoutine(args ArgPeerAuthenticationRequestsProcessor) (*peerAuthenticationRequestsProcessor, error) {
	return newPeerAuthenticationRequestsProcessor(args)
}
