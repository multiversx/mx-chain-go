package chaos

type failureName string

const (
	failureCreatingBlockError                              failureName = "creatingBlockError"
	failureProcessingBlockError                            failureName = "processingBlockError"
	failurePanicOnEpochChange                              failureName = "panicOnEpochChange"
	failureConsensusCorruptSignature                       failureName = "consensusCorruptSignature"
	failureConsensusV1ReturnErrorInCheckSignaturesValidity failureName = "consensusV1ReturnErrorInCheckSignaturesValidity"
	failureConsensusV1DelayBroadcastingFinalBlockAsLeader  failureName = "consensusV1DelayBroadcastingFinalBlockAsLeader"
	failureConsensusV2CorruptLeaderSignature               failureName = "consensusV2CorruptLeaderSignature"
	failureConsensusV2DelayLeaderSignature                 failureName = "consensusV2DelayLeaderSignature"
	failureConsensusV2SkipSendingBlock                     failureName = "consensusV2SkipSendingBlock"
)

type failureExpressionParameterName string

const (
	parameterRandomNumber    failureExpressionParameterName = "randomNumber"
	parameterNow             failureExpressionParameterName = "now"
	parameterNodeDisplayName failureExpressionParameterName = "nodeDisplayName"
	parameterShard           failureExpressionParameterName = "shard"
	parameterEpoch           failureExpressionParameterName = "epoch"
	parameterRound           failureExpressionParameterName = "round"

	parameterNodeIndex     failureExpressionParameterName = "nodeIndex"
	parameterNodePublicKey failureExpressionParameterName = "nodePublicKey"
	parameterConsensusSize failureExpressionParameterName = "consensusSize"
	parameterAmILeader     failureExpressionParameterName = "amILeader"
	parameterBlockNonce    failureExpressionParameterName = "blockNonce"
)

const (
	defaultConfigFilePath = "./config/chaos.json"
)
