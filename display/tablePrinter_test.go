package display

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"testing"
)

func TestCreateTableString_Empty(t *testing.T) {
	str, _ := CreateTableString(
		make([]string, 1), []*LineData{NewLineData(false, make([]string, 0))})

	fmt.Println(str)
}

func TestCreateTableString_OneColumnOneRow(t *testing.T) {
	str, _ := CreateTableString(
		[]string{"h"}, []*LineData{NewLineData(false, []string{"v"})})

	fmt.Println(str)
}

func TestCreateTableString_2ColumnsBigHeader(t *testing.T) {
	str, _ := CreateTableString(
		[]string{"header1233423322343", "sfkjdbsfjksdbjbfskj"}, []*LineData{NewLineData(false, []string{"v"})})

	fmt.Println(str)
}

func TestCreateTableString_2ColumnsBigLine(t *testing.T) {
	str, _ := CreateTableString(
		[]string{"header1", "header2"},
		[]*LineData{NewLineData(false, []string{"aaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbb"})},
	)

	fmt.Println(str)
}

func TestCreateTableString_2RowsLastWithHRBigLine(t *testing.T) {
	str, _ := CreateTableString(
		[]string{"header1", "header2"},
		[]*LineData{
			NewLineData(false, []string{"aaa", "bbb"}),
			NewLineData(true, []string{"ccc", "ddd"}),
		},
	)

	fmt.Println(str)
}

func TestCreateTableString_RandomData(t *testing.T) {
	rdm := rand.New(rand.NewSource(1000))

	header, dataLines := genBigData(rdm)

	str, _ := CreateTableString(header, dataLines)

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
			dataLine.HRafterLine = true
		}
		dataLines = append(dataLines, dataLine)
	}

	return header, dataLines
}

func BenchmarkCreateTableString(b *testing.B) {
	rdm := rand.New(rand.NewSource(1000))
	hdr, lines := genBigData(rdm)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = CreateTableString(hdr, lines)
	}
}
