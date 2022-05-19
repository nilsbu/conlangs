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

	return i.c, i.v, i.fs, i.load(def)
}

// ini is a struct containing the objects created by ParseDefs.
// It's purpose is to allow splitting the parsing process without passing around lots of variables.
type ini struct {
	c  *creator
	v  *rejections
	fs *filters
}

func (init *ini) load(def []byte) error {

	lines := strings.Split(string(def), "\n")
	if err := init.c.findRate(lines); err != nil {
		return err
	}

	type lineparse struct {
		condition func(line string) bool
		body      func(line string) error
	}

	td := &tableData{}

	lps := []lineparse{
		{td.isValid, td.acceptLine},
		{hasPrefix("words:"), init.c.loadWords},
		{hasPrefix("reject:"), init.v.parseLine},
		{hasPrefix("filter:"), init.fs.parseLine},
		{func(line string) bool { return strings.Contains(line, "=") }, init.c.loadNonTerminal},
	}

	tableStart := -1
	for i, line := range lines {
		// discard comments
		line = strings.Split(line, "#")[0]

		for _, lp := range lps {
			if lp.condition(line) {
				if err := lp.body(line); err != nil {
					return errors.Wrapf(err, "in line %v", i)
				} else {
					break
				}
			}
		}
		if err := td.newline(init); err != nil {
			return errors.Wrapf(err, "in lines %v-%v", tableStart, i-1)
		}
	}
	if err := td.newline(init); err != nil {
		return errors.Wrapf(err, "in line %v", len(lines)-1)
	}

	if _, ok := init.c.symbols["#words"]; !ok {
		return fmt.Errorf("def doesn't contain 'words:'")
	}
	return nil
}

func hasPrefix(prefix string) func(string) bool {
	return func(line string) bool {
		if len(line) < len(prefix) {
			return false
		} else {
			return line[:len(prefix)] == prefix
		}
	}
}

type tableData struct {
	lines    []string
	columns  int
	newLines int
}

func (td *tableData) acceptLine(line string) error {
	td.lines = append(td.lines, line)
	td.columns = len(strings.Fields(line[1:]))
	td.newLines = 0
	return nil
}

func (td *tableData) newline(init *ini) error {
	td.newLines++
	if td.newLines > 1 && len(td.lines) > 0 {
		if err := init.loadTable(td.lines); err != nil {
			return err
		}
		td.lines = []string{}
		td.columns = 0
	}
	return nil
}

func (td *tableData) isValid(line string) bool {
	return hasPrefix("%")(line) || len(td.lines) > 0 && len(line) > 0
}

func (init *ini) loadTable(lines []string) error {
	tableHeads := strings.Fields(lines[0][1:])
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) != len(tableHeads)+1 {
			return fmt.Errorf("table doesn't have correct length: columns = %v, but got %v", len(tableHeads), len(fields)-1)
		} else {
			for i, f := range fields[1:] {
				switch f {
				case "+":
					continue
				case "-":
					if err := init.v.add(fields[0] + tableHeads[i]); err != nil {
						return err
					}
				default:
					if err := init.fs.add(fields[0]+tableHeads[i], f); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
