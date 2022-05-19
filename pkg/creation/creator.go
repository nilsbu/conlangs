package creation

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nilsbu/conlangs/pkg/rand"
)

// A Creator is an object that creates random words in a language.
// The words are accessible in two ways:
// Get(i) returns the i-th word in the language
// Choose(rnd) returns a random word
type Creator interface {
	Get(i int) Word
	Choose(rnd rand.Rand) Word
}

type creator struct {
	symbols map[string]*symbols
	// randomRate is a number in range [0-1) that determines the likelihood of an optional symbol (marked by '?') to occur
	randomRate float64
}

func (c *creator) parseWords(line string) error {
	return c.addOptions(c.ensureSymbolExists("#words"), strings.Fields(removeUntil(line, ":")))
}

func removeUntil(line, f string) string {
	return line[strings.Index(line, f)+1:]
}

func (c *creator) parseNonTerminal(line string) error {
	idx := strings.Index(line, "=")
	pre := strings.Fields(line[:idx])

	if len(pre) != 1 {
		return fmt.Errorf("expect one non-terminal before '=' but got '%v'", line[:idx])
	}
	if err := c.addOptions(c.ensureSymbolExists(pre[0]), strings.Fields(line[idx+1:])); err != nil {
		return err
	}

	return nil
}

func (c *creator) findRate(lines []string) (err error) {
	for _, line := range lines {
		if hasPrefix("random-rate:", line) {
			if c.randomRate, err = strconv.ParseFloat(strings.TrimSpace(removeUntil(line, ":")), 64); err != nil {
				return err
			} else {
				c.randomRate /= 100
				if c.randomRate < 0 || c.randomRate > 1 {
					return fmt.Errorf("random-rate must be in range [0, 100] but is %v", c.randomRate)
				}
			}
		}
	}
	return nil
}

func (c *creator) addOptions(nonT *symbols, opts []string) error {
	if len(opts) == 0 {
		return fmt.Errorf("at least one option needs to be given")
	}

	// tmpWeights might be changed when options are expanded
	tmpWeights, err := calcWeights(opts)
	if err != nil {
		return err
	}

	for i, rawopt := range opts {
		sopts, sws := c.expandOption(rawopt, tmpWeights[i])
		for j, opt := range sopts {
			s := []*symbols{}
			nonT.weights = append(nonT.weights, sws[j])
			var word strings.Builder
			for _, char := range opt {
				word.WriteRune(char)
				if char != '$' {
					s = append(s, c.ensureSymbolExists(word.String()))
					word.Reset()
				}
			}
			nonT.options = append(nonT.options, s)
		}

		// A symbol is either terminal or non-terminal. By default ensureSymbolExists() sets a terinal value,
		// so it needs to be removed, when options are set (and the symbol is determined to be non-terminal).
		nonT.terminal = ""
	}

	return nil
}

func (c *creator) expandOption(opt string, weight float64) (opts []string, ws []float64) {
	opts = make([]string, 1)
	ws = []float64{weight}

	for _, char := range opt {
		if char == '?' {
			n := len(opts)
			for i := 0; i < n; i++ {
				opts = append(opts, opts[i])
				opts[i] = opts[i][:len(opts[i])-1]

				ws = append(ws, (1-c.randomRate)*ws[i])
				ws[i] *= c.randomRate
			}
		} else {
			for i := range opts {
				opts[i] += string(char)
			}
		}
	}

	return opts, ws
}

func (c *creator) ensureSymbolExists(name string) *symbols {
	if s, ok := c.symbols[name]; ok {
		return s
	} else {
		s = &symbols{terminal: name}
		c.symbols[name] = s
		return s
	}
}

func (c *creator) Get(i int) Word {
	return Word(c.symbols["#words"].get(i))
}

func (c *creator) Choose(rnd rand.Rand) Word {
	return Word(c.choose(rnd, c.symbols["#words"]))
}

func (c *creator) choose(rnd rand.Rand, s *symbols) string {
	if len(s.options) > 0 {
		opt := s.choose(rnd.Float(1))
		var str strings.Builder
		for _, s2 := range opt {
			str.WriteString(c.choose(rnd, s2))
		}
		return str.String()
	} else {
		return s.terminal
	}
}
