package provider

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go/logger"
	presenter2 "github.com/ElrondNetwork/elrond-go/statusHandler/presenter"
)

var log = logger.GetOrCreate("cmd/termui/provider")

const statusMetricsUrlSuffix = "/node/status"

type responseFromApi struct {
	Response map[string]interface{} `json:"details"`
}

// StatusMetricsProvider is the struct that will handle initializing the presenter and fetching updated metrics from the node
type StatusMetricsProvider struct {
	presenter     PresenterHandler
	nodeAddress   string
	fetchInterval int
}

// NewStatusMetricsProvider will return a new instance of a StatusMetricsProvider
func NewStatusMetricsProvider(nodeAddress string, fetchInterval int) (*StatusMetricsProvider, error) {
	if len(nodeAddress) == 0 {
		return nil, ErrInvalidAddressLength
	}
	if fetchInterval < 1 {
		return nil, ErrInvalidFetchInterval
	}
	presenter := presenter2.NewPresenterStatusHandler()
	return &StatusMetricsProvider{
		presenter:     presenter,
		nodeAddress:   formatUrlAddress(nodeAddress),
		fetchInterval: fetchInterval,
	}, nil
}

// StartUpdatingData will update data from the API at a given interval
func (smp *StatusMetricsProvider) StartUpdatingData() {
	go func() {
		for {
			metricsMap, err := smp.loadMetricsFromApi()
			if err != nil {
				log.Debug("fetch from API",
					"error", err.Error())
			} else {
				smp.applyMetricsToPresenter(metricsMap)
			}

			time.Sleep(time.Duration(smp.fetchInterval) * time.Second)
		}
	}()
}

func (smp *StatusMetricsProvider) loadMetricsFromApi() (map[string]interface{}, error) {
	client := http.Client{}

	statusMetricsUrl := smp.nodeAddress + statusMetricsUrlSuffix
	resp, err := client.Get(statusMetricsUrl)
	if err != nil {
		return nil, err
	}

	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Error("close response body", "error", err.Error())
		}
	}()

	var metricsResponse responseFromApi
	err = json.Unmarshal(responseBytes, &metricsResponse)
	if err != nil {
		return nil, err
	}

	return metricsResponse.Response, nil
}

func (smp *StatusMetricsProvider) applyMetricsToPresenter(metricsMap map[string]interface{}) {
	var err error
	for key, value := range metricsMap {
		err = smp.setPresenterValue(key, value)
		if err != nil {
			log.Debug("termui metric set",
				"error", err.Error())
		}
	}
}

func (smp *StatusMetricsProvider) setPresenterValue(key string, value interface{}) error {
	switch v := value.(type) {
	case float64:
		// json unmarshal treats all the numbers (in a field interface{}) as floats so we need to cast it to uint64
		// because it is the numeric type used by the presenter
		smp.presenter.SetUInt64Value(key, uint64(v))
	case string:
		smp.presenter.SetStringValue(key, v)
	default:
		return ErrTypeAssertionFailed
	}

	return nil
}

// Presenter will return the presenter so it can be used when creating the termui's instance
func (smp *StatusMetricsProvider) Presenter() PresenterHandler {
	return smp.presenter
}

func formatUrlAddress(address string) string {
	httpPrefix := "http://"
	if !strings.HasPrefix(address, httpPrefix) {
		address = httpPrefix + address
	}

	suffix := "/"
	if strings.HasSuffix(address, suffix) {
		address = address[:len(address)-1]
	}

	return address
}
