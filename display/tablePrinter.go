package display

import (
	"strings"

	"github.com/pkg/errors"
)

// LineData represents a displayable table line
type LineData struct {
	Values      []string
	HRafterLine bool
}

// NewLineData creates a new LineData object
func NewLineData(hrAfterLine bool, values []string) *LineData {
	return &LineData{
		Values:      values,
		HRafterLine: hrAfterLine,
	}
}

// CreateTableString creates an ASCII table having header as table header and a LineData slice as table rows
// It automatically resize itself based on the lengths of every cell
func CreateTableString(header []string, data []*LineData) (string, error) {
	if data == nil || header == nil {
		return "", errors.New("nil data/header")
	}

	if len(data) == 0 && len(header) == 0 {
		return "", errors.New("empty data")
	}

	columnsWidths := computeColumnsWidths(header, data)

	builder := &strings.Builder{}
	drawHR(builder, columnsWidths)

	drawLine(builder, columnsWidths, header)
	drawHR(builder, columnsWidths)

	lastLineHadHR := false

	for i := 0; i < len(data); i++ {
		drawLine(builder, columnsWidths, data[i].Values)
		lastLineHadHR = data[i].HRafterLine

		if data[i].HRafterLine {
			drawHR(builder, columnsWidths)
		}
	}

	if !lastLineHadHR {
		drawHR(builder, columnsWidths)
	}

	return builder.String(), nil
}

func computeColumnsWidths(header []string, data []*LineData) []int {
	widths := make([]int, 0)

	for i := 0; i < len(header); i++ {
		if len(widths) <= i {
			widths = append(widths, len(header[i]))
			continue
		}

		if widths[i] < len(header[i]) {
			widths[i] = len(header[i])
		}
	}

	for i := 0; i < len(data); i++ {
		line := data[i]

		for j := 0; j < len(line.Values); j++ {
			if len(widths) <= j {
				widths = append(widths, len(line.Values[j]))
				continue
			}

			if widths[j] < len(line.Values[j]) {
				widths[j] = len(line.Values[j])
			}
		}
	}

	return widths
}

func drawHR(builder *strings.Builder, columnsWidths []int) {
	builder.WriteByte('+')
	for i := 0; i < len(columnsWidths); i++ {
		for j := 0; j < columnsWidths[i]+2; j++ {
			builder.WriteByte('-')
		}
		builder.WriteByte('+')
	}

	builder.Write([]byte{'\r', '\n'})
}

func drawLine(builder *strings.Builder, columnsWidths []int, strings []string) {
	builder.WriteByte('|')

	for i := 0; i < len(columnsWidths); i++ {
		builder.WriteByte(' ')

		lenStr := 0

		if i < len(strings) {
			lenStr = len(strings[i])
			builder.WriteString(strings[i])
		}

		for j := lenStr; j < columnsWidths[i]; j++ {
			builder.WriteByte(' ')
		}

		builder.Write([]byte{' ', '|'})
	}

	builder.Write([]byte{'\r', '\n'})
}
