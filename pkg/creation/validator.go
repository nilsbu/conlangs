package creation

import (
	"regexp"
	"strings"
)

// A Validator evaluates a Word and decides if it is valid in a language.
type Validator interface {
	OK(word Word) bool
}

// rejections evaluates Words by parsing it for a series of regexes that mark invalid words.
type rejections []*regexp.Regexp

func (v *rejections) OK(word Word) bool {
	for _, rx := range *v {
		if rx.MatchString(string(word)) {
			return false
		}
	}
	return true
}

// parseLine parses a line for definitions of rejections
func (v *rejections) parseLine(line string) error {
	for _, rx := range strings.Fields(removeUntil(line, ":")) {
		if rej, err := regexp.Compile(rx); err != nil {
			return err
		} else {
			*v = append(*v, rej)
		}
	}
	return nil
}

func (v *rejections) add(rx string) error {
	if rej, err := regexp.Compile(rx); err != nil {
		return err
	} else {
		*v = append(*v, rej)
		return nil
	}
}
