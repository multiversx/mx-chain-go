package display

import (
	"strings"
	"unicode/utf8"
)

// LineData represents a displayable table line
type LineData struct {
	Values              []string
	HorizontalRuleAfter bool
}

// NewLineData creates a new LineData object
func NewLineData(horizontalRuleAfter bool, values []string) *LineData {
	return &LineData{
		Values:              values,
		HorizontalRuleAfter: horizontalRuleAfter,
	}
}

// CreateTableString creates an ASCII table having header as table header and a LineData slice as table rows
// It automatically resize itself based on the lengths of every cell
func CreateTableString(header []string, data []*LineData) (string, error) {
	err := checkValidity(header, data)

	if err != nil {
		return "", err
	}

	columnsWidths := computeColumnsWidths(header, data)

	builder := &strings.Builder{}
	drawHorizontalRule(builder, columnsWidths)

	drawLine(builder, columnsWidths, header)
	drawHorizontalRule(builder, columnsWidths)

	lastLineHadHR := false

	for i := 0; i < len(data); i++ {
		drawLine(builder, columnsWidths, data[i].Values)
		lastLineHadHR = data[i].HorizontalRuleAfter

		if data[i].HorizontalRuleAfter {
			drawHorizontalRule(builder, columnsWidths)
		}
	}

	if !lastLineHadHR {
		drawHorizontalRule(builder, columnsWidths)
	}

	return builder.String(), nil
}

func checkValidity(header []string, data []*LineData) error {
	if header == nil {
		return ErrNilHeader
	}

	if data == nil {
		return ErrNilDataLines
	}

	if len(data) == 0 && len(header) == 0 {
		return ErrEmptySlices
	}

	maxElemFound := len(header)

	for _, ld := range data {
		if ld == nil {
			return ErrNilLineDataInSlice
		}

		if ld.Values == nil {
			return ErrNilValuesOfLineDataInSlice
		}

		if maxElemFound < len(ld.Values) {
			maxElemFound = len(ld.Values)
		}
	}

	if maxElemFound == 0 {
		return ErrEmptySlices
	}

	return nil
}

func computeColumnsWidths(header []string, data []*LineData) []int {
	widths := make([]int, len(header))

	for i := 0; i < len(header); i++ {
		widths[i] = len(header[i])
	}

	for i := 0; i < len(data); i++ {
		line := data[i]

		for j := 0; j < len(line.Values); j++ {
			if len(widths) <= j {
				widths = append(widths, utf8.RuneCountInString(line.Values[j]))
				continue
			}

			width := utf8.RuneCountInString(line.Values[j])
			if widths[j] < width {
				widths[j] = width
			}
		}
	}

	return widths
}

func drawHorizontalRule(builder *strings.Builder, columnsWidths []int) {
	_ = builder.WriteByte('+')
	for i := 0; i < len(columnsWidths); i++ {
		for j := 0; j < columnsWidths[i]+2; j++ {
			_ = builder.WriteByte('-')
		}
		_ = builder.WriteByte('+')
	}

	_, _ = builder.Write([]byte{'\r', '\n'})
}

func drawLine(builder *strings.Builder, columnsWidths []int, strings []string) {
	_ = builder.WriteByte('|')

	for i := 0; i < len(columnsWidths); i++ {
		_ = builder.WriteByte(' ')

		lenStr := 0

		if i < len(strings) {
			lenStr = utf8.RuneCountInString(strings[i])
			_, _ = builder.WriteString(strings[i])
		}

		for j := lenStr; j < columnsWidths[i]; j++ {
			_ = builder.WriteByte(' ')
		}

		_, _ = builder.Write([]byte{' ', '|'})
	}

	_, _ = builder.Write([]byte{'\r', '\n'})
}
