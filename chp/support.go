package chp

import (
	"fmt"
	"math/rand"
	"github.com/stretchr/testify/assert"
)

func RandomBool() func(i int64) bool {
	return func(i int64) bool {
		return rand.Intn(2) == 1
	}
}

func RandomInt(lower, upper int) func(i int64) int {
	return func(i int64) int {
		low := lower
		high := upper

		// ensure variety in digit-stream length as well
		length := (1<<rand.Intn(32))-1

		if length < high {
			high = length
		}

		if -length > low {
			low = -length
		}

		if high-low > 0 {		
			low += rand.Intn(high-low)
		}
	
		return low
	}
}

func RandomInt64(lower, upper int64) func(i int64) int64 {
	return func(i int64) int64 {
		low := lower
		high := upper

		// ensure variety in digit-stream length as well
		length := (int64(1)<<rand.Int63n(64))-1

		if length < high {
			high = length
		}

		if -length > low {
			low = -length
		}

		if high-low > 0 {		
			low += rand.Int63n(high-low)
		}

		return low
	}
}

func Values[T interface{}](values ...T) func(i int64) T {
	return func(i int64) T {
		return values[i%int64(len(values))]
	}
}

func AreEqual[T interface{}](token int64, values []T) error {
	for i := 1; i < len(values); i++ {
		if !assert.ObjectsAreEqual(values[i], values[0]) {
			return fmt.Errorf("expected %v, found %v at token %d", values[0], values[i], token)
		}
	}
	return nil
}


