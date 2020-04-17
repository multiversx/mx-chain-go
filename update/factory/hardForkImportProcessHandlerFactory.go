package factory

import "github.com/ElrondNetwork/elrond-go/update"

// ArgsHardForkImportProcessHandlerFactory
type ArgsHardForkImportProcessFactory struct {
}

type hardForkFactory struct {
}

// NewHardForkImportProcessFactory will create the components needed for process after hardfork
func NewHardForkImportProcessFactory(args ArgsHardForkImportProcessFactory) (*hardForkFactory, error) {
	return &hardForkFactory{}, nil
}

// Create makes a hardForkImportProcessHandler
func (h *hardForkFactory) Create() (update.HardForkImportProcessHandler, error) {
	return nil, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (h *hardForkFactory) IsInterfaceNil() bool {
	return h == nil
}
