package chaos

type failureName string

const (
	failureProcessingBlockError                            failureName = "processingBlockError"
	failureProcessingTransactionError                      failureName = "processingTransactionError"
	failureConsensusCorruptSignature                       failureName = "consensusCorruptSignature"
	failureConsensusV1SkipWaitingForSignatures             failureName = "consensusV1SkipWaitingForSignatures"
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
	parameterAmILeader     failureExpressionParameterName = "amILeader"
	parameterBlockNonce    failureExpressionParameterName = "blockNonce"
)
