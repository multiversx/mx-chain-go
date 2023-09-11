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

func (s *statusComponentsHolder) OutportHandler() outport.OutportHandler {
	return s.outportHandler
}

func (s *statusComponentsHolder) SoftwareVersionChecker() statistics.SoftwareVersionChecker {
	return s.softwareVersionChecker
}

func (s *statusComponentsHolder) ManagedPeersMonitor() common.ManagedPeersMonitor {
	return s.managedPeerMonitor
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *statusComponentsHolder) IsInterfaceNil() bool {
	return s == nil
}
