package mock

import (
	"bytes"
	"io"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// DatabaseWriterStub --
type DatabaseWriterStub struct {
	DoRequestCalled     func(req *esapi.IndexRequest) error
	DoBulkRequestCalled func(buff *bytes.Buffer, index string) error
}

// DoRequest --
func (dwm *DatabaseWriterStub) DoRequest(req *esapi.IndexRequest) error {
	if dwm.DoRequestCalled != nil {
		return dwm.DoRequestCalled(req)
	}
	return nil
}

// DoBulkRequest --
func (dwm *DatabaseWriterStub) DoBulkRequest(buff *bytes.Buffer, index string) error {
	if dwm.DoBulkRequestCalled != nil {
		return dwm.DoBulkRequestCalled(buff, index)
	}
	return nil
}

// DoMultiGet --
func (dwm *DatabaseWriterStub) DoMultiGet(_ map[string]interface{}, _ string) (map[string]interface{}, error) {
	return nil, nil
}

// CheckAndCreateIndex --
func (dwm *DatabaseWriterStub) CheckAndCreateIndex(_ string, _ io.Reader) error {
	return nil
}
