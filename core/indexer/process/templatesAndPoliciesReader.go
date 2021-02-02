package process

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
)

const (
	withKibanaFolder = "withKibana"
	noKibanaFolder   = "noKibana"
)

type templatesAndPoliciesReader struct {
	path            string
	useKibana       bool
	indexes         []string
	indexesPolicies []string
}

// NewTemplatesAndPoliciesReader will create a new instance of templatesAndPoliciesReader
func NewTemplatesAndPoliciesReader(pathToTemplates string, useKibana bool) *templatesAndPoliciesReader {
	indexes := []string{openDistroIndex, txIndex, blockIndex, miniblocksIndex, tpsIndex, ratingIndex, roundIndex, accountsESDTIndex,
		validatorsIndex, accountsIndex, accountsHistoryIndex, receiptsIndex, scResultsIndex, accountsESDTHistoryIndex}

	indexesPolicies := []string{txPolicy, blockPolicy, miniblocksPolicy, ratingPolicy, roundPolicy,
		validatorsPolicy, accountsHistoryPolicy, accountsESDTHistoryPolicy, receiptsPolicy, scResultsPolicy}

	var templatesPath string
	if useKibana {
		templatesPath = path.Join(pathToTemplates, withKibanaFolder)
	} else {
		templatesPath = path.Join(pathToTemplates, noKibanaFolder)
	}

	return &templatesAndPoliciesReader{
		path:            templatesPath,
		useKibana:       useKibana,
		indexes:         indexes,
		indexesPolicies: indexesPolicies,
	}
}

// GetElasticTemplatesAndPolicies will return elastic templates and policies
func (tpr *templatesAndPoliciesReader) GetElasticTemplatesAndPolicies() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
	indexTemplates := make(map[string]*bytes.Buffer)
	indexPolicies := make(map[string]*bytes.Buffer)
	var err error

	for _, index := range tpr.indexes {
		indexTemplates[index], err = tpr.getTemplateByIndex(index)
		if err != nil {
			return nil, nil, err
		}
	}

	if tpr.useKibana {
		for _, indexPolicy := range tpr.indexesPolicies {
			indexPolicies[indexPolicy], err = tpr.getPolicyByIndex(indexPolicy)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return indexTemplates, indexPolicies, nil
}

func (tpr *templatesAndPoliciesReader) getTemplateByIndex(index string) (*bytes.Buffer, error) {
	indexTemplate := &bytes.Buffer{}

	fileName := fmt.Sprintf("%s.json", index)
	filePath := filepath.Join(tpr.path, fileName)
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("getTemplateByIndex: %w, path %s, error %s", ErrReadTemplatesFile, filePath, err.Error())
	}

	indexTemplate.Grow(len(fileBytes))
	_, err = indexTemplate.Write(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("getTemplateByIndex: %w, path %s, error %s", ErrWriteToBuffer, filePath, err.Error())
	}

	return indexTemplate, nil
}

func (tpr *templatesAndPoliciesReader) getPolicyByIndex(index string) (*bytes.Buffer, error) {
	indexPolicy := &bytes.Buffer{}

	fileName := fmt.Sprintf("%s.json", index)
	filePath := filepath.Join(tpr.path, fileName)
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("getPolicyByIndex: %w, path %s, error %s", ErrReadPolicyFile, filePath, err.Error())
	}

	indexPolicy.Grow(len(fileBytes))
	_, err = indexPolicy.Write(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("getPolicyByIndex: %w, path %s, error %s", ErrWriteToBuffer, filePath, err.Error())
	}

	return indexPolicy, nil
}
