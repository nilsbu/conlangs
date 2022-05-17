package genesis

import (
	"fmt"
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
	options  [][]*nonTerminal
	terminal string
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
			for _, nt2 := range opt {
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
		for _, nt2 := range opt {
			p *= nt2.n()
		}
		if i < p {
			var str strings.Builder
			for _, nt2 := range opt {
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
	nonT.options = make([][]*nonTerminal, len(opts))
	for i, opt := range opts {
		nt := []*nonTerminal{}
		var word strings.Builder
		for _, char := range opt {
			word.WriteRune(char)
			if char != '$' {
				nt2 := c.ensureNT(word.String())
				nt = append(nt, nt2)
				word.Reset()
			}
		}
		nonT.options[i] = nt
		nonT.terminal = ""
	}
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
		opt := nt.options[rnd.Next(len(nt.options))]
		var str strings.Builder
		for _, nt2 := range opt {
			str.WriteString(c.choose(rnd, nt2))
		}
		return str.String()
	} else {
		return nt.terminal
	}
}
