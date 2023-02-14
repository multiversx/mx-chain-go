package config

import "errors"

var errStakingV4StepsNotInOrder = errors.New("staking v4 enable epochs are not in ascending order; expected StakingV4Step1EnableEpoch < StakingV4Step2EnableEpoch < StakingV4Step3EnableEpoch")

var errNoMaxNodesConfigChangeForStakingV4 = errors.New("no MaxNodesChangeEnableEpoch config found for EpochEnable = StakingV4Step3EnableEpoch")
