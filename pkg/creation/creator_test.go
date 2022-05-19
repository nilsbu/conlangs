package creation_test

import (
	"testing"

	cr "github.com/nilsbu/conlangs/pkg/creation"
	"github.com/nilsbu/conlangs/pkg/rand"
)

func TestCreatorGet(t *testing.T) {
	for _, c := range []struct {
		name  string
		defs  string
		ok    bool
		words []cr.Word
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
			[]cr.Word{"bas"},
		},
		{
			"single word",
			"letters: a b s\nwords: bas",
			true,
			[]cr.Word{"bas"},
		},
		{
			"no word",
			"letters: a b d s\nwords: ",
			false, nil,
		},
		{
			"define a custom non-terminal",
			"letters: a b c\nC = b c\nwords: Ca",
			true,
			[]cr.Word{"ba", "ca"},
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
			[]cr.Word{"ba", "ca", "be", "ce"},
		},
		{
			"repeat non-terminal",
			"letters: b c\nC = b c\nwords: CC",
			true,
			[]cr.Word{"bb", "cb", "bc", "cc"},
		},
		{
			"stacked non-terminals",
			"letters: b c a e n\nW = CV na\nC = b c\nV = a e\n\nwords: W",
			true,
			[]cr.Word{"ba", "ca", "be", "ce", "na"},
		},
		{
			"reject",
			"letters: b c a e n\nwords: ba be bc an\nreject: bc",
			true,
			[]cr.Word{"ba", "be", "", "an"},
		},
		{
			"illegal regexp",
			"letters: a b s\nwords: bas\nreject: (",
			false, nil,
		},
		{
			"missing words",
			"letters: a b s\n",
			false, nil,
		},
		{
			"filter",
			"letters: b a e n m p\nC = b n\nV = a e\n\nwords: CVC\nfilter: na > ma;;b$>p",
			true,
			[]cr.Word{"bap", "map", "bep", "nep", "ban", "man", "ben", "nen"},
		},
		{
			"no letters",
			"C = b n\nV = a e\n\nwords: CVC\nfilter: na > ma;;b$>p",
			true,
			[]cr.Word{"bap", "map", "bep", "nep", "ban", "man", "ben", "nen"},
		},
		{
			"broken filter 1",
			"letters: b a e n m p\nC = b n\nV = a e\n\nwords: CVC\nfilter: na;b$>p",
			false, nil,
		},
		{
			"broken filter 2",
			"letters: b a e n m p\nC = b n\nV = a e\n\nwords: CVC\nfilter: n(a>e;b$>p",
			false, nil,
		},
		{
			"$ in macro name",
			"$C = b c\nwords: $C$C",
			true,
			[]cr.Word{"bb", "cb", "bc", "cc"},
		},
		{
			"optional part",
			"C = b c\nV=a e\nwords: CVC?",
			true,
			[]cr.Word{"ba", "ca", "be", "ce", "bab", "cab", "beb", "ceb", "bac", "cac", "bec", "cec"},
		},
		{
			"broken random-rate",
			"C = b c\nV=a e\nwords: CVC?\nrandom-rate:10", // > 1 not permissible
			false, nil,
		},
		{
			"invalid weights 1",
			"words: a:1 b",
			false, nil,
		},
		{
			"invalid weights 2",
			"words: a b:1",
			false, nil,
		},
		{
			"invalid weights 3",
			"words: a:1 b:z",
			false, nil,
		},
		{
			"invalid weights 4",
			"V = a:1 b:2:2\nwords:",
			false, nil,
		},
		{
			"table filter 1",
			"V=a e i\nC=b c d\nwords: CV\n%a e i\nb + - +\nc - - -\nd - + +",
			true,
			[]cr.Word{"ba", "", "", "", "", "de", "bi", "", "di"},
		},
		{
			"table filter 2",
			"V=a e i\nC=b c d\nwords: CV\n%a e i\nb + - +\nc aa - -\nd - + +\n\n",
			true,
			[]cr.Word{"ba", "aa", "", "", "", "de", "bi", "", "di"},
		},
		{
			"table filter false length",
			"V=a e i\nC=b c d\nwords: CV\n%a e i\nb + + - +\nc - - -\nd - + +",
			false, nil,
		},
		{
			"table filter invalid with rejection",
			"V=a e i\nC=b c d\nwords: CV\n%a e i\n( + - +\nc - - -\nd - + +",
			false, nil,
		},
		{
			"table filter invalid with filter",
			"V=a e i\nC=b c d\n%a e i\n( a - +\nc - - -\nd - + +\nwords: CV",
			false, nil,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			creator, validator, filters, err := cr.ParseDefs([]byte(c.defs))
			if c.ok && err != nil {
				t.Error("expected no error but got:", err)
			} else if !c.ok && err == nil {
				t.Error("expected error but none occured")
			} else if c.ok {
				for i, ex := range c.words {
					ac := creator.Get(i)
					ac = filters.Apply(ac)
					if !validator.OK(ac) {
						ac = ""
					}
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
		word    cr.Word
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
		{
			"weight of ? with 0",
			"words: a?b",
			[]float64{0},
			"b",
		},
		{
			"weight of ? with 1",
			"words: a?b",
			[]float64{1},
			"ab",
		},
		{
			"weight of ? with .5",
			"words: a?b\nrandom-rate:.6",
			[]float64{.5},
			"b",
		},
		{
			"weight with :",
			"words: a:1 b:1.2",
			[]float64{.5},
			"b",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			creator, validator, filters, err := cr.ParseDefs([]byte(c.defs))
			if err != nil {
				t.Error("expected no error but got:", err)
			} else {
				var word cr.Word
				rnd := rand.Cycle(c.choices)
				for {
					word = creator.Choose(rnd)
					word = filters.Apply(word)
					if validator.OK(word) {
						break
					}
				}
				if c.word != word {
					t.Errorf("expected '%v' but got '%v'", c.word, word)
				}
			}
		})
	}
}
