package softwareVersion

import (
	"context"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
)

type tagVersion struct {
	TagVersion string `json:"tag_name"`
}

// SoftwareVersionChecker is a component which is used to check if a new software stable tag is available
type SoftwareVersionChecker struct {
	statusHandler             core.AppStatusHandler
	stableTagProvider         StableTagProviderHandler
	mostRecentSoftwareVersion string
	checkInterval             time.Duration
	closeFunc                 func()
}

var log = logger.GetOrCreate("core/statistics")

// NewSoftwareVersionChecker will create an object for software  version checker
func NewSoftwareVersionChecker(
	appStatusHandler core.AppStatusHandler,
	stableTagProvider StableTagProviderHandler,
	pollingIntervalInMinutes int,
) (*SoftwareVersionChecker, error) {
	if check.IfNil(appStatusHandler) {
		return nil, core.ErrNilAppStatusHandler
	}
	if check.IfNil(stableTagProvider) {
		return nil, core.ErrNilStatusTagProvider
	}
	if pollingIntervalInMinutes <= 0 {
		return nil, core.ErrInvalidPollingInterval
	}

	checkInterval := time.Duration(pollingIntervalInMinutes) * time.Minute

	return &SoftwareVersionChecker{
		statusHandler:             appStatusHandler,
		stableTagProvider:         stableTagProvider,
		mostRecentSoftwareVersion: "",
		checkInterval:             checkInterval,
		closeFunc:                 nil,
	}, nil
}

// StartCheckSoftwareVersion will check on a specific interval if a new software version is available
func (svc *SoftwareVersionChecker) StartCheckSoftwareVersion() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	svc.closeFunc = cancelFunc
	go svc.checkSoftwareVersion(ctx)
}

func (svc *SoftwareVersionChecker) checkSoftwareVersion(ctx context.Context) {
	for {
		svc.readLatestStableVersion()

		select {
		case <-ctx.Done():
			log.Debug("softwareVersionChecker's go routine is stopping...")
			return
		case <-time.After(svc.checkInterval):
		}
	}
}

func (svc *SoftwareVersionChecker) readLatestStableVersion() {
	tagVersionFromURL, err := svc.stableTagProvider.FetchTagVersion()
	if err != nil {
		log.Debug("cannot read json with latest stable tag", "error", err)
		return
	}
	if tagVersionFromURL != "" {
		svc.mostRecentSoftwareVersion = tagVersionFromURL
	}

	svc.statusHandler.SetStringValue(core.MetricLatestTagSoftwareVersion, svc.mostRecentSoftwareVersion)
}

// IsInterfaceNil returns true if there is no value under the interface
func (svc *SoftwareVersionChecker) IsInterfaceNil() bool {
	return svc == nil
}

// Close will handle the closing of opened go routines
func (svc *SoftwareVersionChecker) Close() error {
	if svc.closeFunc != nil {
		svc.closeFunc()
	}
	return nil
}
