package genesis

import (
	"strings"
)

type Creator interface {
	InOrder(n int) []Word
}

type Word string

func NewCreator(defs []byte) (Creator, error) {
	c := &creator{nonTerminals: map[string]nonTerminal{}}
	return c, c.load(defs)
}

type creator struct {
	nonTerminals map[string]nonTerminal
}

type nonTerminal struct {
	options  [][]*nonTerminal
	terminal string
}

func (c *creator) load(defs []byte) error {

	lines := strings.Split(string(defs), "\n")
	for _, line := range lines {
		switch {
		case len(line) > 6 && line[:6] == "words:":
			opts := strings.Fields(line[6:])
			nts := make([][]*nonTerminal, len(opts))
			for i, opt := range opts {
				nt := []*nonTerminal{}
				for _, char := range opt {
					nt = append(nt, &nonTerminal{terminal: string(char)})
				}
				nts[i] = nt
			}
			c.nonTerminals["$words"] = nonTerminal{options: nts}
		}
	}

	return nil
}

func (c *creator) InOrder(n int) []Word {
	words := []Word{}
	c.iterate("$words", n, "", &words)
	return words
}

func (c *creator) iterate(nt string, n int, init string, words *[]Word) {
	nonTerminal := c.nonTerminals[nt]
	for i := range nonTerminal.options {
		c.what(i, n, nonTerminal.options, words)
		if len(*words) == n {
			return
		}
	}
}

func (c *creator) what(i, n int, opts [][]*nonTerminal, words *[]Word) {
	opt := opts[i]
	ids := make([]int, len(opt))
	for {
		var chars string
		for _, nt := range opt {
			chars += nt.terminal
		}

		*words = append(*words, Word(chars))
		if len(*words) == n {
			return
		}

		if inc(ids, opt) {
			return
		}
	}
}

func inc(ids []int, opt []*nonTerminal) bool {
	for i := len(ids) - 1; i >= 0; i-- {
		ids[i]++
		if ids[i] >= len(opt[i].options) {
			ids[i] = 0
		} else {
			return false
		}
	}

	return true
}
