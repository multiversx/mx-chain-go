package goroutines

import "strings"

type snapshot struct {
	routines map[string]*RoutineInfo
	filter   func(data string) bool
}

func newSnapshot(data string, filterFunc func(goRoutineData string) bool) (*snapshot, error) {
	s := &snapshot{
		routines: make(map[string]*RoutineInfo),
		filter:   filterFunc,
	}
	err := s.parse(data)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *snapshot) parse(data string) error {
	splitRoutines := strings.Split(data, "\n\n")
	for _, str := range splitRoutines {
		str = strings.Trim(str, " ")
		str = strings.Trim(str, "\n")
		if len(str) == 0 {
			continue
		}

		if !s.filter(str) {
			continue
		}

		ri, err := NewGoRoutine(str)
		if err != nil {
			return err
		}

		s.routines[ri.ID()] = ri
	}

	return nil
}

func (s *snapshot) diff(to *snapshot) []*RoutineInfo {
	diffSlice := make([]*RoutineInfo, 0)
	for routineID, ri := range s.routines {
		if to == nil {
			diffSlice = append(diffSlice, ri)
			continue
		}

		_, found := to.routines[routineID]
		if found {
			continue
		}

		diffSlice = append(diffSlice, ri)
	}

	if to == nil {
		return diffSlice
	}

	for routineID, ri := range to.routines {
		_, found := s.routines[routineID]
		if found {
			continue
		}

		diffSlice = append(diffSlice, ri)
	}

	return diffSlice
}
