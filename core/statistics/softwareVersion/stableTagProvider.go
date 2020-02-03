package softwareVersion

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type stableTagProvider struct {
	stableTagLocation string
}

// NewStableTagProvider returns a new instance of stableTagProvider
func NewStableTagProvider(stableTagLocation string) *stableTagProvider {
	return &stableTagProvider{stableTagLocation: stableTagLocation}
}

// FetchTagVersion will call the provided URL and will fetch the software version
func (stp *stableTagProvider) FetchTagVersion() (string, error) {
	resp, err := http.Get(stp.stableTagLocation)
	if err != nil {
		return "", err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Debug(err.Error())
		}
	}()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}

	respBytes := buf.Bytes()

	var tag tagVersion
	if err = json.Unmarshal(respBytes, &tag); err != nil {
		return "", err
	}

	return tag.TagVersion, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (stp *stableTagProvider) IsInterfaceNil() bool {
	return stp == nil
}
