package statusHandler

// PresenterInterface defines the methods that return information about node
type PresenterInterface interface {
	GetAppVersion() string
	GetPublicKeyTxSign() string
	GetPublicKeyBlockSign() string
	GetShardId() uint64
	GetNodeType() string
	GetCountConsensus() uint64
	GetCountLeader() uint64
	GetCountAcceptedBlocks() uint64
	GetIsSyncing() uint64
	GetTxPoolLoad() uint64
	GetNonce() uint64
	GetProbableHighestNonce() uint64
	GetSynchronizedRound() uint64
	GetRoundTime() uint64
	GetLiveValidatorNodes() uint64
	GetConnectedNodes() uint64
	GetNumConnectedPeers() uint64
	GetCurrentRound() uint64
	GetNumTxInBlock() uint64
	GetNumMiniBlocks() uint64
	GetCrossCheckBlockHeight() string
	GetConsensusState() string
	GetConsensusRoundState() string
	GetCpuLoadPercent() uint64
	GetMemLoadPercent() uint64
	GetTotalMem() uint64
	GetMemUsedByNode() uint64
	GetNetworkRecvPercent() uint64
	GetNetworkRecvBps() uint64
	GetNetworkRecvBpsPeak() uint64
	GetNetworkSentPercent() uint64
	GetNetworkSentBps() uint64
	GetNetworkSentBpsPeak() uint64
	GetLogLines() []string
}
