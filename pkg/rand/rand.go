package rand

import base "math/rand"

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
