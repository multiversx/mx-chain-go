package components

import (
	"time"

	outportCfg "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
)

type statusComponentsHolder struct {
	closeHandler           *closeHandler
	outportHandler         outport.OutportHandler
	softwareVersionChecker statistics.SoftwareVersionChecker
	managedPeerMonitor     common.ManagedPeersMonitor
}

// CreateStatusComponents will create a new instance of status components holder
func CreateStatusComponents(shardID uint32) (factory.StatusComponentsHandler, error) {
	var err error
	instance := &statusComponentsHolder{
		closeHandler: NewCloseHandler(),
	}

	// TODO add drivers to index data
	instance.outportHandler, err = outport.NewOutport(100*time.Millisecond, outportCfg.OutportConfig{
		ShardID: shardID,
	})
	if err != nil {
		return nil, err
	}
	instance.softwareVersionChecker = &mock.SoftwareVersionCheckerMock{}
	instance.managedPeerMonitor = &testscommon.ManagedPeersMonitorStub{}

	instance.collectClosableComponents()

	return instance, nil
}

// OutportHandler will return the outport handler
func (s *statusComponentsHolder) OutportHandler() outport.OutportHandler {
	return s.outportHandler
}

// SoftwareVersionChecker will return the software version checker
func (s *statusComponentsHolder) SoftwareVersionChecker() statistics.SoftwareVersionChecker {
	return s.softwareVersionChecker
}

// ManagedPeersMonitor will return the managed peers monitor
func (s *statusComponentsHolder) ManagedPeersMonitor() common.ManagedPeersMonitor {
	return s.managedPeerMonitor
}

func (s *statusComponentsHolder) collectClosableComponents() {
	s.closeHandler.AddComponent(s.outportHandler)
	s.closeHandler.AddComponent(s.softwareVersionChecker)
}

// Close will call the Close methods on all inner components
func (s *statusComponentsHolder) Close() error {
	return s.closeHandler.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *statusComponentsHolder) IsInterfaceNil() bool {
	return s == nil
}

// Create will do nothing
func (s *statusComponentsHolder) Create() error {
	return nil
}

// CheckSubcomponents will do nothing
func (s *statusComponentsHolder) CheckSubcomponents() error {
	return nil
}

// String will do nothing
func (s *statusComponentsHolder) String() string {
	return ""
}

// SetForkDetector will do nothing
func (s *statusComponentsHolder) SetForkDetector(_ process.ForkDetector) error {
	return nil
}

// StartPolling will do nothing
func (s *statusComponentsHolder) StartPolling() error {
	// todo check if this method
	return nil
}
