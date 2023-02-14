package config

import (
	"fmt"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("configChecker")

func SanityCheckEnableEpochsStakingV4(cfg *Configs) error {
	enableEpochsCfg := cfg.EpochConfig.EnableEpochs
	err := checkStakingV4EpochsOrder(enableEpochsCfg)
	if err != nil {
		return err
	}

	numOfShards := cfg.GeneralConfig.GeneralSettings.GenesisMaxNumberOfShards
	return checkStakingV4MaxNodesChangeCfg(enableEpochsCfg, numOfShards)
}

func checkStakingV4EpochsOrder(enableEpochsCfg EnableEpochs) error {
	stakingV4StepsInOrder := (enableEpochsCfg.StakingV4Step1EnableEpoch < enableEpochsCfg.StakingV4Step2EnableEpoch) &&
		(enableEpochsCfg.StakingV4Step2EnableEpoch < enableEpochsCfg.StakingV4Step3EnableEpoch)

	if !stakingV4StepsInOrder {
		return errStakingV4StepsNotInOrder
	}

	stakingV4StepsInExpectedOrder := (enableEpochsCfg.StakingV4Step1EnableEpoch == enableEpochsCfg.StakingV4Step2EnableEpoch-1) &&
		(enableEpochsCfg.StakingV4Step2EnableEpoch == enableEpochsCfg.StakingV4Step3EnableEpoch-1)
	if !stakingV4StepsInExpectedOrder {
		log.Warn("staking v4 enable epoch steps should be in cardinal order " +
			"(e.g.: StakingV4Step1EnableEpoch = 2, StakingV4Step2EnableEpoch = 3, StakingV4Step3EnableEpoch = 4)" +
			"; can leave them as they are for playground purposes" +
			", but DO NOT use them in production, since system's behavior is undefined")
	}

	return nil
}

func checkStakingV4MaxNodesChangeCfg(enableEpochsCfg EnableEpochs, numOfShards uint32) error {
	maxNodesConfigAdaptedForStakingV4 := false

	for idx, currMaxNodesChangeCfg := range enableEpochsCfg.MaxNodesChangeEnableEpoch {
		if currMaxNodesChangeCfg.EpochEnable == enableEpochsCfg.StakingV4Step3EnableEpoch {

			maxNodesConfigAdaptedForStakingV4 = true
			if idx == 0 {
				log.Warn(fmt.Sprintf("found config change in MaxNodesChangeEnableEpoch for StakingV4Step3EnableEpoch = %d, ", enableEpochsCfg.StakingV4Step3EnableEpoch) +
					"but no previous config change entry in MaxNodesChangeEnableEpoch, DO NOT use this config in production")
			} else {
				prevMaxNodesChange := enableEpochsCfg.MaxNodesChangeEnableEpoch[idx-1]
				err := checkMaxNodesChangedCorrectly(prevMaxNodesChange, currMaxNodesChangeCfg, numOfShards)
				if err != nil {
					return err
				}
			}

			break
		}
	}

	if !maxNodesConfigAdaptedForStakingV4 {
		return fmt.Errorf("%w = %d", errNoMaxNodesConfigChangeForStakingV4, enableEpochsCfg.StakingV4Step3EnableEpoch)
	}

	return nil
}

func checkMaxNodesChangedCorrectly(prevMaxNodesChange MaxNodesChangeConfig, currMaxNodesChange MaxNodesChangeConfig, numOfShards uint32) error {
	if prevMaxNodesChange.NodesToShufflePerShard != currMaxNodesChange.NodesToShufflePerShard {
		log.Warn("previous MaxNodesChangeEnableEpoch.NodesToShufflePerShard != MaxNodesChangeEnableEpoch.NodesToShufflePerShard" +
			" with EnableEpoch = StakingV4Step3EnableEpoch; can leave them as they are for playground purposes," +
			" but DO NOT use them in production, since this will influence rewards")
	}

	totalShuffled := (numOfShards + 1) * prevMaxNodesChange.NodesToShufflePerShard
	expectedMaxNumNodes := prevMaxNodesChange.MaxNumNodes - totalShuffled
	if expectedMaxNumNodes != currMaxNodesChange.MaxNumNodes {
		return fmt.Errorf("expected MaxNodesChangeEnableEpoch.MaxNumNodes for StakingV4Step3EnableEpoch = %d, but got %d",
			expectedMaxNumNodes, currMaxNodesChange.MaxNumNodes)
	}

	return nil
}
