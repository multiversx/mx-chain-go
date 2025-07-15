package main

import (
	"fmt"

	"github.com/klauspost/cpuid/v2"
	"github.com/multiversx/mx-chain-go/config"
)

func checkHardwareRequirements(cfg config.HardwareRequirementsConfig) error {
	cpuFlags, err := parseFeatures(cfg.CPUFlags)
	if err != nil {
		return err
	}

	if !cpuid.CPU.Supports(cpuFlags...) {
		return fmt.Errorf("CPU Flags: Streaming SIMD Extensions 4 required")
	}

	return nil
}

func parseFeatures(features []string) ([]cpuid.FeatureID, error) {
	flags := make([]cpuid.FeatureID, 0)

	for _, cpuFlag := range features {
		featureID := cpuid.ParseFeature(cpuFlag)
		if featureID == cpuid.UNKNOWN {
			return nil, fmt.Errorf("CPU Flags: cpu flag %s not found", cpuFlag)
		}

		flags = append(flags, featureID)
	}

	return flags, nil
}
