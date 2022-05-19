package creation

import (
	"regexp"
	"strings"
)

type Validator interface {
	OK(word Word) bool
}

type validator struct {
	rejections []*regexp.Regexp
}

func (v *validator) OK(word Word) bool {
	for _, rx := range v.rejections {
		if rx.MatchString(string(word)) {
			return false
		}
	}
	return true
}

func (v *validator) loadReject(line string) error {
	for _, rx := range strings.Fields(line[len("reject:"):]) {
		if err := v.addRejection(rx); err != nil {
			return err
		}
	}
	return nil
}

func (v *validator) addRejection(rx string) error {
	if rej, err := regexp.Compile(rx); err != nil {
		return err
	} else {
		v.rejections = append(v.rejections, rej)
		return nil
	}
}
