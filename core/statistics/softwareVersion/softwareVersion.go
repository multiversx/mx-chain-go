package softwareVersion

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/logger"
)

const checkInterval = time.Hour + 5*time.Minute

type tagVersion struct {
	TagVersion string `json:"tag_name"`
}

// SoftwareVersionChecker is a component which is used to check if a new software stable tag is available
type SoftwareVersionChecker struct {
	statusHandler             core.AppStatusHandler
	stableTagProvider         StableTagProviderHandler
	mostRecentSoftwareVersion string
	checkRandInterval         time.Duration
}

var log = logger.GetOrCreate("core/statistics")

// NewSoftwareVersionChecker will create an object for software  version checker
func NewSoftwareVersionChecker(appStatusHandler core.AppStatusHandler, stableTagProvider StableTagProviderHandler) (*SoftwareVersionChecker, error) {
	if check.IfNil(appStatusHandler) {
		return nil, core.ErrNilAppStatusHandler
	}
	if check.IfNil(stableTagProvider) {
		return nil, core.ErrNilStatusTagProvider
	}

	// check interval will be a random duration in the interval [1hour5minutes , 1hour20minutes]
	randBigInt, err := rand.Int(rand.Reader, big.NewInt(15))
	if err != nil {
		return nil, err
	}

	randInt := randBigInt.Int64()
	randInterval := time.Duration(randInt)
	checkRandInterval := checkInterval + randInterval*time.Minute

	return &SoftwareVersionChecker{
		statusHandler:             appStatusHandler,
		stableTagProvider:         stableTagProvider,
		mostRecentSoftwareVersion: "",
		checkRandInterval:         checkRandInterval,
	}, nil
}

// StartCheckSoftwareVersion will check on a specific interval if a new software version is available
func (svc *SoftwareVersionChecker) StartCheckSoftwareVersion() {
	go func() {
		for {
			svc.readLatestStableVersion()
			time.Sleep(svc.checkRandInterval)
		}
	}()
}

func (svc *SoftwareVersionChecker) readLatestStableVersion() {
	tagVersionFromUrl, err := svc.stableTagProvider.FetchTagVersion()
	if err != nil {
		log.Debug("cannot read json with latest stable tag", err)
		return
	}
	if tagVersionFromUrl != "" {
		svc.mostRecentSoftwareVersion = tagVersionFromUrl
	}

	svc.statusHandler.SetStringValue(core.MetricLatestTagSoftwareVersion, svc.mostRecentSoftwareVersion)
}
