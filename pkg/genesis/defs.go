package genesis

import (
	"fmt"
	"strings"
)

type Creator interface {
	N() int
	Get(i int) Word
}

type Word string

func NewCreator(def []byte) (Creator, error) {
	c := &creator{nonTerminals: map[string]*nonTerminal{}}
	return c, c.load(def)
}

type creator struct {
	nonTerminals map[string]*nonTerminal
}

type nonTerminal struct {
	options  [][]*nonTerminal
	terminal string
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

	lines := strings.Split(string(def), "\n")
	for i, line := range lines {
		switch {
		case len(line) > 8 && line[:8] == "letters:":
			for _, opt := range strings.Fields(line[8:]) {
				nt := c.ensureNT(opt)
				nt.terminal = opt
			}

		case len(line) > 6 && line[:6] == "words:":
			if err := c.addOptions(c.ensureNT("$words"), strings.Fields(line[6:])); err != nil {
				return err
			}

		case strings.Contains(line, "="): // TODO using strings.Contains and strings.Index isn't efficient
			idx := strings.Index(line, "=")
			pre := strings.Fields(line[:idx])

			if len(pre) != 1 {
				return fmt.Errorf("in line %v: expect 1 non-terminal before '=' but got '%v'", i, line[:idx])
			}
			if err := c.addOptions(c.ensureNT(pre[0]), strings.Fields(line[idx+1:])); err != nil {
				return err
			}
		}
	}

	if _, ok := c.nonTerminals["$words"]; !ok {
		return fmt.Errorf("def doesn't contain 'words:'")
	}
	return nil
}

func (c *creator) addOptions(nonT *nonTerminal, opts []string) error {
	nonT.options = make([][]*nonTerminal, len(opts))
	for i, opt := range opts {
		nt := []*nonTerminal{}
		for _, char := range opt {
			nt2 := c.ensureNT(string(char))

			nt = append(nt, nt2)
		}
		nonT.options[i] = nt
	}

	return nil
}

func (c *creator) ensureNT(key string) *nonTerminal {
	if nt, ok := c.nonTerminals[key]; ok {
		return nt
	} else {
		nt = &nonTerminal{}
		c.nonTerminals[key] = nt
		return nt
	}
}

func (c *creator) N() int {
	return c.nonTerminals["$words"].n()
}

func (c *creator) Get(i int) Word {
	return Word(c.nonTerminals["$words"].get(i))
}
