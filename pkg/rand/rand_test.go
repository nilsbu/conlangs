package rand_test

import (
	"math"
	"testing"

	"github.com/nilsbu/conlangs/pkg/rand"
)

func TestRandInt(t *testing.T) {
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
			rand.Cycle([]float64{.2, .31, .12, .1}),
			[]int{100, 100, 100, 100, 10, 10},
			[]int{20, 31, 12, 10, 2, 3},
		},
		{
			"Natural",
			rand.Natural(0),
			[]int{100, 100, 100, 3, 3, 3, 1, 2, 2, 2},
			[]int{69, 7, 30, 0, 0, 0, 0, 0, 0, 1},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			for i, max := range c.maxs {
				value := c.rnd.Int(max)
				if c.values[i] != value {
					t.Errorf("position %v: expected %v but got %v", i, c.values[i], value)
				}
			}
		})
	}
}

func TestRandFloat(t *testing.T) {
	for _, c := range []struct {
		name         string
		rnd          rand.Rand
		maxs, values []float64
	}{
		{
			"Flat",
			rand.Flat(0),
			[]float64{4, 4, 5, 10, 100, 100},
			[]float64{3.7807845971764658, 0.979860341175119, 3.2797813259770257, 0.5434383959970039, 36.75872066324585, 28.948043315659277},
		},
		{
			"different Flat",
			rand.Flat(21),
			[]float64{4, 4, 5, 10, 100, 100},
			[]float64{2.912740783303776, 3.3502706276598153, 4.665054276709533, 2.1896582362805264, 77.04460373605728, 47.55644783763324},
		},
		{
			"Cycle",
			rand.Cycle([]float64{.2, .31, .12, .1}),
			[]float64{100, 100, 100, 100, 10, 10},
			[]float64{20, 31, 12, 10, 2, 3.1},
		},
		{
			"Natural",
			rand.Natural(0),
			[]float64{100, 100, 100, 3, 3, 3, 1, 2, 2, 2},
			[]float64{94.51961492941165, 24.496508529377977, 65.59562651954052, 0.16303151879910116, 1.1027616198973755, 0.8684412994697783, 0.19243860967493215, 1.3106643016296649, 1.794339426299602, 0.3347088851181167},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			for i, max := range c.maxs {
				value := c.rnd.Float(max)
				if math.Abs(c.values[i]-value) > 1e-4 {
					t.Errorf("position %v: expected %v but got %v", i, c.values[i], value)
				}
			}
		})
	}
}
