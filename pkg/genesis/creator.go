package genesis

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/nilsbu/conlangs/pkg/rand"
)

// stdRandomRate is the rate at which optional symbols are chosen in case random-rate isn't set.
const stdRandomRate = 0.1

type Creator interface {
	// Get returns the ith word of the language.
	Get(i int) Word
	// Choose returns a random word from the language.
	Choose(rnd rand.Rand) Word
}

// A Word is a valid string of character in a language.
type Word string

// NewCreator creates a Creator according to a .defs file.
// If the file is invalid, it will return an error.
func NewCreator(def []byte) (Creator, error) {
	c := &creator{symbols: map[string]*symbols{}, randomRate: stdRandomRate}
	return c, c.load(def)
}

type creator struct {
	symbols    map[string]*symbols
	rejections []*regexp.Regexp
	filters    []*filter
	// randomRate is a number in range [0-1) that determines the likelihood of an optional symbol (marked by '?') to occur
	randomRate float64
}

// A filter is a rule by which a part of a word is replaced by something else.
// It searches a string using a regular expression.
type filter struct {
	regexp *regexp.Regexp
	new    string
}

func (f *filter) apply(word Word) Word {
	idxs := f.regexp.FindAllStringIndex(string(word), 20)
	for _, idx := range idxs {
		word = word[:idx[0]] + Word(f.new) + word[idx[1]:]
	}
	return word
}

func (c *creator) load(def []byte) error {
	// TODO detect cycles
	lines := strings.Split(string(def), "\n")
	if err := c.findRate(lines); err != nil {
		return err
	}

	tableHeads := []string{}

	for i, line := range lines {
		continueTable := false

		switch {
		case hasPrefix("words:", line):
			if err := c.addOptions(c.ensureSymbolExists("#words"), strings.Fields(line[len("words:"):])); err != nil {
				return err
			}

		case hasPrefix("reject:", line):
			for _, rx := range strings.Fields(line[len("reject:"):]) {
				if err := c.addRejection(rx); err != nil {
					return err
				}
			}

		case strings.Contains(line, "="): // TODO using strings.Contains and strings.Index isn't efficient
			idx := strings.Index(line, "=")
			pre := strings.Fields(line[:idx])

			if len(pre) != 1 {
				return fmt.Errorf("in line %v: expect one non-terminal before '=' but got '%v'", i, line[:idx])
			}
			if err := c.addOptions(c.ensureSymbolExists(pre[0]), strings.Fields(line[idx+1:])); err != nil {
				return err
			}

		case hasPrefix("filter:", line):
			for _, rule := range strings.Split(line[len("filter:"):], ";") {
				if len(rule) == 0 {
					continue
				}
				idx := strings.Index(rule, ">")
				if idx == -1 {
					return fmt.Errorf("rule '%v' doesn't contain '>'", rule)
				}
				pre, pos := strings.TrimSpace(rule[:idx]), strings.TrimSpace(rule[idx+1:])
				if err := c.addFilter(pre, pos); err != nil {
					return err
				}
			}
		case hasPrefix("%", line):
			continueTable = true
			tableHeads = strings.Fields(line[1:])
		case len(tableHeads) > 0:
			continueTable = true

			fields := strings.Fields(line)
			if len(fields) != len(tableHeads)+1 {
				return fmt.Errorf("table doesn't have correct length: columns = %v, but got %v", len(tableHeads), len(fields)-1)
			} else {
				for i, f := range fields[1:] {
					switch f {
					case "+":
						continue
					case "-":
						if err := c.addRejection(fields[0] + tableHeads[i]); err != nil {
							return err
						}
					default:
						if err := c.addFilter(fields[0]+tableHeads[i], f); err != nil {
							return err
						}
					}
				}
			}
		}

		if !continueTable {
			tableHeads = []string{}
		}
	}

	if _, ok := c.symbols["#words"]; !ok {
		return fmt.Errorf("def doesn't contain 'words:'")
	}
	return nil
}

func (c *creator) addRejection(rx string) error {
	if rej, err := regexp.Compile(rx); err != nil {
		return err
	} else {
		c.rejections = append(c.rejections, rej)
		return nil
	}
}

func (c *creator) addFilter(pre, pos string) error {
	if rej, err := regexp.Compile(pre); err != nil {
		return err
	} else {
		c.filters = append(c.filters, &filter{
			regexp: rej,
			new:    pos,
		})
		return nil
	}
}

func (c *creator) findRate(lines []string) (err error) {
	for _, line := range lines {
		if hasPrefix("random-rate:", line) {
			c.randomRate, err = strconv.ParseFloat(strings.TrimSpace(line[len("random-rate:"):]), 64)
			if c.randomRate < 0 || c.randomRate > 1 {
				err = fmt.Errorf("random-rate must be in range [0, 1] but is %v", c.randomRate)
			}
			return
		}
	}
	return nil
}

func hasPrefix(pre, str string) bool {
	if len(str) < len(pre) {
		return false
	} else {
		return str[:len(pre)] == pre
	}
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
	word := Word(c.symbols["#words"].get(i))
	word = c.applyAllFilters(word)
	found := false
	for _, rx := range c.rejections {
		if rx.MatchString(string(word)) {
			found = true
			break
		}
	}
	if !found {
		return word
	} else {
		return ""
	}
}

func (c *creator) applyAllFilters(word Word) Word {
	for _, filter := range c.filters {
		word = filter.apply(word)
	}
	return word
}

func (c *creator) Choose(rnd rand.Rand) Word {
	for { // TODO Break from infinite loop
		word := Word(c.choose(rnd, c.symbols["#words"]))
		word = c.applyAllFilters(word)
		found := false
		for _, rx := range c.rejections {
			if rx.MatchString(string(word)) {
				found = true
				break
			}
		}
		if !found {
			return word
		}
	}
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
