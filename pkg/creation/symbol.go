package creation

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// A symbols is a either a non-terminal, which means it can be replaced by one or more other symbols
// or a terminal, which is one or more characters that are final.
// If it is a non-termial, options will have at least one sequence. Otherwise terminal is non-empty.
type symbols struct {
	options  []sequence
	weights  weights
	terminal string
}

// A sequence is a sequence of symbols.
type sequence []*symbols

// n recursively calculates the number of options that can come from a symbol.
// If the same sequence of characters can be obtained in different ways, those duplicates are not filtered.
func (s *symbols) n() int {
	if len(s.options) == 0 {
		return 1
	} else {
		n := 0
		for _, opt := range s.options {
			p := 1
			for _, s2 := range opt {
				p *= s2.n()
			}
			n += p
		}
		return n
	}
}

// get returns the i-th sequence that is created by the symbol.
// TODO Explain in detail or remove along with Creator.Get()
func (s *symbols) get(i int) string {
	for _, opt := range s.options {
		p := 1
		for _, s2 := range opt {
			p *= s2.n()
		}
		if i < p {
			var str strings.Builder
			for _, s2 := range opt {
				j := i % s2.n()
				i /= s2.n()
				str.WriteString(s2.get(j))
			}
			return str.String()
		} else {
			i -= p
		}
	}
	return s.terminal
}

// choose picks a sequence from the options. It uses p, which is in range [0, s.weights.sum)
// to determine which one.
func (s *symbols) choose(p float64) sequence {
	ws := make([]float64, len(s.options))
	for i := range s.options {
		ws[i] = s.weights[i]
	}
	sum := 0.0
	for i := 0; i < len(s.options)-1; i++ {
		sum += s.weights[i]
		if sum > p {
			return s.options[i]
		}
	}
	return s.options[len(s.options)-1]
}

// weight describes the relative likelihood of options to be chosen.
// It is used in symbols. Weights have to be non-negative and add up to 1.
type weights []float64

// TODO describe algorithm
func calcWeights(opts []string) (weights weights, err error) {
	weights = make([]float64, len(opts))

	explicit := 0 // 0 = uninitialized, 1 = explicit weights, 2 = implicit
	sum := 0.0
	for i := range opts {
		split := strings.Split(opts[i], ":")
		if explicit < 2 {
			if len(split) == 1 {
				if explicit == 1 {
					return weights, fmt.Errorf("either use weights for all options or none")
				} else {
					explicit = 2
					weights[i] = (math.Log(float64(len(opts)+1)) - math.Log(float64(i+1))) / float64(len(opts))
				}
			} else if len(split) == 2 {
				if w, err := strconv.ParseFloat(split[1], 64); err != nil {
					return weights, fmt.Errorf("'%v' has no valid weight", opts[i])
				} else {
					explicit = 1
					weights[i] = w
					opts[i] = split[0]
				}
			} else {
				return weights, fmt.Errorf("'%v' has no valid weight", opts[i])
			}
		} else if len(split) == 1 {
			weights[i] = (math.Log(float64(len(opts)+1)) - math.Log(float64(i+1))) / float64(len(opts))
		} else {
			return weights, fmt.Errorf("either use weights for all options or none")
		}
		sum += weights[i]
	}

	for i := range weights {
		weights[i] /= sum
	}

	return weights, nil
}
