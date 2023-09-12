package processingOnlyNode

import (
	"time"

	outportCfg "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/testscommon"
)

type statusComponentsHolder struct {
	outportHandler         outport.OutportHandler
	softwareVersionChecker statistics.SoftwareVersionChecker
	managedPeerMonitor     common.ManagedPeersMonitor
}

// CreateStatusComponentsHolder will create a new instance of status components holder
func CreateStatusComponentsHolder(shardID uint32) (factory.StatusComponentsHolder, error) {
	var err error
	instance := &statusComponentsHolder{}

	// TODO add drivers to index data
	instance.outportHandler, err = outport.NewOutport(100*time.Millisecond, outportCfg.OutportConfig{
		ShardID: shardID,
	})
	if err != nil {
		return nil, err
	}
	instance.softwareVersionChecker = &mock.SoftwareVersionCheckerMock{}
	instance.managedPeerMonitor = &testscommon.ManagedPeersMonitorStub{}

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

// IsInterfaceNil returns true if there is no value under the interface
func (s *statusComponentsHolder) IsInterfaceNil() bool {
	return s == nil
}
