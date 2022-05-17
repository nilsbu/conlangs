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
		{
			"reject",
			"letters: b c a e n\nwords: ba be bc an\nreject: bc",
			true,
			[]g.Word{"ba", "be", "", "an"}, 4,
		},
		{
			"illegal regexp",
			"letters: a b s\nwords: bas\nreject: (",
			false, nil, 0,
		},
		{
			"missing words",
			"letters: a b s\n",
			false, nil, 0,
		},
		{
			"filter",
			"letters: b a e n m p\nC = b n\nV = a e\n\nwords: CVC\nfilter: na > ma;;b$>p",
			true,
			[]g.Word{"bap", "map", "bep", "nep", "ban", "man", "ben", "nen"}, 8,
		},
		{
			"no letters",
			"C = b n\nV = a e\n\nwords: CVC\nfilter: na > ma;;b$>p",
			true,
			[]g.Word{"bap", "map", "bep", "nep", "ban", "man", "ben", "nen"}, 8,
		},
		{
			"broken filter 1",
			"letters: b a e n m p\nC = b n\nV = a e\n\nwords: CVC\nfilter: na;b$>p",
			false, nil, 0,
		},
		{
			"broken filter 2",
			"letters: b a e n m p\nC = b n\nV = a e\n\nwords: CVC\nfilter: n(a>e;b$>p",
			false, nil, 0,
		},
		{
			"$ in macro name",
			"$C = b c\nwords: $C$C",
			true,
			[]g.Word{"bb", "cb", "bc", "cc"}, 4,
		},
		{
			"optional part",
			"C = b c\nV=a e\nwords: CVC?",
			true,
			[]g.Word{"ba", "ca", "be", "ce", "bab", "cab", "beb", "ceb", "bac", "cac", "bec", "cec"}, 12,
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
		choices []float64
		word    g.Word
	}{
		{
			"single word",
			"letters: a b s\nwords: bas",
			[]float64{0},
			"bas",
		},
		{
			"single word",
			"letters: a b d s\nwords: bas bad",
			[]float64{1},
			"bad",
		},
		{
			"define a custom non-terminal",
			"letters: a b c\nC = b c\nwords: Ca",
			[]float64{0, 1},
			"ca",
		},
		{
			"with two non-terminals",
			"letters: a e b c\nC = b c\nV = a e\nwords: CV",
			[]float64{0, 0, 1},
			"be",
		},
		{
			"stacked non-terminals",
			"letters: b c a e n\nW = CV na\nC = b c\nV = a e\n\nwords: W",
			[]float64{0, 0, 1, 1},
			"ce",
		},
		{
			"reject first option",
			"reject: ce\nletters: b c a e n\nW = CV na\nC = b c\nV = a e\n\nwords: W",
			[]float64{0, 0, 1, 1, 0, 0, 1, 0},
			"ca",
		},
		{
			"filter",
			"letters: a b s x\nfilter:a>x\nwords: bas",
			[]float64{0},
			"bxs",
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
