package genesis_test

import (
	"testing"

	g "github.com/nilsbu/conlangs/pkg/genesis"
	"github.com/nilsbu/conlangs/pkg/rand"
)

func TestCreatorGet(t *testing.T) {
	for _, c := range []struct {
		name  string
		defs  string
		ok    bool
		words []g.Word
		n     int
	}{
		{
			"no defs",
			"",
			false, nil, 0,
		},
		{
			"single word",
			"letters: a b s\nwords: bas",
			true,
			[]g.Word{"bas"}, 1,
		},
		{
			"single word",
			"letters: a b d s\nwords: bas bad",
			true,
			[]g.Word{"bas", "bad"}, 2,
		},
		{
			"define a custom non-terminal",
			"letters: a b c\nC = b c\nwords: Ca",
			true,
			[]g.Word{"ba", "ca"}, 2,
		},
		{
			"incorrect non-terminal definition",
			" = b c\nwords: Ca",
			false, nil, 0,
		},
		{
			"with two non-terminals",
			"letters: a e b c\nC = b c\nV = a e\nwords: CV",
			true,
			[]g.Word{"ba", "ca", "be", "ce"}, 4,
		},
		{
			"repeat non-terminal",
			"letters: b c\nC = b c\nwords: CC",
			true,
			[]g.Word{"bb", "cb", "bc", "cc"}, 4,
		},
		{
			"stacked non-terminals",
			"letters: b c a e n\nW = CV na\nC = b c\nV = a e\n\nwords: W",
			true,
			[]g.Word{"ba", "ca", "be", "ce", "na"}, 5,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			creator, err := g.NewCreator([]byte(c.defs))
			if c.ok && err != nil {
				t.Error("expected no error but got:", err)
			} else if !c.ok && err == nil {
				t.Error("expected error but none occured")
			} else if c.ok {
				n := creator.N()
				if c.n != n {
					t.Fatalf("expected %v words but got %v", c.n, n)
				}

				for i, ex := range c.words {
					ac := creator.Get(i)
					if ex != ac {
						t.Errorf("word %v: expected '%v' but got '%v'", i, ex, ac)
					}
				}
			}
		})
	}
}

func TestCreatorChoose(t *testing.T) {
	for _, c := range []struct {
		name    string
		defs    string
		choices []int
		word    g.Word
	}{
		{
			"single word",
			"letters: a b s\nwords: bas",
			[]int{0},
			"bas",
		},
		{
			"single word",
			"letters: a b d s\nwords: bas bad",
			[]int{1},
			"bad",
		},
		{
			"define a custom non-terminal",
			"letters: a b c\nC = b c\nwords: Ca",
			[]int{0, 1},
			"ca",
		},
		{
			"with two non-terminals",
			"letters: a e b c\nC = b c\nV = a e\nwords: CV",
			[]int{0, 0, 1},
			"be",
		},
		{
			"stacked non-terminals",
			"letters: b c a e n\nW = CV na\nC = b c\nV = a e\n\nwords: W",
			[]int{0, 0, 1, 1},
			"ce",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			creator, err := g.NewCreator([]byte(c.defs))
			if err != nil {
				t.Error("expected no error but got:", err)
			} else {
				word := creator.Choose(rand.Cycle(c.choices))
				if c.word != word {
					t.Errorf("expected '%v' but got '%v'", c.word, word)
				}
			}
		})
	}
}
