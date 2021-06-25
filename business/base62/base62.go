package base62

import (
	"math"
	"strings"

	"github.com/pkg/errors"
)

const (
	base   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length = int(len(base))
)

func Encode(number int) string {
	var eb strings.Builder
	eb.Grow(11)

	for ; number > 0; number = number / length {
		eb.WriteByte(base[(number % length)])
	}

	return eb.String()
}

func Decode(encoded string) (int, error) {
	var number int

	for i, symbol := range encoded {
		bp := strings.IndexRune(base, symbol)

		if bp == -1 {
			return int(bp), errors.New("invalid character: " + string(symbol))
		}
		number += int(bp) * int(math.Pow(float64(length), float64(i)))
	}

	return number, nil
}
