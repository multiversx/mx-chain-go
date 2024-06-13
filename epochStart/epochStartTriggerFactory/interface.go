package epochStartTriggerFactory

import "github.com/multiversx/mx-chain-go/process"

// EpochStartTriggerFactoryHandler defines the interface needed to create an epoch start trigger
type EpochStartTriggerFactoryHandler interface {
	CreateEpochStartTrigger(args ArgsEpochStartTrigger) (process.EpochStartTriggerHandler, error)
	IsInterfaceNil() bool
}
