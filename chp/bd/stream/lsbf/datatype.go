package lsbf

import (
	"math/big"
)

type Int []int64

func FromInt64(value, base int64) Int {
	v := Int{}
	for value != 0 && value != -1 {
		if value < 0 {
			value -= base-1
		}
		digit := value%base
		if value < 0 {
			digit += base-1
		}
		v = append(v, digit)
		value /= base
	}
	if value < 0 {
		v = append(v, base-1)
	} else {
		v = append(v, 0)
	}
	return v
}

func ToInt64(value Int, base int64) int64 {
	if len(value) == 0 {
		return 0
	}

	neg := value[len(value)-1] == base-1

	var v int64 = 0
	var mult int64 = 1
	for _, digit := range value {
		if neg {
			v += (base - digit - 1)*mult
		} else {
			v += digit*mult
		}
		mult *= base
	}

	if neg {
		return -(v+1)
	}
	return v
}

func FromBigInt(value *big.Int, base int64) Int {
	v := Int{}
	zero := big.NewInt(0)
	negOne := big.NewInt(-1)
	bigBase := big.NewInt(base)
	digit := big.NewInt(0)

	cmpz := value.Cmp(zero)
	for cmpz != 0 && value.Cmp(negOne) != 0 {
		value.DivMod(value, bigBase, digit)
		v = append(v, digit.Int64())
		cmpz = value.Cmp(zero)
	}
	if cmpz < 0 {
		v = append(v, base-1)
	} else {
		v = append(v, 0)
	}
	return v
}

func ToBigInt(value Int, base int64) *big.Int {
	v := big.NewInt(0)
	if len(value) == 0 {
		return v
	}

	neg := value[len(value)-1] == base-1

	mult := big.NewInt(1)
	negOne := big.NewInt(-1)
	bigBase := big.NewInt(base)
	bigBase1 := &big.Int{}	
	bigBase1.Add(bigBase, negOne)
	for _, digit := range value {
		bigDigit := big.NewInt(digit)
		if neg {
			bigDigit.Sub(bigBase1, bigDigit)
		}
		bigDigit.Mul(bigDigit, mult)
		v.Add(v, bigDigit)
		mult.Mul(mult, bigBase)
	}

	if neg {
		v.Neg(v)
		v.Add(v, negOne)
	}
	return v
}

