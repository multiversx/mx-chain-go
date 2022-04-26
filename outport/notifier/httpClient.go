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

type HttpClient interface {
	Post(route string, payload interface{}, response interface{}) error
}

type httpClient struct {
	useAuthorization bool
	username         string
	password         string
	baseUrl          string
}

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
		h.setAuthorization(req)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		bodyCloseErr := resp.Body.Close()
		if bodyCloseErr != nil {
			log.Warn("error while trying to close response body", "err", bodyCloseErr.Error())
		}
	}()

	resBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(resBody, &response)
}

func (h *httpClient) getErrorFromStatusCode(statusCode int) error {
	if statusCode == http.StatusBadRequest {
		return ErrHttpFailedRequest(badRequestMessage, statusCode)
	}
	if statusCode == http.StatusUnauthorized {
		return ErrHttpFailedRequest(unauthorizedMessage, statusCode)
	}
	if statusCode == http.StatusInternalServerError {
		return ErrHttpFailedRequest(internalErrMessage, statusCode)
	}
	if statusCode != http.StatusOK {
		return ErrHttpFailedRequest(genericHttpErrMessage, statusCode)
	}

	return nil
}

func (h *httpClient) setAuthorization(req *http.Request) {
	req.SetBasicAuth(h.username, h.password)
}
