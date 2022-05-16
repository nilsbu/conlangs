package rand

import (
	"math"
	base "math/rand"
)

type Rand interface {
	Next(max int) int
}

func Flat(seed int64) Rand {
	return (*flat)(base.New(base.NewSource(seed)))
}

type flat base.Rand

func (rnd *flat) Next(max int) int {
	return (*base.Rand)(rnd).Int() % max
}

func Cycle(values []int) Rand {
	return &cycle{values: values}
}

type cycle struct {
	values []int
	index  int
}

func (rnd *cycle) Next(max int) int {
	val := rnd.values[rnd.index] % max
	rnd.index = (rnd.index + 1) % len(rnd.values)
	return val
}

func Natural(seed int64) Rand {
	return (*natural)(base.New(base.NewSource(seed)))
}

type natural base.Rand

func (rnd *natural) Next(max int) int {
	// TODO Natural weighting should be optimized
	if max == 1 {
		return 0
	}

	weights := make([]float64, max)
	sum := 0.0
	for i := 0; i < max; i++ {
		weights[i] = (math.Log(float64(max+1)) - math.Log(float64(i+1))) / float64(max)
		sum += weights[i]
	}
	r := (*base.Rand)(rnd).Float64() * sum
	sum = 0.0
	for i := 0; i < max-1; i++ {
		sum += weights[i]
		if sum > r {
			return i
		}
	}
	return max - 1
}
