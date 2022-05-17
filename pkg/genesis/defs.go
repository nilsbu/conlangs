package genesis

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/nilsbu/conlangs/pkg/rand"
)

type Creator interface {
	N() int // TODO N() doesn't subtract illegal words
	Get(i int) Word
	Choose(rnd rand.Rand) Word
}

type Word string

func NewCreator(def []byte) (Creator, error) {
	c := &creator{nonTerminals: map[string]*nonTerminal{}}
	return c, c.load(def)
}

type creator struct {
	nonTerminals map[string]*nonTerminal
	rejections   []*regexp.Regexp
	filters      []*filter
}

type nonTerminal struct {
	options   []sequence
	weightSum float64
	terminal  string
}

type sequence struct {
	chars  []*nonTerminal
	weight float64
}

type filter struct {
	regexp *regexp.Regexp
	new    string
}

func (nt *nonTerminal) n() int {
	if len(nt.options) == 0 {
		return 1
	} else {
		n := 0
		for _, opt := range nt.options {
			p := 1
			for _, nt2 := range opt.chars {
				p *= nt2.n()
			}
			n += p
		}
		return n
	}
}

func (nt *nonTerminal) get(i int) string {
	for _, opt := range nt.options {
		p := 1
		for _, nt2 := range opt.chars {
			p *= nt2.n()
		}
		if i < p {
			var str strings.Builder
			for _, nt2 := range opt.chars {
				j := i % nt2.n()
				i /= nt2.n()
				str.WriteString(nt2.get(j))
			}
			return str.String()
		} else {
			i -= p
		}
	}
	return nt.terminal
}

func (c *creator) load(def []byte) error {
	// TODO detect cycles
	lines := strings.Split(string(def), "\n")
	for i, line := range lines {
		switch {
		case hasPrefix("words:", line):
			c.addOptions(c.ensureNT("#words"), strings.Fields(line[len("words:"):]))

		case hasPrefix("reject:", line):
			for _, rx := range strings.Fields(line[len("reject:"):]) {
				if rej, err := regexp.Compile(rx); err != nil {
					return err
				} else {
					c.rejections = append(c.rejections, rej)
				}
			}

		case strings.Contains(line, "="): // TODO using strings.Contains and strings.Index isn't efficient
			idx := strings.Index(line, "=")
			pre := strings.Fields(line[:idx])

			if len(pre) != 1 {
				return fmt.Errorf("in line %v: expect 1 non-terminal before '=' but got '%v'", i, line[:idx])
			}
			c.addOptions(c.ensureNT(pre[0]), strings.Fields(line[idx+1:]))

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
				if rej, err := regexp.Compile(pre); err != nil {
					return err
				} else {
					c.filters = append(c.filters, &filter{
						regexp: rej,
						new:    pos,
					})
				}
			}
		}
	}

	if _, ok := c.nonTerminals["#words"]; !ok {
		return fmt.Errorf("def doesn't contain 'words:'")
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

func (c *creator) addOptions(nonT *nonTerminal, opts []string) {
	ws, sum := weights(len(opts))
	nonT.weightSum = sum

	for i, rawopt := range opts {
		sopts, sws := expandOption(rawopt, ws[i])
		for j, opt := range sopts {
			nt := sequence{chars: []*nonTerminal{}, weight: sws[j]}
			var word strings.Builder
			for _, char := range opt {
				word.WriteRune(char)
				if char != '$' {
					nt2 := c.ensureNT(word.String())
					nt.chars = append(nt.chars, nt2)
					word.Reset()
				}
			}
			nonT.options = append(nonT.options, nt)
		}

		nonT.terminal = ""
	}
}

func expandOption(opt string, weight float64) (opts []string, ws []float64) {
	opts = make([]string, 1)
	ws = []float64{weight}
	factor := 0.1

	for _, char := range opt {
		if char == '?' {
			n := len(opts)
			for i := 0; i < n; i++ {
				opts = append(opts, opts[i])
				opts[i] = opts[i][:len(opts[i])-1]

				ws = append(ws, (1-factor)*ws[i])
				ws[i] *= factor
			}
		} else {
			for i := range opts {
				opts[i] += string(char)
			}
		}
	}

	return opts, ws
}

func weights(max int) (weights []float64, sum float64) {
	weights = make([]float64, max)
	for i := 0; i < max; i++ {
		weights[i] = (math.Log(float64(max+1)) - math.Log(float64(i+1))) / float64(max)
		sum += weights[i]
	}
	return weights, sum
}

func (c *creator) ensureNT(key string) *nonTerminal {
	if nt, ok := c.nonTerminals[key]; ok {
		return nt
	} else {
		nt = &nonTerminal{terminal: key}
		c.nonTerminals[key] = nt
		return nt
	}
}

func (c *creator) N() int {
	return c.nonTerminals["#words"].n()
}

func (c *creator) Get(i int) Word {
	word := Word(c.nonTerminals["#words"].get(i))
	word = c.filter(word)
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

func (c *creator) filter(word Word) Word {
	for _, filter := range c.filters {
		idxs := filter.regexp.FindAllStringIndex(string(word), 20)
		for _, idx := range idxs {
			word = word[:idx[0]] + Word(filter.new) + word[idx[1]:]
		}
	}
	return word
}

func (c *creator) Choose(rnd rand.Rand) Word {
	for { // TODO Break from infinite loop
		word := Word(c.choose(rnd, c.nonTerminals["#words"]))
		word = c.filter(word)
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

func (c *creator) choose(rnd rand.Rand, nt *nonTerminal) string {
	if len(nt.options) > 0 {
		opt := pick(rnd.Float(nt.weightSum), nt.options)
		var str strings.Builder
		for _, nt2 := range opt.chars {
			str.WriteString(c.choose(rnd, nt2))
		}
		return str.String()
	} else {
		return nt.terminal
	}
}

func pick(p float64, opts []sequence) sequence {
	ws := make([]float64, len(opts))
	for i := range opts {
		ws[i] = opts[i].weight
	}
	sum := 0.0
	for i := 0; i < len(opts)-1; i++ {
		sum += opts[i].weight
		if sum > p {
			return opts[i]
		}
	}
	return opts[len(opts)-1]
}
