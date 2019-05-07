package heartbeat

import (
	"encoding/json"
	"errors"
	"time"
)

// Duration is a wrapper of the original Duration struct
// that has JSON marshal and unmarshal capabilities
// golang issue: https://github.com/golang/go/issues/10275
type Duration struct {
	time.Duration
}

// MarshalJSON is called when a json marshal is triggered on this field
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON is called when a json unmarshal is triggered on this field
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}
