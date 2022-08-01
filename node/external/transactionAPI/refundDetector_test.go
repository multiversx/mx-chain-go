package transactionAPI

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRefundDetector_IsRefundShouldDetectRefund(t *testing.T) {
	detector := newRefundDetector()

	require.True(t, detector.isRefund(refundDetectorInput{
		Value:         "1000",
		Data:          []byte("@ok@test"),
		GasLimit:      0,
		ReturnMessage: "",
	}))

	require.True(t, detector.isRefund(refundDetectorInput{
		Value:         "1000",
		Data:          []byte("@ok@test"),
		GasLimit:      0,
		ReturnMessage: "",
	}))

	require.True(t, detector.isRefund(refundDetectorInput{
		Value:         "1000",
		Data:          []byte("@6f6b@test"),
		GasLimit:      0,
		ReturnMessage: "",
	}))

	require.True(t, detector.isRefund(refundDetectorInput{
		Value:         "1000",
		Data:          []byte("foobar"),
		GasLimit:      0,
		ReturnMessage: "gas refund for relayer",
	}))

	require.False(t, detector.isRefund(refundDetectorInput{
		Value: "0",
	}))

	require.False(t, detector.isRefund(refundDetectorInput{
		Value:    "1000",
		Data:     []byte("foobar"),
		GasLimit: 0,
	}))

	require.True(t, detector.isRefund(refundDetectorInput{
		Value:    "1000",
		Data:     []byte("@ok@test"),
		GasLimit: 1,
	}))

	require.True(t, detector.isRefund(refundDetectorInput{
		Value:    "1000",
		Data:     []byte("@6f6b@test"),
		GasLimit: 1,
	}))
}
