package mock

import (
	"bytes"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// DatabaseWriterStub --
type DatabaseWriterStub struct {
	DoRequestCalled     func(req *esapi.IndexRequest) error
	DoBulkRequestCalled func(buff *bytes.Buffer, index string) error
	DoBulkRemoveCalled  func(index string, hashes []string) error
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

// DoBulkRemove -
func (dwm *DatabaseWriterStub) DoBulkRemove(index string, hashes []string) error {
	if dwm.DoBulkRemoveCalled != nil {
		return dwm.DoBulkRemoveCalled(index, hashes)
	}

	return nil
}

// CheckAndCreateIndex --
func (dwm *DatabaseWriterStub) CheckAndCreateIndex(_ string) error {
	return nil
}

// CheckAndCreateAlias -
func (dwm *DatabaseWriterStub) CheckAndCreateAlias(_ string, _ string) error {
	return nil
}

// CheckAndCreateTemplate -
func (dwm *DatabaseWriterStub) CheckAndCreateTemplate(_ string, _ *bytes.Buffer) error {
	return nil
}

// CheckAndCreatePolicy -
func (dwm *DatabaseWriterStub) CheckAndCreatePolicy(_ string, _ *bytes.Buffer) error {
	return nil
}
