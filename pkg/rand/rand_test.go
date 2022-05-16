package rand_test

import (
	"testing"

	"github.com/nilsbu/conlangs/pkg/rand"
)

func TestRand(t *testing.T) {
	for _, c := range []struct {
		name         string
		rnd          rand.Rand
		maxs, values []int
	}{
		{
			"Flat",
			rand.Flat(0),
			[]int{4, 4, 5, 10, 100, 100},
			[]int{1, 0, 2, 4, 2, 63},
		},
		{
			"different Flat",
			rand.Flat(21),
			[]int{4, 4, 5, 10, 100, 100},
			[]int{0, 1, 1, 3, 32, 1},
		},
		{
			"Cycle",
			rand.Cycle([]int{20, 31, 12, 110}),
			[]int{100, 100, 100, 100, 10, 10},
			[]int{20, 31, 12, 10, 0, 1},
		},
		{
			"Natural",
			rand.Flat(0),
			[]int{100, 100, 100, 3, 3, 3},
			[]int{5, 52, 27, 2, 0, 0},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			for i, max := range c.maxs {
				value := c.rnd.Next(max)
				if c.values[i] != value {
					t.Errorf("position %v: expected %v but got %v", i, c.values[i], value)
				}
			}
		})
	}
}
