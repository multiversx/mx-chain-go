package chaos

type failType string

const (
	failTypePanic            failType = "panic"
	failTypeReturnError      failType = "returnError"
	failTypeEarlyReturn      failType = "earlyReturn"
	failTypeCorruptSignature failType = "corruptSignature"
	failTypeSleep            failType = "sleep"
)

var knownFailTypes = map[failType]struct{}{
	failTypePanic:            {},
	failTypeReturnError:      {},
	failTypeEarlyReturn:      {},
	failTypeCorruptSignature: {},
	failTypeSleep:            {},
}

type pointName string

const (
	pointShardBlockCreateBlock  pointName = "shardBlockCreateBlock"
	pointShardBlockProcessBlock pointName = "shardBlockProcessBlock"
	pointEpochConfirmed         pointName = "epochConfirmed"

	pointConsensusV1SubroundSignatureDoSignatureJobWhenSingleKey                      pointName = "consensusV1SubroundSignatureDoSignatureJobWhenSingleKey"
	pointConsensusV1SubroundSignatureDoSignatureJobWhenMultiKey                       pointName = "consensusV1SubroundSignatureDoSignatureJobWhenMultiKey"
	pointConsensusV1SubroundEndRoundCheckSignaturesValidity                           pointName = "consensusV1SubroundEndRoundCheckSignaturesValidity"
	pointConsensusV1SubroundEndRoundDoEndRoundJobByLeaderBeforeBroadcastingFinalBlock pointName = "consensusV1SubroundEndRoundDoEndRoundJobByLeaderBeforeBroadcastingFinalBlock"

	pointConsensusV2SubroundBlockDoBlockJob                      pointName = "consensusV2SubroundBlockDoBlockJob"
	pointConsensusV2SubroundSignatureDoSignatureJobWhenSingleKey pointName = "consensusV2SubroundSignatureDoSignatureJobWhenSingleKey"
	pointConsensusV2SubroundSignatureDoSignatureJobWhenMultiKey  pointName = "consensusV2SubroundSignatureDoSignatureJobWhenMultiKey"
)

var knownPoints = map[pointName]struct{}{
	pointShardBlockCreateBlock:  {},
	pointShardBlockProcessBlock: {},
	pointEpochConfirmed:         {},
	pointConsensusV1SubroundSignatureDoSignatureJobWhenSingleKey:                      {},
	pointConsensusV1SubroundSignatureDoSignatureJobWhenMultiKey:                       {},
	pointConsensusV1SubroundEndRoundCheckSignaturesValidity:                           {},
	pointConsensusV1SubroundEndRoundDoEndRoundJobByLeaderBeforeBroadcastingFinalBlock: {},
	pointConsensusV2SubroundBlockDoBlockJob:                                           {},
	pointConsensusV2SubroundSignatureDoSignatureJobWhenSingleKey:                      {},
	pointConsensusV2SubroundSignatureDoSignatureJobWhenMultiKey:                       {},
}

type failureExpressionParameterName string

const (
	parameterPoint           failureExpressionParameterName = "point"
	parameterRandomNumber    failureExpressionParameterName = "randomNumber"
	parameterNow             failureExpressionParameterName = "now"
	parameterUptime          failureExpressionParameterName = "uptime"
	parameterNodeDisplayName failureExpressionParameterName = "nodeDisplayName"
	parameterShard           failureExpressionParameterName = "shard"
	parameterEpoch           failureExpressionParameterName = "epoch"
	parameterRound           failureExpressionParameterName = "round"

	parameterNodeIndex           failureExpressionParameterName = "nodeIndex"
	parameterNodePublicKey       failureExpressionParameterName = "nodePublicKey"
	parameterConsensusSize       failureExpressionParameterName = "consensusSize"
	parameterAmILeader           failureExpressionParameterName = "amILeader"
	parameterBlockNonce          failureExpressionParameterName = "blockNonce"
	parameterBlockIsStartOfEpoch failureExpressionParameterName = "blockIsStartOfEpoch"
)

const (
	defaultConfigFilePath = "./config/chaos.json"
)
