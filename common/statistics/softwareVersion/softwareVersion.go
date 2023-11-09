package softwareVersion

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type tagVersion struct {
	TagVersion string `json:"tag_name"`
}

// softwareVersionChecker is a component which is used to check if a new software stable tag is available
type softwareVersionChecker struct {
	statusHandler             core.AppStatusHandler
	stableTagProvider         StableTagProviderHandler
	mostRecentSoftwareVersion string
	checkInterval             time.Duration
	closeFunc                 func()
}

var log = logger.GetOrCreate("common/statistics")

// NewSoftwareVersionChecker will create an object for software  version checker
func NewSoftwareVersionChecker(
	appStatusHandler core.AppStatusHandler,
	stableTagProvider StableTagProviderHandler,
	pollingIntervalInMinutes int,
) (*softwareVersionChecker, error) {
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

	return &softwareVersionChecker{
		statusHandler:             appStatusHandler,
		stableTagProvider:         stableTagProvider,
		mostRecentSoftwareVersion: "",
		checkInterval:             checkInterval,
		closeFunc:                 nil,
	}, nil
}

// StartCheckSoftwareVersion will check on a specific interval if a new software version is available
func (svc *softwareVersionChecker) StartCheckSoftwareVersion() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	svc.closeFunc = cancelFunc
	go svc.checkSoftwareVersion(ctx)
}

func (svc *softwareVersionChecker) checkSoftwareVersion(ctx context.Context) {
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

func (svc *softwareVersionChecker) readLatestStableVersion() {
	tagVersionFromURL, err := svc.stableTagProvider.FetchTagVersion()
	if err != nil {
		log.Debug("cannot read json with latest stable tag", "error", err)
		return
	}
	if tagVersionFromURL != "" {
		svc.mostRecentSoftwareVersion = tagVersionFromURL
	}

	svc.statusHandler.SetStringValue(common.MetricLatestTagSoftwareVersion, svc.mostRecentSoftwareVersion)
}

// IsInterfaceNil returns true if there is no value under the interface
func (svc *softwareVersionChecker) IsInterfaceNil() bool {
	return svc == nil
}

// Close will handle the closing of opened go routines
func (svc *softwareVersionChecker) Close() error {
	if svc.closeFunc != nil {
		svc.closeFunc()
	}
	return nil
}
