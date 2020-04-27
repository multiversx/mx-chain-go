package view

// Presenter defines the methods that return information about node
type Presenter interface {
	GetAppVersion() string
	GetNodeName() string
	GetPublicKeyBlockSign() string
	GetShardId() uint64
	GetNodeType() string
	GetPeerType() string
	GetCountConsensus() uint64
	GetCountConsensusAcceptedBlocks() uint64
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
	GetNumTxProcessed() uint64
	GetCurrentBlockHash() string
	GetEpochNumber() uint64
	CalculateTimeToSynchronize() string
	CalculateSynchronizationSpeed() uint64
	GetCurrentRoundTimestamp() uint64
	GetBlockSize() uint64
	GetNumShardHeadersInPool() uint64
	GetNumShardHeadersProcessed() uint64
	GetHighestFinalBlockInShard() uint64
	CheckSoftwareVersion() (bool, string)

	GetTotalRewardsValue() (string, string)
	CalculateRewardsPerHour() string
	GetZeros() string

	// IsInterfaceNil returns true if there is no value under the interface
	IsInterfaceNil() bool
}
