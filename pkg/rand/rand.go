package rand

import (
	"math"
	base "math/rand"
)

type Rand interface {
	Int(max int) int
	Float(max float64) float64
}

func Flat(seed int64) Rand {
	return (*flat)(base.New(base.NewSource(seed)))
}

type flat base.Rand

func (rnd *flat) Int(max int) int {
	return (*base.Rand)(rnd).Int() % max
}

func (rnd *flat) Float(max float64) float64 {
	return (*base.Rand)(rnd).Float64() * max
}

func Cycle(values []float64) Rand {
	return &cycle{values: values}
}

type cycle struct {
	values []float64
	index  int
}

func (rnd *cycle) Int(max int) int {
	val := int(rnd.values[rnd.index] * float64(max))
	rnd.index = (rnd.index + 1) % len(rnd.values)
	return val
}

func (rnd *cycle) Float(max float64) float64 {
	val := rnd.values[rnd.index]
	rnd.index = (rnd.index + 1) % len(rnd.values)
	return val * max

}

func Natural(seed int64) Rand {
	return (*natural)(base.New(base.NewSource(seed)))
}

type natural base.Rand

func (rnd *natural) Int(max int) int {
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

func (rnd *natural) Float(max float64) float64 {
	return (*base.Rand)(rnd).Float64() * max
}
