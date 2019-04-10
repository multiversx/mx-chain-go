package display

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTableString_NilHeaderShouldErr(t *testing.T) {
	str, err := CreateTableString(nil, []*LineData{NewLineData(false, make([]string, 0))})

	assert.Equal(t, "", str)
	assert.Equal(t, ErrNilHeader, err)
}

func TestCreateTableString_NilDataLinesShouldErr(t *testing.T) {
	str, err := CreateTableString(make([]string, 1), nil)

	assert.Equal(t, "", str)
	assert.Equal(t, ErrNilDataLines, err)
}

func TestCreateTableString_EmptySlicesShouldErr(t *testing.T) {
	str, err := CreateTableString(make([]string, 0), make([]*LineData, 0))

	assert.Equal(t, "", str)
	assert.Equal(t, ErrEmptySlices, err)
}

func TestCreateTableString_NilLineDataValueShouldErr(t *testing.T) {
	str, err := CreateTableString(
		make([]string, 0),
		[]*LineData{nil})

	assert.Equal(t, "", str)
	assert.Equal(t, ErrNilLineDataInSlice, err)
}

func TestCreateTableString_LineDataWithNilValuesShouldErr(t *testing.T) {
	str, err := CreateTableString(
		make([]string, 0),
		[]*LineData{NewLineData(true, nil)})

	assert.Equal(t, "", str)
	assert.Equal(t, ErrNilValuesOfLineDataInSlice, err)
}

func TestCreateTableString_EmptySlices2ShouldErr(t *testing.T) {
	str, err := CreateTableString(
		make([]string, 0),
		[]*LineData{NewLineData(true, make([]string, 0))})

	assert.Equal(t, "", str)
	assert.Equal(t, ErrEmptySlices, err)
}

func TestCreateTableString_Empty(t *testing.T) {
	str, err := CreateTableString(
		make([]string, 1), []*LineData{NewLineData(false, make([]string, 0))})

	assert.Nil(t, err)
	fmt.Println(str)
}

func TestCreateTableString_OneColumnOneRow(t *testing.T) {
	str, err := CreateTableString(
		[]string{"h"}, []*LineData{NewLineData(false, []string{"v"})})

	assert.Nil(t, err)
	fmt.Println(str)
}

func TestCreateTableString_2ColumnsBigHeader(t *testing.T) {
	str, err := CreateTableString(
		[]string{"header1233423322343", "sfkjdbsfjksdbjbfskj"}, []*LineData{NewLineData(false, []string{"v"})})

	assert.Nil(t, err)
	fmt.Println(str)
}

func TestCreateTableString_2ColumnsBigLine(t *testing.T) {
	str, err := CreateTableString(
		[]string{"header1", "header2"},
		[]*LineData{NewLineData(false, []string{"aaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbb"})},
	)

	assert.Nil(t, err)
	fmt.Println(str)
}

func TestCreateTableString_2RowsLastWithHRBigLine(t *testing.T) {
	str, err := CreateTableString(
		[]string{"header1", "header2"},
		[]*LineData{
			NewLineData(false, []string{"aaa", "bbb"}),
			NewLineData(true, []string{"ccc", "ddd"}),
		},
	)

	assert.Nil(t, err)
	fmt.Println(str)
}

func TestCreateTableString_RandomData(t *testing.T) {
	rdm := rand.New(rand.NewSource(1000))

	header, dataLines := genBigData(rdm)

	str, err := CreateTableString(header, dataLines)

	assert.Nil(t, err)
	fmt.Println(str)
}

func TestCreateTableString_MoreColumnsInRows(t *testing.T) {
	str, err := CreateTableString(
		[]string{"header1"},
		[]*LineData{
			NewLineData(false, []string{"aaa", "bbb"}),
			NewLineData(true, []string{"ccc", "ddd"}),
		},
	)

	assert.Nil(t, err)
	fmt.Println(str)
}

func generateRandomStrings(howMany int, rdm *rand.Rand, minChars, maxChars int) []string {
	strs := make([]string, howMany)
	for i := 0; i < 12; i++ {
		//between minChars and maxChars
		chars := float32(minChars) + rdm.Float32()*float32(maxChars-minChars)

		buff := make([]byte, int(chars))
		rdm.Read(buff)

		strs[i] = base64.StdEncoding.EncodeToString(buff)
	}

	return strs
}

func genBigData(rdm *rand.Rand) ([]string, []*LineData) {
	header := generateRandomStrings(12, rdm, 5, 12)

	dataLines := make([]*LineData, 0)
	for i := 0; i < 500; i++ {
		dataLine := NewLineData(false, generateRandomStrings(12, rdm, 5, 12))
		if i%15 == 0 {
			dataLine.HorizontalRuleAfter = true
		}
		dataLines = append(dataLines, dataLine)
	}

	return header, dataLines
}

func TestCreateTableString_SpecialCharacters(t *testing.T) {
	header := []string{"column1", "column2"}
	data := make([]*LineData, 0)
	data = append(data, NewLineData(false, []string{"value1", "65.047µs"}))
	data = append(data, NewLineData(false, []string{"value1µµµµµµµµµµµµ", "65µs"}))

	str, err := CreateTableString(header, data)

	assert.Nil(t, err)
	fmt.Println(str)
}

func BenchmarkCreateTableString(b *testing.B) {
	rdm := rand.New(rand.NewSource(1000))
	hdr, lines := genBigData(rdm)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = CreateTableString(hdr, lines)
	}
}
