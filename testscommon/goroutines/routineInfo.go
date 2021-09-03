package goroutines

import (
	"errors"
	"strings"
)

const goroutineMarker = "goroutine"

var errInvalidLine = errors.New("invalid go routine info string: not enough words for the `goroutine n [chan xxx]:` marker line")
var errMissingGoRoutineMarker = errors.New("invalid go routine info string: missing `goroutine` marker")

// RoutineInfo holds the relevant information about a go routine
type RoutineInfo struct {
	id   string
	data string
}

// NewGoRoutine parses the input data and returns a new instance. Errors if the input data is in a wrong format
func NewGoRoutine(data string) (*RoutineInfo, error) {
	ri := &RoutineInfo{}
	err := ri.parse(data)
	if err != nil {
		return nil, err
	}

	return ri, nil
}

// String returns the complete string of the go routine
func (ri *RoutineInfo) String() string {
	return ri.data
}

// ID returns the ID of the go routine
func (ri *RoutineInfo) ID() string {
	return ri.id
}

func (ri *RoutineInfo) parse(data string) error {
	ri.data = data
	var err error
	ri.id, err = ri.extractID(data)

	return err
}

func (ri *RoutineInfo) extractID(data string) (string, error) {
	data = strings.Trim(data, "\n")
	data = strings.Trim(data, " ")
	splitLines := strings.Split(data, "\n")
	words := strings.Split(splitLines[0], " ")
	if len(words) < 2 {
		return "", errInvalidLine
	}
	if words[0] != goroutineMarker {
		return "", errMissingGoRoutineMarker
	}

	return words[1], nil
}
