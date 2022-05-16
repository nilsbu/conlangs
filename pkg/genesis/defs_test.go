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
			false, nil,
		},
		{
			"single word",
			"letters: a b s\nwords: bas",
			true,
			[]g.Word{"bas"},
		},
		{
			"single word",
			"letters: a b d s\nwords: bas bad",
			true,
			[]g.Word{"bas", "bad"},
		},
		{
			"define a custom non-terminal",
			"letters: a b c\nC = b c\nwords: Ca",
			true,
			[]g.Word{"ba", "ca"},
		},
		{
			"incorrect non-terminal definition",
			" = b c\nwords: Ca",
			false, nil,
		},
		{
			"with two non-terminals",
			"letters: a e b c\nC = b c\nV = a e\nwords: CV",
			true,
			[]g.Word{"ba", "ca", "be", "ce"},
		},
		{
			"repeat non-terminal",
			"letters: b c\nC = b c\nwords: CC",
			true,
			[]g.Word{"bb", "cb", "bc", "cc"},
		},
		{
			"stacked non-terminals",
			"letters: b c a e n\nW = CV na\nC = b c\nV = a e\n\nwords: W",
			true,
			[]g.Word{"ba", "ca", "be", "ce", "na"},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			creator, err := g.NewCreator([]byte(c.defs))
			if c.ok && err != nil {
				t.Error("expected no error but got:", err)
			} else if !c.ok && err == nil {
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
