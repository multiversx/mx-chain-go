package connectionStringValidator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectionStringValidator_IsValid(t *testing.T) {
	t.Parallel()

	csv := NewConnectionStringValidator()
	assert.False(t, csv.IsValid("invalid string"))
	assert.False(t, csv.IsValid(""))

	assert.True(t, csv.IsValid("5.22.219.242"))
	assert.True(t, csv.IsValid("2031:0:130F:0:0:9C0:876A:130B"))
	assert.True(t, csv.IsValid("16Uiu2HAm6yvbp1oZ6zjnWsn9FdRqBSaQkbhELyaThuq48ybdojvJ"))
}
func TestConnectionStringValidator_isValidIP(t *testing.T) {
	t.Parallel()

	csv := NewConnectionStringValidator()
	assert.False(t, csv.isValidIP("invalid ip"))
	assert.False(t, csv.isValidIP(""))
	assert.False(t, csv.isValidIP("a.b.c.d"))
	assert.False(t, csv.isValidIP("10.0.0"))
	assert.False(t, csv.isValidIP("10.0"))
	assert.False(t, csv.isValidIP("10"))
	assert.False(t, csv.isValidIP("2031:0:130F:0:0:9C0:876A"))
	assert.False(t, csv.isValidIP("2031:0:130F:0:0:9C0"))
	assert.False(t, csv.isValidIP("2031:0:130F:0:0"))
	assert.False(t, csv.isValidIP("2031:0:130F:0"))
	assert.False(t, csv.isValidIP("2031:0:130F"))
	assert.False(t, csv.isValidIP("2031:0"))
	assert.False(t, csv.isValidIP("16Uiu2HAm6yvbp1oZ6zjnWsn9FdRqBSaQkbhELyaThuq48ybdojvJ"))

	assert.True(t, csv.isValidIP("127.0.0.1"))
	assert.True(t, csv.isValidIP("5.22.219.242"))
	assert.True(t, csv.isValidIP("2031:0:130F:0:0:9C0:876A:130B"))
}

func TestConnectionStringValidator_isValidPeerID(t *testing.T) {
	t.Parallel()

	csv := NewConnectionStringValidator()
	assert.False(t, csv.isValidPeerID("invalid peer id"))
	assert.False(t, csv.isValidPeerID(""))
	assert.False(t, csv.isValidPeerID("blaiu2HAm6yvbp1oZ6zjnWsn9FdRqBSaQkbhELyaThuq48ybdojvJ")) // first 3 chars altered
	assert.False(t, csv.isValidPeerID("16Uiu2HAm6yvbp1oZ6zjnWsn9FdRqBSaQkbhELyaThuq48ybdobla")) // last 3 chars altered
	assert.False(t, csv.isValidPeerID("16Uiu2HAm6yvbp1oZ6zjnWsn9FblaBSaQkbhELyaThuq48ybdojvJ")) // middle chars altered
	assert.False(t, csv.isValidPeerID("5.22.219.242"))

	assert.True(t, csv.isValidPeerID("16Uiu2HAm6yvbp1oZ6zjnWsn9FdRqBSaQkbhELyaThuq48ybdojvJ"))
}
