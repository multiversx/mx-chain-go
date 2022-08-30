package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	contentTypeKey   = "Content-Type"
	contentTypeValue = "application/json"
)

type httpClientHandler interface {
	Post(route string, payload interface{}, response interface{}) error
}

type httpClient struct {
	useAuthorization bool
	username         string
	password         string
	baseUrl          string
}

// HttpClientArgs defines the arguments needed for http client creation
type HttpClientArgs struct {
	UseAuthorization bool
	Username         string
	Password         string
	BaseUrl          string
}

// NewHttpClient creates an instance of httpClient which is a wrapper for http.Client
func NewHttpClient(args HttpClientArgs) *httpClient {
	return &httpClient{
		useAuthorization: args.UseAuthorization,
		username:         args.Username,
		password:         args.Password,
		baseUrl:          args.BaseUrl,
	}
}

// Post can be used to send POST requests. It handles marshalling to/from json
func (h *httpClient) Post(
	route string,
	payload interface{},
	response interface{},
) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{}
	url := fmt.Sprintf("%s%s", h.baseUrl, route)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set(contentTypeKey, contentTypeValue)

	if h.useAuthorization {
		req.SetBasicAuth(h.username, h.password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			bodyCloseErr := resp.Body.Close()
			if bodyCloseErr != nil {
				log.Warn("error while trying to close response body", "err", bodyCloseErr.Error())
			}
		}
	}()

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Warn("httpClient: received HTTP status", "code", resp.StatusCode, "responseBody", string(resBody))
		return fmt.Errorf("HTTP status code: %d, %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return json.Unmarshal(resBody, &response)
}
