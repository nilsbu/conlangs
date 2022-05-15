package genesis_test

import (
	"testing"

	g "github.com/nilsbu/conlangs/pkg/genesis"
)

func TestCreator(t *testing.T) {
	for _, c := range []struct {
		name  string
		defs  string
		ok    bool
		words []g.Word
	}{
		{
			"no defs",
			"",
			false,
			[]g.Word{},
		},
		{
			"single word",
			"words: bas",
			true,
			[]g.Word{"bas"},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			creator, err := g.NewCreator([]byte(c.defs))
			if c.ok && err != nil {
				t.Error("expected no error but got:", err)
			} else if !c.ok && err != nil {
				t.Error("expected error but none occured")
			} else if c.ok {
				words := creator.InOrder(len(c.words))
				if len(c.words) != len(words) {
					t.Fatalf("expected %v words but got %v", len(c.words), len(words))
				}
				for i, word := range c.words {
					if word != words[i] {
						t.Errorf("word %v: expected '%v' but got '%v'", i, word, words[i])
					}
				}
			}
		})
	}
}
