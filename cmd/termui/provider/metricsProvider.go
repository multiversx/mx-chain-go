package provider

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("termui/provider")

const (
	statusMetricsUrlSuffix          = "/node/status"
	bootstrapStatusMetricsUrlSuffix = "/node/bootstrapstatus"
)

type statusMetricsResponseData struct {
	Response map[string]interface{} `json:"metrics"`
}

type responseFromApi struct {
	Data  statusMetricsResponseData `json:"data"`
	Error string                    `json:"error"`
	Code  string                    `json:"code"`
}

// StatusMetricsProvider is the struct that will handle initializing the presenter and fetching updated metrics from the node
type StatusMetricsProvider struct {
	presenter     PresenterHandler
	nodeAddress   string
	fetchInterval int
}

// NewStatusMetricsProvider will return a new instance of a StatusMetricsProvider
func NewStatusMetricsProvider(presenter PresenterHandler, nodeAddress string, fetchInterval int) (*StatusMetricsProvider, error) {
	if len(nodeAddress) == 0 {
		return nil, ErrInvalidAddressLength
	}
	if fetchInterval < 1 {
		return nil, ErrInvalidFetchInterval
	}
	if presenter == nil {
		return nil, ErrNilTermuiPresenter
	}

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
			smp.updateMetrics()
			time.Sleep(time.Duration(smp.fetchInterval) * time.Millisecond)
		}
	}()
}

func (smp *StatusMetricsProvider) updateMetrics() {
	smp.fetchAndApplyMetrics(statusMetricsUrlSuffix)
	smp.fetchAndApplyMetrics(bootstrapStatusMetricsUrlSuffix)
}

func (smp *StatusMetricsProvider) fetchAndApplyMetrics(metricsPath string) {
	metricsMap, err := smp.loadMetricsFromApi(metricsPath)
	if err != nil {
		log.Debug("fetch from API",
			"path", metricsPath,
			"error", err.Error())
	} else {
		smp.applyMetricsToPresenter(metricsMap)
	}
}

func (smp *StatusMetricsProvider) loadMetricsFromApi(metricsPath string) (map[string]interface{}, error) {
	client := http.Client{}

	statusMetricsUrl := smp.nodeAddress + metricsPath
	resp, err := client.Get(statusMetricsUrl)
	if err != nil {
		return nil, err
	}

	responseBytes, err := io.ReadAll(resp.Body)
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

	return metricsResponse.Data.Response, nil
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
		// json unmarshal treats all the numbers (in a field interface{}) as floats, so we need to cast it to uint64
		// because it is the numeric type used by the presenter
		smp.presenter.SetUInt64Value(key, uint64(v))
	case string:
		smp.presenter.SetStringValue(key, v)
	default:
		return ErrTypeAssertionFailed
	}

	return nil
}

func formatUrlAddress(address string) string {
	httpPrefix := "http://"
	if !strings.HasPrefix(address, "http") {
		address = httpPrefix + address
	}

	suffix := "/"
	if strings.HasSuffix(address, suffix) {
		address = address[:len(address)-1]
	}

	return address
}
