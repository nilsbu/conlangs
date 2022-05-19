package creation

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// A Word is a valid string of character in a language.
type Word string

// ParseDefs creates a Creator, Validator and Filter according to a .defs file.
// If the file is invalid, it will return an error.
func ParseDefs(def []byte) (Creator, Validator, Filter, error) {
	i := &ini{
		c:  &creator{symbols: map[string]*symbols{}, randomRate: stdRandomRate},
		v:  &rejections{},
		fs: &filters{},
	}

	return i.c, i.v, i.fs, i.parse(def)
}

// ini is a struct containing the objects created by ParseDefs.
// It's purpose is to allow splitting the parsing process without passing around lots of variables.
type ini struct {
	c  *creator
	v  *rejections
	fs *filters
}

func (init *ini) parse(def []byte) error {
	lines := strings.Split(string(def), "\n")

	// random-rate needs to be processed first, since it gets used in the creation of non-terminals
	// TODO Should random-rate be changeable, making different rates for different non-terminals possible?
	if err := init.c.findRate(lines); err != nil {
		return err
	}

	// bulk of the processing happens here
	if err := init.parseLines(lines); err != nil {
		return err
	}

	// "words:" command is essential since it defines the root
	if _, ok := init.c.symbols["#words"]; !ok {
		return fmt.Errorf("def doesn't contain 'words:'")
	}
	return nil
}

func (init *ini) parseLines(lines []string) error {
	// processing of tables is slightly more complex than other commands since it's multi-line,
	// tableInit contains that
	ti := &tableInit{init: init}

	processRules := []struct {
		// condition identifies lines that belong to a given rule
		condition func(line string) bool
		// rule executes the command
		rule func(line string) error
	}{
		{ti.isValid, ti.acceptLine},
		{func(line string) bool { return hasPrefix("words:", line) }, init.c.loadWords},
		{func(line string) bool { return hasPrefix("reject:", line) }, init.v.parseLine},
		{func(line string) bool { return hasPrefix("filter:", line) }, init.fs.parseLine},
		{func(line string) bool { return strings.Contains(line, "=") }, init.c.loadNonTerminal},
	}

	for i, line := range lines {
		// discard comments
		line = strings.Split(line, "#")[0]

		for _, lp := range processRules {
			if lp.condition(line) {
				if err := lp.rule(line); err != nil {
					return errors.Wrapf(err, "in line %v (\"%v\")", i+1, line)
				} else {
					break
				}
			}
		}
		ti.newline()
	}

	return nil
}

func hasPrefix(prefix string, line string) bool {
	if len(line) < len(prefix) {
		return false
	} else {
		return line[:len(prefix)] == prefix
	}
}

type tableInit struct {
	init     *ini
	head     []string
	newLines int
}

func (ti *tableInit) isValid(line string) bool {
	return hasPrefix("%", line) || (len(ti.head) > 0 && len(line) > 0)
}

func (ti *tableInit) acceptLine(line string) error {
	ti.newLines = 0

	if len(ti.head) == 0 {
		ti.head = strings.Fields(line[1:])
		return nil
	} else {
		head := ti.head

		fields := strings.Fields(line)
		if len(fields) != len(head)+1 {
			return fmt.Errorf("table doesn't have correct length: columns = %v, but got %v", len(head), len(fields)-1)
		} else {
			for i, f := range fields[1:] {
				switch f {
				case "+":
					// combination permitted, nothing to do
					continue
				case "-":
					// combination forbidden, add to validator
					if err := ti.init.v.add(fields[0] + head[i]); err != nil {
						return err
					}
				default:
					// add filter in more complex cases
					if err := ti.init.fs.add(fields[0]+head[i], f); err != nil {
						return err
					}
				}
			}
			return nil
		}
	}
}

func (ti *tableInit) newline() {
	ti.newLines++
	if ti.newLines == 2 {
		ti.head = []string{}
	}
}
