package utils

import (
	"math/big"
	"strconv"
	"time"
)


// CutPrefix cut '0x' prefix
func CutPrefix(data string) string {
	if len(data) > 2 && data[:2] == "0x"{
		return data[2:]
	}
	return data
}

func AddPrefix(data string) string {
	return "0x" + data
}

func HexDecimalToDecimal(data string) (int64, error) {
	return strconv.ParseInt(CutPrefix(data), 16, 64)
}

// HexDecimalToFloat64 lose some accuracy
func HexDecimalToFloat64(data string) float64 {
	n := new(big.Int)
	n.SetString(CutPrefix(data), 16)
	f := new(big.Float).SetInt(n)
	f64, _ := f.Float64()
	return f64
}

func WeiToEth (wei float64) float64 {
	return wei / 1_000_000_000_000_000_000
}

func HexTimeToTime(hexTime string) (timeUTC time.Time, err error) {
	res, err := HexDecimalToDecimal(hexTime)
	if err != nil {
		return
	}

	return time.Unix(res, 0), nil
}

