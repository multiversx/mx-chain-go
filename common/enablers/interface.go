package enablers

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

// EnableEpochsFactory defines enable epochs handler factory behavior
type EnableEpochsFactory interface {
	CreateEnableEpochsHandler(epochConfig config.EpochConfig, epochNotifier process.EpochNotifier) (common.EnableEpochsHandler, error)
	IsInterfaceNil() bool
}
