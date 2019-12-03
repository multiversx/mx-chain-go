package softwareVersion

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/logger"
)

const checkInterval = time.Hour + 5*time.Minute
const stableTagLocation = "https://api.github.com/repos/ElrondNetwork/elrond-go/releases/latest"

type tagVersion struct {
	TagVersion string `json:"tag_name"`
}

type SoftwareVersionChecker struct {
	statusHandler             core.AppStatusHandler
	mostRecentSoftwareVersion string
	checkRandInterval         time.Duration
}

var log = logger.GetOrCreate("core/statistics")

// NewSoftwareVersionChecker will create an object for software  version checker
func NewSoftwareVersionChecker(appStatusHandler core.AppStatusHandler) (*SoftwareVersionChecker, error) {
	if appStatusHandler == nil || appStatusHandler.IsInterfaceNil() {
		return nil, core.ErrNilAppStatusHandler
	}

	// check interval will be random in a interval [1hour, 1hour 15minutes]
	randInterval := time.Duration(rand.Int() % 15)
	checkRandInterval := checkInterval + randInterval*time.Minute

	return &SoftwareVersionChecker{
		statusHandler:             appStatusHandler,
		mostRecentSoftwareVersion: "",
		checkRandInterval:         checkRandInterval,
	}, nil
}

// StartCheckSoftwareVersion will check on a specific interval if a new software version is available
func (svc *SoftwareVersionChecker) StartCheckSoftwareVersion() {
	go func() {
		svc.readLatestStableVersion()
		for {
			select {
			case <-time.After(svc.checkRandInterval):
				svc.readLatestStableVersion()
			}
		}
	}()
}

func (svc *SoftwareVersionChecker) readLatestStableVersion() {
	tagVersion, err := readJSONFromUrl(stableTagLocation)
	if err != nil {
		log.Debug("cannot read json with latest stable tag", err)
		return
	}
	if tagVersion != "" {
		svc.mostRecentSoftwareVersion = tagVersion
	}

	svc.statusHandler.SetStringValue(core.MetricLatestTagSoftwareVersion, svc.mostRecentSoftwareVersion)
}

func readJSONFromUrl(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Debug(err.Error())
		}
	}()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	respBytes := buf.Bytes()

	var tag tagVersion
	if err = json.Unmarshal(respBytes, &tag); err != nil {
		return "", err
	}

	return tag.TagVersion, nil
}
