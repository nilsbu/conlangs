package creation

import (
	"fmt"
	"regexp"
	"strings"
)

// A Filter makes systematic changes to words.
type Filter interface {
	Apply(word Word) Word
}

// filters implements the Filter interface. It represents a series of regexp searches
// that replace sections of a word.
type filters []*filter

func (fs *filters) parseLine(line string) error {
	for _, rule := range strings.Split(removeUntil(line, ":"), ";") {
		if len(rule) == 0 {
			continue
		}
		idx := strings.Index(rule, ">")
		if idx == -1 {
			return fmt.Errorf("rule '%v' doesn't contain '>'", rule)
		}
		pre, pos := strings.TrimSpace(rule[:idx]), strings.TrimSpace(rule[idx+1:])
		if err := fs.add(pre, pos); err != nil {
			return err
		}
	}
	return nil
}

func (fs *filters) add(pre, pos string) error {
	if rej, err := regexp.Compile(pre); err != nil {
		return err
	} else {
		*fs = append(*fs, &filter{
			regexp: rej,
			new:    pos,
		})
		return nil
	}
}

func (fs *filters) Apply(word Word) Word {
	for _, filter := range *fs {
		word = filter.apply(word)
	}
	return word
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
