package softwareVersion

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/logger"
)

const checkInterval = time.Second
const stableTagLocation = "https://api.github.com/repos/ElrondNetwork/elrond-go/releases/latest"

type tagVersion struct {
	TagVersion string `json:"tag_name"`
}

type SoftwareVersionChecker struct {
	statusHandler core.AppStatusHandler
}

var log = logger.DefaultLogger()

// NewSoftwareVersionChecker will create an object for software  version checker
func NewSoftwareVersionChecker(appStatusHandler core.AppStatusHandler) (*SoftwareVersionChecker, error) {
	if appStatusHandler == nil || appStatusHandler.IsInterfaceNil() {
		return nil, core.ErrNilAppStatusHandler
	}

	return &SoftwareVersionChecker{
		statusHandler: appStatusHandler,
	}, nil
}

// StartCheckSoftwareVersion will check on a specific interval if a new software version is available
func (svc *SoftwareVersionChecker) StartCheckSoftwareVersion() {
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

func (svc *SoftwareVersionChecker) readLatestStableVersion() {
	tagVersion, err := readJSONFromUrl(stableTagLocation)
	if err != nil {
		log.Error("cannot read json with latest stable tag", err)
		return
	}

	svc.statusHandler.SetStringValue(core.MetricLatestTagSoftwareVersion, tagVersion)
}

func readJSONFromUrl(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer func() {
		err := resp.Body.Close()
		log.LogIfError(err)
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
