package config

import "errors"

var errStakingV4StepsNotInOrder = errors.New("staking v4 enable epoch steps should be in cardinal order(e.g.: StakingV4Step1EnableEpoch = 2, StakingV4Step2EnableEpoch = 3, StakingV4Step3EnableEpoch = 4)")

var errNoMaxNodesConfigBeforeStakingV4 = errors.New("no previous config change entry in MaxNodesChangeEnableEpoch before entry with EpochEnable = StakingV4Step3EnableEpoch")

var errMismatchNodesToShuffle = errors.New("previous MaxNodesChangeEnableEpoch.NodesToShufflePerShard != MaxNodesChangeEnableEpoch.NodesToShufflePerShard with EnableEpoch = StakingV4Step3EnableEpoch")

var errNoMaxNodesConfigChangeForStakingV4 = errors.New("no MaxNodesChangeEnableEpoch config found for EpochEnable = StakingV4Step3EnableEpoch")

var errInvalidMaxMinNodes = errors.New("number of min nodes with hysteresis > number of max nodes")
