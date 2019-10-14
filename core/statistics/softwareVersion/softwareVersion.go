package softwareVersion

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/appStatusPolling"
)

const checkInterval = time.Hour
const stableTagLocation = "https://api.github.com/repos/ElrondNetwork/elrond-go/releases/latest"

type tagVersion struct {
	TagVersion string `json:"tag_name"`
}

type softwareVersionChecker struct {
	statusHandler core.AppStatusHandler
}

// NewSoftwareVersionChecker will create a object for software  version checker
func NewSoftwareVersionChecker(appStatusHandler core.AppStatusHandler) (*softwareVersionChecker, error) {
	if appStatusHandler == nil || appStatusHandler.IsInterfaceNil() {
		return nil, appStatusPolling.ErrNilAppStatusHandler
	}

	return &softwareVersionChecker{
		statusHandler: appStatusHandler,
	}, nil
}

// StartCheckSoftwareVersion will check on a specific interval if a new software version is available
func (svc *softwareVersionChecker) StartCheckSoftwareVersion() {
	go func() {
		svc.readLatestStableVersion()
		for {
			select {
			case <-time.After(checkInterval):
				svc.readLatestStableVersion()
			}
		}
	}()
}

func (svc *softwareVersionChecker) readLatestStableVersion() {
	tagVersion, err := readJSONFromUrl(stableTagLocation)
	if err == nil {
		svc.statusHandler.SetStringValue(core.MetricLatestTagSoftwareVersion, tagVersion)
	}
}

func readJSONFromUrl(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	respByte := buf.Bytes()

	var tag tagVersion
	if err := json.Unmarshal(respByte, &tag); err != nil {
		return "", err
	}

	return tag.TagVersion, nil
}
