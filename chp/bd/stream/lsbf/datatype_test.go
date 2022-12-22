package lsbf

import (
	"testing"
	"math/big"

	"github.com/stretchr/testify/assert"
)

func TestLsbfInt64(t *testing.T) {
	v := FromInt64(255, 16)
	assert.Equal(t, 3, len(v))
	assert.Equal(t, int64(15), v[0])
	assert.Equal(t, int64(15), v[1])
	assert.Equal(t, int64(0), v[2])
	assert.Equal(t, int64(255), ToInt64(v, 16))

	v = FromInt64(-128, 16)
	assert.Equal(t, 3, len(v))
	assert.Equal(t, int64(0), v[0])
	assert.Equal(t, int64(8), v[1])
	assert.Equal(t, int64(15), v[2])
	assert.Equal(t, int64(-128), ToInt64(v, 16))

	v = FromInt64(-1, 16)
	assert.Equal(t, 1, len(v))
	assert.Equal(t, int64(15), v[0])
	assert.Equal(t, int64(-1), ToInt64(v, 16))

	v = FromInt64(0, 16)
	assert.Equal(t, 1, len(v))
	assert.Equal(t, int64(0), v[0])
	assert.Equal(t, int64(0), ToInt64(v, 16))

	v = FromInt64(-126, 16)
	assert.Equal(t, 3, len(v))
	assert.Equal(t, int64(2), v[0])
	assert.Equal(t, int64(8), v[1])
	assert.Equal(t, int64(15), v[2])
	assert.Equal(t, int64(-126), ToInt64(v, 16))

	v = FromInt64(1100, 16)
	assert.Equal(t, 4, len(v))
	assert.Equal(t, int64(12), v[0])
	assert.Equal(t, int64(4), v[1])
	assert.Equal(t, int64(4), v[2])
	assert.Equal(t, int64(0), v[3])
	assert.Equal(t, int64(1100), ToInt64(v, 16))

	v = FromInt64(-1100, 16)
	assert.Equal(t, 4, len(v))
	assert.Equal(t, int64(4), v[0])
	assert.Equal(t, int64(11), v[1])
	assert.Equal(t, int64(11), v[2])
	assert.Equal(t, int64(15), v[3])
	assert.Equal(t, int64(-1100), ToInt64(v, 16))

	v = FromInt64(-3487609632, 16)
	assert.Equal(t, 9, len(v))
	assert.Equal(t, int64(0), v[0])
	assert.Equal(t, int64(14), v[1])
	assert.Equal(t, int64(12), v[2])
	assert.Equal(t, int64(4), v[3])
	assert.Equal(t, int64(15), v[4])
	assert.Equal(t, int64(1), v[5])
	assert.Equal(t, int64(0), v[6])
	assert.Equal(t, int64(3), v[7])
	assert.Equal(t, int64(15), v[8])
	assert.Equal(t, int64(-3487609632), ToInt64(v, 16))

	v = FromInt64(20148091803270415, 16)
	assert.Equal(t, 15, len(v))
	assert.Equal(t, int64(15), v[0])
	assert.Equal(t, int64(0), v[1])
	assert.Equal(t, int64(13), v[2])
	assert.Equal(t, int64(15), v[3])
	assert.Equal(t, int64(5), v[4])
	assert.Equal(t, int64(13), v[5])
	assert.Equal(t, int64(14), v[6])
	assert.Equal(t, int64(2), v[7])
	assert.Equal(t, int64(5), v[8])
	assert.Equal(t, int64(9), v[9])
	assert.Equal(t, int64(4), v[10])
	assert.Equal(t, int64(9), v[11])
	assert.Equal(t, int64(7), v[12])
	assert.Equal(t, int64(4), v[13])
	assert.Equal(t, int64(0), v[14])
	assert.Equal(t, int64(20148091803270415), ToInt64(v, 16))
}

func TestLsbfBigInt(t *testing.T) {
	v := FromBigInt(big.NewInt(255), 16)
	assert.Equal(t, 3, len(v))
	assert.Equal(t, int64(15), v[0])
	assert.Equal(t, int64(15), v[1])
	assert.Equal(t, int64(0), v[2])
	assert.Equal(t, int64(255), ToBigInt(v, 16).Int64())

	v = FromBigInt(big.NewInt(-128), 16)
	assert.Equal(t, 3, len(v))
	assert.Equal(t, int64(0), v[0])
	assert.Equal(t, int64(8), v[1])
	assert.Equal(t, int64(15), v[2])
	assert.Equal(t, int64(-128), ToBigInt(v, 16).Int64())

	v = FromBigInt(big.NewInt(-1), 16)
	assert.Equal(t, 1, len(v))
	assert.Equal(t, int64(15), v[0])
	assert.Equal(t, int64(-1), ToBigInt(v, 16).Int64())

	v = FromBigInt(big.NewInt(0), 16)
	assert.Equal(t, 1, len(v))
	assert.Equal(t, int64(0), v[0])
	assert.Equal(t, int64(0), ToBigInt(v, 16).Int64())

	v = FromBigInt(big.NewInt(-126), 16)
	assert.Equal(t, 3, len(v))
	assert.Equal(t, int64(2), v[0])
	assert.Equal(t, int64(8), v[1])
	assert.Equal(t, int64(15), v[2])
	assert.Equal(t, int64(-126), ToBigInt(v, 16).Int64())

	v = FromBigInt(big.NewInt(1100), 16)
	assert.Equal(t, 4, len(v))
	assert.Equal(t, int64(12), v[0])
	assert.Equal(t, int64(4), v[1])
	assert.Equal(t, int64(4), v[2])
	assert.Equal(t, int64(0), v[3])
	assert.Equal(t, int64(1100), ToBigInt(v, 16).Int64())

	v = FromBigInt(big.NewInt(-1100), 16)
	assert.Equal(t, 4, len(v))
	assert.Equal(t, int64(4), v[0])
	assert.Equal(t, int64(11), v[1])
	assert.Equal(t, int64(11), v[2])
	assert.Equal(t, int64(15), v[3])
	assert.Equal(t, int64(-1100), ToBigInt(v, 16).Int64())
}
