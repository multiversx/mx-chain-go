package groups

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestParseBoolUrlParam(t *testing.T) {
	c := &gin.Context{
		Request: &http.Request{
			URL: &url.URL{
				RawQuery: "a=true&b=false&c=foobar&d",
			},
		},
	}

	value, err := parseBoolUrlParam(c, "a")
	require.Nil(t, err)
	require.True(t, value)

	value, err = parseBoolUrlParam(c, "b")
	require.Nil(t, err)
	require.False(t, value)

	value, err = parseBoolUrlParam(c, "c")
	require.NotNil(t, err)
	require.False(t, value)

	value, err = parseBoolUrlParam(c, "d")
	require.Nil(t, err)
	require.False(t, value)

	value, err = parseBoolUrlParam(c, "e")
	require.Nil(t, err)
	require.False(t, value)
}

func TestParseUintUrlParam(t *testing.T) {
	c := &gin.Context{
		Request: &http.Request{
			URL: &url.URL{
				RawQuery: "a=7&b=0&c=foobar&d=-1&e=1234567898765432112345678987654321",
			},
		},
	}

	value, err := parseUintUrlParam(c, "a")
	require.Nil(t, err)
	require.Equal(t, uint64(7), value)

	value, err = parseUintUrlParam(c, "b")
	require.Nil(t, err)
	require.Equal(t, uint64(0), value)

	value, err = parseUintUrlParam(c, "c")
	require.NotNil(t, err)
	require.Equal(t, uint64(0), value)

	value, err = parseUintUrlParam(c, "d")
	require.NotNil(t, err)
	require.Equal(t, uint64(0), value)

	value, err = parseUintUrlParam(c, "e")
	require.NotNil(t, err)
	require.Equal(t, uint64(0xffffffffffffffff), value)
}
