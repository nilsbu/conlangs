package creation

import (
	"fmt"
	"strings"
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

	tableStart := -1

	for i, line := range lines {
		continueTable := false

		switch {
		case hasPrefix("words:", line):
			if err := init.c.loadWords(line); err != nil {
				return err
			}

		case hasPrefix("reject:", line):
			if err := init.v.parseLine(line); err != nil {
				return err
			}

		case strings.Contains(line, "="): // TODO using strings.Contains and strings.Index isn't efficient
			if err := init.c.loadNonTerminal(line); err != nil {
				return err
			}

		case hasPrefix("filter:", line):
			if err := init.fs.parseLine(line); err != nil {
				return err
			}

		case hasPrefix("%", line):
			continueTable = true
			tableStart = i
		case tableStart >= 0 && len(line) > 0:
			continueTable = true
		}

		if !continueTable && tableStart >= 0 {
			if err := init.loadTable(lines[tableStart:i]); err != nil {
				return err
			}
			tableStart = -1
		}
	}

	if tableStart >= 0 {
		if err := init.loadTable(lines[tableStart:]); err != nil {
			return err
		}
	}

	if _, ok := init.c.symbols["#words"]; !ok {
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
