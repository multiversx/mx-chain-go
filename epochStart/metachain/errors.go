package metachain

import "errors"

var errNilValidatorsInfoMap = errors.New("received nil shard validators info map")

var errCannotComputeDenominator = errors.New("cannot compute denominator value")

var errNilAuctionListDisplayHandler = errors.New("nil auction list display handler provided")

var errNilTableDisplayHandler = errors.New("nil table display handler provided")

var errNegativeAcceleratorReward = errors.New("negative accelerator reward")

var errAcceleratorRewardsMoreThanTotalRewards = errors.New("accelerator rewards more than total rewards")
